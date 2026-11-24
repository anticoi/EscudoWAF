package cornjob

import (
	"context"
	"fmt"
	"time"

	mongodb "github.com/anticoi/EscudoWAF/pkg/database/mongo"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/anticoi/EscudoWAF/server/service/daemon"
	"github.com/rs/zerolog"
)

// Start initializes and starts the HAProxy stats aggregation service
func Start(runner daemon.ServiceRunner, logger zerolog.Logger) (func(), error) {

	targetList, err := GetLatestTargetList()
	if err != nil {
		return nil, fmt.Errorf("cornjob start failed:  failed to get target list: %w", err)
	}

	if len(targetList) == 0 {
		logger.Warn().Msg("No HAProxy targets found to monitor. Service will start in standby mode.")
	} else {
		logger.Info().Strs("targets", targetList).Msg("Starting HAProxy stats service with initial targets")
	}

	// Crear el servicio de tareas programadas
	cronJobService, err := GetInstance(runner, targetList)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron job service: %w", err)
	}

	// Iniciar la tarea programada
	if err := cronJobService.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cron job service: %w", err)
	}

	// Devolver función de limpieza para que el programa principal la llame al salir
	cleanup := func() {
		logger.Info().Msg("Shutting down HAProxy stats service...")

		// Crear un contexto con timeout para que el cierre no se cuelgue indefinidamente
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Detener servicio
		if err := cronJobService.Stop(); err != nil {
			logger.Error().Err(err).Msg("Error when stopping HAProxy stats cron jobs")
			// Si expira el tiempo, registrar la terminación forzada
			select {
			case <-shutdownCtx.Done():
				logger.Warn().Msg("Forced shutdown of HAProxy stats service due to timeout")
			default:
				// Cierre normal, no hay nada que hacer
			}
		} else {
			logger.Info().Msg("HAProxy stats service shutdown completed successfully")
		}
	}

	if len(targetList) == 0 {
		logger.Info().Msg("HAProxy stats service started in standby mode (no targets)")
	} else {
		logger.Info().Msg("HAProxy stats service started successfully with active monitoring")
	}

	return cleanup, nil
}

func GetLatestTargetList() ([]string, error) {
	// Conectar a la base de datos
	client, err := mongodb.Connect(config.Global.DBConfig.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Obtener la base de datos
	db := client.Database(config.Global.DBConfig.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var site model.Site
	siteList, err := repository.GetAllSites(ctx, db.Collection(site.GetCollectionName()))

	if err != nil {
		return nil, fmt.Errorf("failed to get site list: %w", err)
	}

	// Usar map para garantizar unicidad
	targetMap := make(map[string]struct{})
	// Iterar sobre la lista de sitios para obtener la lista de objetivos
	for _, site := range siteList {
		if !site.ActiveStatus {
			continue
		}
		targetMap[fmt.Sprintf("fe_%d_http", site.ListenPort)] = struct{}{}
		targetMap[fmt.Sprintf("fe_%d_https", site.ListenPort)] = struct{}{}
	}

	// Convertir el map de nuevo a slice
	var targetList []string
	for target := range targetMap {
		targetList = append(targetList, target)
	}

	return targetList, nil
}
