package repository

import (
	"context"
	"errors"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrBlockedIPNotFound = errors.New("El registro de IP bloqueada no existe")
)

// BlockedIPRepository interfaz del repositorio de IPs bloqueadas
type BlockedIPRepository interface {
	GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) ([]model.BlockedIPRecord, int64, error)
	GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error)
	CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error
	DeleteExpiredBlockedIPs(ctx context.Context) (int64, error)
}

// MongoBlockedIPRepository repositorio de IPs bloqueadas implementado en MongoDB
type MongoBlockedIPRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewBlockedIPRepository crea el repositorio de IPs bloqueadas
func NewBlockedIPRepository(db *mongo.Database) BlockedIPRepository {
	var blockedIP model.BlockedIPRecord
	collection := db.Collection(blockedIP.GetCollectionName())
	logger := config.GetRepositoryLogger("blocked_ip")

	// Crear índices
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Índice de direcciones IP
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "ip", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear el índice de direcciones IP")
	}

	// Índice de tiempo de bloqueo (para limpieza de expirados)
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "blocked_until", Value: 1}},
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear el índice de tiempo de bloqueo")
	}

	// Índice compuesto: motivo y hora de bloqueo
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "reason", Value: 1},
			{Key: "blocked_at", Value: -1},
		},
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear el índice compuesto")
	}

	return &MongoBlockedIPRepository{
		collection: collection,
		logger:     logger,
	}
}

// GetBlockedIPs Obtener lista de IPs bloqueadas
func (r *MongoBlockedIPRepository) GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) ([]model.BlockedIPRecord, int64, error) {
	// Construir el filtro de consulta
	filter := r.buildFilter(req)

	// Calcular la paginación
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	} else if size > 100 {
		size = 100
	}
	skip := int64((page - 1) * size)

	// Construir la ordenación
	sort := r.buildSort(req)

	// Establecer las opciones de consulta
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(size)).
		SetSort(sort)

	// Ejecutar la consulta
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al consultar la lista de IPs bloqueadas")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var records []model.BlockedIPRecord
	if err = cursor.All(ctx, &records); err != nil {
		r.logger.Error().Err(err).Msg("Error al analizar la lista de IPs bloqueadas")
		return nil, 0, err
	}

	// Obtener el total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener el total de IPs bloqueadas")
		return nil, 0, err
	}

	return records, total, nil
}

// GetBlockedIPStats Obtener estadísticas de IPs bloqueadas
func (r *MongoBlockedIPRepository) GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error) {
	now := time.Now()

	stats := &dto.BlockedIPStatsResponse{
		ReasonStats: make(map[string]int64),
	}

	// Obtener la cantidad total de bloqueos
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener el total de bloqueos")
		return nil, err
	}
	stats.TotalBlocked = total

	// Obtener la cantidad de bloqueos actualmente activos
	activeFilter := bson.D{{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}}}
	active, err := r.collection.CountDocuments(ctx, activeFilter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener la cantidad de bloqueos activos")
		return nil, err
	}
	stats.ActiveBlocked = active
	stats.ExpiredBlocked = total - active

	// Estadísticas por motivo
	reasonPipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$reason"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, reasonPipeline)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error en las estadísticas por motivo")
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			r.logger.Error().Err(err).Msg("Error al analizar el resultado de estadísticas por motivo")
			continue
		}
		stats.ReasonStats[result.ID] = result.Count
	}

	// Estadísticas por hora de las últimas 24 horas
	last24Hours := now.Add(-24 * time.Hour)
	hourlyStats, err := r.getHourlyStats(ctx, last24Hours, now)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener las estadísticas por hora")
		// No devolver error, solo registrar en el log
	} else {
		stats.Last24HourStats = hourlyStats
	}

	return stats, nil
}

// CreateBlockedIP crea un registro de IP bloqueada
func (r *MongoBlockedIPRepository) CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error {
	_, err := r.collection.InsertOne(ctx, record)
	if err != nil {
		r.logger.Error().Err(err).Str("ip", record.IP).Msg("Error al insertar el registro de IP bloqueada")
		return err
	}
	return nil
}

// DeleteExpiredBlockedIPs elimina registros de IPs bloqueadas expirados
func (r *MongoBlockedIPRepository) DeleteExpiredBlockedIPs(ctx context.Context) (int64, error) {
	now := time.Now()
	filter := bson.D{{Key: "blocked_until", Value: bson.D{{Key: "$lt", Value: now}}}}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al eliminar registros de IPs bloqueadas expirados")
		return 0, err
	}

	r.logger.Info().Int64("count", result.DeletedCount).Msg("Registros de IPs bloqueadas expirados eliminados")
	return result.DeletedCount, nil
}

// buildFilter construye el filtro de consulta
func (r *MongoBlockedIPRepository) buildFilter(req *dto.BlockedIPListRequest) bson.D {
	filter := bson.D{}

	// Filtrado por dirección IP
	if req.IP != "" {
		filter = append(filter, bson.E{Key: "ip", Value: bson.D{{Key: "$regex", Value: req.IP}, {Key: "$options", Value: "i"}}})
	}

	// Filtro por motivo de bloqueo
	if req.Reason != "" {
		filter = append(filter, bson.E{Key: "reason", Value: req.Reason})
	}

	// Filtro de estado
	now := time.Now()
	switch req.Status {
	case "active":
		filter = append(filter, bson.E{Key: "blocked_until", Value: bson.D{{Key: "$gt", Value: now}}})
	case "expired":
		filter = append(filter, bson.E{Key: "blocked_until", Value: bson.D{{Key: "$lte", Value: now}}})
	case "all", "":
		// No añadir filtro de tiempo
	}

	return filter
}

// buildSort construye la ordenación
func (r *MongoBlockedIPRepository) buildSort(req *dto.BlockedIPListRequest) bson.D {
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "blocked_at"
	}

	sortDir := 1 // Ascendente
	if req.SortDir == "desc" {
		sortDir = -1 // Orden descendente
	}

	return bson.D{{Key: sortBy, Value: sortDir}}
}

// getHourlyStats obtiene las estadísticas por hora
func (r *MongoBlockedIPRepository) getHourlyStats(ctx context.Context, start, end time.Time) ([]dto.BlockedIPHourlyStats, error) {
	pipeline := mongo.Pipeline{
		// Filtrar por rango de tiempo
		{{Key: "$match", Value: bson.D{
			{Key: "blocked_at", Value: bson.D{
				{Key: "$gte", Value: start},
				{Key: "$lt", Value: end},
			}},
		}}},
		// Agrupar por hora
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "$dateToString", Value: bson.D{
					{Key: "format", Value: "%Y-%m-%dT%H:00:00Z"},
					{Key: "date", Value: "$blocked_at"},
				}},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		// Ordenar
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []dto.BlockedIPHourlyStats
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			r.logger.Error().Err(err).Msg("Error al analizar el resultado de estadísticas por hora")
			continue
		}

		stats = append(stats, dto.BlockedIPHourlyStats{
			Hour:  result.ID,
			Count: result.Count,
		})
	}

	return stats, nil
}
