package dto

import (
	"time"

	"github.com/anticoi/EscudoWAF/server/model"
)

// CertificateCreateRequest petición de creación de certificado
// @Description Parámetros de la petición para crear un certificado
type CertificateCreateRequest struct {
	Name        string    `json:"name" example:"example-cert"`               // Nombre/alias del certificado
	Description string    `json:"description" example:"Certificado para example.com"`    // Descripción del certificado
	PublicKey   string    `json:"publicKey" binding:"required"`              // Contenido de la clave pública (formato PEM)
	PrivateKey  string    `json:"privateKey" binding:"required"`             // Contenido de la clave privada (formato PEM)
	ExpireDate  time.Time `json:"expireDate" example:"2023-12-31T23:59:59Z"` // Fecha de expiración del certificado
	IssuerName  string    `json:"issuerName" example:"Let's Encrypt"`        // Autoridad emisora
	FingerPrint string    `json:"fingerPrint" example:"AA:BB:CC:DD:..."`     // Huella del certificado
	Domains     []string  `json:"domains" example:"[\"example.com\"]"`       // Lista de dominios vinculados al certificado
}

// CertificateUpdateRequest petición de actualización de certificado
// @Description Parámetros de la petición para actualizar un certificado
type CertificateUpdateRequest struct {
	Name        string    `json:"name,omitempty" example:"example-cert"`               // Nombre/alias del certificado
	Description string    `json:"description,omitempty" example:"Certificado para example.com"`    // Descripción del certificado
	PublicKey   string    `json:"publicKey,omitempty"`                                 // Contenido de la clave pública (formato PEM)
	PrivateKey  string    `json:"privateKey,omitempty"`                                // Contenido de la clave privada (formato PEM)
	ExpireDate  time.Time `json:"expireDate,omitempty" example:"2023-12-31T23:59:59Z"` // Fecha de expiración del certificado
	IssuerName  string    `json:"issuerName,omitempty" example:"Let's Encrypt"`        // Autoridad emisora
	FingerPrint string    `json:"fingerPrint,omitempty" example:"AA:BB:CC:DD:..."`     // Huella del certificado
	Domains     []string  `json:"domains,omitempty" example:"[\"example.com\"]"`       // Lista de dominios vinculados al certificado
}

// CertificateListResponse Respuesta de lista de certificados
// @Description Respuesta de lista de certificados
type CertificateListResponse struct {
	Total int64                    `json:"total"` // Total
	Items []model.CertificateStore `json:"items"` // Lista de certificados
}
