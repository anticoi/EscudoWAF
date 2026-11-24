package model

import (
	"time"

	"github.com/haproxytech/client-native/v6/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// HAProxyStats definición de los campos de datos estadísticos
// Esta estructura solo se usa para conversión y cálculo de diferencias, ya no para almacenamiento
// @Description Modelo de datos estadísticos de HAProxy, para el cálculo y conversión de métricas
type HAProxyStats struct {
	// Estadísticas relacionadas con el tráfico
	// @Description Bytes de tráfico entrante
	Bin int64 `bson:"bin" json:"bin"`
	// @Description Bytes de tráfico saliente
	Bout int64 `bson:"bout" json:"bout"`

	// Estadísticas de códigos de estado de respuesta HTTP
	// @Description Cantidad de respuestas con código de estado 1xx
	Hrsp1xx int64 `bson:"hrsp_1xx" json:"hrsp_1xx"`
	// @Description Cantidad de respuestas con código de estado 2xx
	Hrsp2xx int64 `bson:"hrsp_2xx" json:"hrsp_2xx"`
	// @Description Cantidad de respuestas con código de estado 3xx
	Hrsp3xx int64 `bson:"hrsp_3xx" json:"hrsp_3xx"`
	// @Description Cantidad de respuestas con código de estado 4xx
	Hrsp4xx int64 `bson:"hrsp_4xx" json:"hrsp_4xx"`
	// @Description Cantidad de respuestas con código de estado 5xx
	Hrsp5xx int64 `bson:"hrsp_5xx" json:"hrsp_5xx"`
	// @Description Cantidad de respuestas con otros códigos de estado
	HrspOther int64 `bson:"hrsp_other" json:"hrsp_other"`

	// Estadísticas relacionadas con errores
	// @Description Cantidad de peticiones rechazadas
	Dreq int64 `bson:"dreq" json:"dreq"`
	// @Description Cantidad de respuestas rechazadas
	Dresp int64 `bson:"dresp" json:"dresp"`
	// @Description Cantidad de peticiones con error
	Ereq int64 `bson:"ereq" json:"ereq"`
	// @Description Cantidad de conexiones rechazadas
	Dcon int64 `bson:"dcon" json:"dcon"`
	// @Description Cantidad de sesiones rechazadas
	Dses int64 `bson:"dses" json:"dses"`
	// @Description Cantidad de errores de conexión
	Econ int64 `bson:"econ" json:"econ"`
	// @Description Cantidad de respuestas con error
	Eresp int64 `bson:"eresp" json:"eresp"`

	// Tasa máxima
	// @Description Tasa máxima de peticiones
	ReqRateMax int64 `bson:"req_rate_max" json:"req_rate_max"`
	// @Description Tasa máxima de conexiones
	ConnRateMax int64 `bson:"conn_rate_max" json:"conn_rate_max"`
	// @Description Tasa máxima de transferencia (bytes/segundo)
	RateMax int64 `bson:"rate_max" json:"rate_max"`
	// @Description Número máximo de sesiones concurrentes
	Smax int64 `bson:"smax" json:"smax"`

	// Valor total
	// @Description Total de conexiones
	ConnTot int64 `bson:"conn_tot" json:"conn_tot"`
	// @Description Total de sesiones
	Stot int64 `bson:"stot" json:"stot"`
	// @Description Total de peticiones
	ReqTot int64 `bson:"req_tot" json:"req_tot"`
}

// HAProxyStatsBaseline almacena la línea base de estadísticas de HAProxy
// @Description Línea base de estadísticas de HAProxy, para registrar y restaurar valores base
type HAProxyStatsBaseline struct {
	// @Description Identificador único del registro
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// @Description Nombre del frontend de destino
	TargetName string `bson:"target_name" json:"target_name"`
	// @Description Marca de tiempo del registro
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	// @Description Contador de reinicios de HAProxy
	ResetCount int32 `bson:"reset_count" json:"reset_count"`

	// Aplanar los campos estadísticos directamente en la estructura
	// @Description Bytes de tráfico entrante
	Bin int64 `bson:"bin" json:"bin"`
	// @Description Bytes de tráfico saliente
	Bout int64 `bson:"bout" json:"bout"`
	// @Description Cantidad de respuestas con código de estado 1xx
	Hrsp1xx int64 `bson:"hrsp_1xx" json:"hrsp_1xx"`
	// @Description Cantidad de respuestas con código de estado 2xx
	Hrsp2xx int64 `bson:"hrsp_2xx" json:"hrsp_2xx"`
	// @Description Cantidad de respuestas con código de estado 3xx
	Hrsp3xx int64 `bson:"hrsp_3xx" json:"hrsp_3xx"`
	// @Description Cantidad de respuestas con código de estado 4xx
	Hrsp4xx int64 `bson:"hrsp_4xx" json:"hrsp_4xx"`
	// @Description Cantidad de respuestas con código de estado 5xx
	Hrsp5xx int64 `bson:"hrsp_5xx" json:"hrsp_5xx"`
	// @Description Cantidad de respuestas con otros códigos de estado
	HrspOther int64 `bson:"hrsp_other" json:"hrsp_other"`
	// @Description Cantidad de peticiones rechazadas
	Dreq int64 `bson:"dreq" json:"dreq"`
	// @Description Cantidad de respuestas rechazadas
	Dresp int64 `bson:"dresp" json:"dresp"`
	// @Description Cantidad de peticiones con error
	Ereq int64 `bson:"ereq" json:"ereq"`
	// @Description Cantidad de conexiones rechazadas
	Dcon int64 `bson:"dcon" json:"dcon"`
	// @Description Cantidad de sesiones rechazadas
	Dses int64 `bson:"dses" json:"dses"`
	// @Description Cantidad de errores de conexión
	Econ int64 `bson:"econ" json:"econ"`
	// @Description Cantidad de respuestas con error
	Eresp int64 `bson:"eresp" json:"eresp"`
	// @Description Tasa máxima de peticiones
	ReqRateMax int64 `bson:"req_rate_max" json:"req_rate_max"`
	// @Description Tasa máxima de conexiones
	ConnRateMax int64 `bson:"conn_rate_max" json:"conn_rate_max"`
	// @Description Tasa máxima de transferencia (bytes/segundo)
	RateMax int64 `bson:"rate_max" json:"rate_max"`
	// @Description Número máximo de sesiones concurrentes
	Smax int64 `bson:"smax" json:"smax"`
	// @Description Total de conexiones
	ConnTot int64 `bson:"conn_tot" json:"conn_tot"`
	// @Description Total de sesiones
	Stot int64 `bson:"stot" json:"stot"`
	// @Description Total de peticiones
	ReqTot int64 `bson:"req_tot" json:"req_tot"`
}

// GetCollectionName devuelve el nombre de la colección
// @Description Obtener el nombre de la colección de la base de datos
func (h *HAProxyStatsBaseline) GetCollectionName() string {
	return "haproxy_baseline"
}

// GetStats obtiene HAProxyStats desde la estructura plana
// @Description Obtener la estructura HAProxyStats desde los datos de línea base
func (h *HAProxyStatsBaseline) GetStats() HAProxyStats {
	return HAProxyStats{
		Bin:         h.Bin,
		Bout:        h.Bout,
		Hrsp1xx:     h.Hrsp1xx,
		Hrsp2xx:     h.Hrsp2xx,
		Hrsp3xx:     h.Hrsp3xx,
		Hrsp4xx:     h.Hrsp4xx,
		Hrsp5xx:     h.Hrsp5xx,
		HrspOther:   h.HrspOther,
		Dreq:        h.Dreq,
		Dresp:       h.Dresp,
		Ereq:        h.Ereq,
		Dcon:        h.Dcon,
		Dses:        h.Dses,
		Econ:        h.Econ,
		Eresp:       h.Eresp,
		ReqRateMax:  h.ReqRateMax,
		ConnRateMax: h.ConnRateMax,
		RateMax:     h.RateMax,
		Smax:        h.Smax,
		ConnTot:     h.ConnTot,
		Stot:        h.Stot,
		ReqTot:      h.ReqTot,
	}
}

// SetStats aplica HAProxyStats a la estructura plana
// @Description Aplicar los datos estadísticos a la estructura de línea base
func (h *HAProxyStatsBaseline) SetStats(stats HAProxyStats) {
	h.Bin = stats.Bin
	h.Bout = stats.Bout
	h.Hrsp1xx = stats.Hrsp1xx
	h.Hrsp2xx = stats.Hrsp2xx
	h.Hrsp3xx = stats.Hrsp3xx
	h.Hrsp4xx = stats.Hrsp4xx
	h.Hrsp5xx = stats.Hrsp5xx
	h.HrspOther = stats.HrspOther
	h.Dreq = stats.Dreq
	h.Dresp = stats.Dresp
	h.Ereq = stats.Ereq
	h.Dcon = stats.Dcon
	h.Dses = stats.Dses
	h.Econ = stats.Econ
	h.Eresp = stats.Eresp
	h.ReqRateMax = stats.ReqRateMax
	h.ConnRateMax = stats.ConnRateMax
	h.RateMax = stats.RateMax
	h.Smax = stats.Smax
	h.ConnTot = stats.ConnTot
	h.Stot = stats.Stot
	h.ReqTot = stats.ReqTot
}

// HAProxyMinuteStats almacena estadísticas por minuto de HAProxy - estructura totalmente plana
// @Description Estadísticas por minuto de HAProxy, para almacenar métricas de cada minuto
type HAProxyMinuteStats struct {
	// @Description Identificador único del registro
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// @Description Nombre del frontend de destino
	TargetName string `bson:"target_name" json:"target_name"`
	// @Description Fecha, en formato YYYY-MM-DD
	Date string `bson:"date" json:"date"`
	// @Description Hora (0-23)
	Hour int `bson:"hour" json:"hour"`
	// @Description Minuto (0-59)
	Minute int `bson:"minute" json:"minute"`
	// @Description Grupo de horas (0-4): 0-5h grupo 0, 6-11h grupo 1, 12-17h grupo 2, 18-23h grupo 3
	HourGroupSix int `bson:"hourGroupSix" json:"hourGroupSix"`
	// @Description Marca de tiempo completa
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`

	// Colocar todos los campos de métricas en el primer nivel
	// @Description Bytes de tráfico entrante
	Bin int64 `bson:"bin" json:"bin"`
	// @Description Bytes de tráfico saliente
	Bout int64 `bson:"bout" json:"bout"`
	// @Description Cantidad de respuestas con código de estado 1xx
	Hrsp1xx int64 `bson:"hrsp_1xx" json:"hrsp_1xx"`
	// @Description Cantidad de respuestas con código de estado 2xx
	Hrsp2xx int64 `bson:"hrsp_2xx" json:"hrsp_2xx"`
	// @Description Cantidad de respuestas con código de estado 3xx
	Hrsp3xx int64 `bson:"hrsp_3xx" json:"hrsp_3xx"`
	// @Description Cantidad de respuestas con código de estado 4xx
	Hrsp4xx int64 `bson:"hrsp_4xx" json:"hrsp_4xx"`
	// @Description Cantidad de respuestas con código de estado 5xx
	Hrsp5xx int64 `bson:"hrsp_5xx" json:"hrsp_5xx"`
	// @Description Cantidad de respuestas con otros códigos de estado
	HrspOther int64 `bson:"hrsp_other" json:"hrsp_other"`
	// @Description Cantidad de peticiones rechazadas
	Dreq int64 `bson:"dreq" json:"dreq"`
	// @Description Cantidad de respuestas rechazadas
	Dresp int64 `bson:"dresp" json:"dresp"`
	// @Description Cantidad de peticiones con error
	Ereq int64 `bson:"ereq" json:"ereq"`
	// @Description Cantidad de conexiones rechazadas
	Dcon int64 `bson:"dcon" json:"dcon"`
	// @Description Cantidad de sesiones rechazadas
	Dses int64 `bson:"dses" json:"dses"`
	// @Description Cantidad de errores de conexión
	Econ int64 `bson:"econ" json:"econ"`
	// @Description Cantidad de respuestas con error
	Eresp int64 `bson:"eresp" json:"eresp"`
	// @Description Tasa máxima de peticiones
	ReqRateMax int64 `bson:"req_rate_max" json:"req_rate_max"`
	// @Description Tasa máxima de conexiones
	ConnRateMax int64 `bson:"conn_rate_max" json:"conn_rate_max"`
	// @Description Tasa máxima de transferencia (bytes/segundo)
	RateMax int64 `bson:"rate_max" json:"rate_max"`
	// @Description Número máximo de sesiones concurrentes
	Smax int64 `bson:"smax" json:"smax"`
	// @Description Total de conexiones
	ConnTot int64 `bson:"conn_tot" json:"conn_tot"`
	// @Description Total de sesiones
	Stot int64 `bson:"stot" json:"stot"`
	// @Description Total de peticiones
	ReqTot int64 `bson:"req_tot" json:"req_tot"`
}

// GetCollectionName devuelve el nombre de la colección
// @Description Obtener el nombre de la colección de estadísticas por minuto
func (h *HAProxyMinuteStats) GetCollectionName() string {
	return "haproxy_minute_stats"
}

// GetStats obtiene HAProxyStats desde la estructura plana
// @Description Obtener la estructura HAProxyStats desde las estadísticas por minuto
func (h *HAProxyMinuteStats) GetStats() HAProxyStats {
	return HAProxyStats{
		Bin:         h.Bin,
		Bout:        h.Bout,
		Hrsp1xx:     h.Hrsp1xx,
		Hrsp2xx:     h.Hrsp2xx,
		Hrsp3xx:     h.Hrsp3xx,
		Hrsp4xx:     h.Hrsp4xx,
		Hrsp5xx:     h.Hrsp5xx,
		HrspOther:   h.HrspOther,
		Dreq:        h.Dreq,
		Dresp:       h.Dresp,
		Ereq:        h.Ereq,
		Dcon:        h.Dcon,
		Dses:        h.Dses,
		Econ:        h.Econ,
		Eresp:       h.Eresp,
		ReqRateMax:  h.ReqRateMax,
		ConnRateMax: h.ConnRateMax,
		RateMax:     h.RateMax,
		Smax:        h.Smax,
		ConnTot:     h.ConnTot,
		Stot:        h.Stot,
		ReqTot:      h.ReqTot,
	}
}

// SetStats aplica HAProxyStats a la estructura plana
// @Description Aplicar los datos estadísticos a la estructura de estadísticas por minuto
func (h *HAProxyMinuteStats) SetStats(stats HAProxyStats) {
	h.Bin = stats.Bin
	h.Bout = stats.Bout
	h.Hrsp1xx = stats.Hrsp1xx
	h.Hrsp2xx = stats.Hrsp2xx
	h.Hrsp3xx = stats.Hrsp3xx
	h.Hrsp4xx = stats.Hrsp4xx
	h.Hrsp5xx = stats.Hrsp5xx
	h.HrspOther = stats.HrspOther
	h.Dreq = stats.Dreq
	h.Dresp = stats.Dresp
	h.Ereq = stats.Ereq
	h.Dcon = stats.Dcon
	h.Dses = stats.Dses
	h.Econ = stats.Econ
	h.Eresp = stats.Eresp
	h.ReqRateMax = stats.ReqRateMax
	h.ConnRateMax = stats.ConnRateMax
	h.RateMax = stats.RateMax
	h.Smax = stats.Smax
	h.ConnTot = stats.ConnTot
	h.Stot = stats.Stot
	h.ReqTot = stats.ReqTot
}

// HAProxyRealTimeStats almacena estadísticas en tiempo real de HAProxy
// @Description Estadísticas en tiempo real de HAProxy, para almacenar y consultar métricas en vivo
type HAProxyRealTimeStats struct {
	// @Description Nombre del frontend de destino
	TargetName string `bson:"target_name" json:"target_name"`
	// @Description Nombre de la métrica
	MetricName string `bson:"metric_name" json:"metric_name"`
	// @Description Valor de la métrica
	Value int64 `bson:"value" json:"value"`
	// @Description Marca de tiempo del registro
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// GetCollectionName devuelve el nombre de la colección
// @Description Obtener el nombre de la colección de estadísticas en tiempo real
func (h *HAProxyRealTimeStats) GetCollectionName() string {
	return "haproxy_real_time_stats"
}

// TimeSeriesMetric métrica de series temporales
// @Description Estructura de métricas de series temporales, para almacenar y consultar datos temporales
type TimeSeriesMetric struct {
	// @Description Marca de tiempo del registro
	Timestamp time.Time `bson:"timestamp"`
	// @Description Valor de la métrica
	Value int64 `bson:"value"`
	// @Description Metadatos de la métrica
	Metadata TimeSeriesMeta `bson:"metadata"`
}

// TimeSeriesMeta metadatos de series temporales
// @Description Metadatos de las métricas de series temporales
type TimeSeriesMeta struct {
	// @Description Nombre del destino
	Target string `bson:"target"`
}

// NativeStatsToHAProxyStats convierte NativeStatStats a HAProxyStats
// @Description Convertir estadísticas nativas de HAProxy a la estructura interna
func NativeStatsToHAProxyStats(native *models.NativeStatStats) HAProxyStats {
	stats := HAProxyStats{}

	// Establecer los campos relacionados con tráfico
	if native.Bin != nil {
		stats.Bin = *native.Bin
	}
	if native.Bout != nil {
		stats.Bout = *native.Bout
	}

	// Establecer el código de estado de respuesta HTTP
	if native.Hrsp1xx != nil {
		stats.Hrsp1xx = *native.Hrsp1xx
	}
	if native.Hrsp2xx != nil {
		stats.Hrsp2xx = *native.Hrsp2xx
	}
	if native.Hrsp3xx != nil {
		stats.Hrsp3xx = *native.Hrsp3xx
	}
	if native.Hrsp4xx != nil {
		stats.Hrsp4xx = *native.Hrsp4xx
	}
	if native.Hrsp5xx != nil {
		stats.Hrsp5xx = *native.Hrsp5xx
	}
	if native.HrspOther != nil {
		stats.HrspOther = *native.HrspOther
	}

	// Establecer los campos relacionados con errores
	if native.Dreq != nil {
		stats.Dreq = *native.Dreq
	}
	if native.Dresp != nil {
		stats.Dresp = *native.Dresp
	}
	if native.Ereq != nil {
		stats.Ereq = *native.Ereq
	}
	if native.Dcon != nil {
		stats.Dcon = *native.Dcon
	}
	if native.Dses != nil {
		stats.Dses = *native.Dses
	}
	if native.Econ != nil {
		stats.Econ = *native.Econ
	}
	if native.Eresp != nil {
		stats.Eresp = *native.Eresp
	}

	// Establecer la tasa máxima
	if native.ReqRateMax != nil {
		stats.ReqRateMax = *native.ReqRateMax
	}
	if native.ConnRateMax != nil {
		stats.ConnRateMax = *native.ConnRateMax
	}
	if native.RateMax != nil {
		stats.RateMax = *native.RateMax
	}
	if native.Smax != nil {
		stats.Smax = *native.Smax
	}

	// Establecer el valor total
	if native.ConnTot != nil {
		stats.ConnTot = *native.ConnTot
	}
	if native.Stot != nil {
		stats.Stot = *native.Stot
	}
	if native.ReqTot != nil {
		stats.ReqTot = *native.ReqTot
	}

	return stats
}

// HAProxyStatsToNative convierte HAProxyStats a NativeStatStats
// @Description Convertir la estructura interna al formato nativo de estadísticas de HAProxy
func HAProxyStatsToNative(stats HAProxyStats) *models.NativeStatStats {
	native := &models.NativeStatStats{}

	// Establecer los campos relacionados con tráfico
	bin := stats.Bin
	native.Bin = &bin

	bout := stats.Bout
	native.Bout = &bout

	// Establecer el código de estado de respuesta HTTP
	hrsp1xx := stats.Hrsp1xx
	native.Hrsp1xx = &hrsp1xx

	hrsp2xx := stats.Hrsp2xx
	native.Hrsp2xx = &hrsp2xx

	hrsp3xx := stats.Hrsp3xx
	native.Hrsp3xx = &hrsp3xx

	hrsp4xx := stats.Hrsp4xx
	native.Hrsp4xx = &hrsp4xx

	hrsp5xx := stats.Hrsp5xx
	native.Hrsp5xx = &hrsp5xx

	hrspOther := stats.HrspOther
	native.HrspOther = &hrspOther

	// Establecer los campos relacionados con errores
	dreq := stats.Dreq
	native.Dreq = &dreq

	dresp := stats.Dresp
	native.Dresp = &dresp

	ereq := stats.Ereq
	native.Ereq = &ereq

	dcon := stats.Dcon
	native.Dcon = &dcon

	dses := stats.Dses
	native.Dses = &dses

	econ := stats.Econ
	native.Econ = &econ

	eresp := stats.Eresp
	native.Eresp = &eresp

	// Establecer la tasa máxima
	reqRateMax := stats.ReqRateMax
	native.ReqRateMax = &reqRateMax

	connRateMax := stats.ConnRateMax
	native.ConnRateMax = &connRateMax

	rateMax := stats.RateMax
	native.RateMax = &rateMax

	smax := stats.Smax
	native.Smax = &smax

	// Establecer el valor total
	connTot := stats.ConnTot
	native.ConnTot = &connTot

	stot := stats.Stot
	native.Stot = &stot

	reqTot := stats.ReqTot
	native.ReqTot = &reqTot

	return native
}

// CalculateStatsDelta calcula la diferencia entre dos HAProxyStats
// @Description Calcular la diferencia de estadísticas entre dos puntos temporales
func CalculateStatsDelta(last, current HAProxyStats) HAProxyStats {
	delta := HAProxyStats{}

	// Calcular la diferencia de campos de tráfico
	delta.Bin = safeSubtract(current.Bin, last.Bin)
	delta.Bout = safeSubtract(current.Bout, last.Bout)

	// Calcular la diferencia de códigos de estado de respuesta HTTP
	delta.Hrsp1xx = safeSubtract(current.Hrsp1xx, last.Hrsp1xx)
	delta.Hrsp2xx = safeSubtract(current.Hrsp2xx, last.Hrsp2xx)
	delta.Hrsp3xx = safeSubtract(current.Hrsp3xx, last.Hrsp3xx)
	delta.Hrsp4xx = safeSubtract(current.Hrsp4xx, last.Hrsp4xx)
	delta.Hrsp5xx = safeSubtract(current.Hrsp5xx, last.Hrsp5xx)
	delta.HrspOther = safeSubtract(current.HrspOther, last.HrspOther)

	// Calcular la diferencia de campos relacionados con errores
	delta.Dreq = safeSubtract(current.Dreq, last.Dreq)
	delta.Dresp = safeSubtract(current.Dresp, last.Dresp)
	delta.Ereq = safeSubtract(current.Ereq, last.Ereq)
	delta.Dcon = safeSubtract(current.Dcon, last.Dcon)
	delta.Dses = safeSubtract(current.Dses, last.Dses)
	delta.Econ = safeSubtract(current.Econ, last.Econ)
	delta.Eresp = safeSubtract(current.Eresp, last.Eresp)

	// La tasa máxima usa directamente el valor actual
	delta.ReqRateMax = current.ReqRateMax
	delta.ConnRateMax = current.ConnRateMax
	delta.RateMax = current.RateMax
	delta.Smax = current.Smax

	// Peticiones, conexiones y sesiones usan directamente la diferencia
	delta.ConnTot = safeSubtract(current.ConnTot, last.ConnTot)
	delta.Stot = safeSubtract(current.Stot, last.Stot)
	delta.ReqTot = safeSubtract(current.ReqTot, last.ReqTot)

	return delta
}

// safeSubtract resta segura, garantiza que el resultado no sea negativo
// @Description Operación de resta segura, gestiona reinicios de contadores por reinicios de HAProxy
func safeSubtract(current, last int64) int64 {
	if current < last {
		// Probablemente causado por un reinicio de HAProxy, devolver el valor actual como incremento
		return current
	}
	return current - last
}

// CreateZeroStats crea HAProxyStats con todas las métricas en 0, para registro tras reinicio
// @Description Crear estadísticas en cero para el registro tras un reinicio de HAProxy
func CreateZeroStats() HAProxyStats {
	return HAProxyStats{} // Valor cero de una estructura Go, todos los campos en 0
}

// DetectReset detecta si HAProxy se ha reiniciado
// @Description Detectar si HAProxy se reinició comparando estadísticas de dos puntos temporales
func DetectReset(lastStats, currentStats HAProxyStats) bool {
	// Comprobar bin (bytes entrantes)
	if currentStats.Bin < lastStats.Bin {
		return true
	}

	// Comprobar bout (bytes salientes)
	if currentStats.Bout < lastStats.Bout {
		return true
	}

	// Comprobar conexiones totales
	if currentStats.ConnTot < lastStats.ConnTot {
		return true
	}

	// Comprobar sesiones totales
	if currentStats.Stot < lastStats.Stot {
		return true
	}

	// Comprobar peticiones totales
	if currentStats.ReqTot < lastStats.ReqTot {
		return true
	}

	return false
}
