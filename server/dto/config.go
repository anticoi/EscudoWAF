package dto

import (
	"time"
)

// ConfigPatchRequest petición de actualización parcial de configuración
// @Description Parámetros de la petición para actualización parcial de configuración
type ConfigPatchRequest struct {
	Name            *string          `json:"name,omitempty" binding:"omitempty" example:"AppConfig"`        // Nombre de la configuración
	Engine          *EnginePatchDTO  `json:"engine,omitempty" binding:"omitempty"`                          // Configuración del motor
	Haproxy         *HaproxyPatchDTO `json:"haproxy,omitempty" binding:"omitempty"`                         // Configuración de HAProxy
	IsResponseCheck *bool            `json:"isResponseCheck,omitempty" binding:"omitempty" example:"false"` // Si se inspeccionan las respuestas
	IsDebug         *bool            `json:"isDebug,omitempty" binding:"omitempty" example:"false"`         // Si el modo de depuración está activado
	IsK8s           *bool            `json:"isK8s,omitempty" binding:"omitempty" example:"false"`           // Si está en entorno K8s
}

// EnginePatchDTO DTO de actualización parcial de configuración del motor
type EnginePatchDTO struct {
	Bind            *string                 `json:"bind,omitempty" binding:"omitempty" example:"127.0.0.1:2342"`                      // Dirección de enlace del motor
	UseBuiltinRules *bool                   `json:"useBuiltinRules,omitempty" binding:"omitempty" example:"true"`                     // Si se usan reglas integradas
	ASNDBPath       *string                 `json:"asnDBPath,omitempty" binding:"omitempty" example:"/opt/geoip/GeoLite2-ASN.mmdb"`   // Ruta de la base de datos ASN
	CityDBPath      *string                 `json:"cityDBPath,omitempty" binding:"omitempty" example:"/opt/geoip/GeoLite2-City.mmdb"` // Ruta de la base de datos de ciudades
	AppConfig       []AppConfigPatchDTO     `json:"appConfig,omitempty" binding:"omitempty,dive"`                                     // Lista de configuraciones de aplicación
	FlowController  *FlowControllerPatchDTO `json:"flowController,omitempty" binding:"omitempty"`                                     // Configuración de control de flujo
}

// AppConfigPatchDTO DTO de actualización parcial de configuración
type AppConfigPatchDTO struct {
	Name           *string `json:"name,omitempty" binding:"omitempty" example:"coraza"`          // Nombre de la aplicación
	Directives     *string `json:"directives,omitempty" binding:"omitempty"`                     // Configuración de directivas
	TransactionTTL *int64  `json:"transactionTTL,omitempty" binding:"omitempty" example:"60000"` // Tiempo de espera de la transacción (ms)
	LogLevel       *string `json:"logLevel,omitempty" binding:"omitempty" example:"info"`        // Nivel de log
	LogFile        *string `json:"logFile,omitempty" binding:"omitempty" example:"/dev/stdout"`  // Archivo de log
	LogFormat      *string `json:"logFormat,omitempty" binding:"omitempty" example:"console"`    // Formato de log
}

// HaproxyPatchDTO DTO de actualización parcial de configuración de HAProxy
type HaproxyPatchDTO struct {
	ConfigBaseDir *string `json:"configBaseDir,omitempty" binding:"omitempty" example:"/EscudoWAF"`    // Directorio raíz de archivos de configuración
	HaproxyBin    *string `json:"haproxyBin,omitempty" binding:"omitempty" example:"haproxy"`      // Ruta del binario de HAProxy
	BackupsNumber *int    `json:"backupsNumber,omitempty" binding:"omitempty" example:"5"`         // Cantidad de copias de seguridad
	SpoeAgentAddr *string `json:"spoeAgentAddr,omitempty" binding:"omitempty" example:"127.0.0.1"` // Dirección del agente SPOE
	SpoeAgentPort *int    `json:"spoeAgentPort,omitempty" binding:"omitempty" example:"2342"`      // Puerto del agente SPOE
	Thread        *int    `json:"thread,omitempty" binding:"omitempty,min=0,max=256" example:"4"`  // Número de hilos
}

// FlowControllerPatchDTO DTO de actualización parcial de configuración del controlador de flujo
type FlowControllerPatchDTO struct {
	VisitLimit  *LimitConfigPatchDTO `json:"visitLimit,omitempty" binding:"omitempty"`  // Configuración del límite de frecuencia de acceso
	AttackLimit *LimitConfigPatchDTO `json:"attackLimit,omitempty" binding:"omitempty"` // Configuración del límite de frecuencia de ataques
	ErrorLimit  *LimitConfigPatchDTO `json:"errorLimit,omitempty" binding:"omitempty"`  // Configuración del límite de frecuencia de errores
}

// LimitConfigPatchDTO DTO de actualización parcial de configuración de límites
type LimitConfigPatchDTO struct {
	Enabled        *bool  `json:"enabled,omitempty" binding:"omitempty" example:"true"`         // Si está habilitado
	Threshold      *int64 `json:"threshold,omitempty" binding:"omitempty" example:"100"`        // Umbral
	StatDuration   *int64 `json:"statDuration,omitempty" binding:"omitempty" example:"60"`      // Ventana de tiempo de estadísticas (segundos)
	BlockDuration  *int64 `json:"blockDuration,omitempty" binding:"omitempty" example:"600"`    // Duración del bloqueo (segundos)
	BurstCount     *int64 `json:"burstCount,omitempty" binding:"omitempty" example:"10"`        // Número de peticiones en ráfaga permitidas
	ParamsCapacity *int64 `json:"paramsCapacity,omitempty" binding:"omitempty" example:"10000"` // Capacidad de caché
}

// ConfigResponse Respuesta de configuración
// @Description Respuesta de configuración
type ConfigResponse struct {
	ID              string     `json:"id,omitempty"`    // ID de configuración
	Name            string     `json:"name"`            // Nombre de la configuración
	Engine          EngineDTO  `json:"engine"`          // Configuración del motor
	Haproxy         HaproxyDTO `json:"haproxy"`         // Configuración de HAProxy
	CreatedAt       time.Time  `json:"createdAt"`       // Fecha de creación
	UpdatedAt       time.Time  `json:"updatedAt"`       // Hora de actualización
	IsResponseCheck bool       `json:"isResponseCheck"` // Si se inspeccionan las respuestas
	IsDebug         bool       `json:"isDebug"`         // Si el modo de depuración está activado
	IsK8s           bool       `json:"isK8s"`           // Si está en entorno K8s
}

// EngineDTO DTO de configuración del motor
type EngineDTO struct {
	Bind            string            `json:"bind"`            // Dirección de enlace del motor
	UseBuiltinRules bool              `json:"useBuiltinRules"` // Si se usan reglas integradas
	ASNDBPath       string            `json:"asnDBPath"`       // Ruta de la base de datos ASN
	CityDBPath      string            `json:"cityDBPath"`      // Ruta de la base de datos de ciudades
	AppConfig       []AppConfigDTO    `json:"appConfig"`       // Lista de configuraciones de aplicación
	FlowController  FlowControllerDTO `json:"flowController"`  // Configuración de control de flujo
}

// AppConfigDTO DTO de configuración de aplicación
type AppConfigDTO struct {
	Name           string `json:"name"`                           // Nombre de la aplicación
	Directives     string `json:"directives"`                     // Configuración de directivas
	TransactionTTL int64  `json:"transactionTTL" example:"60000"` // Tiempo de espera de la transacción (ms)
	LogLevel       string `json:"logLevel"`                       // Nivel de log
	LogFile        string `json:"logFile"`                        // Archivo de log
	LogFormat      string `json:"logFormat"`                      // Formato de log
}

// HaproxyDTO DTO de configuración de HAProxy
type HaproxyDTO struct {
	ConfigBaseDir string `json:"configBaseDir"` // Directorio raíz de archivos de configuración
	HaproxyBin    string `json:"haproxyBin"`    // Ruta del binario de HAProxy
	BackupsNumber int    `json:"backupsNumber"` // Cantidad de copias de seguridad
	SpoeAgentAddr string `json:"spoeAgentAddr"` // Dirección del agente SPOE
	SpoeAgentPort int    `json:"spoeAgentPort"` // Puerto del agente SPOE
	Thread        int    `json:"thread"`        // Número de hilos
}

// FlowControllerDTO DTO de configuración del controlador de flujo
type FlowControllerDTO struct {
	VisitLimit  LimitConfigDTO `json:"visitLimit"`  // Configuración del límite de frecuencia de acceso
	AttackLimit LimitConfigDTO `json:"attackLimit"` // Configuración del límite de frecuencia de ataques
	ErrorLimit  LimitConfigDTO `json:"errorLimit"`  // Configuración del límite de frecuencia de errores
}

// LimitConfigDTO DTO de configuración de límites
type LimitConfigDTO struct {
	Enabled        bool  `json:"enabled"`        // Si está habilitado
	Threshold      int64 `json:"threshold"`      // Umbral
	StatDuration   int64 `json:"statDuration"`   // Ventana de tiempo de estadísticas (segundos)
	BlockDuration  int64 `json:"blockDuration"`  // Duración del bloqueo (segundos)
	BurstCount     int64 `json:"burstCount"`     // Número de peticiones en ráfaga permitidas
	ParamsCapacity int64 `json:"paramsCapacity"` // Capacidad de caché
}

// Convertir time.Duration a int64 en milisegundos
func DurationToMillis(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

// Convertir int64 en milisegundos a time.Duration
func MillisToDuration(millis int64) time.Duration {
	return time.Duration(millis) * time.Millisecond
}
