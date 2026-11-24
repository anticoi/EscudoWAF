package internal

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// LogStore define la interfaz de almacenamiento de logs
type LogStore interface {
	Store(log model.WAFLog) error
	Start()
	Close()
}

// LogBatch estructura de lote de logs, para el pool de objetos
type LogBatch struct {
	logs []interface{}
	size int
}

// Pool de objetos de lote, reduce asignaciones de memoria
var batchPool = sync.Pool{
	New: func() interface{} {
		return &LogBatch{
			logs: make([]interface{}, 0, 200), // Preasignar mayor capacidad
		}
	},
}

// RingBuffer buffer circular de alto rendimiento sin bloqueos
type RingBuffer struct {
	buffer   []model.WAFLog
	capacity uint64
	mask     uint64
	_        [56]byte // Relleno de línea de caché, evita false sharing
	head     atomic.Uint64
	_        [56]byte
	tail     atomic.Uint64
	_        [56]byte
}

// NewRingBuffer crea un buffer circular
func NewRingBuffer(size int) *RingBuffer {
	// Asegurar que la capacidad sea potencia de 2
	capacity := uint64(1)
	for capacity < uint64(size) {
		capacity <<= 1
	}

	return &RingBuffer{
		buffer:   make([]model.WAFLog, capacity),
		capacity: capacity,
		mask:     capacity - 1,
	}
}

// Push envía logs sin bloqueos
func (rb *RingBuffer) Push(log model.WAFLog) bool {
	for {
		tail := rb.tail.Load()
		next := (tail + 1) & rb.mask
		head := rb.head.Load()

		if next == head {
			return false // Buffer lleno
		}

		if rb.tail.CompareAndSwap(tail, next) {
			rb.buffer[tail] = log
			return true
		}
		// CAS falló, reintentando
		runtime.Gosched()
	}
}

// PopBatch extrae logs por lotes
func (rb *RingBuffer) PopBatch(batch *LogBatch, maxSize int) int {
	head := rb.head.Load()
	tail := rb.tail.Load()

	// Calcular la cantidad disponible
	available := int((tail - head) & rb.mask)
	if available == 0 {
		return 0
	}

	// Limitar el tamaño del lote
	if available > maxSize {
		available = maxSize
	}

	// Lectura por lotes
	for i := 0; i < available; i++ {
		idx := (head + uint64(i)) & rb.mask
		batch.logs = append(batch.logs, rb.buffer[idx])
	}

	// Actualizar head
	rb.head.Add(uint64(available))
	batch.size = available
	return available
}

// MongoLogStore almacenamiento de logs de alto rendimiento implementado en MongoDB
type MongoLogStore struct {
	mongo       *mongo.Client
	mongoDB     string
	collection  *mongo.Collection
	ringBuffers []*RingBuffer // Múltiples buffers circulares para reducir la contención
	numBuffers  int
	logger      zerolog.Logger
	state       atomic.Uint32 // 0: stopped, 1: running, 2: closing
	wg          sync.WaitGroup

	// Parámetros de optimización de rendimiento
	writerCount    int           // Número de goroutines de escritura
	batchSize      atomic.Int32  // Tamaño de lote dinámico
	batchInterval  time.Duration // Intervalo de procesamiento por lotes
	maxBatchSize   int
	minBatchSize   int
	bufferSelector atomic.Uint64 // Para seleccionar el buffer
}

// Instancia singleton
var (
	mongoLogStoreOnce     sync.Once
	mongoLogStoreInstance *MongoLogStore
)

// StoreConfig configuración de almacenamiento
type StoreConfig struct {
	BufferSize    int
	NumBuffers    int
	WriterCount   int
	BatchInterval time.Duration
	MaxBatchSize  int
	MinBatchSize  int
}

// DefaultConfig configuración por defecto
func DefaultConfig() StoreConfig {
	cores := runtime.NumCPU()
	return StoreConfig{
		BufferSize:    8192,                   // Tamaño de cada buffer
		NumBuffers:    cores,                  // Número de buffers igual a los núcleos de CPU
		WriterCount:   cores / 2,              // Goroutines de escritura: la mitad de los núcleos
		BatchInterval: 100 * time.Millisecond, // Intervalo de procesamiento por lotes de 100ms
		MaxBatchSize:  500,
		MinBatchSize:  50,
	}
}

// NewMongoLogStore crea un almacenamiento de logs en MongoDB (singleton)
func NewMongoLogStore(client *mongo.Client, database, collection string, logger zerolog.Logger) *MongoLogStore {
	config := DefaultConfig()
	return NewMongoLogStoreWithConfig(client, database, collection, config, logger)
}

// NewMongoLogStoreWithConfig crea el almacenamiento con configuración
func NewMongoLogStoreWithConfig(client *mongo.Client, database, collection string, config StoreConfig, logger zerolog.Logger) *MongoLogStore {
	mongoLogStoreOnce.Do(func() {
		if config.NumBuffers <= 0 {
			config.NumBuffers = runtime.NumCPU()
		}
		if config.WriterCount <= 0 {
			config.WriterCount = config.NumBuffers / 2
			if config.WriterCount == 0 {
				config.WriterCount = 1
			}
		}

		store := &MongoLogStore{
			mongo:         client,
			mongoDB:       database,
			collection:    client.Database(database).Collection(collection),
			ringBuffers:   make([]*RingBuffer, config.NumBuffers),
			numBuffers:    config.NumBuffers,
			logger:        logger,
			writerCount:   config.WriterCount,
			batchInterval: config.BatchInterval,
			maxBatchSize:  config.MaxBatchSize,
			minBatchSize:  config.MinBatchSize,
		}

		// Inicializar el buffer circular
		bufferSizePerRing := config.BufferSize / config.NumBuffers
		for i := 0; i < config.NumBuffers; i++ {
			store.ringBuffers[i] = NewRingBuffer(bufferSizePerRing)
		}

		// Establecer el tamaño inicial del lote
		store.batchSize.Store(int32(config.MinBatchSize))

		mongoLogStoreInstance = store
		logger.Info().
			Int("num_buffers", config.NumBuffers).
			Int("writer_count", config.WriterCount).
			Int("buffer_size", config.BufferSize).
			Msg("Crear nueva instancia de MongoLogStore de alto rendimiento")
	})

	return mongoLogStoreInstance
}

// Store almacenamiento no bloqueante de alto rendimiento
func (s *MongoLogStore) Store(log model.WAFLog) error {
	// Comprobación rápida de estado
	if s.state.Load() != 1 {
		return nil // Estado no iniciado, descartar directamente
	}

	// Seleccionar buffer (distribución de carga round-robin)
	idx := s.bufferSelector.Add(1) % uint64(s.numBuffers)
	buffer := s.ringBuffers[idx]

	// Intentar enviar al buffer circular
	if !buffer.Push(log) {
		// Buffer lleno, descartar directamente (la pérdida de logs es aceptable según lo requerido)
		return nil
	}

	return nil
}

// Start inicia el procesamiento de almacenamiento de logs
func (s *MongoLogStore) Start() {
	if !s.state.CompareAndSwap(0, 1) {
		s.logger.Debug().Msg("El procesamiento de logs ya está en ejecución")
		return
	}

	s.logger.Info().Msg("Iniciar procesamiento de logs de alto rendimiento")

	// Iniciar múltiples goroutines de escritura
	for i := 0; i < s.writerCount; i++ {
		s.wg.Add(1)
		go s.writer(i)
	}

	// Iniciar la goroutine de ajuste dinámico
	s.wg.Add(1)
	go s.dynamicAdjuster()
}

// writer goroutine de escritura
func (s *MongoLogStore) writer(id int) {
	defer s.wg.Done()

	// Cada writer se encarga de buffers específicos
	buffersPerWriter := s.numBuffers / s.writerCount
	startIdx := id * buffersPerWriter
	endIdx := startIdx + buffersPerWriter
	if id == s.writerCount-1 {
		endIdx = s.numBuffers // El último writer procesa los buffers restantes
	}

	ticker := time.NewTicker(s.batchInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		if s.state.Load() == 2 {
			// Cerrando, procesando datos restantes
			s.flushAll(startIdx, endIdx)
			return
		}
		// Procesar los buffers asignados
		s.processBatches(startIdx, endIdx)

		// Comprobar si es necesario salir
		if s.state.Load() == 2 {
			return
		}
	}
}

// processBatches procesa los lotes
func (s *MongoLogStore) processBatches(startIdx, endIdx int) {
	batch := batchPool.Get().(*LogBatch)
	defer func() {
		batch.logs = batch.logs[:0]
		batch.size = 0
		batchPool.Put(batch)
	}()

	currentBatchSize := int(s.batchSize.Load())
	totalCollected := 0

	// Recopilar logs de múltiples buffers
	for i := startIdx; i < endIdx && totalCollected < currentBatchSize; i++ {
		buffer := s.ringBuffers[i]
		remaining := currentBatchSize - totalCollected
		count := buffer.PopBatch(batch, remaining)
		totalCollected += count
	}

	// Si se recopilaron logs, ejecutar la escritura por lotes
	if batch.size > 0 {
		s.batchInsert(batch)
	}
}

// batchInsert inserción por lotes (sin reintentos, optimizado para rendimiento)
func (s *MongoLogStore) batchInsert(batch *LogBatch) {
	if batch.size == 0 {
		return
	}

	// Usar timeout corto; si falla, descartar
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Inserción por lotes, ignora errores (pérdida aceptable)
	_, _ = s.collection.InsertMany(ctx, batch.logs[:batch.size])
}

// dynamicAdjuster ajusta dinámicamente el tamaño del lote
func (s *MongoLogStore) dynamicAdjuster() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastCheck := time.Now()
	lastCount := uint64(0)

	for {
		<-ticker.C
		if s.state.Load() == 2 {
			return
		}

		// Calcular la carga actual
		currentCount := uint64(0)
		for _, buffer := range s.ringBuffers {
			head := buffer.head.Load()
			tail := buffer.tail.Load()
			currentCount += (tail - head) & buffer.mask
		}

		// Ajustar el tamaño del lote según la tasa de crecimiento
		elapsed := time.Since(lastCheck).Seconds()
		rate := float64(currentCount-lastCount) / elapsed

		currentBatchSize := s.batchSize.Load()
		newBatchSize := currentBatchSize

		if rate > 1000 { // Alta carga
			newBatchSize = min(currentBatchSize*2, int32(s.maxBatchSize))
		} else if rate < 100 { // Carga baja
			newBatchSize = max(currentBatchSize/2, int32(s.minBatchSize))
		}

		if newBatchSize != currentBatchSize {
			s.batchSize.Store(newBatchSize)
		}

		lastCheck = time.Now()
		lastCount = currentCount
	}
}

// flushAll vacía todos los buffers
func (s *MongoLogStore) flushAll(startIdx, endIdx int) {
	batch := batchPool.Get().(*LogBatch)
	defer func() {
		batch.logs = batch.logs[:0]
		batch.size = 0
		batchPool.Put(batch)
	}()

	// Recopilar tantos logs como sea posible
	for i := startIdx; i < endIdx; i++ {
		buffer := s.ringBuffers[i]
		for {
			count := buffer.PopBatch(batch, s.maxBatchSize)
			if count == 0 {
				break
			}
			if batch.size > 0 {
				s.batchInsert(batch)
				batch.logs = batch.logs[:0]
				batch.size = 0
			}
		}
	}
}

// Close cierra el almacenamiento de logs
func (s *MongoLogStore) Close() {
	// Establecer el estado cerrado
	if !s.state.CompareAndSwap(1, 2) {
		return
	}

	s.logger.Info().Msg("Cerrar el procesamiento de logs de alto rendimiento")

	// Esperar a que terminen todos los writers
	s.wg.Wait()

	// Restablecer el estado
	s.state.Store(0)
}

// Función auxiliar
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
