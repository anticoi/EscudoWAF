// server/service/ip_group.go
package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/dto"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	SystemDefaultBlacklistName = "system_default_blacklist" // Nombre del grupo de lista negra por defecto del sistema
)

var (
	ErrIPGroupNotFound    = errors.New("El grupo de IPs no existe")
	ErrIPGroupNameExists  = errors.New("El nombre del grupo de IPs ya existe")
	ErrSystemIPGroupNoMod = errors.New("No se permite eliminar el grupo de IPs por defecto del sistema")
)

// IPGroupService interfaz del servicio de grupos de IPs
type IPGroupService interface {
	CreateIPGroup(ctx context.Context, req *dto.IPGroupCreateRequest) (*model.IPGroup, error)
	GetIPGroups(ctx context.Context, pageStr, sizeStr string) ([]model.IPGroup, int64, error)
	GetIPGroupByID(ctx context.Context, id bson.ObjectID) (*model.IPGroup, error)
	UpdateIPGroup(ctx context.Context, id bson.ObjectID, req *dto.IPGroupUpdateRequest) (*model.IPGroup, error)
	DeleteIPGroup(ctx context.Context, id bson.ObjectID) error
	AddIPToBlacklist(ctx context.Context, ip string) error
}

// IPGroupServiceImpl implementación del servicio de grupos de IPs
type IPGroupServiceImpl struct {
	ipGroupRepo repository.IPGroupRepository
	logger      zerolog.Logger
}

// NewIPGroupService crea el servicio de grupos de IPs
func NewIPGroupService(ipGroupRepo repository.IPGroupRepository) IPGroupService {
	logger := config.GetServiceLogger("ipgroup")
	return &IPGroupServiceImpl{
		ipGroupRepo: ipGroupRepo,
		logger:      logger,
	}
}

// CreateIPGroup Crear grupo de IPs
func (s *IPGroupServiceImpl) CreateIPGroup(ctx context.Context, req *dto.IPGroupCreateRequest) (*model.IPGroup, error) {
	// Comprobar si el nombre del grupo de IPs ya existe
	if req.Name != "" {
		exists, err := s.ipGroupRepo.CheckIPGroupNameExists(ctx, req.Name, bson.NilObjectID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrIPGroupNameExists
		}
	}

	// Crear nuevo grupo de IPs
	ipGroup := &model.IPGroup{
		Name:  req.Name,
		Items: req.Items,
	}

	// Guardar grupo de IPs
	err := s.ipGroupRepo.CreateIPGroup(ctx, ipGroup)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al crear el grupo de IPs")
		return nil, err
	}

	s.logger.Info().Str("id", ipGroup.ID.Hex()).Str("name", ipGroup.Name).Msg("Grupo de IPs creado con éxito")
	return ipGroup, nil
}

// GetIPGroups Obtener lista de grupos de IPs
func (s *IPGroupServiceImpl) GetIPGroups(ctx context.Context, pageStr, sizeStr string) ([]model.IPGroup, int64, error) {
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 10
	}

	ipGroups, total, err := s.ipGroupRepo.GetIPGroups(ctx, page, size)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al obtener la lista de grupos de IPs")
		return nil, 0, err
	}

	return ipGroups, total, nil
}

// GetIPGroupByID obtiene un grupo de IPs por ID
func (s *IPGroupServiceImpl) GetIPGroupByID(ctx context.Context, id bson.ObjectID) (*model.IPGroup, error) {
	ipGroup, err := s.ipGroupRepo.GetIPGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrIPGroupNotFound) {
			return nil, ErrIPGroupNotFound
		}
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al obtener el grupo de IPs")
		return nil, err
	}

	return ipGroup, nil
}

// UpdateIPGroup Actualizar grupo de IPs
func (s *IPGroupServiceImpl) UpdateIPGroup(ctx context.Context, id bson.ObjectID, req *dto.IPGroupUpdateRequest) (*model.IPGroup, error) {
	// Obtener el grupo de IPs existente
	ipGroup, err := s.ipGroupRepo.GetIPGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrIPGroupNotFound) {
			return nil, ErrIPGroupNotFound
		}
		return nil, err
	}

	// Comprobar si es el grupo de IPs por defecto del sistema
	if ipGroup.Name == SystemDefaultBlacklistName {
		s.logger.Warn().Str("id", id.Hex()).Msg("Se intentó modificar el grupo de IPs por defecto del sistema")
		// Si es el grupo de IPs por defecto del sistema, solo se permite actualizar Items, no Name
		if req.Name != "" && req.Name != SystemDefaultBlacklistName {
			return nil, ErrSystemIPGroupNoMod
		}
	}

	// Comprobar si el nombre del grupo de IPs ya existe (si se va a actualizar el nombre)
	if req.Name != "" && req.Name != ipGroup.Name {
		exists, err := s.ipGroupRepo.CheckIPGroupNameExists(ctx, req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrIPGroupNameExists
		}
		ipGroup.Name = req.Name
	}

	// Actualizar la lista de IPs (solo campos no vacíos)
	if req.Items != nil {
		ipGroup.Items = req.Items
	}

	// Guardar actualización
	err = s.ipGroupRepo.UpdateIPGroup(ctx, ipGroup)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al actualizar el grupo de IPs")
		return nil, err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", ipGroup.Name).Msg("Grupo de IPs actualizado con éxito")
	return ipGroup, nil
}

// DeleteIPGroup Eliminar grupo de IPs
func (s *IPGroupServiceImpl) DeleteIPGroup(ctx context.Context, id bson.ObjectID) error {
	// Comprobar si el grupo de IPs existe
	ipGroup, err := s.ipGroupRepo.GetIPGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrIPGroupNotFound) {
			return ErrIPGroupNotFound
		}
		return err
	}

	// Comprobar si es el grupo de IPs por defecto del sistema
	if ipGroup.Name == SystemDefaultBlacklistName {
		s.logger.Warn().Str("id", id.Hex()).Msg("Se intentó eliminar el grupo de IPs por defecto del sistema")
		return ErrSystemIPGroupNoMod
	}

	// Eliminar grupo de IPs
	err = s.ipGroupRepo.DeleteIPGroup(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar el grupo de IPs")
		return err
	}

	s.logger.Info().Str("id", id.Hex()).Msg("Grupo de IPs eliminado con éxito")
	return nil
}

// AddIPToBlacklist añade una IP a la lista negra por defecto del sistema
func (s *IPGroupServiceImpl) AddIPToBlacklist(ctx context.Context, ip string) error {
	// Buscar el grupo de lista negra por defecto del sistema
	ipGroup, err := s.ipGroupRepo.GetIPGroupByName(ctx, SystemDefaultBlacklistName)
	if err != nil {
		if errors.Is(err, repository.ErrIPGroupNotFound) {
			s.logger.Error().Msg("El grupo de lista negra por defecto del sistema no existe")
			return ErrIPGroupNotFound
		}
		s.logger.Error().Err(err).Msg("Error al obtener el grupo de lista negra por defecto del sistema")
		return err
	}

	// Comprobar si la IP ya existe
	for _, existingIP := range ipGroup.Items {
		if existingIP == ip {
			s.logger.Info().Str("ip", ip).Msg("La IP ya existe en la lista negra")
			return nil // Si ya existe, devolver éxito directamente
		}
	}

	// Añadir IP a la lista negra
	ipGroup.Items = append(ipGroup.Items, ip)

	// Actualizar grupo de IPs
	err = s.ipGroupRepo.UpdateIPGroup(ctx, ipGroup)
	if err != nil {
		s.logger.Error().Err(err).Str("ip", ip).Msg("Error al añadir la IP a la lista negra")
		return err
	}

	s.logger.Info().Str("ip", ip).Msg("IP añadida a la lista negra correctamente")
	return nil
}
