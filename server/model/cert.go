package model

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// CertificateStore representa la tabla de almacén de certificados
type CertificateStore struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // ID del certificado
	Name        string        `bson:"name" json:"name"`                  // Nombre/alias del certificado
	Description string        `bson:"description" json:"description"`    // Descripción del certificado
	PublicKey   string        `bson:"publicKey" json:"publicKey"`        // Contenido de la clave pública (formato PEM)
	PrivateKey  string        `bson:"privateKey" json:"privateKey"`      // Contenido de la clave privada (formato PEM)
	ExpireDate  time.Time     `bson:"expireDate" json:"expireDate"`      // Fecha de expiración del certificado
	IssuerName  string        `bson:"issuerName" json:"issuerName"`      // Autoridad emisora
	FingerPrint string        `bson:"fingerPrint" json:"fingerPrint"`    // Huella del certificado
	Domains     []string      `bson:"domains" json:"domains"`            // Lista de dominios vinculados al certificado
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`        // Fecha de creación
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`        // Hora de actualización
}

// GetCollectionName devuelve el nombre de la colección
func (c *CertificateStore) GetCollectionName() string {
	return "certificate"
}

// NewCertificateStore crea un nuevo certificado con valores por defecto
func NewCertificateStore() *CertificateStore {
	now := time.Now()
	return &CertificateStore{
		Domains:   make([]string, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ValidateCertificateStore valida la validez de la configuración del certificado
func ValidateCertificateStore(cert *CertificateStore) error {
	// Solo comprobar los campos obligatorios
	if cert.PublicKey == "" || cert.PrivateKey == "" {
		return ErrMissingRequiredField
	}
	return nil
}

// Error general
var (
	ErrMissingRequiredField = errors.New("Faltan campos obligatorios")
	ErrInvalidFormat        = errors.New("Formato no válido")
)
