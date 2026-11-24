package model

import "time"

// BlockedIPRecord registro de IP bloqueada
// @Description Información del registro de IP bloqueada
type BlockedIPRecord struct {
	IP           string    `bson:"ip" json:"ip" example:"192.168.1.1" description:"Dirección IP bloqueada"`
	Reason       string    `bson:"reason" json:"reason" example:"high_frequency_attack" description:"Motivo del bloqueo"`
	RequestUri   string    `bson:"request_uri" json:"requestUri" example:"/api/v1/login" description:"URI de la petición"`
	BlockedAt    time.Time `bson:"blocked_at" json:"blockedAt" description:"Hora de inicio del bloqueo"`
	BlockedUntil time.Time `bson:"blocked_until" json:"blockedUntil" description:"Hora de fin del bloqueo"`
}

func (b *BlockedIPRecord) GetCollectionName() string {
	return "blocked_ips"
}
