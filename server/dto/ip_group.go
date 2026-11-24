// server/dto/ip_group.go
package dto

import (
	"github.com/anticoi/EscudoWAF/pkg/model"
)

// IPGroupCreateRequest petición de creación de grupo de IPs
// @Description Parámetros de la petición para crear un grupo de IPs
type IPGroupCreateRequest struct {
	Name  string   `json:"name" binding:"required" example:"Servidor interno"`              // Nombre del grupo de IPs
	Items []string `json:"items" binding:"required" example:"[\"192.168.1.1\"]"` // Lista de direcciones IP o CIDR
}

// IPGroupUpdateRequest petición de actualización de grupo de IPs
// @Description Parámetros de la petición para actualizar un grupo de IPs
type IPGroupUpdateRequest struct {
	Name  string   `json:"name,omitempty" example:"Servidor interno"`              // Nombre del grupo de IPs
	Items []string `json:"items,omitempty" example:"[\"192.168.1.1\"]"` // Lista de direcciones IP o CIDR
}

// IPGroupListResponse Respuesta de lista de grupos de IPs
// @Description Respuesta de lista de grupos de IPs
type IPGroupListResponse struct {
	Total int64           `json:"total"` // Total
	Items []model.IPGroup `json:"items"` // Lista de grupos de IPs
}

// AddIPToBlacklistRequest petición para añadir IP a la lista negra
// @Description Petición para añadir dirección IP o CIDR a la lista negra por defecto del sistema
type AddIPToBlacklistRequest struct {
	IP string `json:"ip" binding:"required" example:"192.168.1.1"` // Dirección IP o CIDR
}
