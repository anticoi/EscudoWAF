package flowcontroller

import (
	"container/heap"
	"context"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RecorderConfig configuración del registrador
type RecorderConfig struct {
	Capacity        int
	CleanupInterval time.Duration
	BatchSize       int
	WriteTimeout    time.Duration
	MaxRetries      int
	ShardCount      int // Número de shards, debe ser potencia de 2
	MetricsEnabled  bool
	WriteQueueSize  int
}

// DefaultConfig configuración por defecto
func DefaultConfig() RecorderConfig {
	return RecorderConfig{
		Capacity:        10000,
		CleanupInterval: time.Minute,
		BatchSize:       100,
		WriteTimeout:    10 * time.Second,
		MaxRetries:      3,
		ShardCount:      16, // 16 shards por defecto
		MetricsEnabled:  true,
		WriteQueueSize:  10000,
	}
}

// Metrics métricas de monitorización
type Metrics struct {
	TotalBlocked    atomic.Uint64
	TotalExpired    atomic.Uint64
	CurrentBlocked  atomic.Uint64
	CleanupDuration atomic.Value // time.Duration
	WriteQueueSize  atomic.Uint64
	CacheHits       atomic.Uint64
	CacheMisses     atomic.Uint64
}

// MemoryBlockedIP registro simplificado de IP en memoria, solo guarda los campos necesarios para consultas
type MemoryBlockedIP struct {
	IP           string    // Dirección IP
	BlockedUntil time.Time // Limitar la hora de fin
}

// IPRecorder interfaz del registrador de IPs
type IPRecorder interface {
	RecordBlockedIP(ip string, reason string, requestUri string, duration time.Duration) error
	IsIPBlocked(ip string) (bool, *model.BlockedIPRecord)
	GetBlockedIPs() ([]model.BlockedIPRecord, error)
	Close() error
	GetMetrics() *Metrics
}

// IPExpiryItem elemento de la cola de prioridad de expiración
type IPExpiryItem struct {
	ip        string
	expiresAt time.Time
	index     int
}

// Pool de objetos, reduce la presión del GC
var ipExpiryItemPool = sync.Pool{
	New: func() interface{} {
		return &IPExpiryItem{}
	},
}

// IPExpiryHeap implementación de cola de prioridad de IPs expiradas
type IPExpiryHeap []*IPExpiryItem

func (h IPExpiryHeap) Len() int { return len(h) }

func (h IPExpiryHeap) Less(i, j int) bool {
	return h[i].expiresAt.Before(h[j].expiresAt)
}

func (h IPExpiryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *IPExpiryHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*IPExpiryItem)
	item.index = n
	*h = append(*h, item)
}

func (h *IPExpiryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

func (h *IPExpiryHeap) Peek() *IPExpiryItem {
	if len(*h) == 0 {
		return nil
	}
	return (*h)[0]
}

func (h *IPExpiryHeap) Update(item *IPExpiryItem, expiresAt time.Time) {
	item.expiresAt = expiresAt
	heap.Fix(h, item.index)
}

// shard estructura de shard - ahora usa registros simplificados en memoria
type shard struct {
	mu          sync.RWMutex
	blockedIPs  map[string]MemoryBlockedIP // Usar registros simplificados en memoria
	expiryItems map[string]*IPExpiryItem
	expiryHeap  IPExpiryHeap
	toDelete    []string // Caché de eliminación reutilizada
}

// MemoryIPRecorder implementación de registrador de IPs en memoria (versión con shards)
type MemoryIPRecorder struct {
	shards          []*shard
	shardMask       uint32
	capacity        int
	config          RecorderConfig
	logger          zerolog.Logger
	cleanupInterval atomic.Value // time.Duration
	stopCleaner     chan struct{}
	Metrics         *Metrics // Público para compartir con MongoIPRecorder
}

// Instancia singleton
var (
	memoryIPRecorderOnce     sync.Once
	memoryIPRecorderInstance *MemoryIPRecorder
)

// fnv32 calcula el hash FNV-32 de una cadena
func fnv32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// NewMemoryIPRecorder crea un registrador de IPs en memoria (singleton)
func NewMemoryIPRecorder(capacity int, logger zerolog.Logger) *MemoryIPRecorder {
	config := DefaultConfig()
	config.Capacity = capacity
	return NewMemoryIPRecorderWithConfig(config, logger)
}

// NewMemoryIPRecorderWithConfig crea un registrador de IPs en memoria con configuración
func NewMemoryIPRecorderWithConfig(config RecorderConfig, logger zerolog.Logger) *MemoryIPRecorder {
	memoryIPRecorderOnce.Do(func() {
		if config.Capacity <= 0 {
			config.Capacity = 10000
		}

		// Asegurar que el número de shards sea potencia de 2
		shardCount := config.ShardCount
		if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
			shardCount = 16
		}

		recorder := &MemoryIPRecorder{
			shards:      make([]*shard, shardCount),
			shardMask:   uint32(shardCount - 1),
			capacity:    config.Capacity,
			config:      config,
			logger:      logger,
			stopCleaner: make(chan struct{}),
			Metrics:     &Metrics{},
		}

		recorder.cleanupInterval.Store(config.CleanupInterval)

		// Inicializar cada shard - usa registros simplificados en memoria
		capacityPerShard := config.Capacity / shardCount
		for i := 0; i < shardCount; i++ {
			s := &shard{
				blockedIPs:  make(map[string]MemoryBlockedIP, capacityPerShard), // Registros simplificados
				expiryItems: make(map[string]*IPExpiryItem, capacityPerShard),
				expiryHeap:  make(IPExpiryHeap, 0, capacityPerShard),
				toDelete:    make([]string, 0, 100),
			}
			heap.Init(&s.expiryHeap)
			recorder.shards[i] = s
		}

		// Iniciar la limpieza adaptativa
		go recorder.adaptiveCleanupLoop()

		memoryIPRecorderInstance = recorder
		logger.Info().
			Int("capacity", config.Capacity).
			Int("shards", shardCount).
			Msg("Crear nueva instancia de MemoryIPRecorder")
	})

	return memoryIPRecorderInstance
}

// getShard obtiene el shard correspondiente a la IP
func (r *MemoryIPRecorder) getShard(ip string) *shard {
	hash := fnv32(ip)
	return r.shards[hash&r.shardMask]
}

// adaptiveCleanupInterval intervalo de limpieza adaptativo
func (r *MemoryIPRecorder) adaptiveCleanupInterval() time.Duration {
	// Calcular el uso total
	var totalUsed int
	for _, s := range r.shards {
		s.mu.RLock()
		totalUsed += len(s.blockedIPs)
		s.mu.RUnlock()
	}

	usage := float64(totalUsed) / float64(r.capacity)

	var interval time.Duration
	switch {
	case usage > 0.9:
		interval = 10 * time.Second
	case usage > 0.7:
		interval = 30 * time.Second
	case usage > 0.5:
		interval = 45 * time.Second
	default:
		interval = r.config.CleanupInterval
	}

	r.cleanupInterval.Store(interval)
	return interval
}

// adaptiveCleanupLoop bucle de limpieza adaptativo
func (r *MemoryIPRecorder) adaptiveCleanupLoop() {
	ticker := time.NewTicker(r.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			r.cleanupExpired()
			duration := time.Since(start)

			if r.config.MetricsEnabled {
				r.Metrics.CleanupDuration.Store(duration)
			}

			// Ajustar el intervalo de la próxima limpieza
			interval := r.adaptiveCleanupInterval()
			ticker.Reset(interval)

		case <-r.stopCleaner:
			return
		}
	}
}

// cleanupExpired limpia los registros de IPs expirados
func (r *MemoryIPRecorder) cleanupExpired() {
	now := time.Now()
	var totalRemoved int

	// Limpiar cada shard en paralelo
	var wg sync.WaitGroup
	removedChan := make(chan int, len(r.shards))

	for _, s := range r.shards {
		wg.Add(1)
		go func(s *shard) {
			defer wg.Done()
			removed := r.cleanupShardExpired(s, now)
			removedChan <- removed
		}(s)
	}

	wg.Wait()
	close(removedChan)

	for removed := range removedChan {
		totalRemoved += removed
	}

	if totalRemoved > 0 {
		r.Metrics.TotalExpired.Add(uint64(totalRemoved))
		r.logger.Debug().
			Int("removed", totalRemoved).
			Msg("Registros de IPs expirados limpiados")
	}
}

// cleanupShardExpired limpia los registros expirados de un shard
func (r *MemoryIPRecorder) cleanupShardExpired(s *shard, now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.toDelete = s.toDelete[:0]

	// Comprobar desde la cima del heap
	for s.expiryHeap.Len() > 0 {
		item := s.expiryHeap.Peek()
		if item.expiresAt.After(now) {
			break
		}

		heap.Pop(&s.expiryHeap)
		s.toDelete = append(s.toDelete, item.ip)

		// Devolver el objeto al pool
		item.ip = ""
		item.expiresAt = time.Time{}
		item.index = -1
		ipExpiryItemPool.Put(item)
	}

	// Eliminación por lotes
	removed := len(s.toDelete)
	for _, ip := range s.toDelete {
		delete(s.blockedIPs, ip)
		delete(s.expiryItems, ip)
	}

	return removed
}

// ensureShardCapacity garantiza que la capacidad del shard no se exceda
func (r *MemoryIPRecorder) ensureShardCapacity(s *shard) {
	capacityPerShard := r.capacity / len(r.shards)
	if len(s.blockedIPs) < capacityPerShard {
		return
	}

	now := time.Now()
	s.toDelete = s.toDelete[:0]

	// Limpiar primero los expirados
	for s.expiryHeap.Len() > 0 && len(s.blockedIPs) >= capacityPerShard {
		item := s.expiryHeap.Peek()
		if item.expiresAt.After(now) {
			break
		}

		heap.Pop(&s.expiryHeap)
		s.toDelete = append(s.toDelete, item.ip)
		ipExpiryItemPool.Put(item)
	}

	// Eliminación por lotes de registros expirados
	for _, ip := range s.toDelete {
		delete(s.blockedIPs, ip)
		delete(s.expiryItems, ip)
	}

	// Si sigue lleno, eliminar el que expira primero
	for len(s.blockedIPs) >= capacityPerShard && s.expiryHeap.Len() > 0 {
		item := heap.Pop(&s.expiryHeap).(*IPExpiryItem)
		delete(s.blockedIPs, item.ip)
		delete(s.expiryItems, item.ip)

		r.logger.Debug().
			Str("ip", item.ip).
			Time("expires_at", item.expiresAt).
			Msg("Capacidad llena, se elimina el registro de IP que expira primero")

		ipExpiryItemPool.Put(item)
	}
}

// RecordBlockedIP registra una IP restringida - en memoria solo guarda los campos necesarios
func (r *MemoryIPRecorder) RecordBlockedIP(ip string, reason string, requestUri string, duration time.Duration) error {
	s := r.getShard(ip)

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(duration)

	// En memoria solo se guardan los campos necesarios
	memoryRecord := MemoryBlockedIP{
		IP:           ip,
		BlockedUntil: expiresAt,
	}

	// Comprobar si la IP ya existe
	if item, exists := s.expiryItems[ip]; exists {
		s.blockedIPs[ip] = memoryRecord
		s.expiryHeap.Update(item, expiresAt)

		r.logger.Info().
			Str("ip", ip).
			Str("reason", reason).
			Time("until", expiresAt).
			Msg("Actualizar registro de restricción de IP")
		return nil
	}

	// Garantizar la capacidad
	r.ensureShardCapacity(s)

	// Añadir nuevo registro
	s.blockedIPs[ip] = memoryRecord

	item := ipExpiryItemPool.Get().(*IPExpiryItem)
	item.ip = ip
	item.expiresAt = expiresAt

	s.expiryItems[ip] = item
	heap.Push(&s.expiryHeap, item)

	r.Metrics.TotalBlocked.Add(1)
	r.Metrics.CurrentBlocked.Add(1)

	r.logger.Info().
		Str("ip", ip).
		Str("reason", reason).
		Time("until", expiresAt).
		Msg("La IP ya está restringida")

	return nil
}

// IsIPBlocked comprueba si una IP está restringida - devuelve resultado simplificado
func (r *MemoryIPRecorder) IsIPBlocked(ip string) (bool, *model.BlockedIPRecord) {
	// Usar defer recover para prevenir cualquier panic
	defer func() {
		if r := recover(); r != nil {
			// Ante un panic, permitir directamente
		}
	}()

	s := r.getShard(ip)

	// Acceso sin bloqueo, se acepta el riesgo de concurrencia
	memoryRecord, exists := s.blockedIPs[ip]
	if !exists {
		r.Metrics.CacheMisses.Add(1)
		return false, nil
	}

	// Hora de la comprobación de seguridad
	now := time.Now()
	if memoryRecord.BlockedUntil.IsZero() || now.After(memoryRecord.BlockedUntil) {
		// Expirado, devolver false directamente sin eliminar (evita bloqueos)
		r.Metrics.CacheMisses.Add(1)
		return false, nil
	}

	r.Metrics.CacheHits.Add(1)

	// Convertir a BlockedIPRecord completo para devolver (solo campos presentes en memoria)
	fullRecord := &model.BlockedIPRecord{
		IP:           memoryRecord.IP,
		BlockedUntil: memoryRecord.BlockedUntil,
		// Campos como Reason y RequestUri no se guardan en memoria, permanecen en cero
	}

	return true, fullRecord
}

// GetBlockedIPs obtiene todas las IPs restringidas - devuelve registros simplificados
func (r *MemoryIPRecorder) GetBlockedIPs() ([]model.BlockedIPRecord, error) {
	now := time.Now()
	records := make([]model.BlockedIPRecord, 0, r.capacity/10)

	// Recopilar los registros de todos los shards
	for _, s := range r.shards {
		s.mu.RLock()
		for _, memoryRecord := range s.blockedIPs {
			if now.Before(memoryRecord.BlockedUntil) {
				// Convertir a BlockedIPRecord completo (solo campos presentes en memoria)
				fullRecord := model.BlockedIPRecord{
					IP:           memoryRecord.IP,
					BlockedUntil: memoryRecord.BlockedUntil,
					// Campos como Reason y RequestUri no se guardan en memoria, permanecen en cero
				}
				records = append(records, fullRecord)
			}
		}
		s.mu.RUnlock()
	}

	return records, nil
}

// GetMetrics obtiene las métricas de monitorización
func (r *MemoryIPRecorder) GetMetrics() *Metrics {
	// Actualizar la cantidad de bloqueos actual
	var current uint64
	for _, s := range r.shards {
		s.mu.RLock()
		current += uint64(len(s.blockedIPs))
		s.mu.RUnlock()
	}
	r.Metrics.CurrentBlocked.Store(current)

	return r.Metrics
}

// Close cierra el registrador y libera recursos
func (r *MemoryIPRecorder) Close() error {
	close(r.stopCleaner)
	return nil
}

// RingBuffer buffer circular
type RingBuffer struct {
	buffer []model.BlockedIPRecord
	head   uint64
	tail   uint64
	size   uint64
	mask   uint64
	mu     sync.Mutex
}

// NewRingBuffer crea un buffer circular
func NewRingBuffer(size int) *RingBuffer {
	// Asegurar que size sea potencia de 2
	actualSize := 1
	for actualSize < size {
		actualSize <<= 1
	}

	return &RingBuffer{
		buffer: make([]model.BlockedIPRecord, actualSize),
		size:   uint64(actualSize),
		mask:   uint64(actualSize - 1),
	}
}

// Push añade un elemento
func (rb *RingBuffer) Push(record model.BlockedIPRecord) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	next := (rb.tail + 1) & rb.mask
	if next == rb.head {
		return false // Buffer lleno
	}

	rb.buffer[rb.tail] = record
	rb.tail = next
	return true
}

// PopBatch extrae elementos por lotes
func (rb *RingBuffer) PopBatch(maxBatch int) []model.BlockedIPRecord {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.head == rb.tail {
		return nil
	}

	batch := make([]model.BlockedIPRecord, 0, maxBatch)
	for i := 0; i < maxBatch && rb.head != rb.tail; i++ {
		batch = append(batch, rb.buffer[rb.head])
		rb.head = (rb.head + 1) & rb.mask
	}

	return batch
}

// Len obtiene la cantidad actual de elementos
func (rb *RingBuffer) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.tail >= rb.head {
		return int(rb.tail - rb.head)
	}
	return int(rb.size - rb.head + rb.tail)
}

// CircuitBreaker interruptor de circuito
type CircuitBreaker struct {
	failures      atomic.Uint32
	successCount  atomic.Uint32
	lastFailure   atomic.Value  // time.Time
	state         atomic.Uint32 // 0: closed, 1: open, 2: half-open
	threshold     uint32
	timeout       time.Duration
	recoveryCount uint32
}

// NewCircuitBreaker crea un circuit breaker
func NewCircuitBreaker(threshold uint32, timeout time.Duration, recoveryCount uint32) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:     threshold,
		timeout:       timeout,
		recoveryCount: recoveryCount,
	}
}

// IsOpen comprueba si el circuit breaker está abierto
func (cb *CircuitBreaker) IsOpen() bool {
	state := cb.state.Load()
	if state == 0 { // closed
		return false
	}

	if state == 1 { // open
		lastFailure, ok := cb.lastFailure.Load().(time.Time)
		if ok && time.Since(lastFailure) > cb.timeout {
			// Intentar entrar en estado semi-abierto
			cb.state.CompareAndSwap(1, 2)
			cb.successCount.Store(0)
			return false
		}
		return true
	}

	// half-open
	return false
}

// RecordSuccess registra un éxito
func (cb *CircuitBreaker) RecordSuccess() {
	state := cb.state.Load()
	if state == 2 { // half-open
		count := cb.successCount.Add(1)
		if count >= cb.recoveryCount {
			cb.state.Store(0) // closed
			cb.failures.Store(0)
		}
	} else if state == 0 { // closed
		cb.failures.Store(0)
	}
}

// RecordFailure registra un fallo
func (cb *CircuitBreaker) RecordFailure() {
	failures := cb.failures.Add(1)
	cb.lastFailure.Store(time.Now())

	if failures >= cb.threshold {
		cb.state.Store(1) // open
	}
}

// MongoIPRecorder registrador de IPs implementado en MongoDB
type MongoIPRecorder struct {
	client         *mongo.Client
	database       string
	collection     string
	memory         *MemoryIPRecorder
	logger         zerolog.Logger
	config         RecorderConfig
	metrics        *Metrics
	circuitBreaker *CircuitBreaker

	// Usar buffer circular en lugar de channel
	writeBuffer     *RingBuffer
	stopWriter      chan struct{}
	avgWriteLatency atomic.Value // time.Duration
}

// Instancia singleton
var (
	mongoIPRecorderOnce     sync.Once
	mongoIPRecorderInstance *MongoIPRecorder
)

// NewMongoIPRecorder crea un registrador de IPs en MongoDB
func NewMongoIPRecorder(client *mongo.Client, database string, capacity int, logger zerolog.Logger) *MongoIPRecorder {
	config := DefaultConfig()
	config.Capacity = capacity
	return NewMongoIPRecorderWithConfig(client, database, config, logger)
}

// NewMongoIPRecorderWithConfig crea un registrador de IPs en MongoDB con configuración
func NewMongoIPRecorderWithConfig(client *mongo.Client, database string, config RecorderConfig, logger zerolog.Logger) *MongoIPRecorder {
	mongoIPRecorderOnce.Do(func() {
		var blockedIPs model.BlockedIPRecord

		// Crear primero el registrador en memoria
		memoryRecorder := NewMemoryIPRecorderWithConfig(config, logger)

		recorder := &MongoIPRecorder{
			client:         client,
			database:       database,
			collection:     blockedIPs.GetCollectionName(),
			memory:         memoryRecorder,
			logger:         logger,
			config:         config,
			metrics:        memoryRecorder.Metrics, // Compartir métricas
			circuitBreaker: NewCircuitBreaker(5, 30*time.Second, 3),
			writeBuffer:    NewRingBuffer(config.WriteQueueSize),
			stopWriter:     make(chan struct{}),
		}

		// Iniciar la escritura por lotes
		go recorder.adaptiveBatchWriteLoop()

		mongoIPRecorderInstance = recorder
		logger.Info().Msg("Crear nueva instancia de MongoIPRecorder")
	})

	return mongoIPRecorderInstance
}

// adaptiveBatchSize ajusta dinámicamente el tamaño del lote
func (r *MongoIPRecorder) adaptiveBatchSize() int {
	latency, ok := r.avgWriteLatency.Load().(time.Duration)
	if !ok {
		return r.config.BatchSize
	}

	switch {
	case latency > 200*time.Millisecond:
		return r.config.BatchSize / 2
	case latency > 100*time.Millisecond:
		return r.config.BatchSize * 3 / 4
	case latency < 50*time.Millisecond:
		return r.config.BatchSize * 2
	default:
		return r.config.BatchSize
	}
}

// adaptiveBatchWriteLoop bucle de escritura por lotes adaptativo
func (r *MongoIPRecorder) adaptiveBatchWriteLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			batchSize := r.adaptiveBatchSize()
			batch := r.writeBuffer.PopBatch(batchSize)
			if len(batch) > 0 {
				r.metrics.WriteQueueSize.Store(uint64(r.writeBuffer.Len()))
				go r.flushBatchWithRetry(batch)
			}

		case <-r.stopWriter:
			// Escribir los datos restantes antes de cerrar
			batch := r.writeBuffer.PopBatch(r.config.BatchSize * 10)
			if len(batch) > 0 {
				r.flushBatchWithRetry(batch)
			}
			return
		}
	}
}

// flushBatchWithRetry escritura por lotes con reintentos
func (r *MongoIPRecorder) flushBatchWithRetry(batch []model.BlockedIPRecord) {
	for retry := 0; retry < r.config.MaxRetries; retry++ {
		start := time.Now()
		err := r.flushBatch(batch)
		latency := time.Since(start)

		// Actualizar la latencia media
		r.avgWriteLatency.Store(latency)

		if err == nil {
			r.circuitBreaker.RecordSuccess()
			return
		}

		r.circuitBreaker.RecordFailure()

		if retry < r.config.MaxRetries-1 {
			// Retroceso exponencial
			time.Sleep(time.Duration(1<<retry) * time.Second)
		}
	}

	r.logger.Error().
		Int("batch_size", len(batch)).
		Msg("Error de escritura por lotes en MongoDB, se alcanzó el máximo de reintentos")
}

// flushBatch escritura por lotes a MongoDB
func (r *MongoIPRecorder) flushBatch(batch []model.BlockedIPRecord) error {
	if len(batch) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.WriteTimeout)
	defer cancel()

	collection := r.client.Database(r.database).Collection(r.collection)

	// Operación de inserción por lotes - inserta nuevos registros directamente
	// Nota: esto conservará el historial completo, considere:
	// 1. Crear índice TTL para limpiar registros expirados: db.collection.createIndex({"blockedUntil": 1}, {expireAfterSeconds: 0})
	// 2. Crear índice para rendimiento de consultas: db.collection.createIndex({"ip": 1, "blockedUntil": -1})
	// 3. Limpiar periódicamente los registros expirados para evitar acumulación excesiva
	models := make([]mongo.WriteModel, len(batch))
	for i, record := range batch {
		models[i] = mongo.NewInsertOneModel().SetDocument(record)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(ctx, models, opts)
	if err != nil {
		r.logger.Error().
			Err(err).
			Int("batch_size", len(batch)).
			Msg("Error al insertar registros de restricción de IP en MongoDB por lotes")
		return err
	}

	r.logger.Debug().
		Int("batch_size", len(batch)).
		Msg("Registros de restricción de IP insertados en MongoDB por lotes con éxito")

	return nil
}

// RecordBlockedIP registra una IP restringida
func (r *MongoIPRecorder) RecordBlockedIP(ip string, reason string, requestUri string, duration time.Duration) error {
	// Registrar primero en memoria
	err := r.memory.RecordBlockedIP(ip, reason, requestUri, duration)
	if err != nil {
		return err
	}

	// Si el circuit breaker está abierto, retornar directamente
	if r.circuitBreaker.IsOpen() {
		r.logger.Warn().
			Str("ip", ip).
			Msg("Circuit breaker de MongoDB abierto, se omite la persistencia")
		return nil
	}

	// Escritura asíncrona al buffer circular
	now := time.Now()
	record := model.BlockedIPRecord{
		IP:           ip,
		Reason:       reason,
		RequestUri:   requestUri,
		BlockedAt:    now,
		BlockedUntil: now.Add(duration),
	}

	if !r.writeBuffer.Push(record) {
		r.logger.Warn().
			Str("ip", ip).
			Msg("Buffer de escritura de MongoDB lleno, registro descartado")
	}

	return nil
}

// IsIPBlocked comprueba si una IP está restringida
func (r *MongoIPRecorder) IsIPBlocked(ip string) (bool, *model.BlockedIPRecord) {
	return r.memory.IsIPBlocked(ip)
}

// GetBlockedIPs obtiene todas las IPs restringidas
func (r *MongoIPRecorder) GetBlockedIPs() ([]model.BlockedIPRecord, error) {
	return r.memory.GetBlockedIPs()
}

// GetMetrics obtiene las métricas - garantiza que las métricas en memoria estén actualizadas
func (r *MongoIPRecorder) GetMetrics() *Metrics {
	// Obtener primero las métricas más recientes del registrador en memoria
	r.memory.GetMetrics()

	// Luego actualizar las métricas relacionadas con MongoDB
	r.metrics.WriteQueueSize.Store(uint64(r.writeBuffer.Len()))

	return r.metrics
}

// Close cierra el registrador y libera recursos
func (r *MongoIPRecorder) Close() error {
	close(r.stopWriter)
	return r.memory.Close()
}
