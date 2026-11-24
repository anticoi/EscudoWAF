// server/agent_server.go
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	cfg "github.com/anticoi/EscudoWAF/coraza-spoa/config"
	"github.com/anticoi/EscudoWAF/coraza-spoa/internal"
	mongodb "github.com/anticoi/EscudoWAF/pkg/database/mongo"
	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/pkg/utils/network"
)

var globalLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

// ServerState representa el estado de ejecución del servidor
type ServerState int

const (
	ServerStopped ServerState = iota // Servicio detenido
	ServerRunning                    // Servicio en ejecución
	ServerError                      // Error del servicio
)

type AgentServer interface {
	Start() error
	Stop() error
	Restart() error
	UpdateApplications() error
	UpdateNetworkAddress(network, address string)
	UpdateLogger(logger zerolog.Logger)
	GetState() ServerState
	GetLastError() error
	GetLatestConfig() (*model.Config, error)
}

// AgentServer gestiona el ciclo de vida del servicio Agent
type AgentServerImpl struct {
	mu           sync.Mutex
	ctx          context.Context
	cancelFunc   context.CancelFunc
	agent        *internal.Agent
	listener     net.Listener
	network      string
	address      string
	applications map[string]*internal.Application
	logger       zerolog.Logger
	state        ServerState
	lastError    error
	mongoURI     string
}

func NewAgentServer(logger zerolog.Logger, mongoURI string) (AgentServer, error) {
	if mongoURI == "" {
		return nil, errors.New("mongoURI is required")
	}

	return &AgentServerImpl{
		logger:   logger,
		state:    ServerStopped,
		mongoURI: mongoURI,
	}, nil
}

// Start inicia el servicio
func (s *AgentServerImpl) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == ServerRunning {
		return errors.New("El servicio ya está en ejecución")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancelFunc = cancel

	globalConfig, err := s.GetLatestConfig()
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed getting latest config")
		return err
	}

	mongoClient, err := mongodb.Connect(s.mongoURI)
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed creating MongoDB client")
		return err
	}

	var wafLog model.WAFLog
	mongoConfig := &internal.MongoConfig{
		Client:     mongoClient,
		Database:   "waf",
		Collection: wafLog.GetCollectionName(),
	}

	var microRule model.MicroRule
	var ipGroup model.IPGroup

	ruleEngineMongoConfig := &internal.MongoDBConfig{
		MongoClient:       mongoClient,
		Database:          "waf",
		RuleCollection:    microRule.GetCollectionName(),
		IPGroupCollection: ipGroup.GetCollectionName(),
	}

	flowControllerConfig := internal.FlowControllerConfig{
		Client:   mongoClient,
		Database: "waf",
	}

	geoIPConfig := internal.GeoIP2Options{
		ASNDBPath:  globalConfig.Engine.ASNDBPath,
		CityDBPath: globalConfig.Engine.CityDBPath,
	}

	appConfigs := globalConfig.Engine.AppConfig

	// Convert model.AppConfig to internal.AppConfig and create applications
	allApps := make(map[string]*internal.Application)
	for _, appConfig := range appConfigs {
		// Crear la configuración de logs
		logConfig := cfg.LogConfig{
			Level:  appConfig.LogLevel,
			File:   appConfig.LogFile,
			Format: appConfig.LogFormat,
		}

		// Crear el logger
		appLogger, err := logConfig.NewLogger()
		if err != nil {
			s.logger.Warn().Err(err).Str("app", appConfig.Name).Msg("Usar logger por defecto")
			appLogger = globalLogger
		}

		// Crear AppConfig interna
		internalAppConfig := internal.AppConfig{
			Directives:     appConfig.Directives,
			ResponseCheck:  globalConfig.IsResponseCheck, // Usar la configuración global de inspección de respuestas
			Logger:         appLogger,
			TransactionTTL: appConfig.TransactionTTL,
		}

		// Crear la aplicación
		application, err := internalAppConfig.NewApplicationWithContext(ctx, internal.ApplicationOptions{
			MongoConfig:          mongoConfig,
			GeoIPConfig:          &geoIPConfig,
			RuleEngineDbConfig:   ruleEngineMongoConfig,
			FlowControllerConfig: &flowControllerConfig,
		}, globalConfig.IsDebug)
		if err != nil {
			s.logger.Fatal().Err(err).Msg("Failed creating application: " + appConfig.Name)

			return err
		}

		allApps[appConfig.Name] = application
	}

	s.applications = allApps
	s.network, s.address = network.NetworkAddressFromBind(globalConfig.Engine.Bind)

	// Crear el listener
	l, err := (&net.ListenConfig{}).Listen(s.ctx, s.network, s.address)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al crear el socket")
		s.state = ServerError
		s.lastError = err
		return err
	}
	s.listener = l

	// Crear instancia de Agent
	s.agent = &internal.Agent{
		Context:      s.ctx,
		Applications: s.applications,
		Logger:       s.logger,
	}

	// Iniciar el servicio en una goroutine en segundo plano
	go func() {
		s.logger.Info().Msg("Iniciando servicio coraza-spoa, dirección de escucha: " + s.address + " " + s.network)
		err := s.agent.Serve(l)

		// Solo registrar como error si no fue un cierre normal
		if err != nil && !errors.Is(err, net.ErrClosed) && !strings.Contains(err.Error(), "use of closed network connection") {
			s.mu.Lock()
			s.state = ServerError
			s.lastError = err
			s.mu.Unlock()
			s.logger.Error().Err(err).Msg("Error del listener, error inesperado")
		} else if err != nil {
			// Error esperado: al cancelar el contexto, s.agent.Serve(l) lanza un error con prefijo 'accepting conn:'
			s.logger.Info().Msg("El listener se cerró correctamente, error esperado: " + err.Error())
		}
	}()

	s.state = ServerRunning
	return nil
}

// Stop detiene el servicio
func (s *AgentServerImpl) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == ServerStopped {
		return errors.New("Servicio no iniciado")
	}

	// Cancelar el contexto
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	// Cerrar el listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Error().Err(err).Msg("Error al cerrar el listener")
			return err
		}
		s.listener = nil
	}

	s.agent = nil
	s.applications = nil
	s.ctx = nil

	s.state = ServerStopped
	s.logger.Info().Msg("Servicio detenido")
	return nil
}

// Restart reinicia el servicio
func (s *AgentServerImpl) Restart() error {
	if err := s.Stop(); err != nil && !errors.Is(err, errors.New("Servicio no iniciado")) {
		return err
	}
	return s.Start()
}

// UpdateApplications actualiza la configuración de aplicaciones, soporta recarga en caliente
func (s *AgentServerImpl) UpdateApplications() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	globalConfig, err := s.GetLatestConfig()
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed getting latest config")
		return err
	}

	mongoClient, err := mongodb.Connect(s.mongoURI)

	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed creating MongoDB client")
		return err
	}

	var wafLog model.WAFLog
	mongoConfig := &internal.MongoConfig{
		Client:     mongoClient,
		Database:   "waf",
		Collection: wafLog.GetCollectionName(),
	}

	var microRule model.MicroRule
	var ipGroup model.IPGroup

	ruleEngineMongoConfig := &internal.MongoDBConfig{
		MongoClient:       mongoClient,
		Database:          "waf",
		RuleCollection:    microRule.GetCollectionName(),
		IPGroupCollection: ipGroup.GetCollectionName(),
	}

	flowControllerConfig := internal.FlowControllerConfig{
		Client:   mongoClient,
		Database: "waf",
	}

	geoIPConfig := internal.GeoIP2Options{
		ASNDBPath:  globalConfig.Engine.ASNDBPath,
		CityDBPath: globalConfig.Engine.CityDBPath,
	}

	// Extraer la lista de AppConfig desde Config
	appConfigs := globalConfig.Engine.AppConfig

	// Convert model.AppConfig to internal.AppConfig and create applications
	allApps := make(map[string]*internal.Application)
	for _, appConfig := range appConfigs {
		// Crear la configuración de logs
		logConfig := cfg.LogConfig{
			Level:  appConfig.LogLevel,
			File:   appConfig.LogFile,
			Format: appConfig.LogFormat,
		}

		// Crear el logger
		appLogger, err := logConfig.NewLogger()
		if err != nil {
			s.logger.Warn().Err(err).Str("app", appConfig.Name).Msg("Usar logger por defecto")
			appLogger = globalLogger
		}

		// Crear AppConfig interna
		internalAppConfig := internal.AppConfig{
			Directives:     appConfig.Directives,
			ResponseCheck:  globalConfig.IsResponseCheck, // Usar la configuración global de inspección de respuestas
			Logger:         appLogger,
			TransactionTTL: appConfig.TransactionTTL,
		}

		// Crear la aplicación
		application, err := internalAppConfig.NewApplicationWithContext(s.ctx, internal.ApplicationOptions{
			MongoConfig:          mongoConfig,
			GeoIPConfig:          &geoIPConfig,
			RuleEngineDbConfig:   ruleEngineMongoConfig,
			FlowControllerConfig: &flowControllerConfig,
		}, globalConfig.IsDebug)

		if err != nil {
			s.logger.Fatal().Err(err).Msg("Failed creating application: " + appConfig.Name)
			return err
		}

		allApps[appConfig.Name] = application
	}

	s.applications = allApps

	// Si el servicio está en ejecución, actualizar en caliente las aplicaciones del Agent
	if s.state == ServerRunning && s.agent != nil && s.ctx != nil {
		s.agent.ReplaceApplications(allApps)
		s.logger.Info().Msg("Configuración de aplicación actualizada")
	}

	return nil
}

// UpdateNetworkAddress actualiza la dirección de red, no soporta recarga en caliente
func (s *AgentServerImpl) UpdateNetworkAddress(network, address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.network = network
	s.address = address
}

// UpdateLogger actualiza el logger, soporta recarga en caliente
func (s *AgentServerImpl) UpdateLogger(logger zerolog.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger = logger
	if s.agent != nil {
		s.agent.Logger = logger
	}
}

// GetState obtiene el estado actual del servicio
func (s *AgentServerImpl) GetState() ServerState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// GetLastError obtiene el último error
func (s *AgentServerImpl) GetLastError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastError
}

func (s *AgentServerImpl) GetLatestConfig() (*model.Config, error) {
	if s.mongoURI == "" {
		return nil, errors.New("mongoURI is required")
	}
	// Conectar a la base de datos
	client, err := mongodb.Connect(s.mongoURI)
	if err != nil {
		return nil, fmt.Errorf("Error al conectar con la base de datos: %w", err)
	}

	var cfg model.Config
	// Obtener la colección de configuración
	db := client.Database("waf")
	collection := db.Collection(cfg.GetCollectionName())

	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Asegurar que los recursos se liberen

	// Consultar la configuración con el nombre indicado
	err = collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: "AppConfig"}},
	).Decode(&cfg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("Registro de configuración no encontrado")
		}
		return nil, fmt.Errorf("Error al obtener la configuración: %w", err)
	}
	return &cfg, nil
}
