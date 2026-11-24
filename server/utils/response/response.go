package response

import (
	"net/http"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/gin-gonic/gin"
)

// WithRequestID añade el ID de petición a la respuesta
func WithRequestID(c *gin.Context, resp model.APIResponse) model.APIResponse {
	requestID, exists := c.Get("RequestID")
	if exists {
		resp.RequestID = requestID.(string)
	}
	return resp
}

// Success devuelve una respuesta exitosa
func Success(c *gin.Context, message string, data interface{}) {
	resp := model.NewSuccessResponse(message, data)
	resp = WithRequestID(c, resp)
	c.JSON(http.StatusOK, resp)
}

// Error devuelve una respuesta de error
func Error(c *gin.Context, apiErr *model.APIError, showErr bool) {
	// Registrar el log de errores
	logger := config.Logger
	if apiErr.Err != nil {
		logger.Error().Err(apiErr.Err).Str("message", apiErr.Message).Int("code", apiErr.Code).Send()
	} else {
		logger.Error().Str("message", apiErr.Message).Int("code", apiErr.Code).Send()
	}

	// Crear respuesta de error
	// err type zero value is nil
	var errToShow error
	if showErr {
		errToShow = apiErr.Err
	}
	resp := model.NewErrorResponse(apiErr.Code, apiErr.Message, errToShow)
	resp = WithRequestID(c, resp)

	// Devolver respuesta de error
	c.JSON(apiErr.Code, resp)
	c.Abort()
}

// BadRequest devuelve error 400
func BadRequest(c *gin.Context, err error, showErr bool) {
	Error(c, model.ErrBadRequest(err), showErr)
}

// Unauthorized devuelve error 401
func Unauthorized(c *gin.Context, err error) {
	Error(c, model.ErrUnauthorized(err), false)
}

// Forbidden devuelve error 403
func Forbidden(c *gin.Context, err error) {
	Error(c, model.ErrForbidden(err), false)
}

// NotFound devuelve error 404
func NotFound(c *gin.Context, err error) {
	Error(c, model.ErrNotFound(err), false)
}

// InternalServerError devuelve error 500
func InternalServerError(c *gin.Context, err error, showErr bool) {
	Error(c, model.ErrInternalServerError(err), showErr)
}
