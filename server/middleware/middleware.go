package middleware

import (
	"net/http"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Logger middleware logs the request/response details
func Logger() gin.HandlerFunc {
	log := config.Logger
	isProduction := config.Global.IsProduction

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate request time
		elapsed := time.Since(start)
		statusCode := c.Writer.Status()

		// En producción, solo registrar errores, advertencias o peticiones lentas
		if !isProduction ||
			statusCode >= 400 ||
			elapsed > 500*time.Millisecond {

			// Elegir el nivel de log según el código de estado
			event := log.Info()
			if statusCode >= 400 && statusCode < 500 {
				event = log.Warn()
			} else if statusCode >= 500 {
				event = log.Error()
			}

			event.Str("method", method).
				Str("path", path).
				Int("status", statusCode).
				// Dur("latency", elapsed). //unidad en milisegundos
				Str("latency", elapsed.String()).
				Msg("HTTP Request")
		}
	}
}

// Cors middleware handles CORS requests
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID middleware generates and attaches a unique ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CustomErrorHandler función simplificada de manejo de errores
func CustomErrorHandler(c *gin.Context, err any) {
	requestID, _ := c.Get("RequestID")
	requestIDStr, _ := requestID.(string)

	// Registrar el log de errores
	config.Logger.Error().
		Interface("error", err).
		Str("request", c.Request.URL.Path).
		Str("requestId", requestIDStr).
		Msg("Recovery from panic")

	// Crear respuesta de error estándar
	errorResp := model.NewErrorResponse(
		http.StatusInternalServerError,
		"Error interno del servidor",
		nil,
		// fmt.Errorf("%v", err),
	)

	// Añadir ID de petición
	errorResp.RequestID = requestIDStr

	// Devolver respuesta de error estándar
	c.JSON(http.StatusInternalServerError, errorResp)
}
