package cornjob

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/service/daemon"
	"github.com/rs/zerolog"
)

// Instancia singleton
var (
	instance *CronJobService
	mu       sync.Mutex // Mutex para doble verificación de bloqueo y otras operaciones
)

// CronJobService servicio de tareas programadas
type CronJobService struct {
	statsJob  *StatsJob          // Tarea programada de estadísticas de HAProxy
	logger    zerolog.Logger     // Logger
	ctx       context.Context    // Contexto
	cancel    context.CancelFunc // Función de cancelación del contexto
	isRunning bool               // Si está en ejecución
	mu        sync.Mutex         // Bloqueo interno de la instancia, para operaciones Start/Stop
}

// Crear nueva instancia de CronJobService (método interno)
func newCronJobService(runner daemon.ServiceRunner, targetList []string) (*CronJobService, error) {
	// Inicializar el logger
	logger := config.GetLogger().With().Str("component", "cronjob-service").Logger()

	// Crear la tarea programada de estadísticas
	statsJob, err := NewStatsJob(runner, targetList)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats job: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &CronJobService{
		statsJob:  statsJob,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: false,
	}, nil
}

// GetInstance obtiene la instancia singleton con doble verificación de bloqueo
func GetInstance(runner daemon.ServiceRunner, targetList []string) (*CronJobService, error) {
	// Primera comprobación - ruta rápida, sin bloqueo
	if instance != nil {
		return instance, nil
	}

	// Bloquear para garantizar seguridad entre hilos
	mu.Lock()
	defer mu.Unlock()

	// Segunda comprobación - verificar de nuevo dentro del bloqueo
	if instance != nil {
		return instance, nil
	}

	// Inicializar la instancia
	var err error
	instance, err = newCronJobService(runner, targetList)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize CronJobService: %w", err)
	}

	return instance, nil
}

// MustGetInstance obtiene la instancia singleton; hace panic si no existe o hay error
func MustGetInstance() *CronJobService {
	if instance == nil {
		panic("CronJobService instance not initialized, call GetInstance first")
	}
	return instance
}

// ResetInstance reinicia el singleton (principalmente para pruebas o cambios de configuración)
func ResetInstance() {
	mu.Lock()
	defer mu.Unlock()

	// Si la instancia existe y está en ejecución, detener primero el servicio
	if instance != nil {
		// No usar el método Stop de la instancia para evitar deadlocks
		// Llamar directamente a la lógica interna para detener el servicio
		if instance.isRunning {
			instance.isRunning = false

			// Crear un context con timeout
			ctx, cancel := context.WithTimeout(instance.ctx, 30*time.Second)
			defer cancel()

			// Detener tarea
			err := instance.statsJob.Stop(ctx)
			if err != nil {
				instance.logger.Error().Err(err).Msg("Failed to stop HAProxy stats job during reset")
			}

			// Cancelar el contexto
			instance.cancel()

			instance.logger.Info().Msg("CronJobService stopped during reset")
		}
	}

	// Restablecer la variable singleton
	instance = nil
}

// UpdateTargetList actualiza la lista de objetivos monitorizados
func (s *CronJobService) UpdateTargetList(targetList []string) {
	s.statsJob.UpdateTargetList(targetList)
	s.logger.Info().Strs("targets", targetList).Msg("Updated HAProxy monitoring target list")
}

// Start inicia todas las tareas programadas
func (s *CronJobService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return errors.New("service is already running")
	}

	// Iniciar la tarea programada de estadísticas de HAProxy
	if err := s.statsJob.Start(s.ctx); err != nil {
		s.logger.Error().Err(err).Msg("Failed to start HAProxy stats job")
		return fmt.Errorf("failed to start stats job: %w", err)
	}

	s.isRunning = true
	s.logger.Info().Msg("All cron jobs started")
	return nil
}

// Stop detiene todas las tareas programadas
func (s *CronJobService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil // Ya está detenido, no hay nada más que hacer
	}

	s.isRunning = false

	// Crear un context con timeout
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// Detener la tarea con un nuevo context
	err := s.statsJob.Stop(ctx)

	// Luego cancelar el context del propio servicio
	s.cancel()

	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to stop HAProxy stats job")
		return fmt.Errorf("failed to stop stats job: %w", err)
	}

	s.logger.Info().Msg("All cron jobs stopped")
	return nil
}

// IsRunning devuelve si el servicio está en ejecución
func (s *CronJobService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}
