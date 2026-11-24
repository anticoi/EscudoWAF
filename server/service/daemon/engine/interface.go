package engine

import (
	"fmt"

	"github.com/anticoi/EscudoWAF/coraza-spoa/pkg/server"
	"github.com/rs/zerolog"
)

type EngineService interface {
	Start() error
	Restart() error
	Stop() error
	Reload() error
}

// NewEngineService crea una nueva instancia del servicio de motor
func NewEngineService(
	logger zerolog.Logger,
	mongoURI string,
) (EngineService, error) {
	// Crear y devolver el servicio de motor
	agent, err := server.NewAgentServer(
		logger,
		mongoURI,
	)

	if err != nil {
		return nil, fmt.Errorf("Error al crear AgentServer: %w", err)
	}

	engineService := &EngineServiceImpl{
		agent: agent,
	}

	return engineService, nil
}

type EngineServiceImpl struct {
	agent server.AgentServer
}

func (s *EngineServiceImpl) Start() error {
	return s.agent.Start()
}

func (s *EngineServiceImpl) Restart() error {
	return s.agent.Restart()
}

func (s *EngineServiceImpl) Stop() error {
	return s.agent.Stop()
}

func (s *EngineServiceImpl) Reload() error {
	return s.agent.UpdateApplications()
}
