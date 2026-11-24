// server/dto/rule.go
package dto

import (
	"encoding/json"
)

// MicroRuleCreateRequest petición de creación de microrregla
// @Description Parámetros de la petición para crear una microrregla
type MicroRuleCreateRequest struct {
	Name      string          `json:"name" binding:"required" example:"Regla de protección contra inyección SQL"`                           // Nombre de la regla
	Type      string          `json:"type" binding:"required,oneof=whitelist blacklist" example:"blacklist"` // Tipo de regla
	Status    string          `json:"status" binding:"required,oneof=enabled disabled" example:"enabled"`    // Estado de la regla
	Priority  int             `json:"priority" binding:"required" example:"100"`                             // Campo de prioridad; a mayor número, mayor prioridad
	Condition json.RawMessage `json:"condition" binding:"required" swaggertype:"object"`                     // Condición de la regla
}

// MicroRuleUpdateRequest petición de actualización de microrregla
// @Description Parámetros de la petición para actualizar una microrregla
type MicroRuleUpdateRequest struct {
	Name      string          `json:"name,omitempty" example:"Regla de protección contra inyección SQL"`                                               // Nombre de la regla
	Type      string          `json:"type,omitempty" binding:"omitempty,oneof=whitelist blacklist" example:"blacklist"` // Tipo de regla
	Status    string          `json:"status,omitempty" binding:"omitempty,oneof=enabled disabled" example:"enabled"`    // Estado de la regla
	Priority  *int            `json:"priority,omitempty" example:"100"`                                                 // Campo de prioridad; a mayor número, mayor prioridad
	Condition json.RawMessage `json:"condition,omitempty" swaggertype:"object"`                                         // Condición de la regla
}

// MicroRuleResponse respuesta de microrregla
// @Description Parámetros de respuesta de la microrregla
type MicroRuleResponse struct {
	ID        string          `json:"id,omitempty" example:"60a763d0f03239868b50e810"`
	Name      string          `json:"name,omitempty" example:"Regla de protección contra inyección SQL"`                                               // Nombre de la regla
	Type      string          `json:"type,omitempty" binding:"omitempty,oneof=whitelist blacklist" example:"blacklist"` // Tipo de regla
	Status    string          `json:"status,omitempty" binding:"omitempty,oneof=enabled disabled" example:"enabled"`    // Estado de la regla
	Priority  *int            `json:"priority,omitempty" example:"100"`                                                 // Campo de prioridad; a mayor número, mayor prioridad
	Condition json.RawMessage `json:"condition,omitempty" swaggertype:"object"`                                         // Condición de la regla
}

// MicroRuleListResponse Respuesta de lista de microrreglas
// @Description Respuesta de lista de microrreglas
type MicroRuleListResponse struct {
	Total int64               `json:"total"` // Total
	Items []MicroRuleResponse `json:"items"` // Lista de microrreglas
}
