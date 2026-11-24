package dto

import "time"

// Constantes de rango de tiempo
const (
	TimeRange24Hours = "24h"
	TimeRange7Days   = "7d"
	TimeRange30Days  = "30d"
)

// StatsRequest petición de datos estadísticos
// @Description Parámetros de la petición de datos estadísticos
type StatsRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"` // Rango de tiempo: 24h, 7d, 30d
}

// RealtimeQPSRequest petición de QPS en tiempo real
// @Description Parámetros de la petición de datos QPS en tiempo real
type RealtimeQPSRequest struct {
	Limit int `json:"limit" form:"limit" binding:"omitempty,min=1,max=60" default:"30" example:"30"` // Límite de puntos de datos, por defecto 30, máximo 60
}

// TimeSeriesDataRequest petición de datos de series temporales
// @Description Parámetros de la petición de datos de series temporales
type TimeSeriesDataRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"`     // Rango de tiempo: 24h, 7d, 30d
	Metric    string `json:"metric" form:"metric" binding:"required,oneof=requests blocks" example:"requests"` // Tipo de métrica: requests (peticiones), blocks (bloqueos)
}

// TrafficTimeSeriesRequest petición de datos de series temporales de tráfico
// @Description Parámetros de la petición de datos de series temporales de tráfico
type TrafficTimeSeriesRequest struct {
	TimeRange string `json:"timeRange" form:"timeRange" binding:"required,oneof=24h 7d 30d" example:"24h"` // Rango de tiempo: 24h, 7d, 30d
}

// OverviewStats estadísticas de resumen
// @Description Estadísticas de resumen con los indicadores clave
type OverviewStats struct {
	// Rango de tiempo
	TimeRange string `json:"timeRange" example:"24h"` // Rango de tiempo de los datos estadísticos

	// Estadísticas de tráfico
	TotalRequests   int64 `json:"totalRequests" example:"123456"`     // Total de peticiones
	InboundTraffic  int64 `json:"inboundTraffic" example:"67890123"`  // Tráfico entrante (bytes)
	OutboundTraffic int64 `json:"outboundTraffic" example:"12345678"` // Tráfico saliente (bytes)
	MaxQPS          int64 `json:"maxQPS" example:"150"`               // QPS máximo

	// Estadísticas de errores
	Error4xx     int64   `json:"error4xx" example:"234"`      // Cantidad de errores 4xx
	Error4xxRate float64 `json:"error4xxRate" example:"0.23"` // Tasa de errores 4xx (porcentaje)
	Error5xx     int64   `json:"error5xx" example:"45"`       // Cantidad de errores 5xx
	Error5xxRate float64 `json:"error5xxRate" example:"0.05"` // Tasa de errores 5xx (porcentaje)

	// Estadísticas de seguridad
	BlockCount    int64 `json:"blockCount" example:"123"`   // Cantidad de bloqueos
	AttackIPCount int64 `json:"attackIPCount" example:"45"` // Cantidad de IPs atacantes
}

// RealtimeQPSData datos QPS en tiempo real
// @Description Punto de datos QPS en tiempo real, con marca de tiempo y valor
type RealtimeQPSData struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:30:45Z"` // Marca de tiempo
	Value     int64     `json:"value" example:"120"`                      // Valor de QPS
}

// RealtimeQPSResponse respuesta de QPS en tiempo real
// @Description Respuesta de datos QPS en tiempo real
type RealtimeQPSResponse struct {
	Data []RealtimeQPSData `json:"data"` // Lista de puntos de datos QPS
}

// TimeSeriesDataPoint punto de datos de serie temporal
// @Description Punto de datos del gráfico de series temporales
type TimeSeriesDataPoint struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"` // Marca de tiempo que indica el momento del punto de datos
	Value     int64     `json:"value" example:"128"`                      // Valor numérico que representa la métrica en ese punto temporal
}

// TrafficDataPoint punto de datos de tráfico
// @Description Punto de datos del gráfico de series temporales de tráfico
type TrafficDataPoint struct {
	Timestamp       time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"` // Marca de tiempo que indica el momento del punto de datos
	InboundTraffic  int64     `json:"inboundTraffic" example:"1024000"`         // Tráfico entrante (bytes)
	OutboundTraffic int64     `json:"outboundTraffic" example:"2048000"`        // Tráfico saliente (bytes)
}

// TimeSeriesResponse respuesta de series temporales
// @Description Respuesta de datos del gráfico de series temporales
type TimeSeriesResponse struct {
	Metric    string                `json:"metric" example:"requests"` // Nombre de la métrica
	TimeRange string                `json:"timeRange" example:"24h"`   // Rango de tiempo
	Data      []TimeSeriesDataPoint `json:"data"`                      // Lista de puntos de datos
}

// TrafficTimeSeriesResponse respuesta de series temporales de tráfico
// @Description Respuesta de datos del gráfico de series temporales de tráfico
type TrafficTimeSeriesResponse struct {
	TimeRange string             `json:"timeRange" example:"24h"` // Rango de tiempo
	Data      []TrafficDataPoint `json:"data"`                    // Lista de puntos de datos de tráfico
}

// CombinedTimeSeriesResponse respuesta de series temporales combinadas
// @Description Datos de series temporales con peticiones y bloqueos
type CombinedTimeSeriesResponse struct {
	TimeRange string             `json:"timeRange" example:"24h"` // Rango de tiempo
	Requests  TimeSeriesResponse `json:"requests"`                // Serie temporal de peticiones
	Blocks    TimeSeriesResponse `json:"blocks"`                  // Serie temporal de bloqueos
}
