// server/service/rule.go
package service

import (
	"context"
	"encoding/json"
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
	SystemDefaultIPBlockRule = "system_default_ip_block" // Nombre de la regla de bloqueo de IP por defecto del sistema
)

var (
	ErrMicroRuleNotFound   = errors.New("La microrregla no existe")
	ErrMicroRuleNameExists = errors.New("El nombre de la microrregla ya existe")
	ErrSystemRuleNoMod     = errors.New("No se permite modificar las reglas por defecto del sistema")
	ErrSystemRuleNoDelete  = errors.New("No se permite eliminar las reglas por defecto del sistema")
)

// MicroRuleService interfaz del servicio de microrreglas
type MicroRuleService interface {
	CreateMicroRule(ctx context.Context, req *dto.MicroRuleCreateRequest) (*model.MicroRule, error)
	GetMicroRules(ctx context.Context, pageStr, sizeStr string) ([]model.MicroRule, int64, error)
	GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error)
	UpdateMicroRule(ctx context.Context, id bson.ObjectID, req *dto.MicroRuleUpdateRequest) (*model.MicroRule, error)
	DeleteMicroRule(ctx context.Context, id bson.ObjectID) error
}

// MicroRuleServiceImpl implementación del servicio de microrreglas
type MicroRuleServiceImpl struct {
	ruleRepo repository.MicroRuleRepository
	logger   zerolog.Logger
}

// NewMicroRuleService crea el servicio de microrreglas
func NewMicroRuleService(ruleRepo repository.MicroRuleRepository) MicroRuleService {
	logger := config.GetServiceLogger("microrule")
	return &MicroRuleServiceImpl{
		ruleRepo: ruleRepo,
		logger:   logger,
	}
}

// CreateMicroRule Crear microrregla
func (s *MicroRuleServiceImpl) CreateMicroRule(ctx context.Context, req *dto.MicroRuleCreateRequest) (*model.MicroRule, error) {
	// Comprobar si el nombre de la regla ya existe
	if req.Name != "" {
		exists, err := s.ruleRepo.CheckMicroRuleNameExists(ctx, req.Name, bson.NilObjectID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrMicroRuleNameExists
		}
	}

	// Convertir condiciones JSON a BSON
	var condition bson.Raw
	if len(req.Condition) > 0 {
		// Usar el parser JSON para convertir JSON a interface{}
		var anyValue interface{}
		if err := json.Unmarshal(req.Condition, &anyValue); err != nil {
			s.logger.Error().Err(err).Msg("Error al analizar las condiciones JSON")
			return nil, err
		}

		// Convertir interface{} a BSON
		bsonData, err := bson.Marshal(anyValue)
		if err != nil {
			s.logger.Error().Err(err).Msg("Error al convertir la condición a BSON")
			return nil, err
		}

		condition = bsonData
	}

	// Crear nueva microrregla
	rule := &model.MicroRule{
		Name:      req.Name,
		Type:      model.RuleType(req.Type),
		Status:    model.RuleStatus(req.Status),
		Priority:  req.Priority,
		Condition: condition,
	}

	// Guardar microrregla
	err := s.ruleRepo.CreateMicroRule(ctx, rule)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al crear la microrregla")
		return nil, err
	}

	s.logger.Info().Str("id", rule.ID.Hex()).Str("name", rule.Name).Msg("Microrregla creada con éxito")
	return rule, nil
}

// GetMicroRules Obtener lista de microrreglas
func (s *MicroRuleServiceImpl) GetMicroRules(ctx context.Context, pageStr, sizeStr string) ([]model.MicroRule, int64, error) {
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 10
	}

	rules, total, err := s.ruleRepo.GetMicroRules(ctx, page, size)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error al obtener la lista de microrreglas")
		return nil, 0, err
	}

	return rules, total, nil
}

// GetMicroRuleByID obtiene una microrregla por ID
func (s *MicroRuleServiceImpl) GetMicroRuleByID(ctx context.Context, id bson.ObjectID) (*model.MicroRule, error) {
	rule, err := s.ruleRepo.GetMicroRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRuleNotFound) {
			return nil, ErrMicroRuleNotFound
		}
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al obtener la microrregla")
		return nil, err
	}

	return rule, nil
}

// UpdateMicroRule Actualizar microrregla
func (s *MicroRuleServiceImpl) UpdateMicroRule(ctx context.Context, id bson.ObjectID, req *dto.MicroRuleUpdateRequest) (*model.MicroRule, error) {
	// Obtener la microrregla existente
	rule, err := s.ruleRepo.GetMicroRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRuleNotFound) {
			return nil, ErrMicroRuleNotFound
		}
		return nil, err
	}

	// Comprobar si es la regla por defecto del sistema
	if rule.Name == SystemDefaultIPBlockRule {
		s.logger.Warn().Str("id", id.Hex()).Msg("Se intentó modificar la regla por defecto del sistema")
		return nil, ErrSystemRuleNoMod
	}

	// Comprobar si el nombre de la regla ya existe (si se va a actualizar el nombre)
	if req.Name != "" && req.Name != rule.Name {
		exists, err := s.ruleRepo.CheckMicroRuleNameExists(ctx, req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrMicroRuleNameExists
		}
		rule.Name = req.Name
	}

	// Actualizar los demás campos (solo campos no vacíos)
	if req.Type != "" {
		rule.Type = model.RuleType(req.Type)
	}
	if req.Status != "" {
		rule.Status = model.RuleStatus(req.Status)
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if len(req.Condition) > 0 {
		// Usar el parser JSON para convertir JSON a interface{}
		var anyValue interface{}
		if err := json.Unmarshal(req.Condition, &anyValue); err != nil {
			s.logger.Error().Err(err).Msg("Error al analizar las condiciones JSON")
			return nil, err
		}

		// Convertir interface{} a BSON
		bsonData, err := bson.Marshal(anyValue)
		if err != nil {
			s.logger.Error().Err(err).Msg("Error al convertir la condición a BSON")
			return nil, err
		}

		rule.Condition = bsonData
	}

	// Guardar actualización
	err = s.ruleRepo.UpdateMicroRule(ctx, rule)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al actualizar la microrregla")
		return nil, err
	}

	s.logger.Info().Str("id", id.Hex()).Str("name", rule.Name).Msg("Microrregla actualizada con éxito")
	return rule, nil
}

// DeleteMicroRule Eliminar microrregla
func (s *MicroRuleServiceImpl) DeleteMicroRule(ctx context.Context, id bson.ObjectID) error {
	// Comprobar si la microrregla existe
	rule, err := s.ruleRepo.GetMicroRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRuleNotFound) {
			return ErrMicroRuleNotFound
		}
		return err
	}

	// Comprobar si es la regla por defecto del sistema
	if rule.Name == SystemDefaultIPBlockRule {
		s.logger.Warn().Str("id", id.Hex()).Msg("Se intentó eliminar la regla por defecto del sistema")
		return ErrSystemRuleNoDelete
	}

	// Eliminar microrregla
	err = s.ruleRepo.DeleteMicroRule(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id.Hex()).Msg("Error al eliminar la microrregla")
		return err
	}

	s.logger.Info().Str("id", id.Hex()).Msg("Microrregla eliminada con éxito")
	return nil
}
