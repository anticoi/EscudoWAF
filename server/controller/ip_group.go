// server/controller/ip_group.go
package controller

import (
	"errors"
	"net/http"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// IPGroupController interfaz del controlador de grupos de IPs
type IPGroupController interface {
	CreateIPGroup(ctx *gin.Context)
	GetIPGroups(ctx *gin.Context)
	GetIPGroupByID(ctx *gin.Context)
	UpdateIPGroup(ctx *gin.Context)
	DeleteIPGroup(ctx *gin.Context)
	AddIPToBlacklist(ctx *gin.Context)
}

// IPGroupControllerImpl implementación del controlador de grupos de IPs
type IPGroupControllerImpl struct {
	ipGroupService service.IPGroupService
	logger         zerolog.Logger
}

// NewIPGroupController crea el controlador de grupos de IPs
func NewIPGroupController(ipGroupService service.IPGroupService) IPGroupController {
	logger := config.GetControllerLogger("ipgroup")
	return &IPGroupControllerImpl{
		ipGroupService: ipGroupService,
		logger:         logger,
	}
}

// CreateIPGroup Crear grupo de IPs
//
//	@Summary		Crear grupo de IPs
//	@Description	Crear un nuevo grupo de direcciones IP para su uso en reglas de IP
//	@Tags			Gestión de grupos de IPs
//	@Accept			json
//	@Produce		json
//	@Param			ipGroup	body	dto.IPGroupCreateRequest	true	"Información del grupo de IPs"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.IPGroup}	"Grupo de IPs creado con éxito"
//	@Failure		400	{object}	model.ErrResponse							"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError				"Acceso prohibido"
//	@Failure		409	{object}	model.ErrResponseDontShowError				"El nombre del grupo de IPs ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"Error interno del servidor"
//	@Router			/api/v1/ip-groups [post]
func (c *IPGroupControllerImpl) CreateIPGroup(ctx *gin.Context) {
	var req dto.IPGroupCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("name", req.Name).Msg("Petición de creación de grupo de IPs")
	ipGroup, err := c.ipGroupService.CreateIPGroup(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "El nombre del grupo de IPs ya existe", err), false)
			return
		}
		c.logger.Error().Err(err).Msg("Error al crear el grupo de IPs")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", ipGroup.ID.Hex()).Str("name", ipGroup.Name).Msg("Grupo de IPs creado con éxito")
	response.Success(ctx, "Grupo de IPs creado con éxito", ipGroup)
}

// GetIPGroups Obtener lista de grupos de IPs
//
//	@Summary		Obtener lista de grupos de IPs
//	@Description	Obtener todos los grupos de IPs, con paginación
//	@Tags			Gestión de grupos de IPs
//	@Produce		json
//	@Param			page	query	int	false	"Número de página"	default(1)
//	@Param			size	query	int	false	"Cantidad por página"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.IPGroupListResponse}	"Lista de grupos de IPs obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/ip-groups [get]
func (c *IPGroupControllerImpl) GetIPGroups(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	c.logger.Info().Str("page", page).Str("size", size).Msg("Petición de lista de grupos de IPs")
	ipGroups, total, err := c.ipGroupService.GetIPGroups(ctx, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al obtener la lista de grupos de IPs")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Int64("total", total).Msg("Lista de grupos de IPs obtenida con éxito")
	response.Success(ctx, "Lista de grupos de IPs obtenida con éxito", gin.H{
		"total": total,
		"items": ipGroups,
	})
}

// GetIPGroupByID Obtener un grupo de IPs
//
//	@Summary		Obtener un grupo de IPs
//	@Description	Obtener detalles del grupo de IPs por ID
//	@Tags			Gestión de grupos de IPs
//	@Produce		json
//	@Param			id	path	string	true	"ID del grupo de IPs"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.IPGroup}	"Detalles del grupo de IPs obtenidos con éxito"
//	@Failure		400	{object}	model.ErrResponse							"Formato de ID no válido"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"Acceso no autorizado"
//	@Failure		404	{object}	model.ErrResponseDontShowError				"El grupo de IPs no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"Error interno del servidor"
//	@Router			/api/v1/ip-groups/{id} [get]
func (c *IPGroupControllerImpl) GetIPGroupByID(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de detalles del grupo de IPs")

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	ipGroup, err := c.ipGroupService.GetIPGroupByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al obtener los detalles del grupo de IPs")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", ipGroup.Name).Msg("Detalles del grupo de IPs obtenidos con éxito")
	response.Success(ctx, "Detalles del grupo de IPs obtenidos con éxito", ipGroup)
}

// UpdateIPGroup Actualizar grupo de IPs
//
//	@Summary		Actualizar grupo de IPs
//	@Description	Actualizar la información del grupo de IPs especificado
//	@Tags			Gestión de grupos de IPs
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string						true	"ID del grupo de IPs"
//	@Param			ipGroup	body	dto.IPGroupUpdateRequest	true	"Información de actualización del grupo de IPs"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.IPGroup}	"Grupo de IPs actualizado con éxito"
//	@Failure		400	{object}	model.ErrResponse							"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError				"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError				"Prohibido operar sobre el grupo de IPs por defecto del sistema"
//	@Failure		404	{object}	model.ErrResponseDontShowError				"El grupo de IPs no existe"
//	@Failure		409	{object}	model.ErrResponseDontShowError				"El nombre del grupo de IPs ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError				"Error interno del servidor"
//	@Router			/api/v1/ip-groups/{id} [put]
func (c *IPGroupControllerImpl) UpdateIPGroup(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.IPGroupUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Str("id", id).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("id", id).Msg("Petición de actualización de grupo de IPs")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	ipGroup, err := c.ipGroupService.UpdateIPGroup(ctx, objectID, &req)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrIPGroupNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "El nombre del grupo de IPs ya existe", err), false)
			return
		} else if errors.Is(err, service.ErrSystemIPGroupNoMod) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "No se permite modificar el nombre del grupo de IPs por defecto del sistema", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al actualizar el grupo de IPs")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", ipGroup.Name).Msg("Grupo de IPs actualizado con éxito")
	response.Success(ctx, "Grupo de IPs actualizado con éxito", ipGroup)
}

// DeleteIPGroup Eliminar grupo de IPs
//
//	@Summary		Eliminar grupo de IPs
//	@Description	Eliminar el grupo de IPs especificado; los grupos por defecto del sistema no se pueden eliminar
//	@Tags			Gestión de grupos de IPs
//	@Produce		json
//	@Param			id	path	string	true	"ID del grupo de IPs"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"Grupo de IPs eliminado con éxito"
//	@Failure		400	{object}	model.ErrResponse				"Formato de ID no válido"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"Prohibido eliminar el grupo de IPs por defecto del sistema"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"El grupo de IPs no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/api/v1/ip-groups/{id} [delete]
func (c *IPGroupControllerImpl) DeleteIPGroup(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de eliminación de grupo de IPs")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	err = c.ipGroupService.DeleteIPGroup(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrSystemIPGroupNoMod) {
			response.Error(ctx, model.NewAPIError(http.StatusForbidden, "No se permite eliminar el grupo de IPs por defecto del sistema", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al eliminar el grupo de IPs")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Msg("Grupo de IPs eliminado con éxito")
	response.Success(ctx, "Grupo de IPs eliminado con éxito", nil)
}

// AddIPToBlacklist añade una IP a la lista negra por defecto del sistema
//
//	@Summary		Añadir IP a la lista negra
//	@Description	Añadir la dirección IP o CIDR especificada al grupo de lista negra por defecto del sistema
//	@Tags			Gestión de grupos de IPs
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.AddIPToBlacklistRequest	true	"Dirección IP o CIDR"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"IP añadida con éxito"
//	@Failure		400	{object}	model.ErrResponse				"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"La lista negra por defecto del sistema no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/api/v1/ip-groups/blacklist/add [post]
func (c *IPGroupControllerImpl) AddIPToBlacklist(ctx *gin.Context) {
	var req dto.AddIPToBlacklistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("ip", req.IP).Msg("Petición para añadir IP a la lista negra")
	err := c.ipGroupService.AddIPToBlacklist(ctx, req.IP)
	if err != nil {
		if errors.Is(err, service.ErrIPGroupNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("ip", req.IP).Msg("Error al añadir la IP a la lista negra")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("ip", req.IP).Msg("IP añadida a la lista negra con éxito")
	response.Success(ctx, "IP añadida a la lista negra con éxito", nil)
}
