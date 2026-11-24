// server/controller/rule.go
package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	pkgmodel "github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MicroRuleController interfaz del controlador de microrreglas
type MicroRuleController interface {
	CreateMicroRule(ctx *gin.Context)
	GetMicroRules(ctx *gin.Context)
	GetMicroRuleByID(ctx *gin.Context)
	UpdateMicroRule(ctx *gin.Context)
	DeleteMicroRule(ctx *gin.Context)
}

// MicroRuleControllerImpl implementación del controlador de microrreglas
type MicroRuleControllerImpl struct {
	ruleService service.MicroRuleService
	logger      zerolog.Logger
}

// NewMicroRuleController crea el controlador de microrreglas
func NewMicroRuleController(ruleService service.MicroRuleService) MicroRuleController {
	logger := config.GetControllerLogger("microrule")
	return &MicroRuleControllerImpl{
		ruleService: ruleService,
		logger:      logger,
	}
}

// BSONToJSON convierte datos BSON a JSON
func BSONToJSON(bsonData bson.Raw) (json.RawMessage, error) {
	if len(bsonData) == 0 {
		return nil, nil
	}

	// Convertir BSON a interface{}
	var anyValue interface{}
	if err := bson.Unmarshal(bsonData, &anyValue); err != nil {
		return nil, err
	}

	// Convertir interface{} a JSON
	jsonData, err := json.Marshal(anyValue)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// ConvertToResponse convierte el modelo a un objeto de respuesta DTO
func ConvertToResponse(rule *pkgmodel.MicroRule) (*dto.MicroRuleResponse, error) {
	var jsonCondition json.RawMessage
	var err error

	if len(rule.Condition) > 0 {
		jsonCondition, err = BSONToJSON(rule.Condition)
		if err != nil {
			return nil, err
		}
	}

	return &dto.MicroRuleResponse{
		ID:        rule.ID.Hex(),
		Name:      rule.Name,
		Type:      string(rule.Type),
		Status:    string(rule.Status),
		Priority:  &rule.Priority,
		Condition: jsonCondition,
	}, nil
}

// CreateMicroRule Crear microrregla
//
//	@Summary		Crear microrregla
//	@Description	Crear una nueva microrregla WAF para filtrar peticiones
//	@Tags			Gestión de reglas
//	@Accept			json
//	@Produce		json
//	@Param			rule	body	dto.MicroRuleCreateRequest	true	"Información de la microrregla"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.MicroRuleResponse}	"Microrregla creada con éxito"
//	@Failure		400	{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError						"Acceso prohibido"
//	@Failure		409	{object}	model.ErrResponseDontShowError						"El nombre de la microrregla ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/micro-rules [post]
func (c *MicroRuleControllerImpl) CreateMicroRule(ctx *gin.Context) {
	var req dto.MicroRuleCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("name", req.Name).Msg("Petición de creación de microrregla")
	rule, err := c.ruleService.CreateMicroRule(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "El nombre de la microrregla ya existe", err), false)
			return
		}
		c.logger.Error().Err(err).Msg("Error al crear la microrregla")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a DTO de respuesta
	resp, err := ConvertToResponse(rule)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al convertir el objeto de respuesta")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", rule.ID.Hex()).Str("name", rule.Name).Msg("Microrregla creada con éxito")
	response.Success(ctx, "Microrregla creada con éxito", resp)
}

// GetMicroRules Obtener lista de microrreglas
//
//	@Summary		Obtener lista de microrreglas
//	@Description	Obtener todas las microrreglas del WAF, con paginación
//	@Tags			Gestión de reglas
//	@Produce		json
//	@Param			page	query	int	false	"Número de página"	default(1)
//	@Param			size	query	int	false	"Cantidad por página"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.MicroRuleListResponse}	"Lista de microrreglas obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"Error interno del servidor"
//	@Router			/api/v1/micro-rules [get]
func (c *MicroRuleControllerImpl) GetMicroRules(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	c.logger.Info().Str("page", page).Str("size", size).Msg("Petición de lista de microrreglas")
	rules, total, err := c.ruleService.GetMicroRules(ctx, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al obtener la lista de microrreglas")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir el objeto de respuesta
	responses := make([]*dto.MicroRuleResponse, len(rules))
	for i, rule := range rules {
		resp, err := ConvertToResponse(&rule)
		if err != nil {
			c.logger.Error().Err(err).Msg("Error al convertir el objeto de respuesta")
			response.InternalServerError(ctx, err, false)
			return
		}
		responses[i] = resp
	}

	c.logger.Info().Int64("total", total).Msg("Lista de microrreglas obtenida con éxito")
	response.Success(ctx, "Lista de microrreglas obtenida con éxito", gin.H{
		"total": total,
		"items": responses,
	})
}

// GetMicroRuleByID Obtener una microrregla
//
//	@Summary		Obtener una microrregla
//	@Description	Obtener detalles de la microrregla por ID
//	@Tags			Gestión de reglas
//	@Produce		json
//	@Param			id	path	string	true	"ID de la microrregla"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.MicroRuleResponse}	"Detalles de la microrregla obtenidos con éxito"
//	@Failure		400	{object}	model.ErrResponse									"Formato de ID no válido"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		404	{object}	model.ErrResponseDontShowError						"La microrregla no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/micro-rules/{id} [get]
func (c *MicroRuleControllerImpl) GetMicroRuleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de detalles de la microrregla")

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	rule, err := c.ruleService.GetMicroRuleByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al obtener los detalles de la microrregla")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a DTO de respuesta
	resp, err := ConvertToResponse(rule)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al convertir el objeto de respuesta")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", rule.Name).Msg("Detalles de la microrregla obtenidos con éxito")
	response.Success(ctx, "Detalles de la microrregla obtenidos con éxito", resp)
}

// UpdateMicroRule Actualizar microrregla
//
//	@Summary		Actualizar microrregla
//	@Description	Actualizar la microrregla especificada; las reglas por defecto del sistema no se pueden modificar
//	@Tags			Gestión de reglas
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string						true	"ID de la microrregla"
//	@Param			rule	body	dto.MicroRuleUpdateRequest	true	"Información de actualización de la microrregla"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.MicroRuleResponse}	"Microrregla actualizada con éxito"
//	@Failure		400	{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError						"Prohibido modificar las reglas por defecto del sistema"
//	@Failure		404	{object}	model.ErrResponseDontShowError						"La microrregla no existe"
//	@Failure		409	{object}	model.ErrResponseDontShowError						"El nombre de la microrregla ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/micro-rules/{id} [put]
func (c *MicroRuleControllerImpl) UpdateMicroRule(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.MicroRuleUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Str("id", id).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("id", id).Msg("Petición de actualización de microrregla")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	rule, err := c.ruleService.UpdateMicroRule(ctx, objectID, &req)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrMicroRuleNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "El nombre de la microrregla ya existe", err), false)
			return
		} else if errors.Is(err, service.ErrSystemRuleNoMod) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "No se permite modificar las reglas por defecto del sistema", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al actualizar la microrregla")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a DTO de respuesta
	resp, err := ConvertToResponse(rule)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al convertir el objeto de respuesta")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", rule.Name).Msg("Microrregla actualizada con éxito")
	response.Success(ctx, "Microrregla actualizada con éxito", resp)
}

// DeleteMicroRule Eliminar microrregla
//
//	@Summary		Eliminar microrregla
//	@Description	Eliminar la microrregla especificada; las reglas por defecto del sistema no se pueden eliminar
//	@Tags			Gestión de reglas
//	@Produce		json
//	@Param			id	path	string	true	"ID de la microrregla"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"Microrregla eliminada con éxito"
//	@Failure		400	{object}	model.ErrResponse				"Formato de ID no válido"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"Prohibido eliminar las reglas por defecto del sistema"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"La microrregla no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/api/v1/micro-rules/{id} [delete]
func (c *MicroRuleControllerImpl) DeleteMicroRule(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de eliminación de microrregla")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	err = c.ruleService.DeleteMicroRule(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrMicroRuleNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrSystemRuleNoDelete) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "No se permite eliminar las reglas por defecto del sistema", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al eliminar la microrregla")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Msg("Microrregla eliminada con éxito")
	response.Success(ctx, "Microrregla eliminada con éxito", nil)
}
