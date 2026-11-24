package service

import (
	"context"
	"errors"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/anticoi/EscudoWAF/server/utils/jwt"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Definición de errores
var (
	ErrUserNotFound     = errors.New("El usuario no existe")
	ErrInvalidPassword  = errors.New("Contraseña incorrecta")
	ErrUserAlreadyExist = errors.New("El usuario ya existe")
	ErrForbidden        = errors.New("Sin permisos")
)

// AuthService interfaz del servicio de autenticación
type AuthService interface {
	Login(ctx context.Context, req dto.UserLoginRequest) (string, *model.User, error)
	ResetPassword(ctx context.Context, userID bson.ObjectID, req dto.UserPasswordResetRequest) error
	CreateUser(ctx context.Context, adminID bson.ObjectID, req dto.UserCreateRequest) (*model.User, error)
	GetUsers(ctx context.Context) ([]*model.User, error)
}

// AuthServiceImpl implementación del servicio de autenticación
type AuthServiceImpl struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	logger   zerolog.Logger
}

// NewAuthService crea el servicio de autenticación
func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) AuthService {
	return &AuthServiceImpl{
		userRepo: userRepo,
		roleRepo: roleRepo,
		logger:   config.GetServiceLogger("auth"),
	}
}

// Login Inicio de sesión
func (s *AuthServiceImpl) Login(ctx context.Context, req dto.UserLoginRequest) (string, *model.User, error) {
	// Buscar usuario
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, ErrUserNotFound
	}

	// Validar la contraseña
	if !user.CheckPassword(req.Password) {
		return "", nil, ErrInvalidPassword
	}

	// Actualizar la hora del último inicio de sesión
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Str("userId", user.ID.Hex()).Msg("Error al actualizar la hora de inicio de sesión")
	}

	// Generar el token
	token, err := jwt.GenerateToken(*user, 24*time.Hour)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// ResetPassword Restablecer contraseña
func (s *AuthServiceImpl) ResetPassword(ctx context.Context, userID bson.ObjectID, req dto.UserPasswordResetRequest) error {
	// Buscar usuario
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Validar la contraseña anterior
	if !user.CheckPassword(req.OldPassword) {
		return ErrInvalidPassword
	}

	// Actualizar la contraseña
	user.Password = req.NewPassword
	if err := user.HashPassword(); err != nil {
		return err
	}

	// Marcar la contraseña como restablecida
	user.NeedReset = false
	user.UpdatedAt = time.Now()

	// Guardar usuario
	return s.userRepo.Update(ctx, user)
}

// CreateUser crea un usuario (solo administradores)
func (s *AuthServiceImpl) CreateUser(ctx context.Context, adminID bson.ObjectID, req dto.UserCreateRequest) (*model.User, error) {
	// Validar los permisos de administrador
	admin, err := s.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if admin == nil || admin.Role != model.RoleAdmin {
		return nil, ErrForbidden
	}

	// Comprobar si el usuario ya existe
	existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExist
	}

	// Crear nuevo usuario
	user := &model.User{
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		NeedReset: true, // El nuevo usuario debe restablecer la contraseña
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generar hash de la contraseña
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	// Guardar usuario
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// No devolver la contraseña
	user.Password = ""
	return user, nil
}

// GetUsers obtiene todos los usuarios (solo administradores)
func (s *AuthServiceImpl) GetUsers(ctx context.Context) ([]*model.User, error) {
	// Implementar la lógica de obtener todos los usuarios
	// Esto requiere añadir un método FindAll a UserRepository
	return s.userRepo.FindAll(ctx)
}
