package repository

import (
	"context"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// UserRepository interfaz del repositorio de usuarios
type UserRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	UpdateLastLogin(ctx context.Context, id bson.ObjectID) error
	FindAll(ctx context.Context) ([]*model.User, error)
	InitAdminUser() error
}

// MongoUserRepository repositorio de usuarios implementado en MongoDB
type MongoUserRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewUserRepository crea el repositorio de usuarios
func NewUserRepository(db *mongo.Database) UserRepository {
	var user model.User
	collection := db.Collection(user.GetCollectionName())
	logger := config.GetRepositoryLogger("user")

	// Crear índices
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Índice único de nombre de usuario
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error al crear el índice de nombres de usuario")
	}

	repo := &MongoUserRepository{
		collection: collection,
		logger:     logger,
	}

	// Inicializar el usuario administrador
	if err := repo.InitAdminUser(); err != nil {
		logger.Error().Err(err).Msg("Error al inicializar el usuario administrador")
	}

	return repo
}

// FindByID busca un usuario por ID
func (r *MongoUserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al consultar el usuario")
		return nil, err
	}
	return &user, nil
}

// FindByUsername busca un usuario por nombre de usuario
func (r *MongoUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.D{{Key: "username", Value: username}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("username", username).Msg("Error al consultar el usuario")
		return nil, err
	}
	return &user, nil
}

// Create crea un usuario
func (r *MongoUserRepository) Create(ctx context.Context, user *model.User) error {
	// Establecer la hora
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// No establecer ID, dejar que MongoDB genere el ObjectID
	// MongoDB crea automáticamente el campo _id al insertar

	// Insertar datos
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		r.logger.Error().Err(err).Str("username", user.Username).Msg("Error al crear el usuario")
		return err
	}

	// Asignar el ObjectID generado al ID de usuario
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		user.ID = oid
	}

	return nil
}

// Update actualiza un usuario
func (r *MongoUserRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: user.ID}}, user)
	if err != nil {
		r.logger.Error().Err(err).Str("id", user.ID.Hex()).Msg("Error al actualizar el usuario")
		return err
	}
	return nil
}

// UpdateLastLogin actualiza la hora del último inicio de sesión
func (r *MongoUserRepository) UpdateLastLogin(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "lastLogin", Value: time.Now()},
			{Key: "updatedAt", Value: time.Now()},
		}}},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al actualizar la hora de inicio de sesión")
		return err
	}
	return nil
}

// FindAll busca todos los usuarios
func (r *MongoUserRepository) FindAll(ctx context.Context) ([]*model.User, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al consultar todos los usuarios")
		return nil, err
	}

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error().Err(err).Msg("Error al analizar la lista de usuarios")
		return nil, err
	}

	return users, nil
}

// InitAdminUser inicializa el usuario administrador
func (r *MongoUserRepository) InitAdminUser() error {
	// Comprobar si ya existe un usuario administrador
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.D{{Key: "role", Value: model.RoleAdmin}})
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al consultar el usuario administrador")
		return err
	}

	// Si no hay usuario administrador, crear uno
	if count == 0 {
		// Crear el usuario administrador por defecto
		adminUser := &model.User{
			Username:  model.RoleAdmin,
			Password:  "admin123", // Contraseña inicial
			Role:      model.RoleAdmin,
			NeedReset: true, // Se requiere restablecer la contraseña
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Generar hash de la contraseña
		if err := adminUser.HashPassword(); err != nil {
			r.logger.Error().Err(err).Msg("Error al generar el hash de la contraseña del administrador")
			return err
		}

		// Guardar usuario
		if err := r.Create(ctx, adminUser); err != nil {
			r.logger.Error().Err(err).Msg("Error al crear el usuario administrador")
			return err
		}

		// Crear el usuario auditor
		auditorUser := &model.User{
			Username:  model.RoleAuditor,
			Password:  "auditor123", // Contraseña inicial
			Role:      model.RoleAuditor,
			NeedReset: true, // Se requiere restablecer la contraseña
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Generar hash de la contraseña
		if err := auditorUser.HashPassword(); err != nil {
			r.logger.Error().Err(err).Msg("Error al generar el hash de la contraseña del auditor")
			return err
		}

		// Guardar usuario
		if err := r.Create(ctx, auditorUser); err != nil {
			r.logger.Error().Err(err).Msg("Error al crear el usuario auditor")
			return err
		}

		// Crear usuario administrador de configuración
		configUser := &model.User{
			Username:  model.RoleConfigurator,
			Password:  "config123", // Contraseña inicial
			Role:      model.RoleConfigurator,
			NeedReset: true, // Se requiere restablecer la contraseña
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Generar hash de la contraseña
		if err := configUser.HashPassword(); err != nil {
			r.logger.Error().Err(err).Msg("Error al generar el hash de la contraseña del administrador de configuración")
			return err
		}

		// Guardar usuario
		if err := r.Create(ctx, configUser); err != nil {
			r.logger.Error().Err(err).Msg("Error al crear el usuario administrador de configuración")
			return err
		}

		r.logger.Info().Msg("Usuario por defecto creado")
	}

	return nil
}
