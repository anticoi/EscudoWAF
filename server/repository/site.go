package repository

import (
	"context"
	"errors"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"

	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrSiteNotFound       = errors.New("El sitio no existe")
	ErrDomainPortExists   = errors.New("La combinación de dominio y puerto ya existe")
	ErrDomainPortConflict = errors.New("La combinación de dominio y puerto ya está en uso por otro sitio")
)

// SiteRepository repositorio de sitios
type SiteRepository interface {
	CreateSite(ctx context.Context, site *model.Site) error
	GetSites(ctx context.Context, page, size int64) ([]model.Site, int64, error)
	GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error)
	UpdateSite(ctx context.Context, site *model.Site) error
	DeleteSite(ctx context.Context, id bson.ObjectID) error
	CheckDomainPortExists(ctx context.Context, site *model.Site) error
	CheckDomainPortConflict(ctx context.Context, site *model.Site) error
}

// SiteRepository repositorio de sitios
type MongoSiteRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewSiteRepository crea el repositorio de sitios
func NewSiteRepository(db *mongo.Database) SiteRepository {
	var site model.Site
	collection := db.Collection(site.GetCollectionName())
	logger := config.GetRepositoryLogger("site")

	return &MongoSiteRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateSite crea un sitio
func (r *MongoSiteRepository) CreateSite(ctx context.Context, site *model.Site) error {
	// Establecer las fechas de creación y actualización
	now := time.Now()
	site.CreatedAt = now
	site.UpdatedAt = now

	// Insertar nuevo sitio
	result, err := r.collection.InsertOne(ctx, site)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al insertar el sitio")
		return err
	}

	// Establecer el ID
	if id, ok := result.InsertedID.(bson.ObjectID); ok {
		site.ID = id
	}

	return nil
}

// GetSites Obtener lista de sitios
func (r *MongoSiteRepository) GetSites(ctx context.Context, page, size int64) ([]model.Site, int64, error) {
	// Calcular la paginación
	skip := (page - 1) * size

	// Establecer las opciones de consulta
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(size).
		SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Ordenar por fecha de creación descendente

	// Ejecutar la consulta
	cursor, err := r.collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al consultar la lista de sitios")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var sites []model.Site
	if err = cursor.All(ctx, &sites); err != nil {
		r.logger.Error().Err(err).Msg("Error al analizar la lista de sitios")
		return nil, 0, err
	}

	// Obtener el total
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener el total de sitios")
		return nil, 0, err
	}

	return sites, total, nil
}

// GetSiteByID obtiene un sitio por ID
func (r *MongoSiteRepository) GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error) {
	var site model.Site
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&site)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrSiteNotFound // Devolver error de no encontrado
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al consultar el sitio")
		return nil, err
	}

	return &site, nil
}

// UpdateSite Actualizar sitio
func (r *MongoSiteRepository) UpdateSite(ctx context.Context, site *model.Site) error {
	// Actualizar sitio
	site.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: site.ID}},
		site,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", site.ID.Hex()).Msg("Error al actualizar el sitio")
		return err
	}

	return nil
}

// DeleteSite Eliminar sitio
func (r *MongoSiteRepository) DeleteSite(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar el sitio")
		return err
	}

	return nil
}

func (r *MongoSiteRepository) CheckDomainPortExists(ctx context.Context, site *model.Site) error {
	// Comprobar si la combinación de dominio y puerto ya existe
	filter := bson.D{
		{Key: "domain", Value: site.Domain},
		{Key: "listenPort", Value: site.ListenPort},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al comprobar si el dominio y puerto del sitio existen")
		return err
	}
	if count > 0 {
		r.logger.Error().Msg("La combinación de dominio y puerto del sitio ya existe")
		return ErrDomainPortExists
	}
	return nil
}

func (r *MongoSiteRepository) CheckDomainPortConflict(ctx context.Context, site *model.Site) error {
	// Comprobar si la combinación de dominio y puerto entra en conflicto con otros sitios
	filter := bson.D{
		{Key: "_id", Value: bson.D{{Key: "$ne", Value: site.ID}}},
		{Key: "domain", Value: site.Domain},
		{Key: "listenPort", Value: site.ListenPort},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al comprobar conflictos de dominio y puerto del sitio")
		return err
	}
	if count > 0 {
		r.logger.Error().Msg("Error al actualizar el sitio, la combinación de dominio y puerto ya existe")
		return ErrDomainPortConflict
	}
	return nil
}

// GetAllSites obtiene todos los sitios, sin paginación
func GetAllSites(ctx context.Context, collection *mongo.Collection) ([]model.Site, error) {
	// Establecer opciones de consulta, ordenar por fecha de creación descendente
	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	// Ejecutar la consulta
	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Error al consultar todos los sitios")
		return nil, err
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var sites []model.Site
	if err = cursor.All(ctx, &sites); err != nil {
		config.Logger.Error().Err(err).Msg("Error al analizar los datos de todos los sitios")
		return nil, err
	}

	return sites, nil
}
