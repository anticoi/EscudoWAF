package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/anticoi/EscudoWAF/server/utils/jwt"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// HasPermission middleware de verificación de permisos
func HasPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el rol del usuario
		role, exists := c.Get("userRole")
		if !exists {
			response.Unauthorized(c, fmt.Errorf("No hay información de rol de usuario en el contexto de la petición"))
			c.Abort()
			return
		}

		// Si es administrador, pasar directamente
		if role == model.RoleAdmin {
			c.Next()
			return
		}

		// Obtener los permisos del usuario
		var userPermissions []string

		roleRepo := c.MustGet("roleRepo").(repository.RoleRepository)
		// Obtener los permisos por defecto del rol
		rolePermissions := model.GetDefaultRolePermissions()[role.(string)]

		if rolePermissions == nil {
			ctx := c.Request.Context()
			roleObj, err := roleRepo.FindByName(ctx, role.(string))
			if err != nil {
				response.InternalServerError(c, err, false)
				c.Abort()
				return
			}
			rolePermissions = roleObj.Permissions
		}

		// Obtener los permisos adicionales del usuario
		extraPermissions, exists := c.Get("userPermissions")
		if exists {
			userPermissions = append(rolePermissions, extraPermissions.([]string)...)
		} else {
			userPermissions = rolePermissions
		}

		// Comprobar si se tienen los permisos necesarios
		hasPermission := false
		for _, perm := range userPermissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.Forbidden(c, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// JWTAuth middleware de autenticación JWT
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el token de la cabecera de la petición
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, fmt.Errorf("Token no proporcionado"))
			c.Abort()
			return
		}

		// Comprobar el formato
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, fmt.Errorf("Formato de token no válido"))
			c.Abort()
			return
		}

		// Analizar el token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			if err == jwt.ErrExpiredToken {
				response.Error(c, model.NewAPIError(http.StatusUnauthorized, "Token expirado", err), false)
			} else {
				response.Unauthorized(c, err)
			}
			c.Abort()
			return
		}

		// Obtener la información del usuario
		userRepo := c.MustGet("userRepo").(repository.UserRepository)
		userID, err := bson.ObjectIDFromHex(claims.UserID)
		if err != nil {
			response.Unauthorized(c, nil)
			c.Abort()
			return
		}
		user, err := userRepo.FindByID(c, userID)
		if err != nil || user == nil {
			response.Unauthorized(c, nil)
			c.Abort()
			return
		}

		// Almacenar la información del usuario en el contexto
		c.Set("userID", user.ID.Hex())
		c.Set("username", user.Username)
		c.Set("userRole", user.Role)
		c.Set("userPermissions", user.Permissions)
		c.Set("needReset", user.NeedReset)

		c.Next()
	}
}

// PasswordResetRequired middleware de verificación de restablecimiento de contraseña
func PasswordResetRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		needReset, exists := c.Get("needReset")
		if !exists {
			response.Unauthorized(c, fmt.Errorf("No existe el indicador de restablecimiento de contraseña"))
			c.Abort()
			return
		}

		// Si se requiere restablecer la contraseña, solo permitir acceso a la interfaz de restablecimiento
		if needReset.(bool) && c.FullPath() != "/api/v1/auth/reset-password" {
			response.Error(c, model.NewAPIError(http.StatusForbidden, "Por favor, restablezca primero la contraseña", nil), false)
			c.Abort()
			return
		}

		c.Next()
	}
}
