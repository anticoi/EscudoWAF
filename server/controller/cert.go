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

// CertificateController interfaz del controlador de certificados
type CertificateController interface {
	CreateCertificate(ctx *gin.Context)
	GetCertificates(ctx *gin.Context)
	GetCertificateByID(ctx *gin.Context)
	UpdateCertificate(ctx *gin.Context)
	DeleteCertificate(ctx *gin.Context)
}

// CertificateControllerImpl implementación del controlador de certificados
type CertificateControllerImpl struct {
	certService service.CertificateService
	logger      zerolog.Logger
}

// NewCertificateController crea el controlador de certificados
func NewCertificateController(certService service.CertificateService) CertificateController {
	logger := config.GetControllerLogger("certificate")
	return &CertificateControllerImpl{
		certService: certService,
		logger:      logger,
	}
}

// CreateCertificate crea un certificado
//
//	@Summary		Crear nuevo certificado
//	@Description	Crear un nuevo certificado SSL/TLS
//	@Tags			Gestión de certificados
//	@Accept			json
//	@Produce		json
//	@Param			certificate	body	dto.CertificateCreateRequest	true	"Información del certificado"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.CertificateStore}	"Certificado creado con éxito"
//	@Failure		400	{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError						"Acceso prohibido"
//	@Failure		409	{object}	model.ErrResponseDontShowError						"El nombre del certificado ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/certificates [post]
func (c *CertificateControllerImpl) CreateCertificate(ctx *gin.Context) {
	var req dto.CertificateCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("name", req.Name).Msg("Petición de creación de certificado")
	cert, err := c.certService.CreateCertificate(ctx, &req)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "El nombre del certificado ya existe", err), false)
			return
		} else if errors.Is(err, service.ErrInvalidCertificate) {
			response.BadRequest(ctx, err, true)
			return
		}
		c.logger.Error().Err(err).Msg("Error al crear el certificado")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a respuesta DTO
	certResponse := model.CertificateStore{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		PublicKey:   cert.PublicKey,
		PrivateKey:  cert.PrivateKey,
		ExpireDate:  cert.ExpireDate,
		IssuerName:  cert.IssuerName,
		FingerPrint: cert.FingerPrint,
		Domains:     cert.Domains,
		CreatedAt:   cert.CreatedAt,
		UpdatedAt:   cert.UpdatedAt,
	}

	c.logger.Info().Str("id", cert.ID.Hex()).Str("name", cert.Name).Msg("Certificado creado con éxito")
	response.Success(ctx, "Certificado creado con éxito", certResponse)
}

// GetCertificates Obtener lista de certificados
//
//	@Summary		Obtener lista de certificados
//	@Description	Obtener todos los certificados SSL/TLS, con paginación
//	@Tags			Gestión de certificados
//	@Produce		json
//	@Param			page	query	int	false	"Número de página"	default(1)
//	@Param			size	query	int	false	"Cantidad por página"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=dto.CertificateListResponse}	"Lista de certificados obtenida con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError							"Acceso no autorizado"
//	@Failure		500	{object}	model.ErrResponseDontShowError							"Error interno del servidor"
//	@Router			/api/v1/certificates [get]
func (c *CertificateControllerImpl) GetCertificates(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	c.logger.Info().Str("page", page).Str("size", size).Msg("Petición de lista de certificados")
	certificates, total, err := c.certService.GetCertificates(ctx, page, size)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error al obtener la lista de certificados")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Int64("total", total).Msg("Lista de certificados obtenida con éxito")
	response.Success(ctx, "Lista de certificados obtenida con éxito", gin.H{
		"total": total,
		"items": certificates,
	})
}

// GetCertificateByID Obtener un certificado
//
//	@Summary		Obtener un certificado
//	@Description	Obtener detalles del certificado por ID
//	@Tags			Gestión de certificados
//	@Produce		json
//	@Param			id	path	string	true	"ID del certificado"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.CertificateStore}	"Detalles del certificado obtenidos con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		404	{object}	model.ErrResponseDontShowError						"El certificado no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/certificates/{id} [get]
func (c *CertificateControllerImpl) GetCertificateByID(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de detalles del certificado")

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	cert, err := c.certService.GetCertificateByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al obtener los detalles del certificado")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a respuesta DTO
	certResponse := model.CertificateStore{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		PublicKey:   cert.PublicKey,
		PrivateKey:  cert.PrivateKey,
		ExpireDate:  cert.ExpireDate,
		IssuerName:  cert.IssuerName,
		FingerPrint: cert.FingerPrint,
		Domains:     cert.Domains,
		CreatedAt:   cert.CreatedAt,
		UpdatedAt:   cert.UpdatedAt,
	}

	c.logger.Info().Str("id", id).Str("name", cert.Name).Msg("Detalles del certificado obtenidos con éxito")
	response.Success(ctx, "Detalles del certificado obtenidos con éxito", certResponse)
}

// UpdateCertificate Actualizar certificado
//
//	@Summary		Actualizar certificado
//	@Description	Actualizar la información del certificado especificado
//	@Tags			Gestión de certificados
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string							true	"ID del certificado"
//	@Param			certificate	body	dto.CertificateUpdateRequest	true	"Información de actualización del certificado"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponse{data=model.CertificateStore}	"Certificado actualizado con éxito"
//	@Failure		400	{object}	model.ErrResponse									"Parámetros de la petición incorrectos"
//	@Failure		401	{object}	model.ErrResponseDontShowError						"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError						"Acceso prohibido"
//	@Failure		404	{object}	model.ErrResponseDontShowError						"El certificado no existe"
//	@Failure		409	{object}	model.ErrResponseDontShowError						"El nombre del certificado ya existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError						"Error interno del servidor"
//	@Router			/api/v1/certificates/{id} [put]
func (c *CertificateControllerImpl) UpdateCertificate(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.CertificateUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Str("id", id).Msg("Error al vincular los parámetros de la petición")
		response.BadRequest(ctx, err, true)
		return
	}

	c.logger.Info().Str("id", id).Msg("Petición de actualización de certificado")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	cert, err := c.certService.UpdateCertificate(ctx, objectID, &req)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNotFound) {
			response.NotFound(ctx, err)
			return
		} else if errors.Is(err, service.ErrCertificateNameExists) {
			response.Error(ctx, model.NewAPIError(http.StatusConflict, "El nombre del certificado ya existe", err), false)
			return
		} else if errors.Is(err, service.ErrInvalidCertificate) {
			response.BadRequest(ctx, err, true)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al actualizar el certificado")
		response.InternalServerError(ctx, err, false)
		return
	}

	// Convertir a respuesta DTO
	certResponse := model.CertificateStore{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		PublicKey:   cert.PublicKey,
		PrivateKey:  cert.PrivateKey,
		ExpireDate:  cert.ExpireDate,
		IssuerName:  cert.IssuerName,
		FingerPrint: cert.FingerPrint,
		Domains:     cert.Domains,
		CreatedAt:   cert.CreatedAt,
		UpdatedAt:   cert.UpdatedAt,
	}

	c.logger.Info().Str("id", id).Str("name", cert.Name).Msg("Certificado actualizado con éxito")
	response.Success(ctx, "Certificado actualizado con éxito", certResponse)
}

// ... existing code ...

// DeleteCertificate Eliminar certificado
//
//	@Summary		Eliminar certificado
//	@Description	Eliminar el certificado SSL/TLS especificado
//	@Tags			Gestión de certificados
//	@Produce		json
//	@Param			id	path	string	true	"ID del certificado"
//	@Security		BearerAuth
//	@Success		200	{object}	model.SuccessResponseNoData		"Certificado eliminado con éxito"
//	@Failure		401	{object}	model.ErrResponseDontShowError	"Acceso no autorizado"
//	@Failure		403	{object}	model.ErrResponseDontShowError	"Acceso prohibido"
//	@Failure		404	{object}	model.ErrResponseDontShowError	"El certificado no existe"
//	@Failure		500	{object}	model.ErrResponseDontShowError	"Error interno del servidor"
//	@Router			/api/v1/certificates/{id} [delete]
func (c *CertificateControllerImpl) DeleteCertificate(ctx *gin.Context) {
	id := ctx.Param("id")

	c.logger.Info().Str("id", id).Msg("Petición de eliminación de certificado")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.logger.Error().Err(err).Str("id", id).Msg("Formato de ID no válido")
		response.BadRequest(ctx, err, true)
		return
	}
	err = c.certService.DeleteCertificate(ctx, objectID)
	if err != nil {
		if errors.Is(err, service.ErrCertificateNotFound) {
			response.NotFound(ctx, err)
			return
		}
		c.logger.Error().Err(err).Str("id", id).Msg("Error al eliminar el certificado")
		response.InternalServerError(ctx, err, false)
		return
	}

	c.logger.Info().Str("id", id).Msg("Certificado eliminado con éxito")
	response.Success(ctx, "Certificado eliminado con éxito", nil)
}
