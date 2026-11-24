package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/anticoi/EscudoWAF/server/config"
	cornjob "github.com/anticoi/EscudoWAF/server/service/cornjob/haproxy"
	"github.com/anticoi/EscudoWAF/server/service/daemon"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/rs/zerolog"
)

// Definición de errores
var (
	ErrRunnerNotRunning     = errors.New("El runner no está en ejecución")
	ErrRunnerAlreadyRunning = errors.New("El runner ya está en ejecución")
)

// RunnerService interfaz del servicio del runner
type RunnerService interface {
	// Obtener el estado del runner
	GetStatus(ctx context.Context) (daemon.ServiceState, error)

	// Operación del runner
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Restart(ctx context.Context) error
	ForceStop(ctx context.Context) error
	Reload(ctx context.Context) error
	// get haproxy stats
	GetStats() (models.NativeStats, error)
}

// RunnerServiceImpl implementación del servicio del runner
type RunnerServiceImpl struct {
	logger zerolog.Logger
	runner daemon.ServiceRunner
}

// NewRunnerService crea el servicio del runner
func NewRunnerService() (RunnerService, error) {
	logger := config.GetServiceLogger("runner")

	// Obtener el servicio ServiceRunner
	runner, err := daemon.GetRunnerService()
	if err != nil {
		logger.Error().Err(err).Msg("Error al obtener ServiceRunner")
		return nil, fmt.Errorf("Error al inicializar el servicio runner: %w", err)
	}

	return &RunnerServiceImpl{
		logger: logger,
		runner: runner,
	}, nil
}

// GetStatus obtiene el estado del runner
func (s *RunnerServiceImpl) GetStatus(ctx context.Context) (daemon.ServiceState, error) {
	return s.runner.GetState(), nil
}

// Start inicia el runner
func (s *RunnerServiceImpl) Start(ctx context.Context) error {
	// Comprobar el estado actual
	if s.runner.GetState() == daemon.ServiceRunning {
		return ErrRunnerAlreadyRunning
	}

	// Actualizar la lista de datos de métricas del servicio haproxy
	targetList, err := cornjob.GetLatestTargetList()
	if err != nil {
		return fmt.Errorf("failed to get target list: %w", err)
	}

	cronJobService, err := cornjob.GetInstance(s.runner, targetList)
	if err != nil {
		return fmt.Errorf("failed to create cron job service: %w", err)
	}
	cronJobService.UpdateTargetList(targetList)

	// Iniciar el servicio
	err = s.runner.StartServices()
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al iniciar el runner")
		return fmt.Errorf("Error al iniciar el runner: %w", err)
	}

	return nil
}

// Stop detiene el runner
func (s *RunnerServiceImpl) Stop(ctx context.Context) error {
	// Comprobar el estado actual
	if s.runner.GetState() != daemon.ServiceRunning {
		return ErrRunnerNotRunning
	}

	// Detener servicio
	err := s.runner.StopServices()
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al detener el runner")
		return fmt.Errorf("Error al detener el runner: %w", err)
	}

	return nil
}

// Restart reinicia el runner
func (s *RunnerServiceImpl) Restart(ctx context.Context) error {
	// Actualizar la lista de datos de métricas del servicio haproxy
	targetList, err := cornjob.GetLatestTargetList()
	if err != nil {
		return fmt.Errorf("failed to get target list: %w", err)
	}

	cronJobService, err := cornjob.GetInstance(s.runner, targetList)
	if err != nil {
		return fmt.Errorf("failed to create cron job service: %w", err)
	}
	cronJobService.UpdateTargetList(targetList)

	// Reiniciar el servicio
	err = s.runner.Restart()
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al reiniciar el runner")
		return fmt.Errorf("Error al reiniciar el runner: %w", err)
	}

	return nil
}

// ForceStop detiene forzosamente el runner
func (s *RunnerServiceImpl) ForceStop(ctx context.Context) error {
	// Detener forzosamente el servicio
	s.runner.ForceStop()
	return nil
}

// Reload recarga en caliente del runner
func (s *RunnerServiceImpl) Reload(ctx context.Context) error {
	// Comprobar el estado actual
	if s.runner.GetState() != daemon.ServiceRunning {
		return ErrRunnerNotRunning
	}

	// Actualizar la lista de datos de métricas del servicio haproxy
	targetList, err := cornjob.GetLatestTargetList()
	if err != nil {
		return fmt.Errorf("failed to get target list: %w", err)
	}

	cronJobService, err := cornjob.GetInstance(s.runner, targetList)
	if err != nil {
		return fmt.Errorf("failed to create cron job service: %w", err)
	}
	cronJobService.UpdateTargetList(targetList)

	// Recarga en caliente del servicio
	err = s.runner.HotReload()
	if err != nil {
		s.logger.Error().Err(err).Msg("Error en la recarga en caliente del runner")
		return fmt.Errorf("Error en la recarga en caliente del runner: %w", err)
	}

	return nil
}

func (s *RunnerServiceImpl) GetStats() (models.NativeStats, error) {
	return s.runner.GetStats()
}
