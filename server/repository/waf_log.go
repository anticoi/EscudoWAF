package repository

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type WAFLogRepository interface {
	AggregateAttackEvents(ctx context.Context, pipeline mongo.Pipeline) ([]dto.AttackEventAggregateResult, error)
	CountAggregateAttackEvents(ctx context.Context, pipeline mongo.Pipeline) (int64, error)
	FindAttackLogs(ctx context.Context, filter bson.D, skip int64, limit int64) ([]model.WAFLog, error)
	CountAttackLogs(ctx context.Context, filter bson.D) (int64, error)
}

type MongoWAFLogRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewWAFLogRepository creates a new WAFLogRepository instance
func NewWAFLogRepository(db *mongo.Database) WAFLogRepository {
	var wafLog model.WAFLog
	collection := db.Collection(wafLog.GetCollectionName())
	logger := config.GetRepositoryLogger("waf_log")

	return &MongoWAFLogRepository{
		collection: collection,
		logger:     logger,
	}
}

// AggregateAttackEvents executes the aggregation pipeline for attack events
func (r *MongoWAFLogRepository) AggregateAttackEvents(
	ctx context.Context,
	pipeline mongo.Pipeline,
) ([]dto.AttackEventAggregateResult, error) {
	// Execute data aggregation
	dataCursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error executing data aggregation: %w", err)
	}
	defer dataCursor.Close(ctx)

	// Process results
	currentTime := time.Now()
	var results []dto.AttackEventAggregateResult

	for dataCursor.Next(ctx) {
		var result struct {
			SrcIP           string        `bson:"srcIp"`
			SrcIPInfo       *model.IPInfo `bson:"srcIpInfo"`
			DstPort         int           `bson:"dstPort"`
			Domain          string        `bson:"domain"`
			Count           int           `bson:"count"`
			FirstAttackTime time.Time     `bson:"firstAttackTime"`
			LastAttackTime  time.Time     `bson:"lastAttackTime"`
			AllTimes        []time.Time   `bson:"allTimes"`
		}

		if err := dataCursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("error decoding data result: %w", err)
		}

		// Create result object

		aggregateResult := dto.AttackEventAggregateResult{
			SrcIP:           result.SrcIP,
			SrcIPInfo:       result.SrcIPInfo,
			DstPort:         result.DstPort,
			Domain:          result.Domain,
			Count:           result.Count,
			FirstAttackTime: result.FirstAttackTime,
			LastAttackTime:  result.LastAttackTime,
		}

		// Check if attack is ongoing (within last 3 minutes)
		timeSinceLastAttack := currentTime.Sub(result.LastAttackTime).Minutes()
		if timeSinceLastAttack < 3 {
			aggregateResult.IsOngoing = true
			// Calculate duration of continuous attack
			aggregateResult.DurationInMinutes = r.calculateAttackDuration(result.AllTimes)
		} else {
			aggregateResult.IsOngoing = false
		}

		results = append(results, aggregateResult)
	}

	if err := dataCursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return results, nil
}

// CountAggregateAttackEvents counts the total number of attack events after aggregation
func (r *MongoWAFLogRepository) CountAggregateAttackEvents(
	ctx context.Context,
	pipeline mongo.Pipeline,
) (int64, error) {
	// Add count stage to pipeline
	countStage := bson.D{{Key: "$count", Value: "total"}}
	countPipeline := append(pipeline, countStage)

	// Execute count aggregation
	countCursor, err := r.collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return 0, fmt.Errorf("error executing count aggregation: %w", err)
	}
	defer countCursor.Close(ctx)

	// Get total count
	var totalCount int64 = 0
	var countResult struct {
		Total int64 `bson:"total"`
	}

	if countCursor.Next(ctx) {
		if err := countCursor.Decode(&countResult); err != nil {
			return 0, fmt.Errorf("error decoding count result: %w", err)
		}
		totalCount = countResult.Total
	}

	return totalCount, nil
}

// FindAttackLogs finds attack logs with the given filter and options
func (r *MongoWAFLogRepository) FindAttackLogs(
	ctx context.Context,
	filter bson.D,
	skip int64,
	limit int64,
) ([]model.WAFLog, error) {
	// Usar options.Find() para crear las opciones
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Los más recientes primero

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error executing find query: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.WAFLog
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("error decoding query results: %w", err)
	}

	return results, nil
}

// CountAttackLogs counts the total number of attack logs matching the filter
func (r *MongoWAFLogRepository) CountAttackLogs(ctx context.Context, filter bson.D) (int64, error) {
	// Establecer opciones de conteo para optimizar el rendimiento en grandes volúmenes
	countOptions := options.Count().
		SetHint(bson.D{{Key: "createdAt", Value: 1}}) // Usar hint de índice de tiempo

	total, err := r.collection.CountDocuments(ctx, filter, countOptions)
	if err != nil {
		return 0, fmt.Errorf("error counting documents: %w", err)
	}
	return total, nil
}

// calculateAttackDuration calculates the duration of a continuous attack
// by finding the longest sequence of attacks with gaps no larger than 5 minutes
func (r *MongoWAFLogRepository) calculateAttackDuration(attackTimes []time.Time) float64 {
	if len(attackTimes) == 0 {
		return 1.0 // Aunque no haya hora de ataque, devolver al menos 1 minuto
	}

	// Manejar el caso de un solo punto de ataque, devolver 1 minuto directamente
	if len(attackTimes) == 1 {
		return 1.0
	}

	// Sort attack times in descending order (newest first)
	sortedTimes := make([]time.Time, len(attackTimes))
	copy(sortedTimes, attackTimes)

	// Sort in reverse chronological order
	sort.Slice(sortedTimes, func(i, j int) bool {
		return sortedTimes[i].After(sortedTimes[j])
	})

	// Comenzar desde la hora del ataque más reciente
	var currentSequenceStart time.Time = sortedTimes[0]
	var previousTime time.Time = sortedTimes[0]
	var foundContinuousSequence bool = false

	// Buscar la hora de ataque más antigua que sigue en la secuencia continua
	for i := 1; i < len(sortedTimes); i++ {
		currentTime := sortedTimes[i]

		// Comprobar si la diferencia con el ataque anterior es menor de 5 minutos
		if previousTime.Sub(currentTime).Minutes() <= 5 {
			// Todavía en la secuencia continua
			previousTime = currentTime
			foundContinuousSequence = true
		} else {
			// Interrupción de la secuencia continua
			break
		}
	}

	// Si no se encontró secuencia continua (el ataque más reciente está a más de 5 minutos de los demás)
	if !foundContinuousSequence {
		return 1.0 // Devolver duración mínima de 1 minuto
	}

	// Calcular la duración (hora más reciente menos la más antigua de la secuencia)
	duration := currentSequenceStart.Sub(previousTime).Minutes()

	// Redondear la duración hacia arriba a al menos 1 minuto
	if duration < 1.0 {
		return 1.0
	}

	// Devolver minutos enteros (redondeo hacia arriba)
	return math.Ceil(duration)
}
