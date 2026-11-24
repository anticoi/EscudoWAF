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
	ErrCertNotFound      = errors.New("El certificado no existe")
	ErrCertNameExists    = errors.New("El nombre del certificado ya existe")
	ErrInvalidCertFormat = errors.New("Formato de certificado no válido")
)

// CertificateRepository interfaz del repositorio de certificados
type CertificateRepository interface {
	CreateCertificate(ctx context.Context, certificate *model.CertificateStore) error
	GetCertificates(ctx context.Context, page, size int64) ([]model.CertificateStore, int64, error)
	GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error)
	UpdateCertificate(ctx context.Context, certificate *model.CertificateStore) error
	DeleteCertificate(ctx context.Context, id bson.ObjectID) error
	CheckCertificateNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error)
}

// MongoCertificateRepository repositorio de certificados implementado en MongoDB
type MongoCertificateRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewCertificateRepository crea el repositorio de certificados
func NewCertificateRepository(db *mongo.Database) CertificateRepository {
	var cert model.CertificateStore
	collection := db.Collection(cert.GetCollectionName())
	logger := config.GetRepositoryLogger("certificate")

	// Crear índices
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Índice único de nombre de certificado
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear el índice de nombres de certificados")
	}

	return &MongoCertificateRepository{
		collection: collection,
		logger:     logger,
	}
}

// CreateCertificate crea un certificado
func (r *MongoCertificateRepository) CreateCertificate(ctx context.Context, certificate *model.CertificateStore) error {
	// Establecer las fechas de creación y actualización
	now := time.Now()
	certificate.CreatedAt = now
	certificate.UpdatedAt = now

	// Insertar nuevo certificado
	result, err := r.collection.InsertOne(ctx, certificate)
	if err != nil {
		r.logger.Error().Err(err).Str("name", certificate.Name).Msg("Error al insertar el certificado")
		return err
	}

	certificate.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetCertificates Obtener lista de certificados
func (r *MongoCertificateRepository) GetCertificates(ctx context.Context, page, size int64) ([]model.CertificateStore, int64, error) {
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
		r.logger.Error().Err(err).Msg("Error al consultar la lista de certificados")
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Resultado del análisis
	var certificates []model.CertificateStore
	if err = cursor.All(ctx, &certificates); err != nil {
		r.logger.Error().Err(err).Msg("Error al analizar la lista de certificados")
		return nil, 0, err
	}

	// Obtener el total
	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener el total de certificados")
		return nil, 0, err
	}

	return certificates, total, nil
}

// GetCertificateByID obtiene un certificado por ID
func (r *MongoCertificateRepository) GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error) {
	var certificate model.CertificateStore
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&certificate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCertNotFound
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al consultar el certificado")
		return nil, err
	}

	return &certificate, nil
}

// UpdateCertificate Actualizar certificado
func (r *MongoCertificateRepository) UpdateCertificate(ctx context.Context, certificate *model.CertificateStore) error {
	// Hora de actualización
	certificate.UpdatedAt = time.Now()

	// Actualizar certificado
	_, err := r.collection.ReplaceOne(
		ctx,
		bson.D{{Key: "_id", Value: certificate.ID}},
		certificate,
	)

	if err != nil {
		r.logger.Error().Err(err).Str("id", certificate.ID.Hex()).Msg("Error al actualizar el certificado")
		return err
	}

	return nil
}

// DeleteCertificate Eliminar certificado
func (r *MongoCertificateRepository) DeleteCertificate(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar el certificado")
		return err
	}

	if result.DeletedCount == 0 {
		return ErrCertNotFound
	}

	return nil
}

// CheckCertificateNameExists comprueba si el nombre del certificado ya existe
func (r *MongoCertificateRepository) CheckCertificateNameExists(ctx context.Context, name string, excludeID bson.ObjectID) (bool, error) {
	filter := bson.D{{Key: "name", Value: name}}

	// Si es una actualización, excluir el ID del certificado actual
	if excludeID != bson.NilObjectID {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$ne", Value: excludeID}}})
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("name", name).Msg("Error al comprobar si el nombre del certificado existe")
		return false, err
	}

	return count > 0, nil
}
