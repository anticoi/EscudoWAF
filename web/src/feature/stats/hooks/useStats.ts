import { useQuery } from '@tanstack/react-query'
import { statsApi } from '@/api/services'
import { TimeRange } from '@/types/stats'
import { useState, useRef, useEffect } from 'react'

const REFRESH_INTERVAL = 60000 // 60 segundos
const QPS_REFRESH_INTERVAL = 5000 // 5 segundos
const QPS_STALETIME = 4000 // 4 segundos

// Definición del tipo de punto de datos QPS
export interface QPSDataPoint {
    timestamp: string
    value: number
}

// Hook de datos del resumen estadístico
export const useOverviewStats = (timeRange: TimeRange) => {
    return useQuery({
        queryKey: ['stats', 'overview', timeRange],
        queryFn: () => statsApi.getOverviewStats(timeRange),
        refetchInterval: REFRESH_INTERVAL, // Usar constantes
    })
}

// Hook de datos QPS en tiempo real - versión de actualización incremental con scroll
export const useRealtimeQPS = (limit: number = 30) => {
    // Almacenamiento de estado local
    const [localQPSData, setLocalQPSData] = useState<QPSDataPoint[]>([])
    const lastTimestampRef = useRef<string | null>(null)
    const isInitializedRef = useRef(false)
    const shouldScrollRef = useRef(true)

    // Consultar datos de la API
    const queryResult = useQuery({
        queryKey: ['stats', 'realtimeQPS', limit],
        queryFn: () => statsApi.getRealtimeQPS(limit),
        refetchInterval: QPS_REFRESH_INTERVAL, // Usar constantes
        refetchIntervalInBackground: true, // Seguir actualizando aunque la página esté inactiva
        staleTime: QPS_STALETIME, // Tiempo que los datos permanecen frescos
        refetchOnWindowFocus: false, // No recargar automáticamente cuando la ventana recibe el foco
    })

    // Procesamiento de datos
    useEffect(() => {
        if (!queryResult.data?.data || queryResult.data.data.length === 0) return

        const apiData = queryResult.data.data

        // Si los datos están vacíos o es la primera carga
        if (localQPSData.length === 0 || !isInitializedRef.current) {
            setLocalQPSData(apiData)
            lastTimestampRef.current = apiData[apiData.length - 1].timestamp
            isInitializedRef.current = true
            return
        }

        // Obtener la marca de tiempo más reciente
        const newestTimestamp = apiData[apiData.length - 1].timestamp

        // Si hay datos nuevos
        if (newestTimestamp !== lastTimestampRef.current && shouldScrollRef.current) {
            // Encontrar los datos incrementales
            const lastIndex = apiData.findIndex(item => item.timestamp === lastTimestampRef.current)

            if (lastIndex !== -1 && lastIndex < apiData.length - 1) {
                // Hay datos nuevos
                const newPoints = apiData.slice(lastIndex + 1)

                setLocalQPSData(prevData => {
                    const newData = [...prevData, ...newPoints]
                    return newData.slice(-limit) // Mantener la cantidad de datos dentro del límite
                })
            } else {
                // La estructura de datos cambió, reemplazar completamente
                setLocalQPSData(apiData)
            }

            // Actualizar la última marca de tiempo
            lastTimestampRef.current = newestTimestamp
        }
    }, [queryResult.data, limit, localQPSData.length])

    // Pausar/reanudar actualizaciones de scroll
    const toggleScroll = (shouldScroll: boolean) => {
        shouldScrollRef.current = shouldScroll
    }

    // Devolver el objeto de resultado extendido
    return {
        ...queryResult,
        localData: localQPSData,
        isInitialized: isInitializedRef.current,
        toggleScroll, // Exponer métodos para controlar el scroll
        refreshInterval: QPS_REFRESH_INTERVAL // Exponer el intervalo de actualización
    }
}

// Hook de datos de series temporales de tráfico
export const useTrafficTimeSeriesData = (timeRange: TimeRange) => {
    return useQuery({
        queryKey: ['stats', 'trafficTimeSeries', timeRange],
        queryFn: () => statsApi.getTrafficTimeSeriesData(timeRange),
        refetchInterval: REFRESH_INTERVAL, // Usar constantes
    })
}

// Hook de datos de series temporales combinadas de peticiones y bloqueos
export const useCombinedTimeSeriesData = (timeRange: TimeRange) => {
    return useQuery({
        queryKey: ['stats', 'combinedTimeSeries', timeRange],
        queryFn: () => statsApi.getCombinedTimeSeriesData(timeRange),
        refetchInterval: REFRESH_INTERVAL, // Usar constantes
    })
}

// Hook de selección de rango de tiempo
export const useTimeRangeSelector = (initialRange: TimeRange = '24h') => {
    const [timeRange, setTimeRange] = useState<TimeRange>(initialRange)

    return {
        timeRange,
        setTimeRange,
    }
}