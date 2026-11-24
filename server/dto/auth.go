package dto

import (
	"github.com/anticoi/EscudoWAF/server/model"
)

// UserLoginRequest petición de inicio de sesión
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserPasswordResetRequest petición de restablecimiento de contraseña
type UserPasswordResetRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UserCreateRequest petición de creación de usuario (solo administradores)
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// UserUpdateRequest petición de actualización de usuario
type UserUpdateRequest struct {
	Username  string `json:"username,omitempty" binding:"omitempty,min=3,max=20"`
	Password  string `json:"password,omitempty" binding:"omitempty,min=6"`
	Role      string `json:"role,omitempty" binding:"omitempty,oneof=admin auditor configurator user"`
	NeedReset *bool  `json:"needReset,omitempty"`
}

// LoginResponseData datos de respuesta de inicio de sesión
type LoginResponseData struct {
	Token string     `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."` // JWT token
	User  model.User `json:"user"`                                    // Información del usuario
}

type ResetPasswordResponseData = model.TSuccessResponse[model.User]

type GetUserInfoResponseData struct {
	ID        string `json:"id" example:"1234567890"`
	Username  string `json:"username" example:"user123"`
	Role      string `json:"role" example:"admin"`
	NeedReset bool   `json:"needReset" example:"false"`
}
