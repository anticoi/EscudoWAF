package cornjob

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/service/daemon"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
)

// StatsJob tarea programada de estadísticas de HAProxy
type StatsJob struct {
	scheduler   gocron.Scheduler
	aggregator  *StatsAggregator
	logger      zerolog.Logger
	realtimeJob gocron.Job // Tarea de estadísticas en tiempo real
	minuteJob   gocron.Job // Tarea de estadísticas por minuto
	targetList  []string   // Lista de objetivos monitorizados
	isRunning   bool       // Si está en ejecución
}

// NewStatsJob crea una nueva tarea programada de estadísticas
func NewStatsJob(runner daemon.ServiceRunner, targetList []string) (*StatsJob, error) {
	logger := config.GetLogger().With().Str("component", "cronjob-haproxy-stats-job").Logger()
	dbName := config.Global.DBConfig.Database
	// Crear el agregador de datos
	aggregator, err := NewStatsAggregator(runner, dbName, targetList)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats aggregator: %w", err)
	}

	// Crear el planificador y establecer la zona horaria
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(time.Local), // Establecer la zona horaria al crear el planificador
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &StatsJob{
		scheduler:  scheduler,
		aggregator: aggregator,
		logger:     logger,
		targetList: targetList,
		isRunning:  false,
	}, nil
}

// UpdateTargetList actualiza la lista de objetivos monitorizados
func (j *StatsJob) UpdateTargetList(targetList []string) {
	// Comprobar cambios de estado anteriores y posteriores
	hadNoTargets := len(j.targetList) == 0
	willHaveNoTargets := len(targetList) == 0

	// Guardar la cantidad de objetivos anterior para el log
	oldTargetCount := len(j.targetList)

	j.targetList = targetList
	j.aggregator.UpdateTargetList(targetList)

	// Manejar las distintas transiciones de estado
	if hadNoTargets && !willHaveNoTargets {
		// De sin objetivo a con objetivo
		j.logger.Info().Strs("targets", targetList).Msg("First targets added, activating monitoring")
	} else if !hadNoTargets && willHaveNoTargets {
		// De con objetivo a sin objetivo
		j.logger.Warn().Int("previous_count", oldTargetCount).Msg("All targets removed, monitoring entering standby mode")
	} else {
		// El resto se actualiza normalmente
		j.logger.Info().
			Int("count", len(targetList)).
			Msg("Updated monitoring target list")
	}
}

// Start inicia la tarea programada
func (j *StatsJob) Start(ctx context.Context) error {
	if j.isRunning {
		return errors.New("job is already running")
	}

	// Inicializar el agregador
	if err := j.aggregator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start aggregator: %w", err)
	}

	// Crear la tarea de estadísticas en tiempo real (cada 5 segundos)
	realtimeJob, err := j.scheduler.NewJob(
		gocron.DurationJob(
			5*time.Second,
		),
		gocron.NewTask(
			func(ctx context.Context) {
				if err := j.aggregator.CollectRealtimeMetrics(ctx); err != nil {
					j.logger.Error().Err(err).Msg("Failed to collect realtime metrics")
				}
			},
			ctx,
		),
	)
	if err != nil {
		// Si falla la creación del Job, detener el agregador ya iniciado
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = j.aggregator.Stop(stopCtx) // Ignorar errores de detención, ya tenemos un error más importante

		return fmt.Errorf("failed to create realtime job: %w", err)
	}
	j.realtimeJob = realtimeJob

	// Crear la tarea de estadísticas por minuto
	minuteJob, err := j.scheduler.NewJob(
		gocron.CronJob(
			"0 * * * * *", // Ejecutar en el segundo 0 de cada minuto
			true,          // withSeconds establecido en true
		),
		gocron.NewTask(
			func(ctx context.Context) {
				if err := j.aggregator.CollectMinuteMetrics(ctx); err != nil {
					j.logger.Error().Err(err).Msg("Failed to collect minute metrics")
				}
			},
			ctx,
		),
	)
	if err != nil {
		// Si falla la creación del Job, detener el agregador y el planificador ya iniciados
		if err := j.scheduler.RemoveJob(realtimeJob.ID()); err != nil {
			j.logger.Error().Err(err).Msg("Failed to remove realtime job")
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = j.aggregator.Stop(stopCtx) // Ignorar errores de detención

		return fmt.Errorf("failed to create minute job: %w", err)
	}
	j.minuteJob = minuteJob

	// Iniciar el planificador
	j.scheduler.Start()
	j.isRunning = true
	j.logger.Info().Msg("HAProxy stats collection jobs started")
	return nil
}

// Stop detiene la tarea programada
func (j *StatsJob) Stop(ctx context.Context) error {
	if !j.isRunning {
		return nil // Ya está detenido, no hay nada más que hacer
	}

	j.isRunning = false
	var errs []error

	// Detener primero el planificador
	if err := j.scheduler.Shutdown(); err != nil {
		j.logger.Error().Err(err).Msg("Failed to shutdown scheduler")
		errs = append(errs, fmt.Errorf("scheduler shutdown error: %w", err))
	}

	// Luego detener el agregador
	if err := j.aggregator.Stop(ctx); err != nil {
		j.logger.Error().Err(err).Msg("Failed to stop aggregator")
		errs = append(errs, fmt.Errorf("aggregator stop error: %w", err))
	}

	j.logger.Info().Msg("HAProxy stats collection jobs stopped")

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors during stop: %v", errs)
	}
	return nil
}
