// server/service/config.go
package service

import (
	"context"
	"errors"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/rs/zerolog"
)

var (
	ErrConfigNotFound = errors.New("La configuración no existe")
)

// ConfigService interfaz del servicio de configuración
type ConfigService interface {
	GetConfig(ctx context.Context) (*model.Config, error)
	PatchConfig(ctx context.Context, req *dto.ConfigPatchRequest) (*model.Config, error)
}

// ConfigServiceImpl implementación del servicio de configuración
type ConfigServiceImpl struct {
	configRepo repository.ConfigRepository
	logger     zerolog.Logger
}

// NewConfigService crea el servicio de configuración
func NewConfigService(configRepo repository.ConfigRepository) ConfigService {
	logger := config.GetServiceLogger("config")
	return &ConfigServiceImpl{
		configRepo: configRepo,
		logger:     logger,
	}
}

// GetConfig obtiene la configuración
func (s *ConfigServiceImpl) GetConfig(ctx context.Context) (*model.Config, error) {
	cfg, err := s.configRepo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return nil, ErrConfigNotFound
		}
		s.logger.Error().Err(err).Msg("Error al obtener la configuración")
		return nil, err
	}

	return cfg, nil
}

// PatchConfig actualización parcial de configuración
func (s *ConfigServiceImpl) PatchConfig(ctx context.Context, req *dto.ConfigPatchRequest) (*model.Config, error) {
	// Obtener la configuración existente
	cfg, err := s.configRepo.GetConfig(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrConfigNotFound) {
			return nil, ErrConfigNotFound
		}
		s.logger.Error().Err(err).Msg("Error al obtener la configuración")
		return nil, err
	}

	if req.IsResponseCheck != nil {
		cfg.IsResponseCheck = *req.IsResponseCheck
	}

	if req.IsDebug != nil {
		cfg.IsDebug = *req.IsDebug
	}

	if req.IsK8s != nil {
		cfg.IsK8s = *req.IsK8s
	}

	// Actualizar la configuración de Engine
	if req.Engine != nil {
		if req.Engine.Bind != nil {
			cfg.Engine.Bind = *req.Engine.Bind
		}

		if req.Engine.UseBuiltinRules != nil {
			cfg.Engine.UseBuiltinRules = *req.Engine.UseBuiltinRules
		}

		if req.Engine.ASNDBPath != nil {
			cfg.Engine.ASNDBPath = *req.Engine.ASNDBPath
		}

		if req.Engine.CityDBPath != nil {
			cfg.Engine.CityDBPath = *req.Engine.CityDBPath
		}

		// Actualizar AppConfig
		if len(req.Engine.AppConfig) > 0 {
			for _, reqApp := range req.Engine.AppConfig {
				// Encontrar la AppConfig correspondiente y actualizarla
				for i, app := range cfg.Engine.AppConfig {
					if reqApp.Name != nil && app.Name == *reqApp.Name {
						// Actualizar campos no vacíos
						if reqApp.Directives != nil {
							cfg.Engine.AppConfig[i].Directives = *reqApp.Directives
						}
						if reqApp.TransactionTTL != nil {
							cfg.Engine.AppConfig[i].TransactionTTL = dto.MillisToDuration(*reqApp.TransactionTTL)
						}
						if reqApp.LogLevel != nil {
							cfg.Engine.AppConfig[i].LogLevel = *reqApp.LogLevel
						}
						if reqApp.LogFile != nil {
							cfg.Engine.AppConfig[i].LogFile = *reqApp.LogFile
						}
						if reqApp.LogFormat != nil {
							cfg.Engine.AppConfig[i].LogFormat = *reqApp.LogFormat
						}
						break
					}
				}
			}
		}

		// Actualizar la configuración de FlowController
		if req.Engine.FlowController != nil {
			// Actualizar la configuración de VisitLimit
			if req.Engine.FlowController.VisitLimit != nil {
				visitLimit := req.Engine.FlowController.VisitLimit
				if visitLimit.Enabled != nil {
					cfg.Engine.FlowController.VisitLimit.Enabled = *visitLimit.Enabled
				}
				if visitLimit.Threshold != nil {
					cfg.Engine.FlowController.VisitLimit.Threshold = *visitLimit.Threshold
				}
				if visitLimit.StatDuration != nil {
					cfg.Engine.FlowController.VisitLimit.StatDuration = *visitLimit.StatDuration
				}
				if visitLimit.BlockDuration != nil {
					cfg.Engine.FlowController.VisitLimit.BlockDuration = *visitLimit.BlockDuration
				}
				if visitLimit.BurstCount != nil {
					cfg.Engine.FlowController.VisitLimit.BurstCount = *visitLimit.BurstCount
				}
				if visitLimit.ParamsCapacity != nil {
					cfg.Engine.FlowController.VisitLimit.ParamsCapacity = *visitLimit.ParamsCapacity
				}
			}

			// Actualizar la configuración de AttackLimit
			if req.Engine.FlowController.AttackLimit != nil {
				attackLimit := req.Engine.FlowController.AttackLimit
				if attackLimit.Enabled != nil {
					cfg.Engine.FlowController.AttackLimit.Enabled = *attackLimit.Enabled
				}
				if attackLimit.Threshold != nil {
					cfg.Engine.FlowController.AttackLimit.Threshold = *attackLimit.Threshold
				}
				if attackLimit.StatDuration != nil {
					cfg.Engine.FlowController.AttackLimit.StatDuration = *attackLimit.StatDuration
				}
				if attackLimit.BlockDuration != nil {
					cfg.Engine.FlowController.AttackLimit.BlockDuration = *attackLimit.BlockDuration
				}
				if attackLimit.BurstCount != nil {
					cfg.Engine.FlowController.AttackLimit.BurstCount = *attackLimit.BurstCount
				}
				if attackLimit.ParamsCapacity != nil {
					cfg.Engine.FlowController.AttackLimit.ParamsCapacity = *attackLimit.ParamsCapacity
				}
			}

			// Actualizar la configuración de ErrorLimit
			if req.Engine.FlowController.ErrorLimit != nil {
				errorLimit := req.Engine.FlowController.ErrorLimit
				if errorLimit.Enabled != nil {
					cfg.Engine.FlowController.ErrorLimit.Enabled = *errorLimit.Enabled
				}
				if errorLimit.Threshold != nil {
					cfg.Engine.FlowController.ErrorLimit.Threshold = *errorLimit.Threshold
				}
				if errorLimit.StatDuration != nil {
					cfg.Engine.FlowController.ErrorLimit.StatDuration = *errorLimit.StatDuration
				}
				if errorLimit.BlockDuration != nil {
					cfg.Engine.FlowController.ErrorLimit.BlockDuration = *errorLimit.BlockDuration
				}
				if errorLimit.BurstCount != nil {
					cfg.Engine.FlowController.ErrorLimit.BurstCount = *errorLimit.BurstCount
				}
				if errorLimit.ParamsCapacity != nil {
					cfg.Engine.FlowController.ErrorLimit.ParamsCapacity = *errorLimit.ParamsCapacity
				}
			}
		}
	}

	// Actualizar la configuración de Haproxy
	if req.Haproxy != nil {
		if req.Haproxy.ConfigBaseDir != nil {
			cfg.Haproxy.ConfigBaseDir = *req.Haproxy.ConfigBaseDir
		}
		if req.Haproxy.HaproxyBin != nil {
			cfg.Haproxy.HaproxyBin = *req.Haproxy.HaproxyBin
		}
		if req.Haproxy.BackupsNumber != nil {
			cfg.Haproxy.BackupsNumber = *req.Haproxy.BackupsNumber
		}
		if req.Haproxy.SpoeAgentAddr != nil {
			cfg.Haproxy.SpoeAgentAddr = *req.Haproxy.SpoeAgentAddr
		}
		if req.Haproxy.SpoeAgentPort != nil {
			cfg.Haproxy.SpoeAgentPort = *req.Haproxy.SpoeAgentPort
		}
		if req.Haproxy.Thread != nil {
			cfg.Haproxy.Thread = *req.Haproxy.Thread
		}
	}

	// Guardar actualización
	err = s.configRepo.UpdateConfig(ctx, cfg)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al actualizar la configuración")
		return nil, err
	}

	s.logger.Info().Str("name", cfg.Name).Msg("Configuración actualizada con éxito")
	return cfg, nil
}
