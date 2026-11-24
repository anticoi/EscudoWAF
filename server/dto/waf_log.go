package dto

import (
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
)

// AttackEventRequset petición de consulta de eventos de ataque
// @Description Estructura de parámetros para la consulta agregada de eventos de ataque, con filtrado multidimensional por IP origen/destino, dominio, puerto y rango de tiempo, y paginación
type AttackEventRequset struct {
	SrcIP     string    `json:"srcIp" form:"srcIp" binding:"omitempty" example:"192.168.1.100"`                                                   // Dirección IP de origen, para rastrear el origen del ataque
	DstIP     string    `json:"dstIp" form:"dstIp" binding:"omitempty" example:"10.0.0.5"`                                                        // Dirección IP de destino, dirección del servidor atacado
	Domain    string    `json:"domain" form:"domain" binding:"omitempty" example:"example.com"`                                                   // Dominio del sitio atacado
	SrcPort   int       `json:"srcPort" form:"srcPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // Puerto de origen desde el que se inició el ataque
	DstPort   int       `json:"dstPort" form:"dstPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // Puerto de destino, puerto del servicio atacado
	StartTime time.Time `json:"startTime" form:"startTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-17T00:00:00Z"` // Hora de inicio de la consulta, formato ISO8601
	EndTime   time.Time `json:"endTime" form:"endTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-18T23:59:59Z"`     // Hora de fin de la consulta, formato ISO8601
	Page      int       `json:"page" form:"page" binding:"omitempty,min=1" default:"1" example:"1"`                                               // Página actual, comenzando en 1
	PageSize  int       `json:"pageSize" form:"pageSize" binding:"omitempty,min=1,max=100" default:"10" example:"10"`                             // Registros por página, máximo 100
}

// AttackLogRequest petición de consulta de logs de ataque
// @Description Estructura de parámetros para la consulta detallada de logs de ataque, con filtrado por ID de regla, IP origen/destino, dominio, puerto, ID de petición y rango de tiempo
type AttackLogRequest struct {
	RuleID    int       `json:"ruleId" form:"ruleId" binding:"omitempty" example:"100012"`                                                        // ID de la regla, identificador de la regla WAF que detectó el ataque
	SrcPort   int       `json:"srcPort" form:"srcPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // Puerto de origen desde el que se inició el ataque
	DstPort   int       `json:"dstPort" form:"dstPort" binding:"omitempty,min=1,max=65535" example:"443"`                                         // Puerto de destino, puerto del servicio atacado
	Domain    string    `json:"domain" form:"domain" binding:"omitempty" example:"example.com"`                                                   // Dominio del sitio atacado
	SrcIP     string    `json:"srcIp" form:"srcIp" binding:"omitempty" example:"192.168.1.100"`                                                   // Dirección IP de origen, para rastrear el origen del ataque
	DstIP     string    `json:"dstIp" form:"dstIp" binding:"omitempty" example:"10.0.0.5"`                                                        // Dirección IP de destino, dirección del servidor atacado
	RequestID string    `json:"requestId" form:"requestId" binding:"omitempty" example:"1234567890"`                                              // ID de la petición, identificador único de la petición HTTP
	StartTime time.Time `json:"startTime" form:"startTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-17T00:00:00Z"` // Hora de inicio de la consulta, formato ISO8601
	EndTime   time.Time `json:"endTime" form:"endTime" binding:"omitempty" time_format:"2006-01-02T15:04:05Z" example:"2024-03-18T23:59:59Z"`     // Hora de fin de la consulta, formato ISO8601
	Page      int       `json:"page" form:"page" binding:"omitempty,min=1" default:"1" example:"1"`                                               // Página actual, comenzando en 1
	PageSize  int       `json:"pageSize" form:"pageSize" binding:"omitempty,min=1,max=100" default:"10" example:"10"`                             // Registros por página, máximo 100
}

// AttackEventAggregateResult resultado de agregación de eventos de ataque
// @Description Resultado de estadísticas agregadas de eventos de ataque, con resumen por IP, dominio y puerto, incluyendo número de ataques, primera y última vez, y duración
type AttackEventAggregateResult struct {
	SrcIP             string        `bson:"srcIp" json:"srcIp" example:"192.168.1.100"`                                    // Dirección IP de origen, dirección del atacante
	SrcIPInfo         *model.IPInfo `bson:"srcIpInfo" json:"srcIpInfo"`                                                    // Dirección IP de origen, dirección del atacante
	DstPort           int           `bson:"dstPort" json:"dstPort" example:"443"`                                          // Puerto de destino, puerto del servicio atacado
	Domain            string        `bson:"domain" json:"domain" example:"example.com"`                                    // Dominio del sitio atacado
	Count             int           `bson:"count" json:"count" example:"15"`                                               // Número total de ataques desde el mismo origen
	FirstAttackTime   time.Time     `bson:"firstAttackTime" json:"firstAttackTime" example:"2024-03-18T08:12:33Z"`         // Hora del primer ataque desde esta IP
	LastAttackTime    time.Time     `bson:"lastAttackTime" json:"lastAttackTime" example:"2024-03-18T08:30:45Z"`           // Hora del último ataque desde esta IP
	DurationInMinutes float64       `bson:"durationInMinutes,omitempty" json:"durationInMinutes,omitempty" example:"18.2"` // Duración del ataque (minutos), lapso desde el primer hasta el último ataque
	IsOngoing         bool          `bson:"isOngoing" json:"isOngoing" example:"true"`                                     // Si está en curso, indica si el ataque continúa
}

// AttackEventResponse respuesta de eventos de ataque
// @Description Estructura de respuesta paginada de consulta de eventos de ataque, con lista de resultados agregados y metadatos de paginación
type AttackEventResponse struct {
	Results     []AttackEventAggregateResult `json:"results"`                 // Lista de resultados agregados, registros de eventos de ataque de la página actual
	TotalCount  int64                        `json:"totalCount" example:"35"` // Total de registros, cantidad de eventos de ataque que cumplen los criterios
	PageSize    int                          `json:"pageSize" example:"10"`   // Tamaño de página, registros por página configurados
	CurrentPage int                          `json:"currentPage" example:"1"` // Página actual, comenzando en 1
	TotalPages  int                          `json:"totalPages" example:"4"`  // Total de páginas, calculado según el total de registros y el tamaño de página
}

// AttackLogResponse respuesta de logs de ataque
// @Description Estructura de respuesta paginada de consulta de logs de ataques, con registros WAF detallados e información de paginación
type AttackLogResponse struct {
	Results     []model.WAFLog `json:"results"`                  // Lista de registros de log, detalles de ataques WAF de la página actual
	TotalCount  int64          `json:"totalCount" example:"128"` // Total de registros, cantidad de logs que cumplen los criterios de búsqueda
	PageSize    int            `json:"pageSize" example:"10"`    // Tamaño de página, registros por página configurados
	CurrentPage int            `json:"currentPage" example:"1"`  // Página actual, comenzando en 1
	TotalPages  int            `json:"totalPages" example:"13"`  // Total de páginas, calculado según el total de registros y el tamaño de página
}
