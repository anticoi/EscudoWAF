package model

import (
	"net/http"
	"time"
)

// TypedResponse respuesta de API con genéricos, para documentación swagger
type TSuccessResponse[T any] struct {
	Code      int       `json:"code" example:"200"`
	Success   bool      `json:"success" example:"true"`
	Message   string    `json:"message,omitempty" example:"Operación exitosa"`
	Data      T         `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	RequestID string    `json:"requestId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SuccessResponse es la estructura de respuesta de API exitosa
// @Description Formato estándar de respuesta de API exitosa
type SuccessResponse struct {
	Code      int         `json:"code" example:"200"`
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message,omitempty" example:"Operación exitosa"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	RequestID string      `json:"requestId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SuccessResponseNoData es la estructura de respuesta de API exitosa sin datos
// @Description Formato estándar de respuesta de API exitosa, sin data
type SuccessResponseNoData struct {
	Code      int       `json:"code" example:"200"`
	Success   bool      `json:"success" example:"true"`
	Message   string    `json:"message,omitempty" example:"Operación exitosa"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	RequestID string    `json:"requestId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ErrResponse es la estructura de respuesta de API de error
// @Description Formato estándar de respuesta de API de error
type ErrResponse struct {
	Code      int       `json:"code" example:"400"`
	Success   bool      `json:"success" example:"false"`
	Message   string    `json:"message" example:"Parámetros de la petición incorrectos"`
	Error     string    `json:"error" example:"Error en los parámetros"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	RequestID string    `json:"requestId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ErrResponseDontShowError es la estructura de respuesta de API de error sin mostrar error
// @Description Formato estándar de respuesta de API de error, sin mostrar error
type ErrResponseDontShowError struct {
	Code      int       `json:"code" example:"400"`
	Success   bool      `json:"success" example:"false"`
	Message   string    `json:"message" example:"Parámetros de la petición incorrectos"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	RequestID string    `json:"requestId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// APIResponse es la estructura unificada de respuesta de la API
// @Description Formato estándar de respuesta de la API
type APIResponse struct {
	Code      int         `json:"code" example:"200"`                                                 // Código de estado HTTP
	Success   bool        `json:"success" example:"true"`                                             // Si tuvo éxito
	Message   string      `json:"message,omitempty" example:"Operación exitosa"`                                   // Mensaje de la respuesta
	Data      interface{} `json:"data,omitempty"`                                                     // Datos de la respuesta
	Error     string      `json:"error,omitempty" example:"Error en los parámetros"`                                     // Mensaje de error (cuando success=false)
	Timestamp time.Time   `json:"timestamp" example:"2023-01-01T12:00:00Z"`                           // Marca de tiempo de la respuesta
	RequestID string      `json:"requestId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // ID de la petición, para rastreo
}

// NewSuccessResponse crea una respuesta exitosa
func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Code:      http.StatusOK,
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse crea una respuesta de error
func NewErrorResponse(code int, message string, err error) APIResponse {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	return APIResponse{
		Code:      code,
		Success:   false,
		Message:   message,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
}

type APIError struct {
	Code    int
	Message string
	Err     error
}

// Error implementa la interfaz error
func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewAPIError crea un error de API
func NewAPIError(code int, message string, err error) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Códigos de estado de error comunes
var (
	ErrBadRequest          = func(err error) *APIError { return NewAPIError(http.StatusBadRequest, "Parámetros de la petición incorrectos", err) }
	ErrUnauthorized        = func(err error) *APIError { return NewAPIError(http.StatusUnauthorized, "Acceso no autorizado", err) }
	ErrForbidden           = func(err error) *APIError { return NewAPIError(http.StatusForbidden, "Acceso prohibido", err) }
	ErrNotFound            = func(err error) *APIError { return NewAPIError(http.StatusNotFound, "El recurso no existe", err) }
	ErrInternalServerError = func(err error) *APIError {
		return NewAPIError(http.StatusInternalServerError, "Error interno del servidor", err)
	}
)
