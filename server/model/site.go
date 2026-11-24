package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// WAFMode define el tipo de modo de trabajo del WAF
type WAFMode string

const (
	WAFModeProtection  WAFMode = "protection"  // Modo de protección
	WAFModeObservation WAFMode = "observation" // Modo de observación
)

// Site representa la configuración de un sitio
type Site struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`                  // ID del sitio
	Name         string        `bson:"name" json:"name"`                                   // Nombre del sitio
	Domain       string        `bson:"domain" json:"domain"`                               // Dominio, p. ej. a.com
	ListenPort   int           `bson:"listenPort" json:"listenPort"`                       // Puerto de escucha, p. ej. 9000
	EnableHTTPS  bool          `bson:"enableHTTPS" json:"enableHTTPS"`                     // Si HTTPS está habilitado
	Certificate  Certificate   `bson:"certificate,omitempty" json:"certificate,omitempty"` // Información del certificado
	Backend      Backend       `bson:"backend" json:"backend"`                             // Configuración del servidor backend
	WAFEnabled   bool          `bson:"wafEnabled" json:"wafEnabled"`                       // Si el WAF está habilitado
	WAFMode      WAFMode       `bson:"wafMode" json:"wafMode"`                             // Modo de protección WAF
	CreatedAt    time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time     `bson:"updatedAt" json:"updatedAt"`
	ActiveStatus bool          `bson:"activeStatus" json:"activeStatus"` // Si el sitio está activo
}

// Certificate representa la información de un certificado
type Certificate struct {
	CertName    string    `bson:"certName" json:"certName"`       // Nombre/alias del certificado
	PublicKey   string    `bson:"publicKey" json:"publicKey"`     // Contenido de la clave pública (formato PEM)
	PrivateKey  string    `bson:"privateKey" json:"privateKey"`   // Contenido de la clave privada (formato PEM)
	ExpireDate  time.Time `bson:"expireDate" json:"expireDate"`   // Fecha de expiración del certificado
	IssuerName  string    `bson:"issuerName" json:"issuerName"`   // Autoridad emisora
	FingerPrint string    `bson:"fingerPrint" json:"fingerPrint"` // Huella del certificado
}

// Backend representa la configuración de un servidor backend
type Backend struct {
	Servers []Server `bson:"servers" json:"servers"` // Lista de servidores
}

// Server representa un servidor backend individual
type Server struct {
	Host  string `bson:"host" json:"host"`   // Dirección del host, p. ej. IP o dominio
	Port  int    `bson:"port" json:"port"`   // Puerto
	IsSSL bool   `bson:"isSSL" json:"isSSL"` // Si SSL está habilitado
}

// IsValidWAFMode comprueba si el modo WAF es válido
func IsValidWAFMode(mode WAFMode) bool {
	return mode == WAFModeProtection || mode == WAFModeObservation
}

// DefaultWAFMode devuelve el modo WAF por defecto
func DefaultWAFMode() WAFMode {
	return WAFModeObservation
}

// GetAllWAFModes devuelve todos los modos WAF disponibles
func GetAllWAFModes() []WAFMode {
	return []WAFMode{WAFModeProtection, WAFModeObservation}
}

// NewSite crea un nuevo sitio con valores por defecto
func NewSite() *Site {
	now := time.Now()
	return &Site{
		WAFMode:      DefaultWAFMode(),
		WAFEnabled:   false,
		EnableHTTPS:  false,
		ActiveStatus: true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Backend: Backend{
			Servers: make([]Server, 0),
		},
	}
}

// ValidateSite valida la validez de la configuración del sitio
func ValidateSite(site *Site) error {
	if !IsValidWAFMode(site.WAFMode) {
		site.WAFMode = DefaultWAFMode()
	}
	return nil
}

// WAFModeFromString convierte una cadena a WAFMode
func WAFModeFromString(s string) WAFMode {
	mode := WAFMode(s)
	if !IsValidWAFMode(mode) {
		return DefaultWAFMode()
	}
	return mode
}

// GetCollectionName devuelve el nombre de la colección
func (r *Site) GetCollectionName() string {
	return "site"
}
