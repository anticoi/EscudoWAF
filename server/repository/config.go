// server/repository/config.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/constant"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrConfigNotFound = errors.New("La configuración no existe")
)

// ConfigRepository interfaz del repositorio de configuración
type ConfigRepository interface {
	GetConfig(ctx context.Context) (*model.Config, error)
	UpdateConfig(ctx context.Context, config *model.Config) error
}

// MongoConfigRepository repositorio de configuración implementado en MongoDB
type MongoConfigRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewConfigRepository crea el repositorio de configuración
func NewConfigRepository(db *mongo.Database) ConfigRepository {
	var cfg model.Config
	collection := db.Collection(cfg.GetCollectionName())
	logger := config.GetRepositoryLogger("config")

	return &MongoConfigRepository{
		collection: collection,
		logger:     logger,
	}
}

// GetConfig obtiene la configuración
func (r *MongoConfigRepository) GetConfig(ctx context.Context) (*model.Config, error) {
	var cfg model.Config

	// Usar constante para el nombre de configuración; si no existe usar "AppConfig" por defecto
	configName := constant.GetString("APP_CONFIG_NAME", "AppConfig")

	// Consultar la configuración por nombre
	err := r.collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: configName}},
	).Decode(&cfg)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConfigNotFound
		}
		r.logger.Error().Err(err).Msg("Error al consultar la configuración")
		return nil, err
	}

	return &cfg, nil
}

// UpdateConfig actualiza la configuración
func (r *MongoConfigRepository) UpdateConfig(ctx context.Context, config *model.Config) error {
	// Hora de actualización
	config.UpdatedAt = time.Now()

	// Buscar por nombre y actualizar la configuración
	filter := bson.D{{Key: "name", Value: config.Name}}
	update := bson.D{{Key: "$set", Value: config}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al actualizar la configuración")
		return err
	}

	if result.MatchedCount == 0 {
		return ErrConfigNotFound
	}

	return nil
}
