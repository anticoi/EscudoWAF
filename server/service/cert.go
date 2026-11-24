package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrCertificateNotFound   = errors.New("El certificado no existe")
	ErrCertificateNameExists = errors.New("El nombre del certificado ya existe")
	ErrInvalidCertificate    = errors.New("Formato de certificado no válido")
)

// CertificateService interfaz del servicio de certificados
type CertificateService interface {
	CreateCertificate(ctx context.Context, req *dto.CertificateCreateRequest) (*model.CertificateStore, error)
	GetCertificates(ctx context.Context, pageStr, sizeStr string) ([]model.CertificateStore, int64, error)
	GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error)
	UpdateCertificate(ctx context.Context, id bson.ObjectID, req *dto.CertificateUpdateRequest) (*model.CertificateStore, error)
	DeleteCertificate(ctx context.Context, id bson.ObjectID) error
}

// CertificateServiceImpl implementación del servicio de certificados
type CertificateServiceImpl struct {
	certRepo repository.CertificateRepository
	logger   zerolog.Logger
}

// NewCertificateService crea el servicio de certificados
func NewCertificateService(certRepo repository.CertificateRepository) CertificateService {
	logger := config.GetServiceLogger("certificate")
	return &CertificateServiceImpl{
		certRepo: certRepo,
		logger:   logger,
	}
}

// CreateCertificate crea un certificado
func (s *CertificateServiceImpl) CreateCertificate(ctx context.Context, req *dto.CertificateCreateRequest) (*model.CertificateStore, error) {
	// Comprobar si el nombre del certificado ya existe
	if req.Name != "" {
		exists, err := s.certRepo.CheckCertificateNameExists(ctx, req.Name, bson.NilObjectID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCertificateNameExists
		}
	}

	// Crear nuevo certificado
	cert := model.NewCertificateStore()
	cert.Name = req.Name
	cert.Description = req.Description
	cert.PublicKey = req.PublicKey
	cert.PrivateKey = req.PrivateKey
	cert.ExpireDate = req.ExpireDate
	cert.IssuerName = req.IssuerName
	cert.FingerPrint = req.FingerPrint
	cert.Domains = req.Domains

	// Validar el certificado
	if err := model.ValidateCertificateStore(cert); err != nil {
		s.logger.Error().Err(err).Msg("Error de validación del certificado")
		return nil, ErrInvalidCertificate
	}

	// Guardar certificado
	err := s.certRepo.CreateCertificate(ctx, cert)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al crear el certificado")
		return nil, err
	}

	s.logger.Info().Str("id", cert.ID.Hex()).Str("name", cert.Name).Msg("Certificado creado con éxito")
	return cert, nil
}

// GetCertificates Obtener lista de certificados
func (s *CertificateServiceImpl) GetCertificates(ctx context.Context, pageStr, sizeStr string) ([]model.CertificateStore, int64, error) {
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 10
	}

	certificates, total, err := s.certRepo.GetCertificates(ctx, page, size)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al obtener la lista de certificados")
		return nil, 0, err
	}

	// Convertir a DTO
	responses := make([]model.CertificateStore, len(certificates))
	for i, cert := range certificates {
		responses[i] = model.CertificateStore{
			ID:          cert.ID,
			Name:        cert.Name,
			Description: cert.Description,
			PublicKey:   cert.PublicKey,
			PrivateKey:  cert.PrivateKey,
			ExpireDate:  cert.ExpireDate,
			IssuerName:  cert.IssuerName,
			FingerPrint: cert.FingerPrint,
			Domains:     cert.Domains,
			CreatedAt:   cert.CreatedAt,
			UpdatedAt:   cert.UpdatedAt,
		}
	}

	return responses, total, nil
}

// GetCertificateByID obtiene un certificado por ID
func (s *CertificateServiceImpl) GetCertificateByID(ctx context.Context, id bson.ObjectID) (*model.CertificateStore, error) {
	cert, err := s.certRepo.GetCertificateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCertNotFound) {
			return nil, ErrCertificateNotFound
		}
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al obtener el certificado")
		return nil, err
	}

	return cert, nil
}

// UpdateCertificate Actualizar certificado
func (s *CertificateServiceImpl) UpdateCertificate(ctx context.Context, id bson.ObjectID, req *dto.CertificateUpdateRequest) (*model.CertificateStore, error) {
	// Obtener el certificado existente
	cert, err := s.certRepo.GetCertificateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCertNotFound) {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}

	// Comprobar si el nombre del certificado ya existe (si se va a actualizar el nombre)
	if req.Name != "" && req.Name != cert.Name {
		exists, err := s.certRepo.CheckCertificateNameExists(ctx, req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCertificateNameExists
		}
		cert.Name = req.Name
	}

	// Actualizar los demás campos (solo campos no vacíos)
	if req.Description != "" {
		cert.Description = req.Description
	}
	if req.PublicKey != "" {
		cert.PublicKey = req.PublicKey
	}
	if req.PrivateKey != "" {
		cert.PrivateKey = req.PrivateKey
	}
	if !req.ExpireDate.IsZero() {
		cert.ExpireDate = req.ExpireDate
	}
	if req.IssuerName != "" {
		cert.IssuerName = req.IssuerName
	}
	if req.FingerPrint != "" {
		cert.FingerPrint = req.FingerPrint
	}
	if req.Domains != nil {
		cert.Domains = req.Domains
	}

	// Validar el certificado
	if err := model.ValidateCertificateStore(cert); err != nil {
		s.logger.Error().Err(err).Msg("Error de validación del certificado")
		return nil, ErrInvalidCertificate
	}

	// Guardar actualización
	err = s.certRepo.UpdateCertificate(ctx, cert)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al actualizar el certificado")
		return nil, err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", cert.Name).Msg("Certificado actualizado con éxito")
	return cert, nil
}

// DeleteCertificate Eliminar certificado
func (s *CertificateServiceImpl) DeleteCertificate(ctx context.Context, id bson.ObjectID) error {
	// Comprobar si el certificado existe
	_, err := s.certRepo.GetCertificateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCertNotFound) {
			return ErrCertificateNotFound
		}
		return err
	}

	// Eliminar certificado
	err = s.certRepo.DeleteCertificate(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar el certificado")
		return err
	}

	s.logger.Info().Str("id", id.Hex()).Msg("Certificado eliminado con éxito")
	return nil
}
