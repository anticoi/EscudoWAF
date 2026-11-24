package service

import (
	"context"
	"fmt"
	"math"
	"time"

	mongodb "github.com/anticoi/EscudoWAF/pkg/database/mongo"
	pkgModel "github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type StatsService interface {
	GetOverviewStats(ctx context.Context, timeRange string) (*dto.OverviewStats, error)
	GetRealtimeQPS(ctx context.Context, limit int) (*dto.RealtimeQPSResponse, error)
	GetTimeSeriesData(ctx context.Context, timeRange string, metric string) (*dto.TimeSeriesResponse, error)
	GetCombinedTimeSeriesData(ctx context.Context, timeRange string) (*dto.CombinedTimeSeriesResponse, error)
	GetTrafficTimeSeriesData(ctx context.Context, timeRange string) (*dto.TrafficTimeSeriesResponse, error)
}

type StatsServiceImpl struct {
	wafLogRepository repository.WAFLogRepository
	dbName           string
	logger           zerolog.Logger
}

func NewStatsService(wafLogRepository repository.WAFLogRepository) StatsService {
	dbName := config.Global.DBConfig.Database
	logger := config.GetServiceLogger("stats")
	return &StatsServiceImpl{
		wafLogRepository: wafLogRepository,
		dbName:           dbName,
		logger:           logger,
	}
}

// GetOverviewStats obtiene las estadísticas de resumen
func (s *StatsServiceImpl) GetOverviewStats(ctx context.Context, timeRange string) (*dto.OverviewStats, error) {
	// Determinar el rango de tiempo
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// Obtener la conexión a la base de datos MongoDB
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener la conexión a la base de datos: %w", err)
	}

	// 1. Obtener los datos estadísticos de HAProxyMinuteStats
	minuteStatsResult, err := s.getHAProxyStatsAggregate(ctx, db, startTime)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener los datos estadísticos: %w", err)
	}

	// 2. Obtener estadísticas de bloqueo del WAF
	blockCount, attackIPCount, err := s.getWAFBlockStats(ctx, startTime)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener las estadísticas de bloqueo del WAF: %w", err)
	}

	// 3. Calcular la tasa de errores
	error4xxRate := 0.0
	error5xxRate := 0.0

	if minuteStatsResult.TotalRequests > 0 {
		error4xxRate = float64(minuteStatsResult.Error4xx) / float64(minuteStatsResult.TotalRequests) * 100
		error5xxRate = float64(minuteStatsResult.Error5xx) / float64(minuteStatsResult.TotalRequests) * 100
	}

	// Construir el resultado
	result := &dto.OverviewStats{
		TimeRange:       timeRange,
		TotalRequests:   minuteStatsResult.TotalRequests,
		InboundTraffic:  minuteStatsResult.InboundTraffic,
		OutboundTraffic: minuteStatsResult.OutboundTraffic,
		MaxQPS:          minuteStatsResult.MaxQPS,
		Error4xx:        minuteStatsResult.Error4xx,
		Error4xxRate:    math.Round(error4xxRate*100) / 100, // Conservar dos decimales
		Error5xx:        minuteStatsResult.Error5xx,
		Error5xxRate:    math.Round(error5xxRate*100) / 100, // Conservar dos decimales
		BlockCount:      blockCount,
		AttackIPCount:   attackIPCount,
	}

	return result, nil
}

// GetRealtimeQPS Obtener datos QPS en tiempo real
func (s *StatsServiceImpl) GetRealtimeQPS(ctx context.Context, limit int) (*dto.RealtimeQPSResponse, error) {
	if limit <= 0 {
		limit = 30 // Obtener 30 puntos por defecto
	}
	if limit > 240 {
		limit = 240 // No más de 240 puntos
	}

	// Obtener la conexión a la base de datos MongoDB
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener la conexión a la base de datos: %w", err)
	}

	// Consultar datos de la base de datos de series temporales
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "metadata.target", Value: "all"},
	}}}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: -1}, // Ordenar por tiempo descendente
	}}}

	limitStage := bson.D{{Key: "$limit", Value: limit}}

	pipeline := mongo.Pipeline{matchStage, sortStage, limitStage}

	// Ejecutar la consulta de agregación
	reqRateColl := db.Collection("req_rate")
	cursor, err := reqRateColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("Error al consultar los datos QPS en tiempo real: %w", err)
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var results []struct {
		Timestamp time.Time `bson:"timestamp"`
		Value     int64     `bson:"value"`
		Metadata  bson.D    `bson:"metadata"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("Error al analizar los datos QPS en tiempo real: %w", err)
	}

	// Construir el resultado
	dataPoints := make([]dto.RealtimeQPSData, 0, len(results))
	// Invertir el orden para que el tiempo vaya de antiguo a reciente
	for i := len(results) - 1; i >= 0; i-- {
		dataPoints = append(dataPoints, dto.RealtimeQPSData{
			Timestamp: results[i].Timestamp,
			Value:     results[i].Value,
		})
	}

	return &dto.RealtimeQPSResponse{
		Data: dataPoints,
	}, nil
}

// GetTimeSeriesData Obtener datos de series temporales
func (s *StatsServiceImpl) GetTimeSeriesData(ctx context.Context, timeRange string, metric string) (*dto.TimeSeriesResponse, error) {
	// Determinar el rango de tiempo
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// Decidir la granularidad de agregación según el rango de tiempo
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// Agregación por hora en las últimas 24 horas
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// Agregación cada 6 horas en 7 días, para más puntos de datos
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// Agregación por día en los últimos 30 días
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("Rango de tiempo no válido: %s", timeRange)
	}

	// Obtener datos según el tipo de métrica
	var dataPoints []dto.TimeSeriesDataPoint

	switch metric {
	case "requests":
		// Serie temporal de peticiones
		dataPoints, err = s.getRequestTimeSeries(ctx, startTime, interval, groupByFields)
	case "blocks":
		// Serie temporal de bloqueos
		dataPoints, err = s.getBlockTimeSeries(ctx, startTime, interval, groupByFields)
	default:
		return nil, fmt.Errorf("Tipo de métrica no válido: %s", metric)
	}

	if err != nil {
		return nil, err
	}

	return &dto.TimeSeriesResponse{
		Metric:    metric,
		TimeRange: timeRange,
		Data:      dataPoints,
	}, nil
}

// GetCombinedTimeSeriesData obtiene series temporales de peticiones y bloqueos
func (s *StatsServiceImpl) GetCombinedTimeSeriesData(ctx context.Context, timeRange string) (*dto.CombinedTimeSeriesResponse, error) {
	// Determinar el rango de tiempo
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// Decidir la granularidad de agregación según el rango de tiempo
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// Agregación por hora en las últimas 24 horas
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// Agregación cada 6 horas en 7 días, para más puntos de datos
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// Agregación por día en los últimos 30 días
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("Rango de tiempo no válido: %s", timeRange)
	}

	// Obtener ambos tipos de datos en paralelo
	requestCh := make(chan struct {
		data []dto.TimeSeriesDataPoint
		err  error
	})
	blockCh := make(chan struct {
		data []dto.TimeSeriesDataPoint
		err  error
	})

	// Obtener los datos de la petición
	go func() {
		data, err := s.getRequestTimeSeries(ctx, startTime, interval, groupByFields)
		requestCh <- struct {
			data []dto.TimeSeriesDataPoint
			err  error
		}{data, err}
	}()

	// Obtener los datos de bloqueos
	go func() {
		data, err := s.getBlockTimeSeries(ctx, startTime, interval, groupByFields)
		blockCh <- struct {
			data []dto.TimeSeriesDataPoint
			err  error
		}{data, err}
	}()

	// Recibir resultados
	requestResult := <-requestCh
	blockResult := <-blockCh

	// Comprobar errores
	if requestResult.err != nil {
		return nil, fmt.Errorf("Error al obtener los datos de la petición: %w", requestResult.err)
	}
	if blockResult.err != nil {
		return nil, fmt.Errorf("Error al obtener los datos de bloqueo: %w", blockResult.err)
	}

	// Devolver el resultado combinado
	return &dto.CombinedTimeSeriesResponse{
		TimeRange: timeRange,
		Requests: dto.TimeSeriesResponse{
			Metric:    "requests",
			TimeRange: timeRange,
			Data:      requestResult.data,
		},
		Blocks: dto.TimeSeriesResponse{
			Metric:    "blocks",
			TimeRange: timeRange,
			Data:      blockResult.data,
		},
	}, nil
}

// GetTrafficTimeSeriesData Obtener datos de series temporales de tráfico
func (s *StatsServiceImpl) GetTrafficTimeSeriesData(ctx context.Context, timeRange string) (*dto.TrafficTimeSeriesResponse, error) {
	// Determinar el rango de tiempo
	startTime, err := s.getTimeRangeStart(timeRange)
	if err != nil {
		return nil, err
	}

	// Decidir la granularidad de agregación según el rango de tiempo
	var interval string
	var groupByFields []string

	switch timeRange {
	case dto.TimeRange24Hours:
		// Agregación por hora en las últimas 24 horas
		interval = "hour"
		groupByFields = []string{"date", "hour"}
	case dto.TimeRange7Days:
		// Agregación cada 6 horas en 7 días, para más puntos de datos
		interval = "6hour"
		groupByFields = []string{"date", "hourGroupSix"}
	case dto.TimeRange30Days:
		// Agregación por día en los últimos 30 días
		interval = "day"
		groupByFields = []string{"date"}
	default:
		return nil, fmt.Errorf("Rango de tiempo no válido: %s", timeRange)
	}

	// Obtener la conexión a la base de datos MongoDB
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener la conexión a la base de datos: %w", err)
	}

	var haproxyMinuteStats model.HAProxyMinuteStats
	collection := db.Collection(haproxyMinuteStats.GetCollectionName())

	// Construir las condiciones de filtro de tiempo
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// Construir las condiciones de agrupación
	var groupStage bson.D

	if interval == "6hour" {
		// Agrupar usando el campo precalculado HourGroupSix
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
			{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	} else {
		// Construir el ID de agrupación convencional
		groupID := bson.D{}
		for _, field := range groupByFields {
			groupID = append(groupID, bson.E{Key: field, Value: fmt.Sprintf("$%s", field)})
		}

		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: groupID},
			{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
			{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // Ordenar por tiempo ascendente
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage}
	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true)

	// Ejecutar la consulta de agregación
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, fmt.Errorf("Error al ejecutar la consulta de agregación de series temporales de tráfico: %w", err)
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var results []struct {
		ID              bson.D    `bson:"_id"`
		InboundTraffic  int64     `bson:"inboundTraffic"`
		OutboundTraffic int64     `bson:"outboundTraffic"`
		Timestamp       time.Time `bson:"timestamp"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("Error al analizar el resultado de series temporales de tráfico: %w", err)
	}

	// Construir los puntos de datos
	dataPoints := make([]dto.TrafficDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TrafficDataPoint{
			Timestamp:       result.Timestamp,
			InboundTraffic:  result.InboundTraffic,
			OutboundTraffic: result.OutboundTraffic,
		})
	}

	return &dto.TrafficTimeSeriesResponse{
		TimeRange: timeRange,
		Data:      dataPoints,
	}, nil
}

// Método auxiliar - obtener la hora de inicio del rango de tiempo
func (s *StatsServiceImpl) getTimeRangeStart(timeRange string) (time.Time, error) {
	now := time.Now()

	switch timeRange {
	case dto.TimeRange24Hours:
		return now.Add(-24 * time.Hour), nil
	case dto.TimeRange7Days:
		return now.AddDate(0, 0, -7), nil
	case dto.TimeRange30Days:
		return now.AddDate(0, 0, -30), nil
	default:
		return time.Time{}, fmt.Errorf("Rango de tiempo no válido: %s", timeRange)
	}
}

// Método auxiliar - resultado de agregación de estadísticas de haproxy
type haproxyStatsAggregateResult struct {
	TotalRequests   int64 // Total de peticiones
	InboundTraffic  int64 // Tráfico entrante
	OutboundTraffic int64 // Tráfico saliente
	MaxQPS          int64 // QPS máximo
	Error4xx        int64 // Cantidad de errores 4xx
	Error5xx        int64 // Cantidad de errores 5xx
}

// Método auxiliar - obtener resultado agregado de estadísticas de haproxy
func (s *StatsServiceImpl) getHAProxyStatsAggregate(ctx context.Context, db *mongo.Database, startTime time.Time) (*haproxyStatsAggregateResult, error) {
	// Crear el pipeline de agregación
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: nil},
		{Key: "totalRequests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
		{Key: "inboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bin"}}},
		{Key: "outboundTraffic", Value: bson.D{{Key: "$sum", Value: "$bout"}}},
		{Key: "maxQPS", Value: bson.D{{Key: "$max", Value: "$req_rate_max"}}},
		{Key: "error4xx", Value: bson.D{{Key: "$sum", Value: "$hrsp_4xx"}}},
		{Key: "error5xx", Value: bson.D{{Key: "$sum", Value: "$hrsp_5xx"}}},
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage}

	// Establecer opciones de agregación para optimizar el rendimiento
	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true).                                                        // Permitir el uso de disco para agregaciones de conjuntos de datos grandes
		SetHint(bson.D{{Key: "target_name", Value: 1}, {Key: "timestamp", Value: 1}}) // Usar el índice compuesto óptimo

	// Ejecutar la consulta de agregación
	var haproxyMinuteStats model.HAProxyMinuteStats
	collection := db.Collection(haproxyMinuteStats.GetCollectionName())

	queryStartTime := time.Now()
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	duration := time.Since(queryStartTime)

	if err != nil {
		s.logger.Error().
			Err(err).
			Dur("duration", duration).
			Time("queryStartTime", startTime).
			Msg("Falló la consulta de agregación de estadísticas de HAProxy")
		return nil, fmt.Errorf("Error al ejecutar la consulta de agregación de estadísticas: %w", err)
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var results []struct {
		TotalRequests   int64 `bson:"totalRequests"`
		InboundTraffic  int64 `bson:"inboundTraffic"`
		OutboundTraffic int64 `bson:"outboundTraffic"`
		MaxQPS          int64 `bson:"maxQPS"`
		Error4xx        int64 `bson:"error4xx"`
		Error5xx        int64 `bson:"error5xx"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		s.logger.Error().
			Err(err).
			Dur("duration", duration).
			Msg("Error al analizar el resultado de agregación de estadísticas de HAProxy")
		return nil, fmt.Errorf("Error al analizar el resultado de agregación estadística: %w", err)
	}

	// Construir el resultado
	result := &haproxyStatsAggregateResult{}
	if len(results) > 0 {
		result.TotalRequests = results[0].TotalRequests
		result.InboundTraffic = results[0].InboundTraffic
		result.OutboundTraffic = results[0].OutboundTraffic
		result.MaxQPS = results[0].MaxQPS
		result.Error4xx = results[0].Error4xx
		result.Error5xx = results[0].Error5xx
	}

	// Registrar logs de monitorización de rendimiento
	s.logger.Debug().
		Dur("duration", duration).
		Int64("totalRequests", result.TotalRequests).
		Int64("inboundTraffic", result.InboundTraffic).
		Int64("outboundTraffic", result.OutboundTraffic).
		Int64("maxQPS", result.MaxQPS).
		Time("queryStartTime", startTime).
		Msg("Consulta de agregación de estadísticas de HAProxy completada")

	// Advertencia de rendimiento
	if duration > 1*time.Second {
		s.logger.Warn().
			Dur("duration", duration).
			Time("queryStartTime", startTime).
			Msg("La consulta de agregación de estadísticas de HAProxy tardó demasiado")
	}

	return result, nil
}

// Método auxiliar - obtener estadísticas de bloqueos del WAF
func (s *StatsServiceImpl) getWAFBlockStats(ctx context.Context, startTime time.Time) (int64, int64, error) {
	// Construir las condiciones de filtro de tiempo
	timeFilter := bson.D{
		{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}

	// Obtener el total de bloqueos
	blockCount, err := s.wafLogRepository.CountAttackLogs(ctx, timeFilter)
	if err != nil {
		return 0, 0, fmt.Errorf("Error al calcular el total de bloqueos: %w", err)
	}

	// Obtener la conexión a MongoDB y ejecutar directamente la agregación para el número de IPs atacantes distintas
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return 0, 0, fmt.Errorf("Error al obtener la conexión a la base de datos: %w", err)
	}

	var wafLog pkgModel.WAFLog
	collection := db.Collection(wafLog.GetCollectionName())

	// Pipeline de agregación optimizado: calcula directamente IPs distintas
	ipCountPipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{
				{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
			}},
		},
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$srcIp"},
			}},
		},
		{
			{Key: "$count", Value: "uniqueIPs"},
		},
	}

	// Establecer opciones de agregación para optimizar consultas en grandes volúmenes de datos
	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true).                                                  // Permitir el uso de disco para agregaciones de conjuntos de datos grandes
		SetHint(bson.D{{Key: "srcIp", Value: 1}, {Key: "createdAt", Value: 1}}) // Usar hint de índice compuesto

	// Ejecutar la consulta de agregación
	cursor, err := collection.Aggregate(ctx, ipCountPipeline, aggregateOptions)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al calcular la cantidad de IPs atacantes")
		return blockCount, 0, nil
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var result struct {
		UniqueIPs int64 `bson:"uniqueIPs"`
	}

	attackIPCount := int64(0)
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			s.logger.Error().Err(err).Msg("Error al analizar el resultado de cantidad de IPs atacantes")
			return blockCount, 0, nil
		}
		attackIPCount = result.UniqueIPs
	}

	// Registrar estadísticas para monitorización de rendimiento
	s.logger.Debug().
		Int64("blockCount", blockCount).
		Int64("attackIPCount", attackIPCount).
		Time("startTime", startTime).
		Msg("Consulta de estadísticas de bloqueo del WAF completada")

	return blockCount, attackIPCount, nil
}

// Método auxiliar - obtener la serie temporal de peticiones
func (s *StatsServiceImpl) getRequestTimeSeries(ctx context.Context, startTime time.Time, interval string, groupByFields []string) ([]dto.TimeSeriesDataPoint, error) {
	// Obtener la conexión a la base de datos MongoDB
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener la conexión a la base de datos: %w", err)
	}

	var haproxyMinuteStats model.HAProxyMinuteStats
	collection := db.Collection(haproxyMinuteStats.GetCollectionName())

	// Construir las condiciones de filtro de tiempo
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "target_name", Value: "all"},
		{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// Construir las condiciones de agrupación
	var groupStage bson.D

	if interval == "6hour" {
		// Agrupar usando el campo precalculado HourGroupSix
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "requests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	} else {
		// Construir el ID de agrupación convencional
		groupID := bson.D{}
		for _, field := range groupByFields {
			groupID = append(groupID, bson.E{Key: field, Value: fmt.Sprintf("$%s", field)})
		}

		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: groupID},
			{Key: "requests", Value: bson.D{{Key: "$sum", Value: "$req_tot"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$timestamp"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // Ordenar por tiempo ascendente
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage}

	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true)

	// Ejecutar la consulta de agregación
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, fmt.Errorf("Error al ejecutar la consulta de agregación de series temporales de peticiones: %w", err)
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var results []struct {
		ID        bson.D    `bson:"_id"`
		Requests  int64     `bson:"requests"`
		Timestamp time.Time `bson:"timestamp"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("Error al analizar el resultado de series temporales de peticiones: %w", err)
	}

	// Construir los puntos de datos
	dataPoints := make([]dto.TimeSeriesDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TimeSeriesDataPoint{
			Timestamp: result.Timestamp,
			Value:     result.Requests,
		})
	}

	return dataPoints, nil
}

// Método auxiliar - obtener la serie temporal de bloqueos
func (s *StatsServiceImpl) getBlockTimeSeries(ctx context.Context, startTime time.Time, interval string, groupByFields []string) ([]dto.TimeSeriesDataPoint, error) {
	// Obtener la conexión a la base de datos MongoDB
	db, err := mongodb.GetDatabase(s.dbName)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener la conexión a la base de datos: %w", err)
	}

	collection := db.Collection("waf_log")

	// Construir las condiciones de filtro de tiempo
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: startTime}}},
	}}}

	// Decidir cómo agrupar según el intervalo
	var groupStage bson.D

	if interval == "hour" {
		// Agrupar por hora
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hour", Value: "$hour"},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	} else if interval == "6hour" {
		// Agrupar con el campo precalculado HourGroupSix
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: "$date"},
				{Key: "hourGroup", Value: "$hourGroupSix"},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	} else {
		// Agrupar por fecha
		groupStage = bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "date", Value: "$date"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "timestamp", Value: bson.D{{Key: "$min", Value: "$createdAt"}}},
		}}}
	}

	sortStage := bson.D{{Key: "$sort", Value: bson.D{
		{Key: "timestamp", Value: 1}, // Ordenar por tiempo ascendente
	}}}

	// Limitar los elementos devueltos - optimiza el rendimiento
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "count", Value: 1},
		{Key: "timestamp", Value: 1},
	}}}

	pipeline := mongo.Pipeline{matchStage, groupStage, sortStage, projectStage}

	aggregateOptions := options.Aggregate().
		SetAllowDiskUse(true)

	// Ejecutar la consulta de agregación
	cursor, err := collection.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, fmt.Errorf("Error al ejecutar la consulta de agregación de series temporales de bloqueos del WAF: %w", err)
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var results []struct {
		Count     int64     `bson:"count"`
		Timestamp time.Time `bson:"timestamp"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("Error al analizar el resultado de series temporales de bloqueos del WAF: %w", err)
	}

	// Construir los puntos de datos
	dataPoints := make([]dto.TimeSeriesDataPoint, 0, len(results))
	for _, result := range results {
		dataPoints = append(dataPoints, dto.TimeSeriesDataPoint{
			Timestamp: result.Timestamp,
			Value:     result.Count,
		})
	}

	return dataPoints, nil
}
