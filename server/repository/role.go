package repository

import (
	"context"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RoleRepository interfaz del repositorio de roles
type RoleRepository interface {
	FindByID(ctx context.Context, id string) (*model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
	FindAll(ctx context.Context) ([]*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
	InitDefaultRoles() error
}

// MongoRoleRepository repositorio de roles implementado en MongoDB
type MongoRoleRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewRoleRepository crea el repositorio de roles
func NewRoleRepository(db *mongo.Database) RoleRepository {
	var role model.Role
	collection := db.Collection(role.GetCollectionName())
	logger := config.GetRepositoryLogger("role")

	repo := &MongoRoleRepository{
		collection: collection,
		logger:     logger,
	}

	// Inicializar los roles por defecto
	if err := repo.InitDefaultRoles(); err != nil {
		logger.Error().Err(err).Msg("Error al inicializar los roles por defecto")
	}

	return repo
}

// InitDefaultRoles inicializa los roles por defecto
func (r *MongoRoleRepository) InitDefaultRoles() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener el mapeo de permisos de roles por defecto
	rolePermissions := model.GetDefaultRolePermissions()

	// Descripción del rol
	roleDescriptions := map[string]string{
		model.RoleAdmin:        "Administrador del sistema, con todos los permisos",
		model.RoleAuditor:      "Auditor, encargado de ver y auditar los logs del sistema",
		model.RoleConfigurator: "Administrador de configuración, responsable de la gestión de la configuración del sistema",
		model.RoleUser:         "Usuario normal, permisos de operación básicos",
	}

	// Comprobar y crear cada rol por defecto
	for roleName, permissions := range rolePermissions {
		// Comprobar si el rol ya existe
		count, err := r.collection.CountDocuments(ctx, bson.D{{Key: "name", Value: roleName}})
		if err != nil {
			r.logger.Error().Err(err).Str("role", roleName).Msg("Error al consultar el rol")
			continue
		}

		// Si el rol no existe, crearlo
		if count == 0 {
			role := &model.Role{
				Name:        roleName,
				Description: roleDescriptions[roleName],
				Permissions: permissions,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			_, err := r.collection.InsertOne(ctx, role)
			if err != nil {
				r.logger.Error().Err(err).Str("role", roleName).Msg("Error al crear el rol")
				continue
			}

			r.logger.Info().Str("role", roleName).Msg("Rol por defecto creado")
		}
	}

	return nil
}

// Implementar los demás métodos del repositorio...
func (r *MongoRoleRepository) FindByID(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", id).Msg("Error al consultar el rol")
		return nil, err
	}
	return &role, nil
}

func (r *MongoRoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("name", name).Msg("Error al consultar el rol")
		return nil, err
	}
	return &role, nil
}

func (r *MongoRoleRepository) FindAll(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role

	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al consultar todos los roles")
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var role model.Role
		if err := cursor.Decode(&role); err != nil {
			r.logger.Error().Err(err).Msg("Error al decodificar los datos del rol")
			continue
		}
		roles = append(roles, &role)
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error().Err(err).Msg("Error al recorrer los datos de roles")
		return nil, err
	}

	return roles, nil
}

func (r *MongoRoleRepository) Create(ctx context.Context, role *model.Role) error {
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, role)
	if err != nil {
		r.logger.Error().Err(err).Str("name", role.Name).Msg("Error al crear el rol")
		return err
	}
	return nil
}

func (r *MongoRoleRepository) Update(ctx context.Context, role *model.Role) error {
	role.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: role.ID}}, role)
	if err != nil {
		r.logger.Error().Err(err).Str("id", role.ID).Msg("Error al actualizar el rol")
		return err
	}
	return nil
}

func (r *MongoRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		r.logger.Error().Err(err).Str("id", id).Msg("Error al eliminar el rol")
		return err
	}
	return nil
}
