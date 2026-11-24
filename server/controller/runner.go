package controller

import (
	"errors"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/service/daemon"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RunnerController interfaz del controlador del runner
type RunnerController interface {
	GetStatus(ctx *gin.Context)
	Control(ctx *gin.Context)
}

// RunnerControllerImpl implementación del controlador del runner
type RunnerControllerImpl struct {
	runnerService service.RunnerService
	logger        zerolog.Logger
}

// NewRunnerController crea el controlador del runner
func NewRunnerController(runnerService service.RunnerService) RunnerController {
	logger := config.GetControllerLogger("runner")
	return &RunnerControllerImpl{
		runnerService: runnerService,
		logger:        logger,
	}
}

// getStateString convierte ServiceState a cadena
func getStateString(state daemon.ServiceState) string {
	switch state {
	case daemon.ServiceRunning:
		return "running"
	case daemon.ServiceStopped:
		return "stopped"
	case daemon.ServiceError:
		return "error"
	default:
		return "unknown"
	}
}

// isStateRunning comprueba si el estado es en ejecución
func isStateRunning(state daemon.ServiceState) bool {
	return state == daemon.ServiceRunning
}

// getSuccessMessage devuelve el mensaje de éxito según el tipo de operación
func getSuccessMessage(action string) string {
	switch action {
	case "start":
		return "El runner se inició con éxito"
	case "stop":
		return "El runner se detuvo con éxito"
	case "restart":
		return "El runner se reinició con éxito"
	case "force_stop":
		return "El runner se detuvo forzosamente"
	case "reload":
		return "Configuración del runner recargada"
	default:
		return "Operación exitosa"
	}
}

// toRunnerStatusResponse convierte el estado a objeto de respuesta
func toRunnerStatusResponse(state daemon.ServiceState) dto.RunnerStatusResponse {
	return dto.RunnerStatusResponse{
		State:     getStateString(state),
		IsRunning: isStateRunning(state),
	}
}

// buildControlResponse construye la respuesta de control
func buildControlResponse(action string, state daemon.ServiceState) dto.RunnerControlResponse {
	return dto.RunnerControlResponse{
		Success: true,
		Action:  action,
		Message: getSuccessMessage(action),
		State:   getStateString(state),
	}
}

// GetStatus obtiene el estado del runner
//
//	@Summary		Obtener el estado del runner en segundo plano
//	@Description	Obtener el estado de ejecución del runner del WAF
//	@Tags			Gestión del runner
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.RunnerStatusResponse}	"Estado del runner obtenido con éxito"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"Error interno del servidor"
//	@Router			/api/runner/status [get]
func (c *RunnerControllerImpl) GetStatus(ctx *gin.Context) {
	// Obtener el estado del runner
	state, err := c.runnerService.GetStatus(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al obtener el estado del runner")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Construir la respuesta
	resp := toRunnerStatusResponse(state)

	response.Success(ctx, "Estado del runner obtenido con éxito", resp)
}

// Control controla el runner
//
//	@Summary		Controlar el runner en segundo plano
//	@Description	Ejecutar operación de inicio, parada, reinicio, parada forzada o recarga en caliente
//	@Tags			Gestión del runner
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.RunnerControlRequest	true	"Petición de control del runner"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.RunnerControlResponse}	"Operación exitosa"
//	@Failure		400	{object}	model.ErrResponse										"Parámetros de la petición incorrectos"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"Error interno del servidor"
//	@Router			/api/runner/control [post]
func (c *RunnerControllerImpl) Control(ctx *gin.Context) {
	var req dto.RunnerControlRequest

	// Vincular los parámetros de la petición
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("action", req.Action).Msg("Petición de control del runner")

	var err error

	// Ejecutar la operación correspondiente según el tipo
	switch req.Action {
	case "start":
		err = c.runnerService.Start(ctx)
	case "stop":
		err = c.runnerService.Stop(ctx)
	case "restart":
		err = c.runnerService.Restart(ctx)
	case "force_stop":
		err = c.runnerService.ForceStop(ctx)
	case "reload":
		err = c.runnerService.Reload(ctx)
	default:
		c.logger.Warn().Str("action", req.Action).Msg("Tipo de operación no soportada")
		response.BadRequest(ctx, errors.New("Tipo de operación no soportada"), true)
		return
	}

	// Manejar el resultado de la operación
	if err != nil {
		if errors.Is(err, service.ErrRunnerNotRunning) {
			c.logger.Warn().Str("action", req.Action).Msg("El runner no está en ejecución")
			response.BadRequest(ctx, err, true)
			return
		} else if errors.Is(err, service.ErrRunnerAlreadyRunning) {
			c.logger.Warn().Str("action", req.Action).Msg("El runner ya está en ejecución")
			response.BadRequest(ctx, err, true)
			return
		}

		c.logger.Error().Err(err).Str("action", req.Action).Msg("Operación del runner fallida")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Obtener el estado más reciente tras una operación exitosa
	state, _ := c.runnerService.GetStatus(ctx)

	// Construir la respuesta
	resp := buildControlResponse(req.Action, state)

	c.logger.Info().Str("action", req.Action).Str("state", resp.State).Msg("Operación del runner exitosa")
	response.Success(ctx, "Operación exitosa", resp)
}
