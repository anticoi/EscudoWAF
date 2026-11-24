package controller

import (
	"errors"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// BlockedIPController interfaz del controlador de IPs bloqueadas
type BlockedIPController interface {
	GetBlockedIPs(ctx *gin.Context)
	GetBlockedIPStats(ctx *gin.Context)
	CleanupExpiredBlockedIPs(ctx *gin.Context)
}

// BlockedIPControllerImpl implementación del controlador de IPs bloqueadas
type BlockedIPControllerImpl struct {
	blockedIPService service.BlockedIPService
	logger           zerolog.Logger
}

// NewBlockedIPController crea el controlador de IPs bloqueadas
func NewBlockedIPController(blockedIPService service.BlockedIPService) BlockedIPController {
	logger := config.GetControllerLogger("blocked_ip")
	return &BlockedIPControllerImpl{
		blockedIPService: blockedIPService,
		logger:           logger,
	}
}

// GetBlockedIPs Obtener lista de IPs bloqueadas
//
//	@Summary		Obtener lista de IPs bloqueadas
//	@Description	Obtener la lista de IPs bloqueadas, con paginación, filtrado y ordenación
//	@Tags			Gestión de IPs bloqueadas
//	@Produce		json
//	@Param			page	query	int		false	"Número de página, comenzando en 1"								default(1)	minimum(1)
//	@Param			size	query	int		false	"Cantidad por página, máximo 100"							default(10)	minimum(1)	maximum(100)
//	@Param			ip		query	string	false	"Filtrar por dirección IP, admite coincidencia difusa"							example(192.168.1.1)
//	@Param			reason	query	string	false	"Filtrar por motivo de bloqueo"								example(high_frequency_attack)
//	@Param			status	query	string	false	"Filtro de estado: active - activo, expired - expirado, all - todos"	default(all)		Enums(active, expired, all)
//	@Param			sortBy	query	string	false	"Campo de ordenación"									default(blocked_at)	Enums(blocked_at, blocked_until, ip)
//	@Param			sortDir	query	string	false	"Dirección de ordenación: asc - ascendente, desc - descendente"					default(desc)		Enums(asc, desc)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.BlockedIPListResponse}	"Lista de IPs bloqueadas obtenida con éxito"
//	@Failure		400	{object}	model.ErrResponse										"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError							"Acceso prohibido"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"Error interno del servidor"
//	@Router			/api/v1/blocked-ips [get]
func (c *BlockedIPControllerImpl) GetBlockedIPs(ctx *gin.Context) {
	var req dto.BlockedIPListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().
		Int("page", req.Page).
		Int("size", req.Size).
		Str("ip", req.IP).
		Str("reason", req.Reason).
		Str("status", req.Status).
		Str("sortBy", req.SortBy).
		Str("sortDir", req.SortDir).
		Msg("Petición de lista de IPs bloqueadas")

	result, err := c.blockedIPService.GetBlockedIPs(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPageSize) {
			response.BadRequest(ctx, err, true)
			return
		}
		c.logger.Error().Err(err).Msg("Error al obtener la lista de IPs bloqueadas")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().
		Int64("total", result.Total).
		Int("items_count", len(result.Items)).
		Int("page", result.Page).
		Int("pages", result.Pages).
		Msg("Lista de IPs bloqueadas obtenida con éxito")

	response.Success(ctx, "Lista de IPs bloqueadas obtenida con éxito", result)
}

// GetBlockedIPStats Obtener estadísticas de IPs bloqueadas
//
//	@Summary		Obtener estadísticas de IPs bloqueadas
//	@Description	Obtener estadísticas de IPs bloqueadas: total, activas, expiradas, por motivo y por hora
//	@Tags			Gestión de IPs bloqueadas
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.BlockedIPStatsResponse}	"Información estadística obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError							"Acceso prohibido"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"Error interno del servidor"
//	@Router			/api/v1/blocked-ips/stats [get]
func (c *BlockedIPControllerImpl) GetBlockedIPStats(ctx *gin.Context) {
	c.logger.Info().Msg("Petición de estadísticas de IPs bloqueadas")

	stats, err := c.blockedIPService.GetBlockedIPStats(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al obtener las estadísticas de IPs bloqueadas")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().
		Int64("total_blocked", stats.TotalBlocked).
		Int64("active_blocked", stats.ActiveBlocked).
		Int64("expired_blocked", stats.ExpiredBlocked).
		Int("reason_types", len(stats.ReasonStats)).
		Int("hourly_stats", len(stats.Last24HourStats)).
		Msg("Estadísticas de IPs bloqueadas obtenidas con éxito")

	response.Success(ctx, "Información estadística obtenida con éxito", stats)
}

// CleanupExpiredBlockedIPs Limpiar registros de IPs bloqueadas expirados
//
//	@Summary		Limpiar registros de IPs bloqueadas expirados
//	@Description	Eliminar registros de IPs bloqueadas expirados para liberar espacio
//	@Tags			Gestión de IPs bloqueadas
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.BlockedIPCleanupResponse}	"Limpieza completada"
//	@Failure		401	{object}	model.ErrResponseDontShowError								"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError								"Acceso prohibido"
//	@Failure		500	{object}	model.ErrResponseDontShowError								"Error interno del servidor"
//	@Router			/api/v1/blocked-ips/cleanup [delete]
func (c *BlockedIPControllerImpl) CleanupExpiredBlockedIPs(ctx *gin.Context) {
	c.logger.Info().Msg("Petición de limpieza de registros de IPs bloqueadas expirados")

	deletedCount, err := c.blockedIPService.CleanupExpiredBlockedIPs(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al limpiar los registros de IPs bloqueadas expirados")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Int64("deleted_count", deletedCount).Msg("Registros de IPs bloqueadas expirados limpiados con éxito")

	cleanupResponse := dto.BlockedIPCleanupResponse{
		DeletedCount: deletedCount,
		Message:      "Registros de IPs bloqueadas expirados limpiados con éxito",
	}

	response.Success(ctx, "Limpieza completada", cleanupResponse)
}
