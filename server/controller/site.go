package controller

import (
	"errors"
	"net/http"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/anticoi/EscudoWAF/server/service"
	"github.com/anticoi/EscudoWAF/server/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SiteController interfaz del controlador de sitios
type SiteController interface {
	CreateSite(ctx *gin.Context)
	GetSites(ctx *gin.Context)
	GetSiteByID(ctx *gin.Context)
	UpdateSite(ctx *gin.Context)
	DeleteSite(ctx *gin.Context)
}

// SiteControllerImpl implementación del controlador de sitios
type SiteControllerImpl struct {
	siteService service.SiteService
	logger      zerolog.Logger
}

// NewSiteController crea el controlador de sitios
func NewSiteController(siteService service.SiteService) SiteController {
	logger := config.GetControllerLogger("site")
	return &SiteControllerImpl{
		siteService: siteService,
		logger:      logger,
	}
}

// CreateSite crea un sitio
//
//	@Summary		Crear nuevo sitio
//	@Description	Crear una nueva configuración de sitio
//	@Tags			Gestión de sitios
//	@Accept			json
//	@Produce		json
//	@Param			site	body	dto.CreateSiteRequest	true	"Información del sitio"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.SiteResponse}	"Sitio creado con éxito"
//	@Failure		400	{object}	model.ErrResponse								"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError					"Acceso prohibido"
//	@Failure		409	{object}	model.ErrResponseDontShowError					"La combinación de dominio y puerto ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"Error interno del servidor"
//	@Router			/api/v1/site [post]
func (c *SiteControllerImpl) CreateSite(ctx *gin.Context) {
	var req dto.CreateSiteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("name", req.Name).Str("domain", req.Domain).Msg("Petición de creación de sitio")

	site, err := c.siteService.CreateSite(ctx, &req)
	if err != nil {
		if errors.Is(err, repository.ErrDomainPortExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "La combinación de dominio y puerto ya existe", err), false)
			return
		}
		c.logger.Error().Err(err).Msg("Error al crear el sitio")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", site.ID.Hex()).Str("name", site.Name).Msg("Sitio creado con éxito")
	response.Success(ctx, "Sitio creado con éxito", site)
}

// GetSites Obtener lista de sitios
//
//	@Summary		Obtener lista de sitios
//	@Description	Obtener la lista de todas las configuraciones de sitio
//	@Tags			Gestión de sitios
//	@Produce		json
//	@Param			page	query	int	false	"Número de página"	default(1)
//	@Param			size	query	int	false	"Cantidad por página"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.SiteListResponse}	"Lista de sitios obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/site [get]
func (c *SiteControllerImpl) GetSites(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	c.logger.Info().Str("page", page).Str("size", size).Msg("Petición de lista de sitios")
	sites, total, err := c.siteService.GetSites(ctx, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al obtener la lista de sitios")
		response.InternalServerError(ctx, err, false)
		return
	}

	if sites == nil {
		sites = []model.Site{}
	}

	c.logger.Info().Int64("total", total).Msg("Lista de sitios obtenida con éxito")
	response.Success(ctx, "Lista de sitios obtenida con éxito", gin.H{
		"total": total,
		"items": sites,
	})
}

// GetSiteByID Obtener un sitio
//
//	@Summary		Obtener un sitio
//	@Description	Obtener detalles del sitio por ID
//	@Tags			Gestión de sitios
//	@Produce		json
//	@Param			id	path	string	true	"ID del sitio"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.SiteResponse}	"Detalles del sitio obtenidos con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"Acceso no autorizado"
//	@Failure		404	{object}	model.ErrResponseDontShowError					"El sitio no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"Error interno del servidor"
//	@Router			/api/v1/site/{id} [get]
func (c *SiteControllerImpl) GetSiteByID(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de detalles del sitio")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	site, err := c.siteService.GetSiteByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, repository.ErrSiteNotFound) {
			response.Error(ctx, model.NewAPIError(http.StatusNotFound, "El sitio no existe", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al obtener los detalles del sitio")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", site.Name).Msg("Detalles del sitio obtenidos con éxito")
	response.Success(ctx, "Detalles del sitio obtenidos con éxito", site)
}

// UpdateSite Actualizar sitio
//
//	@Summary		Actualizar sitio
//	@Description	Actualizar la configuración del sitio especificado
//	@Tags			Gestión de sitios
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string					true	"ID del sitio"
//	@Param			site	body	dto.UpdateSiteRequest	true	"Información de actualización del sitio"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.SiteResponse}	"Sitio actualizado con éxito"
//	@Failure		400	{object}	model.ErrResponse								"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError					"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError					"Acceso prohibido"
//	@Failure		404	{object}	model.ErrResponseDontShowError					"El sitio no existe"
//	@Failure		409	{object}	model.ErrResponseDontShowError					"La combinación de dominio y puerto ya está en uso por otro sitio"
//	@Failure		500	{object}	model.ErrResponseDontShowError					"Error interno del servidor"
//	@Router			/api/v1/site/{id} [put]
func (c *SiteControllerImpl) UpdateSite(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.UpdateSiteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Str("id", id).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("id", id).Msg("Petición de actualización de sitio")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	site, err := c.siteService.UpdateSite(ctx, objectID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrSiteNotFound) {
			response.Error(ctx, model.NewAPIError(http.StatusNotFound, "El sitio no existe", err), false)
			return
		} else if errors.Is(err, repository.ErrDomainPortConflict) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "La combinación de dominio y puerto ya está en uso por otro sitio", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al actualizar el sitio")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Str("name", site.Name).Msg("Sitio actualizado con éxito")
	response.Success(ctx, "Sitio actualizado con éxito", site)
}

// DeleteSite Eliminar sitio
//
//	@Summary		Eliminar sitio
//	@Description	Eliminar la configuración de sitio especificada
//	@Tags			Gestión de sitios
//	@Produce		json
//	@Param			id	path	string	true	"ID del sitio"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"Sitio eliminado con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"Acceso prohibido"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"El sitio no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/api/v1/site/{id} [delete]
func (c *SiteControllerImpl) DeleteSite(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de eliminación de sitio")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	err = c.siteService.DeleteSite(ctx, objectID)
	if err != nil {
		if errors.Is(err, repository.ErrSiteNotFound) {
			response.Error(ctx, model.NewAPIError(http.StatusNotFound, "El sitio no existe", err), false)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al eliminar el sitio")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Msg("Sitio eliminado con éxito")
	response.Success(ctx, "Sitio eliminado con éxito", nil)
}
