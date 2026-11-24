import { useEffect, useRef } from 'react'
import { useRealtimeQPS } from '@/feature/stats/hooks/useStats'
import * as echarts from 'echarts'

/**
 * Componente de gráfico QPS en tiempo real para pantalla grande
 * Adaptar al layout inferior, usar más puntos de datos
 * Versión optimizada: efecto de brillo y experiencia visual mejorados
 */
export function DashboardQPSChart() {
    const chartRef = useRef<HTMLDivElement>(null)
    const chartInstanceRef = useRef<echarts.ECharts | null>(null)

    // Usar 120 puntos de datos, adecuado para pantallas grandes
    const { localData, isInitialized } = useRealtimeQPS(120)

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

    // Inicializar y actualizar el gráfico
    useEffect(() => {
        if (!chartRef.current) return

        // Si ya existe una instancia del gráfico, no es necesario recrearla
        if (!chartInstanceRef.current) {
            chartInstanceRef.current = echarts.init(chartRef.current)
        }

        // Si no hay datos, mostrar gráfico vacío
        if (localData.length === 0) {
            chartInstanceRef.current.setOption({
                grid: {
                    left: '2%',
                    right: '2%',
                    bottom: '20%',
                    top: '5%',
                    containLabel: false
                },
                xAxis: {
                    type: 'category',
                    data: [],
                    axisLabel: {
                        color: 'rgba(255, 255, 255, 0.8)',
                        fontSize: 10
                    },
                    axisLine: {
                        lineStyle: {
                            color: 'rgba(255, 255, 255, 0.3)'
                        }
                    }
                },
                yAxis: {
                    type: 'value',
                    axisLabel: {
                        show: false
                    },
                    splitLine: {
                        show: false
                    }
                },
                series: [{
                    type: 'bar',
                    data: []
                }]
            })
            return
        }

        // Configuración de colores para pantalla grande - versión mejorada
        const CHART_PRIMARY_COLOR = '#a071da' // Tema púrpura
        const CHART_SECONDARY_COLOR = 'rgba(160, 113, 218, 0.3)' // Púrpura semitransparente
        const CHART_GLOW_COLOR = 'rgba(160, 113, 218, 0.8)' // Color de brillo
        const CHART_BRIGHT_COLOR = '#c299f0' // Púrpura más brillante para la parte superior del degradado

        // Procesar los datos, asegurar que los valores 0 también muestren pequeñas barras
        const processedData = localData.map(item => {
            // Si el valor es 0, usar un mínimo de 0.1 para mostrar el efecto de brillo
            return item.value === 0 ? 0.1 : item.value
        })

        // Establecer las opciones del gráfico
        const option: echarts.EChartsOption = {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow',
                    shadowStyle: {
                        color: 'rgba(160, 113, 218, 0.15)'
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
                        return `<div style="text-shadow: 0 0 8px rgba(160, 113, 218, 0.8);">${formatTimeLabel(item.timestamp)}: <span style="font-weight: bold; color: ${CHART_BRIGHT_COLOR}; text-shadow: 0 0 10px rgba(160, 113, 218, 1);">${item.value}</span> QPS</div>`
                    }
                    return ''
                },
                backgroundColor: 'transparent',
                borderColor: 'transparent',
                textStyle: {
                    color: '#fff',
                    fontSize: 12,
                    fontWeight: 'normal'
                },
                extraCssText: 'box-shadow: none; backdrop-filter: none;'
            },
            grid: {
                left: '2%',
                right: '2%',
                bottom: '20%',
                top: '5%',
                containLabel: false
            },
            xAxis: {
                type: 'category',
                boundaryGap: true,
                data: localData.map(item => formatTimeLabel(item.timestamp)),
                axisLabel: {
                    show: true,
                    color: 'rgba(255, 255, 255, 0.8)',
                    fontSize: 10,
                    interval: Math.floor(localData.length / 8) // Mostrar unas 8 etiquetas
                },
                axisLine: {
                    show: true,
                    lineStyle: {
                        color: 'rgba(255, 255, 255, 0.3)'
                    }
                },
                axisTick: {
                    show: false
                },
                splitLine: {
                    show: false
                }
            },
            yAxis: {
                type: 'value',
                axisLabel: {
                    show: false
                },
                axisLine: {
                    show: false
                },
                axisTick: {
                    show: false
                },
                splitLine: {
                    show: false
                },
                max: function (value) {
                    // Dejar un margen superior en el gráfico, valor máximo +20%
                    return Math.max(10, value.max * 1.2)
                },
                min: 0
            },
            series: [
                {
                    name: 'QPS',
                    type: 'bar',
                    barWidth: '60%',
                    itemStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            { offset: 0, color: CHART_BRIGHT_COLOR },
                            { offset: 0.3, color: CHART_PRIMARY_COLOR },
                            { offset: 1, color: CHART_SECONDARY_COLOR }
                        ]),
                        borderRadius: [4, 4, 0, 0], // Aumentar el redondeo de esquinas
                        // Efecto de brillo mejorado - sombras multicapa
                        shadowColor: CHART_GLOW_COLOR,
                        shadowBlur: 20, // Aumentar el radio de desenfoque
                        shadowOffsetY: 0,
                        shadowOffsetX: 0
                    },
                    // Añadir efecto de brillo para el estado de énfasis
                    emphasis: {
                        itemStyle: {
                            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                { offset: 0, color: '#d4b3ff' },
                                { offset: 0.3, color: '#b388e6' },
                                { offset: 1, color: 'rgba(160, 113, 218, 0.5)' }
                            ]),
                            shadowColor: CHART_GLOW_COLOR,
                            shadowBlur: 30, // Brillo más intenso al pasar el cursor
                            shadowOffsetY: 0,
                            shadowOffsetX: 0,
                            borderColor: CHART_BRIGHT_COLOR,
                            borderWidth: 1
                        }
                    },
                    data: processedData,
                    animationDuration: 300,
                    animationEasing: 'cubicOut'
                }
            ],
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
    }, [localData, isInitialized])

    // Limpiar la instancia del gráfico al desmontar el componente
    useEffect(() => {
        return () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.dispose()
                chartInstanceRef.current = null
            }
        }
    }, [])

    return (
        <div className="h-full pb-12">
            <div className="flex items-center mb-2 ml-8">
                <div className="flex items-center gap-2">
                    <span className="text-xl font-bold text-white text-shadow-glow-white">{currentQPS}</span>
                    <span className="text-white text-xs text-shadow-glow-white">QPS</span>
                </div>
            </div>
            <div
                ref={chartRef}
                className='w-full h-full'
            />
        </div>
    )
} 