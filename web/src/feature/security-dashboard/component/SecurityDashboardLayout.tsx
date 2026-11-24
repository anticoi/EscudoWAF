import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Globe3DMap } from './globe3D-map'
import { WAFAttackTrajectory } from './globe3D-map/types'
import { StatCard } from './StatCard'
import { AttackIPList } from './AttackIPList'
import { RealtimeAttackList } from './RealtimeAttackList'
import { DashboardQPSChart } from './DashboardQPSChart'
import { useSecurityDashboard } from '../hooks/useSecurityDashboard'
import { Maximize2, Minimize2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { OverviewStats } from '@/types/stats'
import { AttackEventAggregateResult } from '@/types/log'
import { useDebounce } from '@/hooks/useDebounce'
import { geoDistance } from 'd3-geo';

/**
 * Componente de layout de pantalla grande de seguridad
 * Incluye barra de título superior, tarjetas de estadísticas a la izquierda, lista de ataques en tiempo real a la derecha, gráfico QPS inferior y globo 3D de fondo
 */
export const SecurityDashboardLayout: React.FC = () => {
    const { t, i18n } = useTranslation()
    const [currentTime, setCurrentTime] = useState(new Date())
    const [isFullscreen, setIsFullscreen] = useState(false)

    // Gestión de estado independiente - solo actualiza cuando los datos cambian realmente
    const [globeData, setGlobeData] = useState<WAFAttackTrajectory[]>([])
    const globeDataDebounce = useDebounce(globeData, 10000) // Debounce de 10 segundos; debe ser mayor que el intervalo de sondeo
    const [statsData, setStatsData] = useState<OverviewStats | null>(null)
    const [realtimeAttacksData, setRealtimeAttacksData] = useState<AttackEventAggregateResult[]>([])
    const [attackIPsData, setAttackIPsData] = useState<AttackEventAggregateResult[]>([])

    // Ref para detectar cambios de datos - evita renderizados infinitos
    const prevDataRef = useRef<{
        overviewStats: OverviewStats | null
        attackEvents: typeof attackEvents.data | null
    }>({
        overviewStats: null,
        attackEvents: null
    })

    // Obtener los datos del panel grande
    const {
        overviewStats,
        attackEvents,
        realtimeAttacks,
        attackIPs
    } = useSecurityDashboard()

    // Actualizador de hora - actualiza la hora cada segundo
    useEffect(() => {
        const timeUpdateRef = setInterval(() => {
            setCurrentTime(new Date())
        }, 1000)

        return () => {
            clearInterval(timeUpdateRef)
        }
    }, [])

    // Escuchar cambios de estado de pantalla completa
    useEffect(() => {
        const handleFullscreenChange = () => {
            setIsFullscreen(!!document.fullscreenElement)
        }

        document.addEventListener('fullscreenchange', handleFullscreenChange)
        return () => {
            document.removeEventListener('fullscreenchange', handleFullscreenChange)
        }
    }, [])

    // Formatear la visualización de la hora
    const formatCurrentTime = (date: Date) => {
        const locale = i18n.language === 'zh' ? 'zh-CN' : 'en-US'
        return date.toLocaleString(locale, {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        }).replace(/\//g, '-')
    }

    // Cambio a pantalla completa
    const toggleFullscreen = () => {
        if (!document.fullscreenElement) {
            document.documentElement.requestFullscreen()
        } else {
            document.exitFullscreen()
        }
    }

    // Escuchar cambios en estadísticas, solo actualizar estado cuando cambian realmente
    useEffect(() => {
        const currentStats = overviewStats.data
        const prevStats = prevDataRef.current.overviewStats

        // Comprobar si los datos estadísticos cambiaron
        const statsChanged = JSON.stringify(currentStats) !== JSON.stringify(prevStats)

        if (statsChanged) {
            console.log('Los datos estadísticos cambiaron, actualizando estado...')
            prevDataRef.current.overviewStats = currentStats || null
            setStatsData(currentStats || null)
        }
    }, [overviewStats.data])


    // Escuchar cambios en datos de eventos de ataque, actualizar el globo 3D y la lista de IPs atacantes
    useEffect(() => {
        const currentAttackEvents = attackEvents.data
        const prevAttackEvents = prevDataRef.current.attackEvents

        // Comprobar si los datos de eventos de ataque cambiaron (los datos ya están normalizados en el hook)
        const attackEventsChanged = JSON.stringify(currentAttackEvents) !== JSON.stringify(prevAttackEvents)


        if (!attackEventsChanged) {
            return // Los datos no cambiaron, no actualizar
        }

        console.log('Los datos de eventos de ataque cambiaron, actualizando el globo 3D y las listas de IPs atacantes y ataques en tiempo real...')
        prevDataRef.current.attackEvents = currentAttackEvents

        if (!currentAttackEvents?.results || currentAttackEvents?.results?.length === 0) {
            setGlobeData([])
            setAttackIPsData([])
            setRealtimeAttacksData([])
            return
        }

        // Actualizar los datos del globo 3D
        const newTrajectoryData = currentAttackEvents.results
            .filter(event => event.srcIpInfo?.location?.latitude) // Mostrar solo ataques con información de ubicación
            .map((event, index) => ({
                type: "waf_attack",
                order: index + 1,
                from: `${event.srcIp} (${event.srcIpInfo.location.latitude.toFixed(2)}, ${event.srcIpInfo.location.longitude.toFixed(2)})`,
                to: "Centro de protección WAF",
                flightCode: event.domain,
                date: event.firstAttackTime,
                status: event.isOngoing,
                startLat: event.srcIpInfo.location.latitude,
                startLng: event.srcIpInfo.location.longitude,
                endLat: 30.274084, // Coordenadas de Hangzhou
                endLng: 120.155070,
                // arcAlt: Math.min(0.3, Math.max(0.05, event.count / 500)),
                arcAlt: geoDistance([event.srcIpInfo.location.longitude, event.srcIpInfo.location.latitude], [120.155070, 30.274084]), // Calcular la distancia geográfica entre origen y destino (la línea genera la altura de la curva Bézier según la distancia)
                colorIndex: Math.abs(event.srcIp.split('.').reduce((a, b) => a + parseInt(b), 0)) % 8 // Generar un índice de color fijo basado en la dirección IP
            }))

        setGlobeData(newTrajectoryData)

        // Actualizar los datos de la lista de IPs atacantes
        setAttackIPsData(attackIPs)
        setRealtimeAttacksData(realtimeAttacks)
    }, [attackEvents.data])

    return (
        <div
            className={`relative w-full bg-gradient-to-br from-[#0d0c27] via-[#1a1336] to-[#2d1b54] overflow-hidden ${isFullscreen
                ? 'fixed inset-0 z-[9999] h-screen w-screen'
                : 'h-screen'
                }`}
            style={isFullscreen ? {
                margin: 0,
                padding: 0,
                width: '100vw',
                height: '100vh'
            } : undefined}
        >
            {/* Fondo de globo 3D - usa datos de estado independientes */}
            <div className="absolute inset-0 z-0 w-full h-full">
                <Globe3DMap wafAttackTrajectoryData={globeDataDebounce} />
            </div>

            {/* Barra de título superior */}
            <div className="absolute top-0 left-0 right-0 z-10 h-24">
                <div className="flex items-center justify-between h-full px-6">
                    <div className="flex items-center ml-2">
                        <div className="font-bold text-2xl gap-2 flex">
                            <span className="text-[#E8DFFF] dark:text-[#F0EBFF] text-shadow-glow-purple transition-all duration-300">EscudoWAF</span>
                            <span className="text-[#8ED4FF] dark:text-[#A5DEFF] text-shadow-glow-blue transition-all duration-300">{t('securityDashboard.title')}</span>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="text-white text-lg font-mono text-shadow-glow-white">
                            {formatCurrentTime(currentTime)}
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={toggleFullscreen}
                            className="text-white transition-all duration-200"
                            title={isFullscreen ? "Salir de pantalla completa" : "Entrar en pantalla completa"}
                        >
                            {isFullscreen ? (
                                <Minimize2 className="w-4 h-4" />
                            ) : (
                                <Maximize2 className="w-4 h-4" />
                            )}
                        </Button>
                    </div>
                </div>
            </div>

            {/* Área de tarjetas de estadísticas izquierda - usa datos de estado independientes */}
            <div className="absolute left-6 top-28 z-10 w-64 flex flex-col">
                {/* Las tres primeras tarjetas de estadísticas */}
                <div className="flex-none space-y-0.5">
                    <StatCard
                        title={t('securityDashboard.stats.blockCount24h')}
                        value={statsData?.blockCount || 0}
                    />
                    <StatCard
                        title={t('securityDashboard.stats.attackIPCount24h')}
                        value={statsData?.attackIPCount || 0}
                    />
                    <StatCard
                        title={t('securityDashboard.stats.totalRequests24h')}
                        value={statsData?.totalRequests || 0}
                    />
                </div>

                {/* Cuarta tarjeta - lista de IPs atacantes, usa datos de estado independientes */}
                <div className="flex-1 min-h-0 mt-20">
                    <AttackIPList
                        attackIPs={attackIPsData}
                        isLoading={attackEvents.isLoading}
                    />
                </div>
            </div>

            {/* Lista de ataques en tiempo real a la derecha - usa datos de estado independientes */}
            <div className="absolute right-6 top-28 bottom-48 z-10 w-80">
                <RealtimeAttackList
                    realtimeAttacks={realtimeAttacksData}
                    isLoading={attackEvents.isLoading}
                />
            </div>

            {/* Gráfico QPS inferior */}
            <div className="absolute left-2 right-2 bottom-2 z-10 h-32">
                <DashboardQPSChart />
            </div>
        </div>
    )
} 