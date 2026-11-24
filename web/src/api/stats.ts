import { get } from './index'
import {
  OverviewStats,
  RealtimeQPSResponse,
  TimeSeriesResponse,
  CombinedTimeSeriesResponse,
  TrafficTimeSeriesResponse,
  TimeRange
} from '@/types/stats'

// Ruta base de la API de estadísticas
const BASE_URL = '/stats'

/**
 * Servicios de API relacionados con estadísticas
 */
export const statsApi = {
  /**
   * Obtener datos de resumen estadístico
   * @param timeRange rango de tiempo
   * @returns datos de resumen estadístico
   */
  getOverviewStats: (timeRange: TimeRange = '24h'): Promise<OverviewStats> => {
    return get<OverviewStats>(`${BASE_URL}/overview`, {
      params: { timeRange }
    })
  },

  /**
   * Obtener datos QPS en tiempo real
   * @param limit límite de puntos de datos, por defecto 30
   * @returns datos QPS en tiempo real
   */
  getRealtimeQPS: (limit: number = 30): Promise<RealtimeQPSResponse> => {
    return get<RealtimeQPSResponse>(`${BASE_URL}/realtime-qps`, {
      params: { limit }
    })
  },

  /**
   * Obtener datos de series temporales
   * @param timeRange rango de tiempo
   * @param metric tipo de métrica
   * @returns datos de series temporales
   */
  getTimeSeriesData: (
    timeRange: TimeRange = '24h',
    metric: 'requests' | 'blocks' = 'requests'
  ): Promise<TimeSeriesResponse> => {
    return get<TimeSeriesResponse>(`${BASE_URL}/time-series`, {
      params: { timeRange, metric }
    })
  },

  /**
   * Obtener datos de series temporales combinadas (peticiones y bloqueos)
   * @param timeRange rango de tiempo
   * @returns datos de series temporales combinadas
   */
  getCombinedTimeSeriesData: (timeRange: TimeRange = '24h'): Promise<CombinedTimeSeriesResponse> => {
    return get<CombinedTimeSeriesResponse>(`${BASE_URL}/combined-time-series`, {
      params: { timeRange }
    })
  },

  /**
   * Obtener datos de series temporales de tráfico
   * @param timeRange rango de tiempo
   * @returns datos de series temporales de tráfico
   */
  getTrafficTimeSeriesData: (timeRange: TimeRange = '24h'): Promise<TrafficTimeSeriesResponse> => {
    return get<TrafficTimeSeriesResponse>(`${BASE_URL}/traffic-time-series`, {
      params: { timeRange }
    })
  }
}