// server/repository/ip_group.go
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
	ErrIPGroupNotFound = errors.New("El grupo de IPs no existe")
)

// IPGroupRepository interfaz del repositorio de grupos de IPs
type IPGroupRepository interface {
	CreateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error
	GetIPGroups(ctx context.Context, page, size int64) ([]model.IPGroup, int64, error)
	GetIPGroupByID(ctx context.Context, id bson.ObjectID) (*model.IPGroup, error)
	GetIPGroupByName(ctx context.Context, name string) (*model.IPGroup, error)
	UpdateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error
	DeleteIPGroup(ctx context.Context, id bson.ObjectID) error
	CheckIPGroupNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error)
}

// MongoIPGroupRepository repositorio de grupos de IPs implementado en MongoDB
type MongoIPGroupRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewIPGroupRepository crea el repositorio de grupos de IPs
func NewIPGroupRepository(db *mongo.Database) IPGroupRepository {
	var ipGroup model.IPGroup
	collection := db.Collection(ipGroup.GetCollectionName())
	logger := config.GetRepositoryLogger("ipgroup")

	// Crear índices
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Índice único de nombres de grupos de IPs
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear el índice de nombres de grupos de IPs")
	}

	return &MongoIPGroupRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateIPGroup Crear grupo de IPs
func (r *MongoIPGroupRepository) CreateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error {
	// Insertar nuevo grupo de IPs
	result, err := r.collection.InsertOne(ctx, ipGroup)
	if err != nil {
		r.logger.Error().Err(err).Str("name", ipGroup.Name).Msg("Error al insertar el grupo de IPs")
		return err
	}

	ipGroup.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetIPGroups Obtener lista de grupos de IPs
func (r *MongoIPGroupRepository) GetIPGroups(ctx context.Context, page, size int64) ([]model.IPGroup, int64, error) {
	// Calcular la paginación
	skip := (page - 1) * size

	// Establecer las opciones de consulta
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "name", Value: 1}}) // Ordenar por nombre ascendente

	// Ejecutar la consulta
	cursor, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al consultar la lista de grupos de IPs")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var ipGroups []model.IPGroup
	if err = cursor.All(ctx, &ipGroups); err != nil {
		r.logger.Error().Err(err).Msg("Error al analizar la lista de grupos de IPs")
		return nil, 0, err
	}

	// Obtener el total
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener el total de grupos de IPs")
		return nil, 0, err
	}

	return ipGroups, total, nil
}

// GetIPGroupByID obtiene un grupo de IPs por ID
func (r *MongoIPGroupRepository) GetIPGroupByID(ctx context.Context, id bson.ObjectID) (*model.IPGroup, error) {
	var ipGroup model.IPGroup
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&ipGroup)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrIPGroupNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al consultar el grupo de IPs")
		return nil, err
	}

	return &ipGroup, nil
}

// GetIPGroupByName obtiene un grupo de IPs por nombre
func (r *MongoIPGroupRepository) GetIPGroupByName(ctx context.Context, name string) (*model.IPGroup, error) {
	var ipGroup model.IPGroup
	err := r.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&ipGroup)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrIPGroupNotFound
		}
		r.logger.Error().Err(err).Str("name", name).Msg("Error al consultar el grupo de IPs por nombre")
		return nil, err
	}

	return &ipGroup, nil
}

// UpdateIPGroup Actualizar grupo de IPs
func (r *MongoIPGroupRepository) UpdateIPGroup(ctx context.Context, ipGroup *model.IPGroup) error {
	// Actualizar grupo de IPs
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: ipGroup.ID}},
		ipGroup,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", ipGroup.ID.Hex()).Msg("Error al actualizar el grupo de IPs")
		return err
	}

	return nil
}

// DeleteIPGroup Eliminar grupo de IPs
func (r *MongoIPGroupRepository) DeleteIPGroup(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar el grupo de IPs")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrIPGroupNotFound
	}

	return nil
}

// CheckIPGroupNameExists comprueba si el nombre del grupo de IPs ya existe
func (r *MongoIPGroupRepository) CheckIPGroupNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error) {
	filter := bson.D{{Key: "name", Value: name}}

	// Si es una actualización, excluir el ID del grupo de IPs actual
	if excludeID != bson.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$ne", Value: excludeID}}})
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("name", name).Msg("Error al comprobar si el nombre del grupo de IPs existe")
		return false, err
	}

	return count > 0, nil
}
