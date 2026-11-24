package flowcontroller

import (
	"context"
	"fmt"
	"sync"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/anticoi/EscudoWAF/pkg/model"
)

// FlowControlConfig define la configuración de control de flujo
type FlowControlConfig struct {
	// Configuración de límite de acceso de alta frecuencia
	VisitLimit struct {
		Enabled        bool          // Si está habilitado
		Threshold      int64         // Umbral
		StatDuration   time.Duration // Ventana de tiempo de estadísticas
		BlockDuration  time.Duration // Duración del bloqueo
		BurstCount     int64         // Número de peticiones en ráfaga
		ParamsCapacity int64         // Capacidad de caché
	}

	// Configuración de límite de ataques de alta frecuencia
	AttackLimit struct {
		Enabled        bool          // Si está habilitado
		Threshold      int64         // Umbral
		StatDuration   time.Duration // Ventana de tiempo de estadísticas
		BlockDuration  time.Duration // Duración del bloqueo
		BurstCount     int64         // Número de peticiones en ráfaga
		ParamsCapacity int64         // Capacidad de caché
	}

	// Configuración de límite de errores de alta frecuencia
	ErrorLimit struct {
		Enabled        bool          // Si está habilitado
		Threshold      int64         // Umbral
		StatDuration   time.Duration // Ventana de tiempo de estadísticas
		BlockDuration  time.Duration // Duración del bloqueo
		BurstCount     int64         // Número de peticiones en ráfaga
		ParamsCapacity int64         // Capacidad de caché
	}
}

// FlowController procesador de control de flujo
type FlowController struct {
	config      FlowControlConfig // Configuración
	logger      zerolog.Logger    // Log
	ipRecorder  IPRecorder        // Registrador de IPs
	initialized bool              // Si está inicializado
	mutex       sync.Mutex        // Mutex
}

// Constante de nombre de recurso
const (
	ResourceVisit  = "waf:visit"  // Acceder al recurso
	ResourceAttack = "waf:attack" // Recursos de ataque
	ResourceError  = "waf:error"  // Recurso de error
)

// ConvertFromModelConfig convierte la configuración del modelo a configuración de control de flujo
func ConvertFromModelConfig(modelConfig model.FlowControlConfig) FlowControlConfig {
	config := FlowControlConfig{}

	// Configuración de límite de acceso
	config.VisitLimit.Enabled = modelConfig.VisitLimit.Enabled
	config.VisitLimit.Threshold = modelConfig.VisitLimit.Threshold
	config.VisitLimit.StatDuration = time.Duration(modelConfig.VisitLimit.StatDuration) * time.Second
	config.VisitLimit.BlockDuration = time.Duration(modelConfig.VisitLimit.BlockDuration) * time.Second
	config.VisitLimit.BurstCount = modelConfig.VisitLimit.BurstCount
	config.VisitLimit.ParamsCapacity = modelConfig.VisitLimit.ParamsCapacity

	// Configuración de límite de ataques
	config.AttackLimit.Enabled = modelConfig.AttackLimit.Enabled
	config.AttackLimit.Threshold = modelConfig.AttackLimit.Threshold
	config.AttackLimit.StatDuration = time.Duration(modelConfig.AttackLimit.StatDuration) * time.Second
	config.AttackLimit.BlockDuration = time.Duration(modelConfig.AttackLimit.BlockDuration) * time.Second
	config.AttackLimit.BurstCount = modelConfig.AttackLimit.BurstCount
	config.AttackLimit.ParamsCapacity = modelConfig.AttackLimit.ParamsCapacity

	// Configuración de límite de errores
	config.ErrorLimit.Enabled = modelConfig.ErrorLimit.Enabled
	config.ErrorLimit.Threshold = modelConfig.ErrorLimit.Threshold
	config.ErrorLimit.StatDuration = time.Duration(modelConfig.ErrorLimit.StatDuration) * time.Second
	config.ErrorLimit.BlockDuration = time.Duration(modelConfig.ErrorLimit.BlockDuration) * time.Second
	config.ErrorLimit.BurstCount = modelConfig.ErrorLimit.BurstCount
	config.ErrorLimit.ParamsCapacity = modelConfig.ErrorLimit.ParamsCapacity

	return config
}

// Añadir instancia singleton y bloqueo
var (
	flowControllerInstance *FlowController
	flowControllerMutex    sync.Mutex
)

// NewFlowControllerFromMongoConfig crea un procesador de control de flujo desde la config de Mongo (singleton)
// @Summary Crear procesador de control de flujo desde la configuración de MongoDB
// @Description Cargar configuración desde MongoDB y crear o actualizar el procesador de control de flujo, con patrón singleton
// @Param client *mongo.Client - cliente MongoDB
// @Param database string - nombre de la base de datos
// @Param logger zerolog.Logger - logger
// @Param recorder IPRecorder - registrador de IPs
// @Return *FlowController - procesador de control de flujo creado
// @Return error - información del error
func NewFlowControllerFromMongoConfig(client *mongo.Client, database string, logger zerolog.Logger, recorder IPRecorder) (*FlowController, error) {
	flowControllerMutex.Lock()
	defer flowControllerMutex.Unlock()

	// Intentar cargar la configuración
	config, err := loadFlowControlConfig(client, database, logger)
	if err != nil {
		return nil, err
	}

	// Si la instancia ya existe, actualizar la configuración
	if flowControllerInstance != nil {
		logger.Info().Msg("Actualizar la configuración del procesador de control de flujo existente")
		flowControllerInstance.UpdateConfig(config)
		return flowControllerInstance, nil
	}

	// Crear nueva instancia
	logger.Info().Msg("Crear nueva instancia del procesador de control de flujo")
	fc := NewFlowController(config, logger, recorder)
	flowControllerInstance = fc
	return fc, nil
}

// Cargar la configuración de control de flujo desde MongoDB
func loadFlowControlConfig(client *mongo.Client, database string, logger zerolog.Logger) (FlowControlConfig, error) {
	var cfg model.Config
	db := client.Database(database)
	collection := db.Collection(cfg.GetCollectionName())

	// Consultar la configuración
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Consultar la configuración con el nombre indicado
	err := collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: "AppConfig"}},
	).Decode(&cfg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return FlowControlConfig{}, fmt.Errorf("Registro de configuración no encontrado")
		}
		return FlowControlConfig{}, fmt.Errorf("Error al obtener la configuración: %w", err)
	}

	return ConvertFromModelConfig(cfg.Engine.FlowController), nil
}

// UpdateConfig actualiza la configuración de control de flujo y recarga las reglas
func (fc *FlowController) UpdateConfig(config FlowControlConfig) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	// Actualizar configuración
	fc.config = config

	// Recargar reglas
	if fc.initialized {
		// Vaciar las reglas existentes
		hotspot.ClearRules()

		// Reconfigurar los distintos tipos de reglas de control de flujo
		fc.setupAllRules()

		fc.logger.Info().Msg("Regla de control de flujo actualizada")
	}
}

// NewFlowController crea un nuevo procesador de control de flujo
func NewFlowController(config FlowControlConfig, logger zerolog.Logger, recorder IPRecorder) *FlowController {
	return &FlowController{
		config:     config,
		logger:     logger,
		ipRecorder: recorder,
	}
}

// Initialize inicializa el procesador de control de flujo
func (fc *FlowController) Initialize() error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	if fc.initialized {
		return nil
	}

	// Crear la configuración de Sentinel
	conf := config.NewDefaultConfig()

	// Configuración de la aplicación
	conf.Sentinel.App.Name = "escudo-waf-flow-controller"
	conf.Sentinel.App.Type = 1 // Tipo API Gateway

	// Configuración de logs
	conf.Sentinel.Log.Dir = "/tmp/sentinel-logs"
	conf.Sentinel.Log.UsePid = true // Usar PID para evitar conflictos entre procesos

	// Configuración de cantidad de archivos de log
	conf.Sentinel.Log.Metric.MaxFileCount = 3

	// Configuración de exportación de métricas
	conf.Sentinel.Exporter.Metric.HttpAddr = ":8719"    // Puerto de métricas por defecto de Sentinel
	conf.Sentinel.Exporter.Metric.HttpPath = "/metrics" // Ruta de métricas estilo Prometheus

	// Registrar información de configuración
	fc.logger.Info().
		Str("app_name", conf.Sentinel.App.Name).
		Int32("app_type", conf.Sentinel.App.Type).
		Str("log_dir", conf.Sentinel.Log.Dir).
		Bool("log_use_pid", conf.Sentinel.Log.UsePid).
		Uint32("log_max_files", conf.Sentinel.Log.Metric.MaxFileCount).
		Str("metrics_addr", conf.Sentinel.Exporter.Metric.HttpAddr).
		Str("metrics_path", conf.Sentinel.Exporter.Metric.HttpPath).
		Msg("Inicializar configuración de Sentinel")

	// Inicializar Sentinel
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		return fmt.Errorf("Error al inicializar Sentinel: %v", err)
	}

	// Configurar los distintos tipos de reglas de control de flujo
	fc.setupAllRules()

	fc.initialized = true
	fc.logger.Info().Msg("Inicialización del sistema de control de flujo completada")
	return nil
}

func (fc *FlowController) setupAllRules() {
	var allRules []*hotspot.Rule

	// Añadir regla de límite de acceso
	if fc.config.VisitLimit.Enabled {
		allRules = append(allRules, &hotspot.Rule{
			Resource:          ResourceVisit,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        0, // Primer parámetro, la IP
			Threshold:         fc.config.VisitLimit.Threshold,
			BurstCount:        fc.config.VisitLimit.BurstCount,
			DurationInSec:     int64(fc.config.VisitLimit.StatDuration.Seconds()),
			ParamsMaxCapacity: fc.config.VisitLimit.ParamsCapacity,
		})
	}

	// Añadir regla de límite de ataques
	if fc.config.AttackLimit.Enabled {
		allRules = append(allRules, &hotspot.Rule{
			Resource:          ResourceAttack,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        0, // Primer parámetro, la IP
			Threshold:         fc.config.AttackLimit.Threshold,
			BurstCount:        fc.config.AttackLimit.BurstCount,
			DurationInSec:     int64(fc.config.AttackLimit.StatDuration.Seconds()),
			ParamsMaxCapacity: fc.config.AttackLimit.ParamsCapacity,
		})
	}

	// Añadir regla de límite de errores
	if fc.config.ErrorLimit.Enabled {
		allRules = append(allRules, &hotspot.Rule{
			Resource:          ResourceError,
			MetricType:        hotspot.QPS,
			ControlBehavior:   hotspot.Reject,
			ParamIndex:        0, // Primer parámetro, la IP
			Threshold:         fc.config.ErrorLimit.Threshold,
			BurstCount:        fc.config.ErrorLimit.BurstCount,
			DurationInSec:     int64(fc.config.ErrorLimit.StatDuration.Seconds()),
			ParamsMaxCapacity: fc.config.ErrorLimit.ParamsCapacity,
		})
	}

	// Cargar todas las reglas de una vez
	_, err := hotspot.LoadRules(allRules)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Error al cargar las reglas de limitación de hot spot")
	} else {
		fc.logger.Info().Msgf("Todas las reglas de limitación cargadas, cantidad de reglas activas: %d", len(allRules))

		fc.logger.Info().
			Int64("threshold", fc.config.VisitLimit.Threshold).
			Int64("burstCount", fc.config.VisitLimit.BurstCount).
			Int64("durationInSec", int64(fc.config.VisitLimit.StatDuration.Seconds())).
			Bool("enabled", fc.config.VisitLimit.Enabled).
			Msg("Reglas de limitación de acceso cargadas con éxito")

		fc.logger.Info().
			Int64("threshold", fc.config.AttackLimit.Threshold).
			Int64("burstCount", fc.config.AttackLimit.BurstCount).
			Int64("durationInSec", int64(fc.config.AttackLimit.StatDuration.Seconds())).
			Bool("enabled", fc.config.AttackLimit.Enabled).
			Msg("Reglas de limitación de ataques cargadas con éxito")

		fc.logger.Info().
			Int64("threshold", fc.config.ErrorLimit.Threshold).
			Int64("burstCount", fc.config.ErrorLimit.BurstCount).
			Int64("durationInSec", int64(fc.config.ErrorLimit.StatDuration.Seconds())).
			Bool("enabled", fc.config.ErrorLimit.Enabled).
			Msg("Reglas de limitación de errores cargadas con éxito")
	}
}

// CheckVisit comprueba si la petición de la IP está permitida
func (fc *FlowController) CheckVisit(ip string, requestUri string) (bool, error) {
	if !fc.initialized {
		if err := fc.Initialize(); err != nil {
			return true, err
		}
	}

	// Usar limitación por parámetro de hot spot, pasando la IP como primer parámetro
	entry, blockError := sentinel.Entry(ResourceVisit,
		sentinel.WithArgs(ip),
		sentinel.WithTrafficType(base.Inbound),
	)

	if blockError != nil {
		// Registrar las IPs restringidas
		fc.ipRecorder.RecordBlockedIP(ip, "high_frequency_visit", requestUri, fc.config.VisitLimit.BlockDuration)
		fc.logger.Warn().
			Str("ip", ip).
			Str("reason", "high_frequency_visit").
			Dur("block_duration", fc.config.VisitLimit.BlockDuration).
			Msg("Acceso restringido por IP")
		return false, nil
	}

	// No olvidar liberar los recursos
	defer entry.Exit()
	return true, nil
}

// RecordAttack registra la detección de ataque de una IP, devuelve si quedó restringida
func (fc *FlowController) RecordAttack(ip string, requestUri string) (bool, error) {
	if !fc.initialized {
		if err := fc.Initialize(); err != nil {
			return false, err
		}
	}

	// Usar limitación por parámetro de hot spot, pasando la IP como primer parámetro
	entry, blockError := sentinel.Entry(ResourceAttack,
		sentinel.WithArgs(ip),
		sentinel.WithTrafficType(base.Inbound),
	)

	if blockError != nil {
		// Registrar las IPs restringidas
		fc.ipRecorder.RecordBlockedIP(ip, "high_frequency_attack", requestUri, fc.config.AttackLimit.BlockDuration)
		fc.logger.Warn().
			Str("ip", ip).
			Str("reason", "high_frequency_attack").
			Dur("block_duration", fc.config.AttackLimit.BlockDuration).
			Msg("IP restringida por ataques de alta frecuencia")
		return true, nil
	}

	// No olvidar liberar los recursos
	defer entry.Exit()
	return false, nil
}

// RecordError registra una respuesta de error de una IP, devuelve si quedó restringida
func (fc *FlowController) RecordError(ip string, requestUri string) (bool, error) {
	if !fc.initialized {
		if err := fc.Initialize(); err != nil {
			return false, err
		}
	}

	// Usar limitación por parámetro de hot spot, pasando la IP como primer parámetro
	entry, blockError := sentinel.Entry(ResourceError,
		sentinel.WithArgs(ip),
		sentinel.WithTrafficType(base.Inbound),
	)

	if blockError != nil {
		// Registrar las IPs restringidas
		fc.ipRecorder.RecordBlockedIP(ip, "high_frequency_error", requestUri, fc.config.ErrorLimit.BlockDuration)
		fc.logger.Warn().
			Str("ip", ip).
			Str("reason", "high_frequency_error").
			Dur("block_duration", fc.config.ErrorLimit.BlockDuration).
			Msg("IP restringida por errores de alta frecuencia")
		return true, nil
	}

	// No olvidar liberar los recursos
	defer entry.Exit()
	return false, nil
}

// Close Cerrar el sistema de control de flujo
// @Summary Cerrar el sistema de control de flujo
// @Description Liberar los recursos del sistema de control de flujo, incluyendo el cierre del registrador de IPs
// @Return error información del error
func (fc *FlowController) Close() error {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	fc.logger.Info().Msg("Cerrando el sistema de control de flujo")

	// Vaciar las reglas de limitación
	if fc.initialized {
		// Vaciar las reglas de parámetros de hot spot
		hotspot.ClearRules()
		fc.initialized = false
	}

	// Cerrar el registrador de IPs
	if fc.ipRecorder != nil {
		if err := fc.ipRecorder.Close(); err != nil {
			fc.logger.Error().Err(err).Msg("Error al cerrar el registrador de IPs")
			return err
		}
	}

	fc.logger.Info().Msg("Sistema de control de flujo cerrado")
	return nil
}
