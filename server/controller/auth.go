package controller

import (
	"errors"
	"net/http"

	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AuthController controlador de autenticación
type AuthController interface {
	Login(ctx *gin.Context)
	ResetPassword(ctx *gin.Context)
	CreateUser(ctx *gin.Context)
	GetUsers(ctx *gin.Context)
	GetUserInfo(ctx *gin.Context)
	DeleteUser(ctx *gin.Context)
	UpdateUser(ctx *gin.Context)
}

type AuthControllerImpl struct {
	authService service.AuthService
}

// NewAuthController crea el controlador de autenticación
func NewAuthController(authService service.AuthService) AuthController {
	return &AuthControllerImpl{
		authService: authService,
	}
}

// Login Inicio de sesión
//
//	@Summary		Inicio de sesión
//	@Description	Inicio de sesión y obtención de token JWT
//	@Tags			Autenticación
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UserLoginRequest								true	"Información de inicio de sesión"
//	@Success		200		{object}	model.SuccessResponse{data=dto.LoginResponseData}	"Inicio de sesión exitoso"
//	@Failure		400		{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		401		{object}	model.ErrResponseDontShowError						"Nombre de usuario o contraseña incorrectos"
//	@Failure		500		{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/auth/login [post]
func (c *AuthControllerImpl) Login(ctx *gin.Context) {
	var req dto.UserLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	// Iniciar sesión
	token, user, err := c.authService.Login(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidPassword) {
			response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "Nombre de usuario o contraseña incorrectos", err), false)
			return
		}
		response.InternalServerError(ctx, err, false)
		return
	}

	// Devolver el token y la información del usuario
	response.Success(ctx, "Inicio de sesión exitoso", gin.H{
		"token": token,
		"user":  user,
	})
}

// ResetPassword Restablecer contraseña
//
//	@Summary		Restablecer contraseña
//	@Description	El usuario restablece su propia contraseña
//	@Tags			Autenticación
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.UserPasswordResetRequest	true	"Información de restablecimiento de contraseña"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"Contraseña restablecida con éxito"
//	@Failure		400	{object}	model.ErrResponseDontShowError	"Parámetros incorrectos o contraseña actual incorrecta"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/auth/reset-password [post]
func (c *AuthControllerImpl) ResetPassword(ctx *gin.Context) {
	var req dto.UserPasswordResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, false)
		return
	}

	// Obtener el ID de usuario del contexto
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, nil)
		return
	}

	// Restablecer contraseña
	userID, err := bson.ObjectIDFromHex(userID.(string))
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	err = c.authService.ResetPassword(ctx, userID.(bson.ObjectID), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			response.Error(ctx, model.NewAPIError(http.StatusBadRequest, "La contraseña actual es incorrecta", err), false)
			return
		}
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Contraseña restablecida con éxito", nil)
}

// CreateUser crea un usuario
//
//	@Summary		Crear nuevo usuario
//	@Description	El administrador crea un nuevo usuario
//	@Tags			Gestión de usuarios
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.UserCreateRequest	true	"Información de creación de usuario"
//	@Security		BearerAuth
//	@Success		200	{object}	dto.ResetPasswordResponseData	"Usuario creado con éxito"
//	@Failure		400	{object}	model.ErrResponse				"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"Acceso prohibido"
//	@Failure		409	{object}	model.ErrResponseDontShowError	"El nombre de usuario ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/users [post]
func (c *AuthControllerImpl) CreateUser(ctx *gin.Context) {
	var req dto.UserCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	// Obtener el ID de administrador del contexto
	adminID, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, nil)
		return
	}

	adminID, err := bson.ObjectIDFromHex(adminID.(string))
	if err != nil {
		response.Unauthorized(ctx, nil)
		return
	}

	// Crear usuario
	user, err := c.authService.CreateUser(ctx, adminID.(bson.ObjectID), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExist) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "El nombre de usuario ya existe", err), false)
			return
		} else if errors.Is(err, service.ErrForbidden) {
			response.Forbidden(ctx, err)
			return
		}
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Usuario creado con éxito", user)
}

// GetUsers Obtener todos los usuarios
//
//	@Summary		Obtener todos los usuarios
//	@Description	Obtener la lista de todos los usuarios del sistema
//	@Tags			Gestión de usuarios
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=[]model.User}	"Lista de usuarios obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError				"Acceso prohibido"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"Error interno del servidor"
//	@Router			/users [get]
func (c *AuthControllerImpl) GetUsers(ctx *gin.Context) {
	// Obtener todos los usuarios
	users, err := c.authService.GetUsers(ctx)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Lista de usuarios obtenida con éxito", users)
}

// GetUserInfo obtiene información del usuario
//
//	@Summary		Obtener información del usuario actual
//	@Description	Obtener información detallada del usuario autenticado
//	@Tags			Autenticación
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.GetUserInfoResponseData}	"Información del usuario obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"Acceso no autorizado"
//	@Router			/auth/me [get]
func (c *AuthControllerImpl) GetUserInfo(ctx *gin.Context) {
	// Obtener la información del usuario del contexto
	userID, _ := ctx.Get("userID")
	username, _ := ctx.Get("username")
	userRole, _ := ctx.Get("userRole")
	needReset, _ := ctx.Get("needReset")

	response.Success(ctx, "Información del usuario obtenida con éxito", gin.H{
		"id":        userID,
		"username":  username,
		"role":      userRole,
		"needReset": needReset,
	})
}

// DeleteUser Eliminar usuario
//
//	@Summary		Eliminar usuario
//	@Description	El administrador elimina el usuario especificado
//	@Tags			Gestión de usuarios
//	@Produce		json
//	@Param			id	path	string	true	"ID de usuario"
//	@Security		BearerAuth
//	@Success		200	{object}	model.APIResponse	"Usuario eliminado con éxito"
//	@Failure		401	{object}	model.APIResponse	"Acceso no autorizado"
//	@Failure		403	{object}	model.APIResponse	"Acceso prohibido"
//	@Failure		404	{object}	model.APIResponse	"El usuario no existe"
//	@Failure		500	{object}	model.APIResponse	"Error interno del servidor"
//	@Router			/users/{id} [delete]
func (c *AuthControllerImpl) DeleteUser(ctx *gin.Context) {
	// Por implementar
}

// UpdateUser actualiza un usuario
//
//	@Summary		Actualizar información del usuario
//	@Description	El administrador actualiza la información del usuario especificado
//	@Tags			Gestión de usuarios
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string					true	"ID de usuario"
//	@Param			request	body	dto.UserUpdateRequest	true	"Información de actualización del usuario"
//	@Security		BearerAuth
//	@Success		200	{object}	model.APIResponse	"Usuario actualizado con éxito"
//	@Failure		400	{object}	model.APIResponse	"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.APIResponse	"Acceso no autorizado"
//	@Failure		403	{object}	model.APIResponse	"Acceso prohibido"
//	@Failure		404	{object}	model.APIResponse	"El usuario no existe"
//	@Failure		500	{object}	model.APIResponse	"Error interno del servidor"
//	@Router			/users/{id} [put]
func (c *AuthControllerImpl) UpdateUser(ctx *gin.Context) {
	// Por implementar
}
