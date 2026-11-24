// Tipo de rango de tiempo
export type TimeRange = '24h' | '7d' | '30d';

// Estadísticas de resumen
export interface OverviewStats {
  timeRange: TimeRange;
  totalRequests: number;
  inboundTraffic: number;
  outboundTraffic: number;
  maxQPS: number;
  error4xx: number;
  error4xxRate: number;
  error5xx: number;
  error5xxRate: number;
  blockCount: number;
  attackIPCount: number;
}

// Punto de datos QPS en tiempo real
export interface RealtimeQPSData {
  timestamp: string;
  value: number;
}

// Respuesta QPS en tiempo real
export interface RealtimeQPSResponse {
  data: RealtimeQPSData[];
}

// Punto de datos de serie temporal
export interface TimeSeriesDataPoint {
  timestamp: string;
  value: number;
}

// Respuesta de series temporales de datos de peticiones y bloqueos
export interface TimeSeriesResponse {
  metric: 'requests' | 'blocks';
  timeRange: TimeRange;
  data: TimeSeriesDataPoint[];
}

// Respuesta combinada de series temporales (peticiones y bloqueos)
export interface CombinedTimeSeriesResponse {
  timeRange: TimeRange;
  requests: TimeSeriesResponse;
  blocks: TimeSeriesResponse;
}

// Punto de datos de tráfico
export interface TrafficDataPoint {
  timestamp: string;
  inboundTraffic: number;
  outboundTraffic: number;
}

// Respuesta de series temporales de tráfico
export interface TrafficTimeSeriesResponse {
  timeRange: TimeRange;
  data: TrafficDataPoint[];
}