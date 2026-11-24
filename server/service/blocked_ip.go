package service

import (
	"context"
	"errors"
	"math"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/rs/zerolog"
)

var (
	ErrBlockedIPNotFound = errors.New("El registro de IP bloqueada no existe")
	ErrInvalidPageSize   = errors.New("Parámetros de paginación no válidos")
)

// BlockedIPService interfaz del servicio de IPs bloqueadas
type BlockedIPService interface {
	GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) (*dto.BlockedIPListResponse, error)
	GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error)
	CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error
	CleanupExpiredBlockedIPs(ctx context.Context) (int64, error)
}

// BlockedIPServiceImpl implementación del servicio de IPs bloqueadas
type BlockedIPServiceImpl struct {
	blockedIPRepo repository.BlockedIPRepository
	logger        zerolog.Logger
}

// NewBlockedIPService crea el servicio de IPs bloqueadas
func NewBlockedIPService(blockedIPRepo repository.BlockedIPRepository) BlockedIPService {
	logger := config.GetServiceLogger("blocked_ip")
	return &BlockedIPServiceImpl{
		blockedIPRepo: blockedIPRepo,
		logger:        logger,
	}
}

// GetBlockedIPs Obtener lista de IPs bloqueadas
func (s *BlockedIPServiceImpl) GetBlockedIPs(ctx context.Context, req *dto.BlockedIPListRequest) (*dto.BlockedIPListResponse, error) {
	// Validar los parámetros de paginación
	if err := s.validatePaginationParams(req); err != nil {
		return nil, err
	}

	// Establecer valores por defecto
	s.setDefaultParams(req)

	s.logger.Info().
		Int("page", req.Page).
		Int("size", req.Size).
		Str("ip", req.IP).
		Str("reason", req.Reason).
		Str("status", req.Status).
		Msg("Petición de lista de IPs bloqueadas")

	// Llamar a la capa de repositorio
	records, total, err := s.blockedIPRepo.GetBlockedIPs(ctx, req)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al obtener la lista de IPs bloqueadas")
		return nil, err
	}

	// Convertir a DTO de respuesta
	items := dto.MapToResponseList(records)

	// Calcular el total de páginas
	pages := int(math.Ceil(float64(total) / float64(req.Size)))
	if pages == 0 {
		pages = 1
	}

	response := &dto.BlockedIPListResponse{
		Total: total,
		Items: items,
		Page:  req.Page,
		Size:  req.Size,
		Pages: pages,
	}

	s.logger.Info().
		Int64("total", total).
		Int("pages", pages).
		Int("items_count", len(items)).
		Msg("Lista de IPs bloqueadas obtenida con éxito")

	return response, nil
}

// GetBlockedIPStats Obtener estadísticas de IPs bloqueadas
func (s *BlockedIPServiceImpl) GetBlockedIPStats(ctx context.Context) (*dto.BlockedIPStatsResponse, error) {
	s.logger.Info().Msg("Petición de estadísticas de IPs bloqueadas")

	stats, err := s.blockedIPRepo.GetBlockedIPStats(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al obtener las estadísticas de IPs bloqueadas")
		return nil, err
	}

	s.logger.Info().
		Int64("total_blocked", stats.TotalBlocked).
		Int64("active_blocked", stats.ActiveBlocked).
		Int64("expired_blocked", stats.ExpiredBlocked).
		Int("reason_types", len(stats.ReasonStats)).
		Int("hourly_stats", len(stats.Last24HourStats)).
		Msg("Estadísticas de IPs bloqueadas obtenidas con éxito")

	return stats, nil
}

// CreateBlockedIP crea un registro de IP bloqueada
func (s *BlockedIPServiceImpl) CreateBlockedIP(ctx context.Context, record *model.BlockedIPRecord) error {
	if record.IP == "" {
		return errors.New("La dirección IP no puede estar vacía")
	}

	if record.BlockedUntil.Before(record.BlockedAt) {
		return errors.New("La hora de fin del bloqueo no puede ser anterior a la de inicio")
	}

	s.logger.Info().
		Str("ip", record.IP).
		Str("reason", record.Reason).
		Str("request_uri", record.RequestUri).
		Time("blocked_at", record.BlockedAt).
		Time("blocked_until", record.BlockedUntil).
		Msg("Petición de creación de registro de IP bloqueada")

	err := s.blockedIPRepo.CreateBlockedIP(ctx, record)
	if err != nil {
		s.logger.Error().Err(err).Str("ip", record.IP).Msg("Error al crear el registro de IP bloqueada")
		return err
	}

	s.logger.Info().Str("ip", record.IP).Str("reason", record.Reason).Msg("Registro de IP bloqueada creado con éxito")
	return nil
}

// CleanupExpiredBlockedIPs Limpiar registros de IPs bloqueadas expirados
func (s *BlockedIPServiceImpl) CleanupExpiredBlockedIPs(ctx context.Context) (int64, error) {
	s.logger.Info().Msg("Iniciando limpieza de registros de IPs bloqueadas expirados")

	deletedCount, err := s.blockedIPRepo.DeleteExpiredBlockedIPs(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al limpiar los registros de IPs bloqueadas expirados")
		return 0, err
	}

	s.logger.Info().Int64("deleted_count", deletedCount).Msg("Limpieza de registros de IPs bloqueadas expirados completada")
	return deletedCount, nil
}

// validatePaginationParams valida los parámetros de paginación
func (s *BlockedIPServiceImpl) validatePaginationParams(req *dto.BlockedIPListRequest) error {
	if req.Page < 0 {
		return ErrInvalidPageSize
	}
	if req.Size < 0 || req.Size > 100 {
		return ErrInvalidPageSize
	}
	return nil
}

// setDefaultParams establece los parámetros por defecto
func (s *BlockedIPServiceImpl) setDefaultParams(req *dto.BlockedIPListRequest) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	if req.Status == "" {
		req.Status = "all"
	}
	if req.SortBy == "" {
		req.SortBy = "blocked_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}
}
