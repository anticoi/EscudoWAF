package dto

import (
	"time"

	"github.com/anticoi/EscudoWAF/server/model"
)

// CreateSiteRequest petición de creación de sitio
// @Description Parámetros de la petición para crear un sitio
type CreateSiteRequest struct {
	Name         string          `json:"name" binding:"required" example:"my-site"`                                      // Nombre del sitio
	Domain       string          `json:"domain" binding:"required,domain" example:"example.com"`                         // Dominio
	ListenPort   int             `json:"listenPort" binding:"required,min=1,max=65535" example:"8080"`                   // Puerto de escucha
	EnableHTTPS  bool            `json:"enableHTTPS" example:"false"`                                                    // Si HTTPS está habilitado
	Certificate  *CertificateDTO `json:"certificate,omitempty" binding:"omitempty,required_if=EnableHTTPS true"`         // Información del certificado
	Backend      BackendDTO      `json:"backend" binding:"required"`                                                     // Configuración del servidor backend
	WAFEnabled   bool            `json:"wafEnabled" example:"false"`                                                     // Si el WAF está habilitado
	WAFMode      string          `json:"wafMode" binding:"omitempty,oneof=protection observation" example:"observation"` // Modo WAF
	ActiveStatus bool            `json:"activeStatus" example:"true"`                                                    // Estado del sitio
}

// UpdateSiteRequest petición de actualización de sitio
// @Description Parámetros de la petición para actualizar un sitio
type UpdateSiteRequest struct {
	Name         string          `json:"name,omitempty" binding:"omitempty" example:"my-site"`                           // Nombre del sitio
	Domain       string          `json:"domain,omitempty" binding:"omitempty,domain" example:"example.com"`              // Dominio
	ListenPort   int             `json:"listenPort,omitempty" binding:"omitempty,min=1,max=65535" example:"8080"`        // Puerto de escucha
	EnableHTTPS  bool            `json:"enableHTTPS" example:"false"`                                                    // Si HTTPS está habilitado
	Certificate  *CertificateDTO `json:"certificate,omitempty" binding:"omitempty,required_if=EnableHTTPS true"`         // Información del certificado
	Backend      *BackendDTO     `json:"backend,omitempty" binding:"omitempty"`                                          // Configuración del servidor backend
	WAFEnabled   bool            `json:"wafEnabled" example:"false"`                                                     // Si el WAF está habilitado
	WAFMode      string          `json:"wafMode" binding:"omitempty,oneof=protection observation" example:"observation"` // Modo WAF
	ActiveStatus bool            `json:"activeStatus" example:"true"`                                                    // Estado del sitio
}

// CertificateDTO DTO de certificado
type CertificateDTO struct {
	CertName    string    `json:"certName" binding:"required" example:"my-cert"`         // Nombre del certificado
	PublicKey   string    `json:"publicKey" binding:"required"`                          // Contenido de la clave pública
	PrivateKey  string    `json:"privateKey" binding:"required"`                         // Contenido de la clave privada
	ExpireDate  time.Time `json:"expireDate" binding:"required"`                         // Hora de expiración
	IssuerName  string    `json:"issuerName" binding:"required" example:"Let's Encrypt"` // Autoridad emisora
	FingerPrint string    `json:"fingerPrint" binding:"required"`                        // Huella del certificado
}

// BackendDTO DTO de configuración de servidor backend
type BackendDTO struct {
	Servers []ServerDTO `json:"servers" binding:"required,min=1,dive"` // Lista de servidores, se requiere al menos uno
}

// ServerDTO DTO de servidor
type ServerDTO struct {
	Host  string `json:"host" binding:"required" example:"backend.example.com"` // Dirección del host
	Port  int    `json:"port" binding:"required,min=1,max=65535" example:"80"`  // Puerto
	IsSSL bool   `json:"isSSL" example:"false"`                                 // Si SSL está habilitado
}

// SiteResponse respuesta de sitio
// @Description Respuesta de información del sitio
type SiteResponse struct {
	model.Site
}

// SiteListResponse Respuesta de lista de sitios
// @Description Respuesta de lista de sitios
type SiteListResponse struct {
	Total int64        `json:"total"` // Total
	Items []model.Site `json:"items"` // Lista de sitios
}
