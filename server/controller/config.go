// server/controller/config.go
package controller

import (
	"errors"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ConfigController interfaz del controlador de configuración
type ConfigController interface {
	GetConfig(ctx *gin.Context)
	PatchConfig(ctx *gin.Context)
}

// ConfigControllerImpl implementación del controlador de configuración
type ConfigControllerImpl struct {
	configService service.ConfigService
	logger        zerolog.Logger
}

// NewConfigController crea el controlador de configuración
func NewConfigController(configService service.ConfigService) ConfigController {
	logger := config.GetControllerLogger("config")
	return &ConfigControllerImpl{
		configService: configService,
		logger:        logger,
	}
}

// GetConfig obtiene la configuración
//
//	@Summary		Obtener la configuración del sistema
//	@Description	Obtener la configuración actual del sistema
//	@Tags			Gestión de configuración
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.ConfigResponse}	"Configuración obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError					"Acceso prohibido"
//	@Failure		404	{object}	model.ErrResponseDontShowError					"La configuración no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"Error interno del servidor"
//	@Router			/api/v1/config [get]
func (c *ConfigControllerImpl) GetConfig(ctx *gin.Context) {
	cfg, err := c.configService.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Msg("Error al obtener la configuración")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a respuesta DTO
	configResponse := mapConfigToDTO(cfg)

	response.Success(ctx, "Configuración obtenida con éxito", configResponse)
}

// PatchConfig actualización parcial de configuración
//
//	@Summary		Actualizar la configuración del sistema
//	@Description	Actualizar la configuración del sistema mediante PATCH
//	@Tags			Gestión de configuración
//	@Accept			json
//	@Produce		json
//	@Param			config	body	dto.ConfigPatchRequest	true	"Información de actualización de la configuración"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.ConfigResponse}	"Configuración actualizada con éxito"
//	@Failure		400	{object}	model.ErrResponse								"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError					"Acceso prohibido"
//	@Failure		404	{object}	model.ErrResponseDontShowError					"La configuración no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"Error interno del servidor"
//	@Router			/api/v1/config [patch]
func (c *ConfigControllerImpl) PatchConfig(ctx *gin.Context) {
	var req dto.ConfigPatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	cfg, err := c.configService.PatchConfig(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Msg("Error al actualizar la configuración")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a respuesta DTO
	configResponse := mapConfigToDTO(cfg)

	response.Success(ctx, "Configuración actualizada con éxito", configResponse)
}

// mapConfigToDTO convierte el modelo a DTO
func mapConfigToDTO(cfg *model.Config) dto.ConfigResponse {
	// Convertir el modelo de configuración a DTO de respuesta
	engineDTO := dto.EngineDTO{
		Bind:            cfg.Engine.Bind,
		UseBuiltinRules: cfg.Engine.UseBuiltinRules,
		ASNDBPath:       cfg.Engine.ASNDBPath,
		CityDBPath:      cfg.Engine.CityDBPath,
		AppConfig:       make([]dto.AppConfigDTO, len(cfg.Engine.AppConfig)),
		FlowController: dto.FlowControllerDTO{
			VisitLimit: dto.LimitConfigDTO{
				Enabled:        cfg.Engine.FlowController.VisitLimit.Enabled,
				Threshold:      cfg.Engine.FlowController.VisitLimit.Threshold,
				StatDuration:   cfg.Engine.FlowController.VisitLimit.StatDuration,
				BlockDuration:  cfg.Engine.FlowController.VisitLimit.BlockDuration,
				BurstCount:     cfg.Engine.FlowController.VisitLimit.BurstCount,
				ParamsCapacity: cfg.Engine.FlowController.VisitLimit.ParamsCapacity,
			},
			AttackLimit: dto.LimitConfigDTO{
				Enabled:        cfg.Engine.FlowController.AttackLimit.Enabled,
				Threshold:      cfg.Engine.FlowController.AttackLimit.Threshold,
				StatDuration:   cfg.Engine.FlowController.AttackLimit.StatDuration,
				BlockDuration:  cfg.Engine.FlowController.AttackLimit.BlockDuration,
				BurstCount:     cfg.Engine.FlowController.AttackLimit.BurstCount,
				ParamsCapacity: cfg.Engine.FlowController.AttackLimit.ParamsCapacity,
			},
			ErrorLimit: dto.LimitConfigDTO{
				Enabled:        cfg.Engine.FlowController.ErrorLimit.Enabled,
				Threshold:      cfg.Engine.FlowController.ErrorLimit.Threshold,
				StatDuration:   cfg.Engine.FlowController.ErrorLimit.StatDuration,
				BlockDuration:  cfg.Engine.FlowController.ErrorLimit.BlockDuration,
				BurstCount:     cfg.Engine.FlowController.ErrorLimit.BurstCount,
				ParamsCapacity: cfg.Engine.FlowController.ErrorLimit.ParamsCapacity,
			},
		},
	}

	// Convertir la configuración de la aplicación
	for i, app := range cfg.Engine.AppConfig {
		engineDTO.AppConfig[i] = dto.AppConfigDTO{
			Name:           app.Name,
			Directives:     app.Directives,
			TransactionTTL: dto.DurationToMillis(app.TransactionTTL),
			LogLevel:       app.LogLevel,
			LogFile:        app.LogFile,
			LogFormat:      app.LogFormat,
		}
	}

	// Convertir la configuración de Haproxy
	haproxyDTO := dto.HaproxyDTO{
		ConfigBaseDir: cfg.Haproxy.ConfigBaseDir,
		HaproxyBin:    cfg.Haproxy.HaproxyBin,
		BackupsNumber: cfg.Haproxy.BackupsNumber,
		SpoeAgentAddr: cfg.Haproxy.SpoeAgentAddr,
		SpoeAgentPort: cfg.Haproxy.SpoeAgentPort,
		Thread:        cfg.Haproxy.Thread,
	}

	return dto.ConfigResponse{
		Name:            cfg.Name,
		Engine:          engineDTO,
		Haproxy:         haproxyDTO,
		CreatedAt:       cfg.CreatedAt,
		UpdatedAt:       cfg.UpdatedAt,
		IsResponseCheck: cfg.IsResponseCheck,
		IsDebug:         cfg.IsDebug,
		IsK8s:           cfg.IsK8s,
	}
}
