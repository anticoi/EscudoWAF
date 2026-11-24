package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// WAFLog representa un registro de evento de seguridad
// @Description Registro completo de eventos de seguridad del WAF, con información detallada de detección y protección de ataques
type WAFLog struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`                                                                                                     // Identificador único del log
	RequestID    string        `json:"requestId" bson:"requestId" example:"a1b2c3d4e5f6"`                                                                                     // Identificador único de la petición
	RuleID       int           `json:"ruleId" bson:"ruleId" example:"10086"`                                                                                                  // ID de la regla que se disparó
	SecLangRaw   string        `json:"secLangRaw" bson:"secLangRaw" example:"SecRule REQUEST_HEADERS:User-Agent \"@rx (?:scanner)\" \"id:1008,phase:1,severity:'CRITICAL'\""` // Definición original de la regla de seguridad
	Severity     int           `json:"severity" bson:"severity" example:"2"`                                                                                                  // Nivel de severidad del evento (0-5)
	Phase        int           `json:"phase" bson:"phase" example:"1"`                                                                                                        // Fase de procesamiento de la petición
	SecMark      string        `json:"secMark" bson:"secMark" example:"web_scanner"`                                                                                          // Marca de seguridad
	Accuracy     int           `json:"accuracy" bson:"accuracy" example:"9"`                                                                                                  // Precisión de coincidencia de la regla (0-10)
	Payload      string        `json:"payload" bson:"payload" example:"Scanner/1.0"`                                                                                          // Carga del ataque
	URI          string        `json:"uri" bson:"uri" example:"/api/v1/users"`                                                                                                // Ruta del URI de la petición
	SrcIP        string        `json:"srcIp" bson:"srcIp" example:"192.168.1.1"`                                                                                              // Dirección IP de origen
	SrcIPInfo    *IPInfo       `json:"srcIpInfo,omitempty" bson:"srcIpInfo,omitempty"`                                                                                        // Información de geolocalización de la IP de origen
	DstIP        string        `json:"dstIp" bson:"dstIp" example:"10.0.0.1"`                                                                                                 // Dirección IP de destino
	ClientIP     string        `json:"clientIp" bson:"clientIp" example:"192.168.1.1"`                                                                                        // Dirección IP de origen
	ServerIP     string        `json:"serverIp" bson:"serverIp" example:"10.0.0.1"`                                                                                           // Dirección IP de destino
	SrcPort      int           `json:"srcPort" bson:"srcPort" example:"52134"`                                                                                                // Puerto de origen
	DstPort      int           `json:"dstPort" bson:"dstPort" example:"443"`                                                                                                  // Puerto de destino
	Domain       string        `json:"domain" bson:"domain" example:"api.example.com"`                                                                                        // Dominio de destino
	Logs         []Log         `json:"logs" bson:"logs"`                                                                                                                      // Entradas de log asociadas
	Message      string        `json:"message" bson:"message" example:"Detección de escáner malicioso"`                                                                                              // Mensaje descriptivo del evento
	Request      string        `json:"request" bson:"request" example:"GET /api/v1/users HTTP/1.1\nHost: api.example.com\nUser-Agent: Scanner/1.0"`                           // Petición HTTP original
	Response     string        `json:"response" bson:"response" example:"HTTP/1.1 403 Forbidden\nContent-Type: text/html\nContent-Length: 146"`                               // Respuesta HTTP original
	Date         string        `json:"date" bson:"date"`
	Hour         int           `json:"hour" bson:"hour"`
	HourGroupSix int           `json:"hourGroupSix" bson:"hourGroupSix" example:"0"`
	Minute       int           `json:"minute" bson:"minute"`
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt" example:"2024-03-18T08:12:33Z"` // Marca de tiempo del evento
}

// Log representa una entrada de log individual
// @Description Registro detallado de coincidencias de reglas del WAF, con información de activación y log original
type Log struct {
	Message    string `json:"message" bson:"message" example:"Detección de escáner malicioso"`                                                                                                                                                                                                                                       // Mensaje de log
	Payload    string `json:"payload" bson:"payload" example:"Scanner/1.0"`                                                                                                                                                                                                                                   // Carga del ataque
	RuleID     int    `json:"ruleId" bson:"ruleId" example:"10086"`                                                                                                                                                                                                                                           // ID de la regla
	Severity   int    `json:"severity" bson:"severity" example:"2"`                                                                                                                                                                                                                                           // Nivel de severidad (0-5)
	Phase      int    `json:"phase" bson:"phase" example:"1"`                                                                                                                                                                                                                                                 // Fase de procesamiento de la petición
	SecMark    string `json:"secMark" bson:"secMark" example:"web_scanner"`                                                                                                                                                                                                                                   // Marca de seguridad
	Accuracy   int    `json:"accuracy" bson:"accuracy" example:"9"`                                                                                                                                                                                                                                           // Precisión de coincidencia de la regla (0-10)
	SecLangRaw string `json:"secLangRaw" bson:"secLangRaw" example:"SecRule REQUEST_HEADERS:User-Agent \"@rx (?:scanner)\" \"id:1008,phase:1,severity:'CRITICAL'\""`                                                                                                                                          // Definición original de la regla de seguridad
	LogRaw     string `json:"logRaw" bson:"logRaw" example:"[2024-03-18 08:12:33] [error] ModSecurity: Access denied with code 403 (phase 1). Matched \"Operator 'Rx' with parameter '(?:scanner)' against variable 'REQUEST_HEADERS:User-Agent'\" [id \"10086\"] [msg \"Detección de escáner malicioso\"] [severity \"CRITICAL\"]"` // Datos de log originales
}

// GetCollectionName devuelve el nombre de la colección MongoDB de WAFLog
// @Description Obtener el nombre de la colección de MongoDB para los logs del WAF
func (wafLog *WAFLog) GetCollectionName() string {
	return "waf_log"
}
