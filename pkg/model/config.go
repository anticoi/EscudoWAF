package model

import (
	"time"
)

// Config configuración del sistema WAF
//
//	@Description	Información de configuración global del sistema WAF
type Config struct {
	Name            string        `bson:"name" json:"name" example:"AppConfig" description:"Nombre de la configuración"`
	Engine          EngineConfig  `bson:"engine" json:"engine" description:"Configuración del motor"`
	Haproxy         HaproxyConfig `bson:"haproxy" json:"haproxy" description:"Configuración de HAProxy"`
	CreatedAt       time.Time     `bson:"createdAt" json:"createdAt" description:"Fecha de creación"`
	UpdatedAt       time.Time     `bson:"updatedAt" json:"updatedAt" description:"Hora de actualización"`
	IsResponseCheck bool          `bson:"isResponseCheck" json:"isResponseCheck" description:"Si se inspeccionan las respuestas"`
	IsDebug         bool          `bson:"isDebug" json:"isDebug" description:"Si el modo de depuración está activado"`
	IsK8s           bool          `bson:"isK8s" json:"isK8s" description:"Si está en entorno K8s"`
}

// EngineConfig configuración del motor
//
//	@Description	Información de configuración del motor WAF
type EngineConfig struct {
	Bind            string            `bson:"bind" json:"bind" example:"0.0.0.0:9000" description:"Dirección de enlace"`
	UseBuiltinRules bool              `bson:"useBuiltinRules" json:"useBuiltinRules" description:"Si se usan reglas integradas"`
	ASNDBPath       string            `bson:"asnDBPath" json:"asnDBPath" example:"/opt/geoip/GeoLite2-ASN.mmdb" description:"Ruta de la base de datos ASN"`
	CityDBPath      string            `bson:"cityDBPath" json:"cityDBPath" example:"/opt/geoip/GeoLite2-City.mmdb" description:"Ruta de la base de datos de ciudades"`
	AppConfig       []AppConfig       `bson:"appConfig" json:"appConfig" description:"Lista de configuraciones de aplicación"`
	FlowController  FlowControlConfig `bson:"flowController" json:"flowController" description:"Configuración de control de flujo"`
}

// AppConfig configuración de la aplicación
//
//	@Description	Configuración de aplicación WAF
type AppConfig struct {
	Name           string        `bson:"name" json:"name" example:"default" description:"Nombre de la aplicación"`
	Directives     string        `bson:"directives" json:"directives" description:"Directivas de Coraza"`
	TransactionTTL time.Duration `bson:"transactionTTL" json:"transactionTTL" example:"10s" description:"Tiempo de espera de la transacción"`
	LogLevel       string        `bson:"logLevel" json:"logLevel" example:"info" description:"Nivel de log"`
	LogFile        string        `bson:"logFile" json:"logFile" example:"/var/log/waf.log" description:"Ruta del archivo de log"`
	LogFormat      string        `bson:"logFormat" json:"logFormat" example:"json" description:"Formato de log"`
}

// HaproxyConfig configuración de HAProxy
//
//	@Description	Configuración relacionada con HAProxy
type HaproxyConfig struct {
	ConfigBaseDir string `bson:"configBaseDir" json:"configBaseDir" example:"/etc/haproxy" description:"Directorio base de configuración"`
	HaproxyBin    string `bson:"haproxyBin" json:"haproxyBin" example:"/usr/sbin/haproxy" description:"Ruta del ejecutable de HAProxy"`
	BackupsNumber int    `bson:"backupsNumber" json:"backupsNumber" example:"5" description:"Cantidad de copias de seguridad"`
	SpoeAgentAddr string `bson:"spoeAgentAddr" json:"spoeAgentAddr" example:"127.0.0.1" description:"Dirección del agente SPOE"`
	SpoeAgentPort int    `bson:"spoeAgentPort" json:"spoeAgentPort" example:"9000" description:"Puerto del agente SPOE"`
	Thread        int    `bson:"thread" json:"thread" example:"4" description:"Número de hilos"`
}

// FlowControlConfig define la configuración de control de flujo, para almacenar en la base de datos
//
//	@Description	Configuración de control de flujo del WAF
type FlowControlConfig struct {
	// Configuración de límite de acceso de alta frecuencia
	VisitLimit struct {
		Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"Si la restricción de acceso está habilitada"`
		Threshold      int64 `bson:"threshold" json:"threshold" example:"100" description:"Umbral de acceso, número máximo de peticiones por minuto"`
		StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"Ventana de tiempo (segundos)"`
		BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"600" description:"Duración del bloqueo (segundos)"`
		BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"10" description:"Número de peticiones en ráfaga permitidas"`
		ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"Capacidad de caché, número máximo de IPs en caché"`
	} `bson:"visitLimit" json:"visitLimit" description:"Configuración del límite de frecuencia de acceso"`

	// Configuración de límite de ataques de alta frecuencia
	AttackLimit struct {
		Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"Si la restricción de ataques está habilitada"`
		Threshold      int64 `bson:"threshold" json:"threshold" example:"5" description:"Umbral de ataque, número máximo de ataques por minuto"`
		StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"Ventana de tiempo (segundos)"`
		BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"3600" description:"Duración del bloqueo (segundos)"`
		BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"2" description:"Número de ataques en ráfaga permitidos"`
		ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"Capacidad de caché, número máximo de IPs en caché"`
	} `bson:"attackLimit" json:"attackLimit" description:"Configuración del límite de frecuencia de ataques"`

	// Configuración de límite de errores de alta frecuencia
	ErrorLimit struct {
		Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"Si la restricción de errores está habilitada"`
		Threshold      int64 `bson:"threshold" json:"threshold" example:"20" description:"Umbral de errores, número máximo de errores por minuto"`
		StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"Ventana de tiempo (segundos)"`
		BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"1800" description:"Duración del bloqueo (segundos)"`
		BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"5" description:"Número de errores en ráfaga permitidos"`
		ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"Capacidad de caché, número máximo de IPs en caché"`
	} `bson:"errorLimit" json:"errorLimit" description:"Configuración del límite de frecuencia de errores"`
}

// GetDefaultFlowControlConfig devuelve la configuración de control de flujo por defecto
//
//	@Summary		Obtener la configuración de control de flujo por defecto
//	@Description	Obtener la configuración de control de flujo por defecto del sistema
//	@Return			Configuración de control de flujo por defecto FlowControlConfig
func GetDefaultFlowControlConfig() FlowControlConfig {
	return FlowControlConfig{
		VisitLimit: struct {
			Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"Si la restricción de acceso está habilitada"`
			Threshold      int64 `bson:"threshold" json:"threshold" example:"100" description:"Umbral de acceso, número máximo de peticiones por minuto"`
			StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"Ventana de tiempo (segundos)"`
			BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"600" description:"Duración del bloqueo (segundos)"`
			BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"10" description:"Número de peticiones en ráfaga permitidas"`
			ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"Capacidad de caché, número máximo de IPs en caché"`
		}{
			Enabled:        false,
			Threshold:      100,   // 100 peticiones por minuto
			StatDuration:   60,    // Ventana estadística de 1 minuto
			BlockDuration:  600,   // Bloqueo de 10 minutos
			BurstCount:     10,    // Permitir ráfagas de 10
			ParamsCapacity: 10000, // Cachear 10.000 IPs
		},
		AttackLimit: struct {
			Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"Si la restricción de ataques está habilitada"`
			Threshold      int64 `bson:"threshold" json:"threshold" example:"5" description:"Umbral de ataque, número máximo de ataques por minuto"`
			StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"Ventana de tiempo (segundos)"`
			BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"3600" description:"Duración del bloqueo (segundos)"`
			BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"2" description:"Número de ataques en ráfaga permitidos"`
			ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"Capacidad de caché, número máximo de IPs en caché"`
		}{
			Enabled:        false,
			Threshold:      5,     // 5 ataques por minuto
			StatDuration:   60,    // Ventana estadística de 1 minuto
			BlockDuration:  3600,  // Bloqueo de 1 hora
			BurstCount:     2,     // Permitir ráfagas de 2
			ParamsCapacity: 10000, // Cachear 10.000 IPs
		},
		ErrorLimit: struct {
			Enabled        bool  `bson:"enabled" json:"enabled" example:"true" description:"Si la restricción de errores está habilitada"`
			Threshold      int64 `bson:"threshold" json:"threshold" example:"20" description:"Umbral de errores, número máximo de errores por minuto"`
			StatDuration   int64 `bson:"statDuration" json:"statDuration" example:"60" description:"Ventana de tiempo (segundos)"`
			BlockDuration  int64 `bson:"blockDuration" json:"blockDuration" example:"1800" description:"Duración del bloqueo (segundos)"`
			BurstCount     int64 `bson:"burstCount" json:"burstCount" example:"5" description:"Número de errores en ráfaga permitidos"`
			ParamsCapacity int64 `bson:"paramsCapacity" json:"paramsCapacity" example:"10000" description:"Capacidad de caché, número máximo de IPs en caché"`
		}{
			Enabled:        false,
			Threshold:      20,    // 20 errores por minuto
			StatDuration:   60,    // Ventana estadística de 1 minuto
			BlockDuration:  1800,  // Bloqueo de 30 minutos
			BurstCount:     5,     // Permitir ráfagas de 5
			ParamsCapacity: 10000, // Cachear 10.000 IPs
		},
	}
}

// GetCollectionName obtiene el nombre de la colección
//
//	@Summary		Obtener el nombre de la colección de MongoDB
//	@Description	Obtener el nombre de la colección de configuración en MongoDB
//	@Return			string nombre de la colección
func (c *Config) GetCollectionName() string {
	return "config"
}
