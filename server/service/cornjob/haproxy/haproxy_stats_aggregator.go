package cornjob

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	mongodb "github.com/anticoi/EscudoWAF/pkg/database/mongo"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/service/daemon"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// StatsData estructura que registra los últimos datos estadísticos para calcular diferencias
type StatsData struct {
	TargetName string
	LastStats  model.HAProxyStats
	LastTime   time.Time
	ResetCount int // Para rastrear el número de reinicios
}

// StatsAggregator agregador de estadísticas de HAProxy
type StatsAggregator struct {
	runner          daemon.ServiceRunner
	dbName          string
	lastStats       map[string]*StatsData // Usar el nombre del backend como clave
	lastStatsLoaded bool                  // Marca si los datos base ya se cargaron de la base de datos
	targetFilter    map[string]bool       // Lista de backends filtrada
	log             zerolog.Logger
	isRunning       bool // Indica si el agregador está en ejecución
}

// NewStatsAggregator crea un nuevo agregador de datos
func NewStatsAggregator(runner daemon.ServiceRunner, dbName string, targetList []string) (*StatsAggregator, error) {
	logger := config.GetLogger().With().Str("component", "cronjob-haproxy-stats-aggregator").Logger()
	// Inicializar targetFilter
	targetFilter := make(map[string]bool)
	for _, target := range targetList {
		targetFilter[target] = true
	}

	agg := &StatsAggregator{
		runner:          runner,
		dbName:          dbName,
		lastStats:       make(map[string]*StatsData),
		lastStatsLoaded: false,
		targetFilter:    targetFilter,
		log:             logger,
		isRunning:       false,
	}

	// Asegurar que existan las colecciones e índices necesarios
	err := agg.ensureCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collections: %w", err)
	}

	return agg, nil
}

// UpdateTargetList actualiza la lista de backends monitorizados
func (a *StatsAggregator) UpdateTargetList(targetList []string) {
	// Crear nuevo filtro
	newFilter := make(map[string]bool)
	for _, target := range targetList {
		newFilter[target] = true
	}

	// Comprobar cambios de estado
	hadNoTargets := len(a.targetFilter) == 0
	willHaveNoTargets := len(newFilter) == 0
	hasNewTargets := len(newFilter) > 0

	// Buscar objetivos que ya no se monitorizan
	for target := range a.targetFilter {
		if !newFilter[target] {
			a.log.Info().Str("target", target).Msg("Removing target from monitoring")
			// Eliminar de memoria, pero conservar la línea base en la base de datos (por si se vuelve a añadir)
			delete(a.lastStats, target)
		}
	}

	// Actualizar el filtro
	prevFilter := a.targetFilter
	a.targetFilter = newFilter

	// Manejar la transición de con objetivo a sin objetivo
	if !hadNoTargets && willHaveNoTargets {
		a.log.Warn().Int("previous_targets", len(prevFilter)).Msg("All targets removed, aggregator entering standby mode")

		// Vaciar las estadísticas para reducir el uso de memoria
		a.lastStats = make(map[string]*StatsData)

		// Restablecer el estado lastStatsLoaded para inicializar correctamente al añadir el próximo objetivo
		a.lastStatsLoaded = false

		// Guardar el estado final en la base de datos y marcar como completado
		if a.isRunning {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := a.saveLastStats(ctx); err != nil {
				a.log.Error().Err(err).Msg("Failed to save final stats when entering standby mode")
			}
		}
	} else if hadNoTargets && hasNewTargets && a.isRunning {
		// Si pasa de sin objetivo a con objetivo, inicializar el estado
		a.log.Info().Msg("Targets added to previously empty target list, initializing stats collection")

		// Crear el contexto para la inicialización
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Intentar inicializar los datos estadísticos
		if err := a.initializeStats(ctx); err != nil {
			a.log.Error().Err(err).Msg("Failed to initialize stats after targets were added")
		}
	} else {
		// Registrar los objetivos añadidos
		for target := range newFilter {
			if !prevFilter[target] {
				a.log.Info().Str("target", target).Msg("Adding new target to monitoring")
				// Los nuevos objetivos se inicializan automáticamente en la próxima recopilación de datos
			}
		}
	}
}

// ensureCollections garantiza que existan las colecciones e índices necesarios
func (a *StatsAggregator) ensureCollections() error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Crear el contexto
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Obtener los nombres de todas las colecciones
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Definición de las colecciones a comprobar
	var haproxyStatsBaseline model.HAProxyStatsBaseline
	var haproxyMinuteStats model.HAProxyMinuteStats
	baselineCollName := haproxyStatsBaseline.GetCollectionName()
	minuteStatsCollName := haproxyMinuteStats.GetCollectionName()

	// Comprobar si existe la colección de series temporales
	requiredCollections := []string{
		baselineCollName,
		minuteStatsCollName,
		"conn_rate",
		"scur",
		"rate",
		"req_rate",
	}

	// Usar slices.Contains para comprobar que existen todas las colecciones necesarias
	allCollectionsExist := true
	for _, collName := range requiredCollections {
		if !slices.Contains(collections, collName) {
			allCollectionsExist = false
			break
		}
	}

	// Si existen todas las colecciones necesarias, la base de datos se considera inicializada
	if allCollectionsExist {
		a.log.Info().
			Strs("requiredCollections", requiredCollections).
			Msg("All required collections exist, skipping initialization")
		return nil
	}

	// Si no existen todas las colecciones, ejecutar la inicialización completa
	a.log.Info().Msg("Some required collections are missing, initializing database")

	// Crear índices de la colección de datos base
	baselineColl := db.Collection(baselineCollName)
	indexOpts := options.Index().SetUnique(true)
	_, err = baselineColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "target_name", Value: 1}},
		Options: indexOpts,
	})
	if err != nil {
		return fmt.Errorf("failed to create index for baseline collection: %w", err)
	}
	a.log.Info().Msg("Created baseline collection index")

	// Crear índices de la colección de estadísticas por minuto
	minuteStatsColl := db.Collection(minuteStatsCollName)
	_, err = minuteStatsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// Índice de marca de tiempo - soporta consultas por rango de tiempo
			Keys: bson.D{{Key: "timestamp", Value: 1}},
		},
		{
			// Índice principal de consulta - índice principal por destino y tiempo
			Keys: bson.D{
				{Key: "target_name", Value: 1},
				{Key: "timestamp", Value: 1},
			},
		},
		{
			// Índice de agregación por hora
			Keys: bson.D{
				{Key: "target_name", Value: 1},
				{Key: "date", Value: 1},
				{Key: "hour", Value: 1},
			},
		},
		{
			// Índice de agregación por grupos de seis horas
			Keys: bson.D{
				{Key: "target_name", Value: 1},
				{Key: "date", Value: 1},
				{Key: "hourGroupSix", Value: 1},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create indexes for minute stats collection: %w", err)
	}
	a.log.Info().Msg("Created minute stats collection indexes")

	// Crear la colección de series temporales de estadísticas en tiempo real
	realtimeMetrics := []string{"conn_rate", "scur", "rate", "req_rate"}
	for _, metric := range realtimeMetrics {
		// Usar slices.Contains para comprobar si la colección ya existe
		if !slices.Contains(collections, metric) {
			// Crear la colección de series temporales
			timeSeriesOpts := options.TimeSeries().
				SetTimeField("timestamp").
				SetMetaField("metadata").
				SetGranularity("seconds")
			createOpts := options.CreateCollection().
				SetTimeSeriesOptions(timeSeriesOpts).
				SetExpireAfterSeconds(3600) // Establecer expiración de 1 hora
			if err := db.CreateCollection(ctx, metric, createOpts); err != nil {
				return fmt.Errorf("failed to create timeseries collection %s: %w", metric, err)
			}
			a.log.Info().Msgf("Created timeseries collection %s", metric)
		}
	}

	a.log.Info().Msg("Database initialization completed successfully")
	return nil
}

// Start inicia el agregador
func (a *StatsAggregator) Start(ctx context.Context) error {
	if a.isRunning {
		return errors.New("aggregator is already running")
	}
	a.isRunning = true

	// Usar un contexto con timeout para las operaciones de inicialización
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Intentar cargar los datos base desde la base de datos
	if err := a.loadLastStats(initCtx); err != nil {
		a.log.Warn().Err(err).Msg("Failed to load lastStats, will initialize new baseline")
		// Si falla la carga, inicializar una nueva línea base
		if err := a.initializeStats(initCtx); err != nil {
			a.isRunning = false
			return fmt.Errorf("failed to initialize stats: %w", err)
		}
	} else {
		a.log.Info().Msg("Successfully loaded lastStats data from database")
		a.lastStatsLoaded = true
		// Obtener los datos más recientes y comprobar si HAProxy se reinició
		if err := a.checkHAProxyResetOnStartup(initCtx); err != nil {
			a.log.Warn().Err(err).Msg("Error checking HAProxy reset on startup, continuing anyway")
		}
	}

	return nil
}

// Stop detiene el agregador
func (a *StatsAggregator) Stop(ctx context.Context) error {
	if !a.isRunning {
		return nil // Ya está detenido, no hay nada más que hacer
	}
	a.isRunning = false

	// Crear un contexto con timeout
	saveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Guardar el último estado
	if err := a.saveLastStats(saveCtx); err != nil {
		a.log.Warn().Err(err).Msg("Failed to save lastStats during shutdown")
		// No devolver error, continuar con el cierre
	}

	a.log.Info().Msg("Aggregator successfully stopped")
	return nil
}

// loadLastStats carga los datos base de lastStats desde MongoDB
func (a *StatsAggregator) loadLastStats(ctx context.Context) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	var haproxyStatsBaseline model.HAProxyStatsBaseline
	collection := db.Collection(haproxyStatsBaseline.GetCollectionName())

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to query lastStats: %w", err)
	}
	defer cursor.Close(ctx)

	var baselines []model.HAProxyStatsBaseline
	if err = cursor.All(ctx, &baselines); err != nil {
		return fmt.Errorf("failed to decode lastStats documents: %w", err)
	}

	if len(baselines) == 0 {
		return errors.New("no baseline data found")
	}

	loadCount := 0
	for _, doc := range baselines {
		targetName := doc.TargetName
		if targetName == "" {
			a.log.Warn().Interface("doc", doc).Msg("Invalid target_name in baseline document")
			continue
		}

		// Cargar en memoria
		a.lastStats[targetName] = &StatsData{
			TargetName: targetName,
			LastStats:  doc.GetStats(), // Usar el nuevo método para obtener estadísticas
			LastTime:   doc.Timestamp,
			ResetCount: int(doc.ResetCount),
		}

		loadCount++
	}

	a.log.Info().Int("count", loadCount).Msg("Loaded target baselines from database")
	if loadCount == 0 {
		return errors.New("no valid baseline data loaded")
	}
	return nil
}

// saveLastStats guarda los datos base de lastStats en MongoDB
func (a *StatsAggregator) saveLastStats(ctx context.Context) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	var haproxyStatsBaseline model.HAProxyStatsBaseline
	collection := db.Collection(haproxyStatsBaseline.GetCollectionName())

	var errs []error
	for targetName, stats := range a.lastStats {
		// Crear el documento base
		doc := model.HAProxyStatsBaseline{
			TargetName: targetName,
			Timestamp:  stats.LastTime,
			ResetCount: int32(stats.ResetCount),
		}
		// Establecer las estadísticas
		doc.SetStats(stats.LastStats)

		// Guardar con operación upsert
		filter := bson.M{"target_name": targetName}
		update := bson.M{"$set": doc}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to save lastStats for %s: %w", targetName, err))
			a.log.Error().Err(err).Str("target", targetName).Msg("Failed to save baseline for target")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while saving baselines: %v", errs)
	}
	return nil
}

// checkHAProxyResetOnStartup comprueba al inicio si HAProxy se reinició
func (a *StatsAggregator) checkHAProxyResetOnStartup(ctx context.Context) error {
	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("checkHAProxyResetOnStartup: failed to get stats")
	}

	resetDetected := false
	newTargets := make(map[string]bool)

	// Encontrar todos los backends activos actuales
	for _, stat := range stats.Stats {
		if stat.Type == "frontend" {
			newTargets[stat.Name] = true
		}
	}

	// Comprobar los backends monitorizados actuales
	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.targetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		currentStats := model.NativeStatsToHAProxyStats(stat.Stats)
		lastStat, exists := a.lastStats[stat.Name]

		if !exists {
			// Nuevo backend, inicializar
			a.lastStats[stat.Name] = &StatsData{
				TargetName: stat.Name,
				LastStats:  currentStats,
				LastTime:   time.Now(),
				ResetCount: 0,
			}
			a.log.Info().Str("target", stat.Name).Msg("New target detected during startup")
			continue
		}

		// Detectar si se reinició
		if model.DetectReset(lastStat.LastStats, currentStats) {
			a.log.Warn().
				Str("target", stat.Name).
				Int("reset_count", lastStat.ResetCount+1).
				Msg("Detected HAProxy reset for target during startup")

			resetDetected = true

			// Guardar registro con incremento cero
			err := a.saveMinuteMetrics(ctx, stat.Name, model.CreateZeroStats(), time.Now())

			if err != nil {
				a.log.Error().
					Err(err).
					Str("target", stat.Name).
					Msg("Failed to save zero metrics for target after reset")
			}

			// Actualizar el contador de reinicios y la línea base
			lastStat.ResetCount++
			lastStat.LastStats = currentStats
			lastStat.LastTime = time.Now()
		}
	}

	// Comprobar backends que ya no existen
	for targetName := range a.lastStats {
		if a.targetFilter[targetName] && !newTargets[targetName] {
			a.log.Warn().
				Str("target", targetName).
				Msg("Target is in monitoring list but not found in HAProxy stats")
		}
	}

	if resetDetected {
		a.log.Warn().Msg("HAProxy reset detected during startup, metrics will be adjusted")
		// Guardar el estado tras el reinicio
		if err := a.saveLastStats(ctx); err != nil {
			a.log.Error().Err(err).Msg("Failed to save lastStats after reset detection")
			return fmt.Errorf("failed to save state after reset: %w", err)
		}
	} else {
		a.log.Info().Msg("No HAProxy reset detected during startup")
	}

	return nil
}

// initializeStats inicializa los datos estadísticos
func (a *StatsAggregator) initializeStats(ctx context.Context) error {
	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("initializeStats: failed to get stats")
	}

	now := time.Now()
	a.lastStats = make(map[string]*StatsData) // Borrar los datos existentes

	for _, stat := range stats.Stats {
		if stat.Type == "frontend" && a.targetFilter[stat.Name] && stat.Stats != nil {
			currentStats := model.NativeStatsToHAProxyStats(stat.Stats)

			a.lastStats[stat.Name] = &StatsData{
				TargetName: stat.Name,
				LastStats:  currentStats,
				LastTime:   now,
				ResetCount: 0,
			}
		}
	}

	// Comprobar si hay al menos un objetivo
	if len(a.lastStats) == 0 {
		a.log.Warn().Msg("No valid targets found to initialize, aggregator will wait for targets to be added")
		return nil // Devolver nil en lugar de error, permite iniciar el servicio en estado inactivo
	}

	// Guardar la línea base inicial en la base de datos
	if err := a.saveLastStats(ctx); err != nil {
		a.log.Error().Err(err).Msg("Failed to save initial lastStats")
		return fmt.Errorf("failed to save initial state: %w", err)
	}

	a.log.Info().Int("count", len(a.lastStats)).Msg("Statistics initialized for backends")
	return nil
}

// CollectRealtimeMetrics recopila métricas en tiempo real
func (a *StatsAggregator) CollectRealtimeMetrics(ctx context.Context) error {
	if !a.isRunning {
		return errors.New("aggregator is not running")
	}

	// Si no hay objetivos, omitir la recopilación
	if len(a.targetFilter) == 0 {
		a.log.Info().Msg("Skipping realtime metrics collection: no targets configured")
		return nil
	}

	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("CollectRealtimeMetrics: failed to get stats")
	}

	return a.processRealtimeMetrics(ctx, stats, time.Now())
}

// processRealtimeMetrics procesa las métricas en tiempo real
func (a *StatsAggregator) processRealtimeMetrics(ctx context.Context, stats models.NativeStats, t time.Time) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Para almacenar datos agregados
	totalConnRate := int64(0)
	totalScur := int64(0)
	totalRate := int64(0)
	totalReqRate := int64(0)

	// Preparar el documento para guardar en series temporales
	documents := map[string][]model.TimeSeriesMetric{
		"conn_rate": {},
		"scur":      {},
		"rate":      {},
		"req_rate":  {},
	}

	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.targetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		// Recopilar métricas en tiempo real de cada backend
		if stat.Stats.ConnRate != nil {
			totalConnRate += *stat.Stats.ConnRate
			documents["conn_rate"] = append(documents["conn_rate"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.ConnRate,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}

		if stat.Stats.Scur != nil {
			totalScur += *stat.Stats.Scur
			documents["scur"] = append(documents["scur"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.Scur,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}

		if stat.Stats.Rate != nil {
			totalRate += *stat.Stats.Rate
			documents["rate"] = append(documents["rate"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.Rate,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}

		if stat.Stats.ReqRate != nil {
			totalReqRate += *stat.Stats.ReqRate
			documents["req_rate"] = append(documents["req_rate"], model.TimeSeriesMetric{
				Timestamp: t,
				Value:     *stat.Stats.ReqRate,
				Metadata:  model.TimeSeriesMeta{Target: stat.Name},
			})
		}
	}

	// Añadir las métricas totales de todos los backends
	documents["conn_rate"] = append(documents["conn_rate"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalConnRate,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	documents["scur"] = append(documents["scur"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalScur,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	documents["rate"] = append(documents["rate"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalRate,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	documents["req_rate"] = append(documents["req_rate"], model.TimeSeriesMetric{
		Timestamp: t,
		Value:     totalReqRate,
		Metadata:  model.TimeSeriesMeta{Target: "all"},
	})

	// Guardar en la colección de series temporales de MongoDB
	var errs []error
	for metric, docs := range documents {
		if len(docs) > 0 {
			collection := db.Collection(metric)

			// Convertir a formato apto para InsertMany
			insertDocs := make([]interface{}, len(docs))
			for i, doc := range docs {
				insertDocs[i] = doc
			}

			_, err := collection.InsertMany(ctx, insertDocs)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to insert %s metrics: %w", metric, err))
				a.log.Error().Err(err).Str("metric", metric).Msg("Failed to insert metrics")
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while inserting metrics: %v", errs)
	}
	return nil
}

// CollectMinuteMetrics recopila métricas por minuto
func (a *StatsAggregator) CollectMinuteMetrics(ctx context.Context) error {
	if !a.isRunning {
		return errors.New("aggregator is not running")
	}

	// Si no hay objetivos, omitir la recopilación
	if len(a.targetFilter) == 0 {
		a.log.Info().Msg("Skipping minute metrics collection: no targets configured")
		return nil
	}

	stats, err := a.runner.GetStats()
	if err != nil {
		a.log.Error().Err(err).Msg("CollectMinuteMetrics: failed to get stats")
	}

	t := time.Now()
	err = a.processMinuteMetrics(ctx, stats, t)
	if err != nil {
		return fmt.Errorf("failed to process minute metrics: %w", err)
	}

	// Guardar datos base tras cada ciclo, para perder como máximo un ciclo ante un fallo
	saveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := a.saveLastStats(saveCtx); err != nil {
		a.log.Error().Err(err).Msg("Failed to save lastStats after minute metrics")
		// No devolver error, continuar la ejecución
	}

	return nil
}

// processMinuteMetrics procesa las métricas por minuto
func (a *StatsAggregator) processMinuteMetrics(ctx context.Context, stats models.NativeStats, t time.Time) error {
	// Para agregar las métricas totales de todos los backends
	aggregateStats := model.HAProxyStats{}

	// Crear el mapa de backends activos actual
	activeTargets := make(map[string]bool)
	for _, stat := range stats.Stats {
		if stat.Type == "frontend" {
			activeTargets[stat.Name] = true
		}
	}

	// Detectar y registrar backends en la lista de monitorización que ya no están activos
	for target := range a.targetFilter {
		if !activeTargets[target] {
			a.log.Warn().Str("target", target).Msg("Target is in monitoring list but not active")
		}
	}

	// Procesar por cada backend
	var errs []error
	for _, stat := range stats.Stats {
		if stat.Type != "frontend" || !a.targetFilter[stat.Name] || stat.Stats == nil {
			continue
		}

		currentStats := model.NativeStatsToHAProxyStats(stat.Stats)
		lastStat, exists := a.lastStats[stat.Name]

		if !exists {
			// Nuevo backend, inicializar
			a.lastStats[stat.Name] = &StatsData{
				TargetName: stat.Name,
				LastStats:  currentStats,
				LastTime:   t,
				ResetCount: 0,
			}
			a.log.Info().Str("target", stat.Name).Msg("New target detected during metrics collection")
			continue // Omitir el cálculo de diferencia, es la primera vez que se ve este backend
		}

		// Detectar reinicio
		if model.DetectReset(lastStat.LastStats, currentStats) {
			a.log.Warn().
				Str("target", stat.Name).
				Int("reset_count", lastStat.ResetCount+1).
				Msg("Detected HAProxy reset for target")

			// Registrar incremento cero en lugar de omitir
			zeroMetrics := model.CreateZeroStats()

			err := a.saveMinuteMetrics(ctx, stat.Name, zeroMetrics, t)

			if err != nil {
				a.log.Error().
					Err(err).
					Str("target", stat.Name).
					Msg("Failed to save zero metrics for target after reset")
				errs = append(errs, fmt.Errorf("failed to save zero metrics for %s: %w", stat.Name, err))
			}

			// Actualizar el contador de reinicios y la línea base
			lastStat.ResetCount++
			lastStat.LastStats = currentStats
			lastStat.LastTime = t

			// También añadir valores cero a las estadísticas agregadas - no afectan la agregación
			continue
		}

		// Calcular la diferencia y guardar
		deltas := model.CalculateStatsDelta(lastStat.LastStats, currentStats)

		// Guardar la diferencia de un backend individual
		err := a.saveMinuteMetrics(ctx, stat.Name, deltas, t)

		if err != nil {
			a.log.Error().
				Err(err).
				Str("target", stat.Name).
				Msg("Failed to save minute metrics for target")
			errs = append(errs, fmt.Errorf("failed to save metrics for %s: %w", stat.Name, err))
		}

		// Acumular en las estadísticas agregadas
		aggregateStats.Bin += deltas.Bin
		aggregateStats.Bout += deltas.Bout

		aggregateStats.Hrsp1xx += deltas.Hrsp1xx
		aggregateStats.Hrsp2xx += deltas.Hrsp2xx
		aggregateStats.Hrsp3xx += deltas.Hrsp3xx
		aggregateStats.Hrsp4xx += deltas.Hrsp4xx
		aggregateStats.Hrsp5xx += deltas.Hrsp5xx
		aggregateStats.HrspOther += deltas.HrspOther

		aggregateStats.Dreq += deltas.Dreq
		aggregateStats.Dresp += deltas.Dresp
		aggregateStats.Ereq += deltas.Ereq
		aggregateStats.Dcon += deltas.Dcon
		aggregateStats.Dses += deltas.Dses
		aggregateStats.Econ += deltas.Econ
		aggregateStats.Eresp += deltas.Eresp

		// Para valores máximos, usar la suma de todos los backends
		aggregateStats.ReqRateMax += deltas.ReqRateMax
		aggregateStats.ConnRateMax += deltas.ConnRateMax
		aggregateStats.RateMax += deltas.RateMax
		aggregateStats.Smax += deltas.Smax

		// Para valores totales, usar la suma de todos los backends
		aggregateStats.ConnTot += deltas.ConnTot
		aggregateStats.Stot += deltas.Stot
		aggregateStats.ReqTot += deltas.ReqTot

		// Actualizar los datos estadísticos anteriores
		lastStat.LastStats = currentStats
		lastStat.LastTime = t
	}

	// Guardar las métricas totales de todos los backends
	err := a.saveMinuteMetrics(ctx, "all", aggregateStats, t)

	if err != nil {
		a.log.Error().Err(err).Msg("Failed to save aggregate minute metrics")
		errs = append(errs, fmt.Errorf("failed to save aggregate metrics: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while processing minute metrics: %v", errs)
	}
	return nil
}

// saveMinuteMetrics guarda las métricas por minuto en MongoDB - usa estructura plana
func (a *StatsAggregator) saveMinuteMetrics(ctx context.Context, targetName string, metrics model.HAProxyStats, timestamp time.Time) error {
	db, err := mongodb.GetDatabase(a.dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Crear la estructura de documento plana
	minuteStats := model.HAProxyMinuteStats{
		TargetName:   targetName,
		Date:         timestamp.Format("2006-01-02"),
		Hour:         timestamp.Hour(),
		HourGroupSix: timestamp.Hour() / 6,
		Minute:       timestamp.Minute(),
		Timestamp:    timestamp,
	}
	// Establecer las estadísticas
	minuteStats.SetStats(metrics)

	// Insertar en la colección de la base de datos
	collection := db.Collection(minuteStats.GetCollectionName())
	_, err = collection.InsertOne(ctx, minuteStats)
	if err != nil {
		return fmt.Errorf("failed to insert minute stats: %w", err)
	}

	return nil
}
