package controller

import (
	"time"

	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
)

type WAFLogController interface {
	GetAttackEvents(ctx *gin.Context)
	GetAttackLogs(ctx *gin.Context)
}

type WAFLogControllerImpl struct {
	wafLogService service.WAFLogService
}

// NewWAFLogController crea una nueva instancia del controlador de logs del WAF
func NewWAFLogController(wafLogService service.WAFLogService) WAFLogController {
	return &WAFLogControllerImpl{
		wafLogService: wafLogService,
	}
}

// GetAttackEvents godoc
//
//	@Summary		Obtener eventos de ataque agregados
//	@Description	Estadísticas de eventos de ataque agregados por IP de origen, puerto destino y dominio, con filtrado multidimensional y paginación
//	@Tags			Registros de seguridad del WAF
//	@Accept			json
//	@Produce		json
//	@Param			srcIp		query		string												false	"Dirección IP de origen, dirección del atacante"
//	@Param			dstIp		query		string												false	"Dirección IP de destino, dirección del servidor atacado"
//	@Param			domain		query		string												false	"Dominio del sitio atacado"
//	@Param			srcPort		query		integer												false	"Puerto de origen desde el que se inició el ataque"
//	@Param			dstPort		query		integer												false	"Puerto de destino, puerto del servicio atacado"
//	@Param			startTime	query		string												false	"Hora de inicio de la consulta (formato ISO8601, p. ej.: 2024-03-17T00:00:00Z)"
//	@Param			endTime		query		string												false	"Hora de fin de la consulta (formato ISO8601, p. ej.: 2024-03-18T23:59:59Z)"
//	@Param			page		query		integer												false	"Página actual, comenzando en 1 (por defecto: 1)"
//	@Param			pageSize	query		integer												false	"Registros por página, máximo 100 (por defecto: 10)"
//	@Success		200			{object}	model.SuccessResponse{data=dto.AttackEventResponse}	"Éxito"
//	@Failure		400			{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		500			{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/waf/logs/events [get]
func (c *WAFLogControllerImpl) GetAttackEvents(ctx *gin.Context) {
	var req dto.AttackEventRequset

	// Usar ShouldBindQuery para vincular los parámetros de consulta
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	// Usar la zona horaria UTC al establecer valores por defecto
	if req.StartTime.IsZero() {
		// Por defecto: hace 24 horas, zona horaria UTC
		req.StartTime = time.Now().UTC().Add(-24 * time.Hour)
	}
	if req.EndTime.IsZero() {
		// Por defecto: hora actual, zona horaria UTC
		req.EndTime = time.Now().UTC()
	}

	// Establecer los parámetros de paginación por defecto
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100 // Limitar la cantidad máxima por página
	}

	// Llamar al servicio
	result, err := c.wafLogService.GetAttackEvents(ctx, req, page, pageSize)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Eventos de ataque obtenidos con éxito", result)
}

// GetAttackLogs godoc
//
//	@Summary		Obtener logs de ataque detallados
//	@Description	Consultar registros detallados de ataques del WAF, con filtros múltiples y paginación, por ID de regla, IP, dominio, puerto y rango de tiempo
//	@Tags			Registros de seguridad del WAF
//	@Accept			json
//	@Produce		json
//	@Param			ruleId		query		integer												false	"ID de la regla WAF que detectó el ataque"
//	@Param			srcIp		query		string												false	"Dirección IP de origen, dirección del atacante"
//	@Param			dstIp		query		string												false	"Dirección IP de destino, dirección del servidor atacado"
//	@Param			domain		query		string												false	"Dominio del sitio atacado"
//	@Param			srcPort		query		integer												false	"Puerto de origen desde el que se inició el ataque"
//	@Param			dstPort		query		integer												false	"Puerto de destino, puerto del servicio atacado"
//	@Param			requestId	query		string												false	"ID de la petición, identificador único de la petición HTTP"
//	@Param			startTime	query		string												false	"Hora de inicio de la consulta (formato ISO8601, p. ej.: 2024-03-17T00:00:00Z)"
//	@Param			endTime		query		string												false	"Hora de fin de la consulta (formato ISO8601, p. ej.: 2024-03-18T23:59:59Z)"
//	@Param			page		query		integer												false	"Página actual, comenzando en 1 (por defecto: 1)"
//	@Param			pageSize	query		integer												false	"Registros por página, máximo 100 (por defecto: 10)"
//	@Success		200			{object}	model.SuccessResponse{data=dto.AttackLogResponse}	"Éxito"
//	@Failure		400			{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		500			{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/waf/logs [get]
func (c *WAFLogControllerImpl) GetAttackLogs(ctx *gin.Context) {
	var req dto.AttackLogRequest

	// Usar ShouldBindQuery para vincular los parámetros de consulta
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, err, true)
		return
	}

	// Usar la zona horaria UTC al establecer valores por defecto
	if req.StartTime.IsZero() {
		// Por defecto: hace 24 horas, zona horaria UTC
		req.StartTime = time.Now().UTC().Add(-24 * time.Hour)
	}
	if req.EndTime.IsZero() {
		// Por defecto: hora actual, zona horaria UTC
		req.EndTime = time.Now().UTC()
	}

	// Establecer los parámetros de paginación por defecto
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100 // Limitar la cantidad máxima por página
	}

	// Llamar al servicio
	result, err := c.wafLogService.GetAttackLogs(ctx, req, page, pageSize)
	if err != nil {
		response.InternalServerError(ctx, err, false)
		return
	}

	response.Success(ctx, "Logs de ataque obtenidos con éxito", result)
}
