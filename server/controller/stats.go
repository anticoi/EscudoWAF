package controller

import (
	"fmt"
	"strconv"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type StatsController interface {
	GetStats(ctx *gin.Context)
	GetOverviewStats(ctx *gin.Context)
	GetRealtimeQPS(ctx *gin.Context)
	GetTimeSeriesData(ctx *gin.Context)
	GetCombinedTimeSeriesData(ctx *gin.Context)
	GetTrafficTimeSeriesData(ctx *gin.Context)
}

type StatsControllerImpl struct {
	runnerService service.RunnerService
	statsService  service.StatsService
	logger        zerolog.Logger
}

func NewStatsController(runnerService service.RunnerService, statsService service.StatsService) StatsController {
	logger := config.GetControllerLogger("stats")
	return &StatsControllerImpl{
		runnerService: runnerService,
		statsService:  statsService,
		logger:        logger,
	}
}

// GetStats obtiene los datos estadísticos originales
//
//	@Summary		Obtener datos estadísticos originales de HAProxy
//	@Description	Obtener estadísticas originales de HAProxy
//	@Tags			Información estadística
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse			"Datos estadísticos obtenidos con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/stats [get]
func (c *StatsControllerImpl) GetStats(ctx *gin.Context) {
	stats, err := c.runnerService.GetStats()
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al obtener las estadísticas de HAProxy")
		response.InternalServerError(ctx, err, false)
		return
	}
	response.Success(ctx, "Datos estadísticos obtenidos con éxito", stats)
}

// GetOverviewStats Obtener datos de resumen estadístico
//
//	@Summary		Obtener datos de resumen estadístico
//	@Description	Obtener datos de resumen estadístico en un rango de tiempo: peticiones, tráfico, tasa de errores, etc.
//	@Tags			Información estadística
//	@Produce		json
//	@Param			timeRange	query	string	true	"Rango de tiempo: 24h (24 horas), 7d (7 días), 30d (30 días)"	Enums(24h, 7d, 30d)	default(24h)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.OverviewStats}	"Resumen estadístico obtenido con éxito"
//	@Failure		400	{object}	model.ErrResponse								"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"Error interno del servidor"
//	@Router			/api/v1/stats/overview [get]
func (c *StatsControllerImpl) GetOverviewStats(ctx *gin.Context) {
	// Analizar los parámetros de la petición
	var req dto.StatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición de resumen estadístico")
		response.BadRequest(ctx, err, true)
		return
	}

	// Llamar al servicio
	stats, err := c.statsService.GetOverviewStats(ctx, req.TimeRange)
	if err != nil {
		c.logger.Error().Err(err).Str("timeRange", req.TimeRange).Msg("Error al obtener los datos de resumen estadístico")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Resumen estadístico obtenido con éxito", stats)
}

// GetRealtimeQPS Obtener datos QPS en tiempo real
//
//	@Summary		Obtener datos QPS en tiempo real
//	@Description	Obtener los puntos de datos QPS en tiempo real más recientes
//	@Tags			Información estadística
//	@Produce		json
//	@Param			limit	query	int	false	"Cantidad de puntos de datos devueltos, por defecto 30, máximo 60"	default(30)	minimum(1)	maximum(60)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.RealtimeQPSResponse}	"Datos QPS en tiempo real obtenidos con éxito"
//	@Failure		400	{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/stats/realtime-qps [get]
func (c *StatsControllerImpl) GetRealtimeQPS(ctx *gin.Context) {
	// Analizar los parámetros de la petición
	limitStr := ctx.DefaultQuery("limit", "30")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 30 // 30 puntos por defecto
	}
	if limit > 240 {
		limit = 240 // Máximo 240 puntos
	}

	// Llamar al servicio
	data, err := c.statsService.GetRealtimeQPS(ctx, limit)
	if err != nil {
		c.logger.Error().Err(err).Int("limit", limit).Msg("Error al obtener los datos QPS en tiempo real")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Datos QPS en tiempo real obtenidos con éxito", data)
}

// GetTimeSeriesData Obtener datos de series temporales
//
//	@Summary		Obtener datos de series temporales
//	@Description	Obtener datos de series temporales por rango de tiempo y tipo de métrica para gráficos
//	@Tags			Información estadística
//	@Produce		json
//	@Param			timeRange	query	string	true	"Rango de tiempo: 24h (24 horas), 7d (7 días), 30d (30 días)"	Enums(24h, 7d, 30d)		default(24h)
//	@Param			metric		query	string	true	"Tipo de métrica: requests (peticiones), blocks (bloqueos)"	Enums(requests, blocks)	default(requests)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.TimeSeriesResponse}	"Datos de series temporales obtenidos con éxito"
//	@Failure		400	{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/stats/time-series [get]
func (c *StatsControllerImpl) GetTimeSeriesData(ctx *gin.Context) {
	// Analizar los parámetros de la petición
	var req dto.TimeSeriesDataRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición de series temporales")
		response.BadRequest(ctx, err, true)
		return
	}

	// Comprobar los parámetros de la petición
	if req.TimeRange == "" {
		req.TimeRange = dto.TimeRange24Hours // 24 horas por defecto
	}

	if req.Metric == "" {
		req.Metric = "requests" // Número de peticiones por defecto
	}

	// Llamar al servicio
	data, err := c.statsService.GetTimeSeriesData(ctx, req.TimeRange, req.Metric)
	if err != nil {
		c.logger.Error().Err(err).
			Str("timeRange", req.TimeRange).
			Str("metric", req.Metric).
			Msg("Error al obtener los datos de series temporales")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Datos de series temporales obtenidos con éxito", data)
}

// GetCombinedTimeSeriesData Obtener datos de series temporales combinadas
//
//	@Summary		Obtener datos de series temporales combinadas
//	@Description	Obtener series temporales de peticiones y bloqueos para mostrar en gráficos
//	@Tags			Información estadística
//	@Produce		json
//	@Param			timeRange	query	string	true	"Rango de tiempo: 24h (24 horas), 7d (7 días), 30d (30 días)"	Enums(24h, 7d, 30d)	default(24h)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.CombinedTimeSeriesResponse}	"Datos de series temporales combinadas obtenidos con éxito"
//	@Failure		400	{object}	model.ErrResponse											"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError								"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError								"Error interno del servidor"
//	@Router			/api/v1/stats/combined-time-series [get]
func (c *StatsControllerImpl) GetCombinedTimeSeriesData(ctx *gin.Context) {
	// Analizar los parámetros de la petición
	timeRange := ctx.DefaultQuery("timeRange", dto.TimeRange24Hours)

	// Validar los parámetros de rango de tiempo
	if timeRange != dto.TimeRange24Hours && timeRange != dto.TimeRange7Days && timeRange != dto.TimeRange30Days {
		response.BadRequest(ctx, fmt.Errorf("Rango de tiempo no válido: %s", timeRange), true)
		return
	}

	// Llamar al servicio
	data, err := c.statsService.GetCombinedTimeSeriesData(ctx, timeRange)
	if err != nil {
		c.logger.Error().Err(err).
			Str("timeRange", timeRange).
			Msg("Error al obtener los datos de series temporales combinadas")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Datos de series temporales combinadas obtenidos con éxito", data)
}

// GetTrafficTimeSeriesData Obtener datos de series temporales de tráfico
//
//	@Summary		Obtener datos de series temporales de tráfico
//	@Description	Obtener series temporales de tráfico entrante y saliente en un rango de tiempo para gráficos
//	@Tags			Información estadística
//	@Produce		json
//	@Param			timeRange	query	string	true	"Rango de tiempo: 24h (24 horas), 7d (7 días), 30d (30 días)"	Enums(24h, 7d, 30d)	default(24h)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.TrafficTimeSeriesResponse}	"Datos de series temporales de tráfico obtenidos con éxito"
//	@Failure		400	{object}	model.ErrResponse											"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError								"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError								"Error interno del servidor"
//	@Router			/api/v1/stats/traffic-time-series [get]
func (c *StatsControllerImpl) GetTrafficTimeSeriesData(ctx *gin.Context) {
	// Analizar los parámetros de la petición
	var req dto.TrafficTimeSeriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición de series temporales de tráfico")
		response.BadRequest(ctx, err, true)
		return
	}

	// Comprobar los parámetros de la petición
	if req.TimeRange == "" {
		req.TimeRange = dto.TimeRange24Hours // 24 horas por defecto
	}

	// Llamar al servicio
	data, err := c.statsService.GetTrafficTimeSeriesData(ctx, req.TimeRange)
	if err != nil {
		c.logger.Error().Err(err).
			Str("timeRange", req.TimeRange).
			Msg("Error al obtener los datos de series temporales de tráfico")
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Datos de series temporales de tráfico obtenidos con éxito", data)
}
