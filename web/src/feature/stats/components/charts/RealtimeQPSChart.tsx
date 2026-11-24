import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useRealtimeQPS } from '../../hooks/useStats'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { History } from 'lucide-react'
import { AnimatedIcon } from '@/components/ui/animation/components/animated-icon'
import * as echarts from 'echarts'
import { useTheme } from '@/provider/theme-context'
import WaveIcon from '@/components/icon/wave-icon'

export function RealtimeQPSChart() {
    const { t } = useTranslation()
    const chartRef = useRef<HTMLDivElement>(null)
    const chartInstanceRef = useRef<echarts.ECharts | null>(null)
    const { theme } = useTheme()
    const [isResetAnimating, setIsResetAnimating] = useState(false)
    // Comprobar si es modo oscuro
    const isDarkMode = theme === 'dark'

    // Usar el hook modificado para obtener datos, ahora incluye refreshInterval
    const { localData, isLoading, isInitialized, refreshInterval } = useRealtimeQPS(30)

    // Valor QPS actual - tomar el punto de datos más reciente
    const currentQPS = localData.length > 0 ? localData[localData.length - 1].value : 0

    // Formatear la marca de tiempo - mostrar hora:minutos:segundos
    const formatTimeLabel = (timestamp: string) => {
        const date = new Date(timestamp)
        return date.toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        })
    }

    // when data initialized and refreshInterval is set, start the animation
    useEffect(() => {
        if (!refreshInterval || !isInitialized) return

        // Crear temporizador sincronizado con el ciclo de actualización de datos
        const animationTimer = setInterval(() => {
            // Disparar la animación
            setIsResetAnimating(true)

            // Terminar la animación tras 1 segundo
            setTimeout(() => {
                setIsResetAnimating(false)
            }, 1000)
        }, refreshInterval)

        // Función de limpieza
        return () => {
            clearInterval(animationTimer)
        }
    }, [refreshInterval, isInitialized])


    // Inicializar y actualizar el gráfico
    useEffect(() => {
        if (!chartRef.current || localData.length === 0) return

        // Si ya existe una instancia del gráfico, no es necesario recrearla
        if (!chartInstanceRef.current) {
            chartInstanceRef.current = echarts.init(chartRef.current)
        }

        // Usar el color correspondiente según el tema
        const CHART_PRIMARY_COLOR = isDarkMode ? '#b394e9' : '#9d76db' // Tema púrpura
        const CHART_SECONDARY_COLOR = isDarkMode ? 'rgba(179, 148, 233, 0.5)' : 'rgba(157, 118, 219, 0.5)' // Púrpura semitransparente

        // Procesar los datos, asegurar que los valores 0 también muestren pequeñas barras
        const processedData = localData.map(item => {
            // Si el valor es 0, usar un mínimo de 0.05 para que se muestre
            return item.value === 0 ? 0.05 : item.value
        })

        // Establecer las opciones del gráfico
        const option: echarts.EChartsOption = {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow',
                    shadowStyle: {
                        color: isDarkMode ? 'rgba(179, 148, 233, 0.15)' : 'rgba(157, 118, 219, 0.05)'
                    }
                },
                formatter: function (params) {
                    if (!params || !Array.isArray(params) || params.length === 0) {
                        return ''
                    }

                    const dataIndex = params[0].dataIndex
                    if (typeof dataIndex === 'number' && dataIndex >= 0 && dataIndex < localData.length) {
                        const item = localData[dataIndex]
                        // Mostrar el valor real en lugar del procesado
                        return `${formatTimeLabel(item.timestamp)}: <span style="font-weight: bold; color: ${CHART_PRIMARY_COLOR}">${item.value}</span> QPS`
                    }
                    return ''
                },
                backgroundColor: isDarkMode ? 'rgba(36, 37, 46, 0.95)' : 'rgba(255, 255, 255, 0.9)',
                borderColor: isDarkMode ? 'rgba(179, 148, 233, 0.2)' : 'rgba(157, 118, 219, 0.2)',
                textStyle: {
                    color: isDarkMode ? '#e0e0e0' : '#333'
                }
            },
            grid: {
                left: 0,  // Eliminar completamente el margen izquierdo
                right: 0, // Eliminar completamente el margen derecho
                bottom: 0, // Quitar el margen inferior
                top: '5%',
                containLabel: false // No reservar espacio para etiquetas
            },
            xAxis: {
                type: 'category',
                boundaryGap: true,
                data: localData.map(item => formatTimeLabel(item.timestamp)),
                axisLabel: {
                    show: false // Ocultar las etiquetas del eje X
                },
                axisLine: {
                    show: false // Ocultar la línea del eje X
                },
                axisTick: {
                    show: false // Ocultar las marcas de graduación
                },
                splitLine: {
                    show: false // Ocultar las líneas divisorias
                }
            },
            yAxis: {
                type: 'value',
                show: false, // Ocultar completamente el eje Y
                max: function (value) {
                    // Dejar un margen superior en el gráfico, valor máximo +20%
                    return Math.max(10, value.max * 1.2)
                },
                min: 0 // Asegurar que el eje Y empiece en 0 para que las barras pequeñas se vean
            },
            series: [
                {
                    name: 'QPS',
                    type: 'bar',
                    barWidth: '65%',  // Ajustar el ancho de las barras al 65%
                    barCategoryGap: '10%',  // Aumentar apropiadamente el espacio entre categorías
                    barGap: '0%',  // Eliminar el espacio entre barras
                    itemStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: CHART_PRIMARY_COLOR },
                            { offset: 1, color: CHART_SECONDARY_COLOR }
                        ]),
                        borderRadius: [4, 4, 0, 0], // Esquinas superiores redondeadas del gráfico de barras
                        shadowColor: isDarkMode ? 'rgba(179, 148, 233, 0.6)' : 'rgba(157, 118, 219, 0.4)',
                        shadowBlur: isDarkMode ? 15 : 10
                    },
                    data: processedData, // Usar los datos procesados
                    animationDuration: 300,
                    animationEasing: 'cubicOut',
                    animationDelay: function (idx) {
                        return idx * 5 // Animación rápida
                    }
                }
            ],
            // Añadir línea discontinua inferior
            markLine: {
                silent: true,
                symbol: 'none',
                lineStyle: {
                    color: isDarkMode ? 'rgba(179, 148, 233, 0.25)' : 'rgba(157, 118, 219, 0.15)',
                    type: 'dashed'
                },
                data: [
                    {
                        yAxis: 0 // Línea en el punto 0 del eje Y
                    }
                ]
            },
            animation: true
        }

        // Opciones de la aplicación
        chartInstanceRef.current.setOption(option)

        // Redibujar el gráfico al cambiar el tamaño de la ventana
        const handleResize = () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.resize()
            }
        }

        window.addEventListener('resize', handleResize)

        // Función de limpieza
        return () => {
            window.removeEventListener('resize', handleResize)
        }
    }, [localData, isInitialized, isDarkMode])

    return (
        <Card className="border-none shadow-none">
            <CardHeader className="flex flex-row items-center justify-between p-4 pb-2">
                <div className="flex items-center gap-3">
                    <CardTitle className="text-lg font-medium dark:text-shadow-glow-white">
                        {t('stats.realtimeQPS')}
                    </CardTitle>
                    <div className="flex items-center bg-gray-50 dark:bg-gray-800 rounded-md py-1 px-3">
                        <span className="text-primary dark:text-primary mr-2 mb-2">
                            {isDarkMode ? <WaveIcon width={20} height={20} /> : <WaveIcon width={20} height={20} color="#B68AFE" />}
                        </span>
                        <span className="font-medium text-gray-800 dark:text-white">{currentQPS}</span>
                    </div>
                </div>
                <div className="h-8 w-8 flex items-center justify-center text-gray-500 dark:text-gray-400 qps-refresh-icon">
                    <AnimatedIcon animationVariant="continuous-spin" isAnimating={isResetAnimating}>
                        <History className="h-5 w-5" />
                    </AnimatedIcon>
                </div>
            </CardHeader>
            <CardContent className="px-4 pb-4 pt-0">
                <div className="h-[180px]">
                    <div
                        ref={chartRef}
                        style={{ width: '100%', height: '100%' }}
                        className={isLoading && localData.length === 0 ? "opacity-50" : ""}
                    />
                </div>
            </CardContent>
        </Card>
    )
}