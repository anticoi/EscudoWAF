// server/repository/rule.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrRuleNotFound = errors.New("La regla no existe")
)

// MicroRuleRepository interfaz del repositorio de microrreglas
type MicroRuleRepository interface {
	CreateMicroRule(ctx context.Context, rule *model.MicroRule) error
	GetMicroRules(ctx context.Context, page, size int64) ([]model.MicroRule, int64, error)
	GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error)
	GetMicroRuleByName(ctx context.Context, name string) (*model.MicroRule, error)
	UpdateMicroRule(ctx context.Context, rule *model.MicroRule) error
	DeleteMicroRule(ctx context.Context, id bson.ObjectID) error
	CheckMicroRuleNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error)
}

// MongoMicroRuleRepository repositorio de microrreglas implementado en MongoDB
type MongoMicroRuleRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewMicroRuleRepository crea el repositorio de microrreglas
func NewMicroRuleRepository(db *mongo.Database) MicroRuleRepository {
	var rule model.MicroRule
	collection := db.Collection(rule.GetCollectionName())
	logger := config.GetRepositoryLogger("microrule")

	// Crear índices
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Índice único de nombre de regla
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear el índice de nombres de reglas")
	}

	return &MongoMicroRuleRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateMicroRule Crear microrregla
func (r *MongoMicroRuleRepository) CreateMicroRule(ctx context.Context, rule *model.MicroRule) error {
	// Insertar nueva microrregla
	result, err := r.collection.InsertOne(ctx, rule)
	if err != nil {
		r.logger.Error().Err(err).Str("name", rule.Name).Msg("Error al insertar la microrregla")
		return err
	}

	rule.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetMicroRules Obtener lista de microrreglas
func (r *MongoMicroRuleRepository) GetMicroRules(ctx context.Context, page, size int64) ([]model.MicroRule, int64, error) {
	// Calcular la paginación
	skip := (page - 1) * size

	// Establecer opciones de consulta, ordenar por prioridad descendente
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "priority", Value: -1}})

	// Ejecutar la consulta
	cursor, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al consultar la lista de microrreglas")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var rules []model.MicroRule
	if err = cursor.All(ctx, &rules); err != nil {
		r.logger.Error().Err(err).Msg("Error al analizar la lista de microrreglas")
		return nil, 0, err
	}

	// Obtener el total
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener el total de microrreglas")
		return nil, 0, err
	}

	return rules, total, nil
}

// GetMicroRuleByID obtiene una microrregla por ID
func (r *MongoMicroRuleRepository) GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error) {
	var rule model.MicroRule
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRuleNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al consultar la microrregla")
		return nil, err
	}

	return &rule, nil
}

// GetMicroRuleByName obtiene una microrregla por nombre
func (r *MongoMicroRuleRepository) GetMicroRuleByName(ctx context.Context, name string) (*model.MicroRule, error) {
	var rule model.MicroRule
	err := r.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRuleNotFound
		}
		r.logger.Error().Err(err).Str("name", name).Msg("Error al consultar la microrregla por nombre")
		return nil, err
	}

	return &rule, nil
}

// UpdateMicroRule Actualizar microrregla
func (r *MongoMicroRuleRepository) UpdateMicroRule(ctx context.Context, rule *model.MicroRule) error {
	// Actualizar microrregla
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: rule.ID}},
		rule,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", rule.ID.Hex()).Msg("Error al actualizar la microrregla")
		return err
	}

	return nil
}

// DeleteMicroRule Eliminar microrregla
func (r *MongoMicroRuleRepository) DeleteMicroRule(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar la microrregla")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrRuleNotFound
	}

	return nil
}

// CheckMicroRuleNameExists comprueba si el nombre de la microrregla ya existe
func (r *MongoMicroRuleRepository) CheckMicroRuleNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error) {
	filter := bson.D{{Key: "name", Value: name}}

	// Si es una actualización, excluir el ID de la regla actual
	if excludeID != bson.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$ne", Value: excludeID}}})
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("name", name).Msg("Error al comprobar si el nombre de la microrregla existe")
		return false, err
	}

	return count > 0, nil
}
