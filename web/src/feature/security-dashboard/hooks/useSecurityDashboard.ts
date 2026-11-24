import { useOverviewStats } from '@/feature/stats/hooks/useStats'
import { useAttackEvents } from '@/feature/log/hook/useAttackEvents'
import { useMemo, useRef, useEffect, useState } from 'react'

const POLLING_INTERVAL = 5000 // Sondeo cada 5 segundos

/**
 * Hook de gestión de datos de la pantalla grande de seguridad
 * Proporcionar estadísticas y datos de eventos de ataque en tiempo real
 */
export const useSecurityDashboard = () => {
    const pollingTimerRef = useRef<number | null>(null)
    
    // Gestionar los parámetros de tiempo en el estado para que cada sondeo los actualice
    const [timeParams, setTimeParams] = useState(() => ({
        startTime: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
        endTime: new Date().toISOString()
    }))

    // Obtener datos del resumen estadístico de 24 horas
    const overviewStats = useOverviewStats('24h')

    // Construir parámetros de consulta con los parámetros de tiempo del estado
    const attackEventsParams = useMemo(() => ({
        pageSize: 50, // Obtener más datos para mostrar
        startTime: timeParams.startTime,
        endTime: timeParams.endTime
    }), [timeParams.startTime, timeParams.endTime])

    // Obtener datos de eventos de ataque de 24 horas
    const attackEvents = useAttackEvents(attackEventsParams)

    // Función para actualizar los parámetros de tiempo
    const updateTimeParams = () => {
        setTimeParams({
            startTime: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
            endTime: new Date().toISOString()
        })
    }

    // Establecer sondeo de 5 segundos
    useEffect(() => {
        // Limpiar el sondeo existente
        if (pollingTimerRef.current !== null) {
            clearInterval(pollingTimerRef.current)
            pollingTimerRef.current = null
        }

        // Establecer un nuevo sondeo - intervalo de 5 segundos
        pollingTimerRef.current = window.setInterval(() => {
            // Actualizar los parámetros de tiempo para obtener los datos de las últimas 24 horas
            // La actualización de parámetros de tiempo dispara automáticamente una nueva petición de useAttackEvents, sin refetch manual
            updateTimeParams()
            
            // Solo es necesario refrescar las estadísticas manualmente, no dependen de parámetros de tiempo
            overviewStats.refetch()
        }, POLLING_INTERVAL)

        // Limpiar el sondeo al desmontar el componente
        return () => {
            if (pollingTimerRef.current !== null) {
                clearInterval(pollingTimerRef.current)
            }
        }
    }, [overviewStats.refetch])

    // Filtrar eventos de ataque con geolocalización, para la lista de ataques en tiempo real
    const realtimeAttacks = useMemo(() => {
        if (!attackEvents.data?.results) return []

        return attackEvents.data.results
            .filter(event => event.srcIpInfo?.city && event.srcIpInfo?.city?.nameZh && event.srcIpInfo?.city?.nameEn) // Mostrar solo ataques con información de ciudad
            .slice(0, 9) // Mostrar solo los 10 más recientes
    }, [attackEvents.data])

    // Lista de IPs atacantes, para la tarjeta izquierda 4
    const attackIPs = useMemo(() => {
        if (!attackEvents.data?.results) return []

        return attackEvents.data.results
            .filter(event => event.srcIpInfo?.city) // Mostrar solo ataques con información de ciudad
            .slice(0, 5) // Mostrar solo los 5 más recientes
    }, [attackEvents.data])

    return {
        // Datos estadísticos
        overviewStats: {
            data: overviewStats.data,
            isLoading: overviewStats.isLoading,
            error: overviewStats.error
        },
        // Datos de eventos de ataque
        attackEvents: {
            data: attackEvents.data,
            isLoading: attackEvents.isLoading,
            error: attackEvents.error
        },
        // Datos procesados
        realtimeAttacks,
        attackIPs,
        // Intervalo de sondeo
        pollingInterval: POLLING_INTERVAL
    }
} 