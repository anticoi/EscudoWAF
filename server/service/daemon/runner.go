package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	mongodb "github.com/anticoi/EscudoWAF/pkg/database/mongo"
	"github.com/haproxytech/client-native/v6/models"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/anticoi/EscudoWAF/server/service/daemon/engine"
	"github.com/anticoi/EscudoWAF/server/service/daemon/haproxy"
	"github.com/rs/zerolog"
)

type ServiceState int

const (
	ServiceStopped ServiceState = iota
	ServiceRunning
	ServiceError
)

type ServiceRunner interface {
	StartServices() error
	StopServices() error
	ForceStop()
	Restart() error
	HotReload() error
	GetState() ServiceState
	GetStats() (models.NativeStats, error)
}

// ServiceRunner gestiona y coordina todos los servicios en segundo plano
type ServiceRunnerImpl struct {
	haproxyService haproxy.HAProxyService
	engineService  engine.EngineService
	ctx            context.Context
	cancel         context.CancelFunc
	logger         *zerolog.Logger
	errChan        chan error
	haproxyDone    chan struct{} // Notificar que el servicio HAProxy se detuvo
	engineDone     chan struct{} // Notificar que el servicio Engine se detuvo
	state          ServiceState
}

// Implementación del patrón singleton
var (
	instance ServiceRunner
	once     sync.Once
	initErr  error
)

// GetRunnerService obtiene la instancia singleton de ServiceRunner
func GetRunnerService() (ServiceRunner, error) {
	once.Do(func() {
		instance, initErr = newServiceRunner()
	})
	return instance, initErr
}

// NewServiceRunner crea un nuevo ejecutor de servicios
func newServiceRunner() (ServiceRunner, error) {
	logger := config.GetLogger().With().Str("component", "runner").Logger()

	haproxyService, err := haproxy.NewHAProxyService("", "", nil)
	if err != nil {
		logger.Error().Err(err).Msg("Error al inicializar el servicio HAProxy")
		return nil, fmt.Errorf("Error al inicializar el servicio HAProxy: %w", err)
	}

	engineLogger := logger.With().Str("component", "engine").Logger()
	// Crear el servicio Engine
	engineService, err := engine.NewEngineService(engineLogger, config.Global.DBConfig.URI)
	if err != nil {
		logger.Error().Err(err).Msg("Error al inicializar el servicio Engine")
		return nil, fmt.Errorf("Error al inicializar el servicio Engine: %w", err)
	}

	return &ServiceRunnerImpl{
		haproxyService: haproxyService,
		engineService:  engineService,
		logger:         &logger,
		state:          ServiceStopped,
	}, nil
}

// StartServices inicia todos los servicios
func (r *ServiceRunnerImpl) StartServices() error {
	// Comprobar si el servicio ya está en ejecución
	if r.state == ServiceRunning {
		return fmt.Errorf("El servicio ya está en ejecución")
	}

	// Crear nuevo contexto y función de cancelación
	r.ctx, r.cancel = context.WithCancel(context.Background())

	// Crear nuevo canal
	r.errChan = make(chan error, 10)
	r.haproxyDone = make(chan struct{})
	r.engineDone = make(chan struct{})

	// Iniciar el servicio HAProxy
	go func() {
		defer close(r.haproxyDone) // Cerrar el canal cuando el servicio se detiene

		client, err := mongodb.Connect(config.Global.DBConfig.URI)
		if err != nil {
			r.logger.Error().Err(err).Msg("runnner start services failed to connect to database")
			r.errChan <- err
			return
		}

		// Obtener la base de datos
		db := client.Database(config.Global.DBConfig.Database)

		var site model.Site
		siteList, err := repository.GetAllSites(r.ctx, db.Collection(site.GetCollectionName()))
		if err != nil {
			r.logger.Error().Err(err).Msg("Error al obtener la lista de sitios")
			r.errChan <- err
			return
		}

		r.logger.Info().Msg("Iniciando el servicio HAProxy...")

		if err = r.haproxyService.RemoveConfig(); err != nil {
			r.logger.Error().Err(err).Msg("Error al eliminar la configuración de HAProxy")
			r.errChan <- err
			return
		}

		if err = r.haproxyService.InitSpoeConfig(); err != nil {
			r.logger.Error().Err(err).Msg("Error al inicializar la configuración SPOE de HAProxy")
			r.errChan <- err
			return
		}

		if err = r.haproxyService.InitHAProxyConfig(); err != nil {
			r.logger.Error().Err(err).Msg("Error al inicializar la configuración de HAProxy")
			r.errChan <- err
			return
		}

		if err = r.haproxyService.AddCorazaBackend(); err != nil {
			r.logger.Error().Err(err).Msg("Error al añadir el backend de Coraza")
			r.errChan <- err
			return
		}

		if err = r.haproxyService.CreateHAProxyCrtStore(); err != nil {
			r.logger.Error().Err(err).Msg("Error al crear el almacén de certificados de HAProxy")
			r.errChan <- err
			return
		}

		for i, site := range siteList {
			if err := r.haproxyService.AddSiteConfig(site); err != nil {
				r.logger.Error().Err(err).Msgf("Error al añadir la configuración de sitio %d", i)
				// Seguir procesando otros sitios, sin devolver error
			}
		}

		if err := r.haproxyService.Start(); err != nil {
			r.logger.Error().Err(err).Msg("Error al iniciar el servicio HAProxy")
			r.errChan <- err
			return
		}

		// Esperar la señal de parada
		<-r.ctx.Done()
		r.logger.Info().Msg("Señal de parada recibida, deteniendo el servicio HAProxy")
		if err := r.haproxyService.Stop(); err != nil {
			r.logger.Error().Err(err).Msg("Error al detener el servicio HAProxy")
			r.errChan <- err
		}
	}()

	// Iniciar el servicio Engine
	go func() {
		defer close(r.engineDone) // Cerrar el canal cuando el servicio se detiene

		r.logger.Info().Msg("Iniciando servicio Engine...")
		if err := r.engineService.Start(); err != nil {
			r.logger.Error().Err(err).Msg("Error al iniciar el servicio Engine")
			r.errChan <- err
			return
		}

		// Esperar la señal de parada
		<-r.ctx.Done()
		r.logger.Info().Msg("Señal de parada recibida, deteniendo el servicio Engine")
		if err := r.engineService.Stop(); err != nil {
			r.logger.Error().Err(err).Msg("Error al detener el servicio Engine")
			r.errChan <- err
		}
	}()

	// Escuchar el canal de errores y devolver el primer error si ocurre
	select {
	case err := <-r.errChan:
		r.logger.Error().Err(err).Msg("Error durante el inicio del servicio")
		r.cancel() // Cancelar el contexto, notifica a todos los servicios que se detengan
		r.state = ServiceError
		return err
	case <-time.After(2 * time.Second): // Dar algo de tiempo de arranque al servicio
		r.logger.Info().Msg("Todos los servicios iniciados")
		r.state = ServiceRunning
		return nil
	}
}

// StopServices detiene todos los servicios
func (r *ServiceRunnerImpl) StopServices() error {
	// Comprobar si el servicio está en ejecución
	if r.state != ServiceRunning {
		return fmt.Errorf("El servicio no está en ejecución")
	}

	r.logger.Info().Msg("Deteniendo todos los servicios...")

	// 1. Primero cancelar el contexto para notificar a todas las operaciones que lo usan
	if r.cancel != nil {
		r.cancel()
	}

	// 2. Esperar la señal de detención de los servicios
	timeoutDuration := 15 * time.Second

	// Monitorizar la detención del servicio HAProxy
	haproxyOk := make(chan struct{})
	go func() {
		select {
		case <-r.haproxyDone:
			r.logger.Info().Msg("El servicio HAProxy se detuvo correctamente")
		case <-time.After(timeoutDuration):
			r.logger.Warn().Msg("Tiempo de espera agotado al detener el servicio HAProxy")
			// Intentar detener forzosamente
			if err := r.haproxyService.Stop(); err != nil {
				r.logger.Error().Err(err).Msg("Error al forzar la detención del servicio HAProxy")
			}
		}
		close(haproxyOk)
	}()

	// Monitorizar la detención del servicio Engine
	engineOk := make(chan struct{})
	go func() {
		select {
		case <-r.engineDone:
			r.logger.Info().Msg("El servicio Engine se detuvo correctamente")
		case <-time.After(timeoutDuration):
			r.logger.Warn().Msg("Tiempo de espera agotado al detener el servicio Engine")
			// Intentar detener forzosamente
			if err := r.engineService.Stop(); err != nil {
				r.logger.Error().Err(err).Msg("Error al forzar la detención del servicio Engine")
			}
		}
		close(engineOk)
	}()

	// Esperar a que todos los servicios se detengan
	<-haproxyOk
	<-engineOk

	var stopErr error
	select {
	case err := <-r.errChan:
		r.logger.Error().Err(err).Msg("Error durante la detención del servicio")
		stopErr = err
	default:
		// Sin errores
	}

	// Realizar el reinicio del servicio HAProxy
	if err := r.haproxyService.Reset(); err != nil {
		r.logger.Error().Err(err).Msg("Error al reiniciar el servicio HAProxy")
		if stopErr == nil {
			stopErr = err
		}
	}

	// Liberar recursos
	r.ctx = nil
	r.cancel = nil
	r.errChan = nil
	r.haproxyDone = nil
	r.engineDone = nil

	// Actualizar el estado
	r.state = ServiceStopped

	if stopErr != nil {
		return stopErr
	}

	r.logger.Info().Msg("Todos los servicios detenidos correctamente")

	return nil
}

// ForceStop detiene forzosamente todos los servicios
func (r *ServiceRunnerImpl) ForceStop() {
	r.logger.Info().Msg("Forzando la detención de todos los servicios...")

	// 1. Primero cancelar el contexto para notificar a todas las operaciones que lo usan
	if r.cancel != nil {
		r.cancel()
	}

	// 2. Detener directamente cada servicio
	if r.haproxyService != nil {
		r.logger.Info().Msg("Forzando la detención del servicio HAProxy...")
		if err := r.haproxyService.Stop(); err != nil {
			r.logger.Error().Err(err).Msg("Error al forzar la detención del servicio HAProxy")
		}
	}

	// Realizar el reinicio del servicio HAProxy
	if err := r.haproxyService.Reset(); err != nil {
		r.logger.Error().Err(err).Msg("Error al reiniciar el servicio HAProxy")
	}

	if r.engineService != nil {
		r.logger.Info().Msg("Forzando la detención del servicio Engine...")
		if err := r.engineService.Stop(); err != nil {
			r.logger.Error().Err(err).Msg("Error al forzar la detención del servicio Engine")
		}
	}

	// Liberar recursos
	r.ctx = nil
	r.cancel = nil
	r.errChan = nil
	r.haproxyDone = nil
	r.engineDone = nil

	// Actualizar el estado
	r.state = ServiceStopped

	r.logger.Info().Msg("Todos los servicios detenidos forzosamente")
}

// Restart reinicia todos los servicios
func (r *ServiceRunnerImpl) Restart() error {
	// Detener servicio
	if r.state == ServiceRunning {
		if err := r.StopServices(); err != nil {
			r.logger.Error().Err(err).Msg("Error al detener el servicio durante el reinicio")
			return err
		}
	}

	// Iniciar el servicio
	return r.StartServices()
}

func (r *ServiceRunnerImpl) HotReload() error {
	// Comprobar si el servicio está en ejecución
	if r.state != ServiceRunning {
		return fmt.Errorf("El servicio no está en ejecución, no se puede recargar en caliente")
	}
	r.logger.Info().Msg("Iniciando recarga en caliente...")

	// // reload haproxy config

	// Realizar el reinicio del servicio HAProxy
	if err := r.haproxyService.Reset(); err != nil {
		r.logger.Error().Err(err).Msg("Error al reiniciar el servicio HAProxy")
		return err
	}

	client, err := mongodb.Connect(config.Global.DBConfig.URI)
	if err != nil {
		r.logger.Error().Err(err).Msg("hot reload failed to connect to database")
		return err
	}

	// Obtener la base de datos
	db := client.Database(config.Global.DBConfig.Database)

	var site model.Site
	siteList, err := repository.GetAllSites(r.ctx, db.Collection(site.GetCollectionName()))
	if err != nil {
		r.logger.Error().Err(err).Msg("Error al obtener la lista de sitios en la recarga en caliente")
		return err
	}

	r.logger.Info().Msg("Iniciando recarga en caliente de la configuración de HAProxy...")
	if err = r.haproxyService.HotReloadRemoveConfig(); err != nil {
		r.logger.Error().Err(err).Msg("Error al eliminar la configuración de HAProxy en la recarga en caliente")
		return err
	}

	if err = r.haproxyService.InitSpoeConfig(); err != nil {
		r.logger.Error().Err(err).Msg("Error al inicializar la configuración SPOE de HAProxy")
		return err
	}

	if err = r.haproxyService.InitHAProxyConfig(); err != nil {
		r.logger.Error().Err(err).Msg("Error al inicializar la configuración de HAProxy")
		return err
	}

	if err = r.haproxyService.AddCorazaBackend(); err != nil {
		r.logger.Error().Err(err).Msg("Error al añadir el backend de Coraza")
		return err
	}

	if err = r.haproxyService.CreateHAProxyCrtStore(); err != nil {
		r.logger.Error().Err(err).Msg("Error al crear el almacén de certificados de HAProxy")
		return err
	}

	for i, site := range siteList {
		if err := r.haproxyService.AddSiteConfig(site); err != nil {
			r.logger.Error().Err(err).Msgf("Error al añadir la configuración de sitio %d", i)
			// Seguir procesando otros sitios, sin devolver error
		}
	}

	if err := r.haproxyService.Reload(); err != nil {
		r.logger.Error().Err(err).Msg("Error en la recarga en caliente de la configuración de HAProxy")
		return err
	}

	// reload engine config

	if err := r.engineService.Reload(); err != nil {
		r.logger.Error().Err(err).Msg("Error en la recarga en caliente de la configuración de Engine")
		return err
	}

	r.logger.Info().Msg("Recarga en caliente exitosa")

	return nil
}

// GetState obtiene el estado actual del servicio
func (r *ServiceRunnerImpl) GetState() ServiceState {
	return r.state
}

// GetStats obtiene la información estadística de HAProxy
func (r *ServiceRunnerImpl) GetStats() (models.NativeStats, error) {
	if r.haproxyService == nil {
		return models.NativeStats{}, fmt.Errorf("haproxy service not initialized")
	}
	return r.haproxyService.GetStats()
}
