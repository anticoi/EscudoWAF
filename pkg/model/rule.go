package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// RuleType tipo de regla
//
//	@Description	Tipo de regla, indica si es lista blanca o lista negra
type RuleType string

const (
	WhitelistRule RuleType = "whitelist" // Reglas de lista blanca
	BlacklistRule RuleType = "blacklist" // Reglas de lista negra
)

// RuleStatus estado de la regla
//
//	@Description	Estado de la regla, indica si está habilitada o deshabilitada
type RuleStatus string

const (
	RuleEnabled  RuleStatus = "enabled"  // Regla habilitada
	RuleDisabled RuleStatus = "disabled" // Regla deshabilitada
)

// MicroRule representa la información de una microrregla del WAF
// @Description Información de microrregla WAF, con nombre, tipo, estado, prioridad y condiciones
type MicroRule struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"60d21b4367d0d8992e89e964"` // Identificador único de la regla
	Name     string        `json:"name" bson:"name" example:"Regla de protección contra inyección SQL"`                                 // Nombre de la regla
	Type     RuleType      `json:"type" bson:"type" example:"blacklist"`                                 // Tipo de regla
	Status   RuleStatus    `json:"status" bson:"status" example:"enabled"`                               // Estado de la regla
	Priority int           `json:"priority" bson:"priority" example:"100"`                               // Campo de prioridad; a mayor número, mayor prioridad
	// @Schema(type=object, example={"type":"composite","operator":"AND","conditions":[{"type":"simple","target":"source_ip","match_type":"in_ipgroup","match_value":"blocked_ips"},{"type":"simple","target":"path","match_type":"regex","match_value":"^/admin/.*$"}]})
	Condition bson.Raw `json:"condition" bson:"condition" swaggertype:"object"`
}

func (r *MicroRule) GetCollectionName() string {
	return "micro_rule"
}
