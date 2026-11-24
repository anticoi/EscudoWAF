package service

import (
	"context"
	"strconv"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type SiteService interface {
	CreateSite(ctx context.Context, req *dto.CreateSiteRequest) (*model.Site, error)
	GetSites(ctx context.Context, pageStr, sizeStr string) ([]model.Site, int64, error)
	GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error)
	UpdateSite(ctx context.Context, id bson.ObjectID, req *dto.UpdateSiteRequest) (*model.Site, error)
	DeleteSite(ctx context.Context, id bson.ObjectID) error
}

// SiteService servicio de sitios
type SiteServiceImpl struct {
	siteRepo repository.SiteRepository
	logger   zerolog.Logger
}

// NewSiteService crea el servicio de sitios
func NewSiteService(siteRepo repository.SiteRepository) SiteService {
	logger := config.GetServiceLogger("site")
	return &SiteServiceImpl{
		siteRepo: siteRepo,
		logger:   logger,
	}
}

// CreateSite crea un sitio
func (s *SiteServiceImpl) CreateSite(ctx context.Context, req *dto.CreateSiteRequest) (*model.Site, error) {
	// Crear nuevo sitio
	site := model.NewSite()
	site.Name = req.Name
	site.Domain = req.Domain
	site.ListenPort = req.ListenPort
	site.EnableHTTPS = req.EnableHTTPS
	site.WAFEnabled = req.WAFEnabled
	site.WAFMode = model.WAFModeFromString(req.WAFMode)
	site.ActiveStatus = req.ActiveStatus
	// Establecer el servidor backend
	site.Backend.Servers = make([]model.Server, len(req.Backend.Servers))
	for i, server := range req.Backend.Servers {
		site.Backend.Servers[i] = model.Server{
			Host: server.Host,
			Port: server.Port,
		}
	}

	// Si HTTPS está habilitado, configurar la información del certificado
	if req.EnableHTTPS && req.Certificate != nil {
		site.Certificate = model.Certificate{
			CertName:    req.Certificate.CertName,
			PublicKey:   req.Certificate.PublicKey,
			PrivateKey:  req.Certificate.PrivateKey,
			ExpireDate:  req.Certificate.ExpireDate,
			IssuerName:  req.Certificate.IssuerName,
			FingerPrint: req.Certificate.FingerPrint,
		}
	}

	// Validar la configuración del sitio
	if err := model.ValidateSite(site); err != nil {
		s.logger.Error().Err(err).Msg("Error de validación del sitio")
		return nil, err
	}

	// Comprobar si el dominio y puerto ya existen
	err := s.siteRepo.CheckDomainPortExists(ctx, site)
	if err != nil {
		return nil, err
	}

	// Guardar sitio
	err = s.siteRepo.CreateSite(ctx, site)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al crear el sitio")
		return nil, err
	}

	s.logger.Info().Str("name", site.Name).Str("domain", site.Domain).Msg("Sitio creado con éxito")
	return site, nil
}

// GetSites Obtener lista de sitios
func (s *SiteServiceImpl) GetSites(ctx context.Context, pageStr, sizeStr string) ([]model.Site, int64, error) {
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 10
	}

	sites, total, err := s.siteRepo.GetSites(ctx, page, size)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al obtener la lista de sitios")
		return nil, 0, err
	}

	return sites, total, nil
}

// GetSiteByID obtiene un sitio por ID
func (s *SiteServiceImpl) GetSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error) {
	site, err := s.siteRepo.GetSiteByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return site, nil
}

// UpdateSite Actualizar sitio
func (s *SiteServiceImpl) UpdateSite(ctx context.Context, id bson.ObjectID, req *dto.UpdateSiteRequest) (*model.Site, error) {

	err := s.siteRepo.CheckDomainPortConflict(ctx, &model.Site{
		ID:         id,
		Domain:     req.Domain,
		ListenPort: req.ListenPort,
	})

	if err != nil {
		return nil, err
	}

	// Obtener el sitio existente
	site, err := s.siteRepo.GetSiteByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Información de actualización del sitio
	if req.Name != "" {
		site.Name = req.Name
	}
	if req.Domain != "" {
		site.Domain = req.Domain
	}
	if req.ListenPort != 0 {
		site.ListenPort = req.ListenPort
	}

	// Actualizar la configuración de HTTPS
	site.EnableHTTPS = req.EnableHTTPS
	site.WAFEnabled = req.WAFEnabled
	if req.WAFMode != "" {
		site.WAFMode = model.WAFModeFromString(req.WAFMode)
	}
	site.ActiveStatus = req.ActiveStatus

	// Actualizar el servidor backend
	if req.Backend != nil && len(req.Backend.Servers) > 0 {
		site.Backend.Servers = make([]model.Server, len(req.Backend.Servers))
		for i, server := range req.Backend.Servers {
			site.Backend.Servers[i] = model.Server{
				Host:  server.Host,
				Port:  server.Port,
				IsSSL: server.IsSSL,
			}
		}
	}

	// Información de actualización del certificado
	if req.EnableHTTPS && req.Certificate != nil {
		site.Certificate = model.Certificate{
			CertName:    req.Certificate.CertName,
			PublicKey:   req.Certificate.PublicKey,
			PrivateKey:  req.Certificate.PrivateKey,
			ExpireDate:  req.Certificate.ExpireDate,
			IssuerName:  req.Certificate.IssuerName,
			FingerPrint: req.Certificate.FingerPrint,
		}
	}

	// Validar la configuración del sitio
	if err := model.ValidateSite(site); err != nil {
		s.logger.Error().Err(err).Msg("Error de validación del sitio")
		return nil, err
	}

	// Guardar actualización
	err = s.siteRepo.UpdateSite(ctx, site)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al actualizar el sitio")
		return nil, err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", site.Name).Msg("Sitio actualizado con éxito")
	return site, nil
}

// DeleteSite Eliminar sitio
func (s *SiteServiceImpl) DeleteSite(ctx context.Context, id bson.ObjectID) error {
	// Comprobar si el sitio existe
	site, err := s.siteRepo.GetSiteByID(ctx, id)
	if err != nil {
		return err
	}

	// Eliminar sitio
	err = s.siteRepo.DeleteSite(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar el sitio")
		return err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", site.Name).Msg("Sitio eliminado con éxito")
	return nil
}
