package dto

import (
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
)

// BlockedIPListRequest parámetros de la petición de lista de IPs bloqueadas
// @Description Parámetros de la petición de lista de IPs bloqueadas
type BlockedIPListRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`                                        // Número de página
	Size    int    `form:"size" binding:"omitempty,min=1,max=100" example:"10"`                               // Cantidad por página
	IP      string `form:"ip" binding:"omitempty" example:"192.168.1.1"`                                      // Filtrado por dirección IP
	Reason  string `form:"reason" binding:"omitempty" example:"high_frequency_attack"`                        // Filtro por motivo de bloqueo
	Status  string `form:"status" binding:"omitempty,oneof=active expired all" example:"active"`              // Filtro de estado: active - activo, expired - expirado, all - todos
	SortBy  string `form:"sortBy" binding:"omitempty,oneof=blocked_at blocked_until ip" example:"blocked_at"` // Campo de ordenación
	SortDir string `form:"sortDir" binding:"omitempty,oneof=asc desc" example:"desc"`                         // Dirección de ordenación
}

// BlockedIPResponse respuesta de IP bloqueada
// @Description Detalles de la IP bloqueada
type BlockedIPResponse struct {
	IP           string    `json:"ip" example:"192.168.1.1"`                    // Dirección IP bloqueada
	Reason       string    `json:"reason" example:"high_frequency_attack"`      // Motivo del bloqueo
	RequestUri   string    `json:"requestUri" example:"/api/v1/login"`          // URI de la petición
	BlockedAt    time.Time `json:"blockedAt" example:"2023-12-01T10:00:00Z"`    // Hora de inicio del bloqueo
	BlockedUntil time.Time `json:"blockedUntil" example:"2023-12-01T11:00:00Z"` // Hora de fin del bloqueo
	IsActive     bool      `json:"isActive" example:"true"`                     // Si sigue bloqueado
	RemainingTTL int64     `json:"remainingTTL" example:"3600"`                 // Tiempo de bloqueo restante (segundos)
}

// BlockedIPListResponse respuesta de lista de IPs bloqueadas
// @Description Respuesta paginada de la lista de IPs bloqueadas
type BlockedIPListResponse struct {
	Total int64               `json:"total" example:"100"` // Cantidad total
	Items []BlockedIPResponse `json:"items"`               // Lista de IPs
	Page  int                 `json:"page" example:"1"`    // Página actual
	Size  int                 `json:"size" example:"10"`   // Cantidad por página
	Pages int                 `json:"pages" example:"10"`  // Total de páginas
}

// BlockedIPStatsResponse respuesta de estadísticas de IPs bloqueadas
// @Description Estadísticas de IPs bloqueadas
type BlockedIPStatsResponse struct {
	TotalBlocked    int64                  `json:"totalBlocked" example:"1000"`  // Cantidad total de bloqueos
	ActiveBlocked   int64                  `json:"activeBlocked" example:"50"`   // Cantidad de bloqueos actualmente activos
	ExpiredBlocked  int64                  `json:"expiredBlocked" example:"950"` // Cantidad de bloqueos expirados
	ReasonStats     map[string]int64       `json:"reasonStats"`                  // Estadísticas por motivo
	Last24HourStats []BlockedIPHourlyStats `json:"last24HourStats"`              // Estadísticas de las últimas 24 horas
}

// BlockedIPHourlyStats estadísticas por hora
// @Description Estadísticas de bloqueo por hora
type BlockedIPHourlyStats struct {
	Hour  string `json:"hour" example:"2023-12-01T10:00:00Z"` // Punto de tiempo por horas
	Count int64  `json:"count" example:"10"`                  // Cantidad de bloqueos en esa hora
}

// MapToResponse convierte el modelo a DTO de respuesta
func (r *BlockedIPResponse) MapFromModel(record *model.BlockedIPRecord) {
	r.IP = record.IP
	r.Reason = record.Reason
	r.RequestUri = record.RequestUri
	r.BlockedAt = record.BlockedAt
	r.BlockedUntil = record.BlockedUntil

	// Calcular si sigue bloqueado
	now := time.Now()
	r.IsActive = now.Before(record.BlockedUntil)

	// Calcular el tiempo de bloqueo restante
	if r.IsActive {
		r.RemainingTTL = int64(record.BlockedUntil.Sub(now).Seconds())
	} else {
		r.RemainingTTL = 0
	}
}

// MapToResponseList convierte la lista de modelos a lista de DTOs de respuesta
func MapToResponseList(records []model.BlockedIPRecord) []BlockedIPResponse {
	responses := make([]BlockedIPResponse, len(records))
	for i, record := range records {
		responses[i].MapFromModel(&record)
	}
	return responses
}

// BlockedIPCleanupResponse respuesta de limpieza de registros de IPs bloqueadas expirados
// @Description Datos de respuesta de la limpieza de registros de IPs bloqueadas expirados
type BlockedIPCleanupResponse struct {
	DeletedCount int64  `json:"deletedCount" example:"25"`        // Cantidad de registros eliminados
	Message      string `json:"message" example:"Registros de IPs bloqueadas expirados limpiados con éxito"` // Mensaje del resultado de la operación
}
