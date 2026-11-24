import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import { useResizeObserver } from '@/feature/stats/hooks/use-resize-observer'
import { useTheme } from "@/provider/theme-context"

interface EChartWrapperProps {
    options: echarts.EChartsOption
    loading?: boolean
    height?: number | string
    className?: string
}

export function EChartWrapper({
    options,
    loading = false,
    height = 300,
    className = '',
}: EChartWrapperProps) {
    const chartRef = useRef<HTMLDivElement>(null)
    const chartInstanceRef = useRef<echarts.ECharts | null>(null)
    const { theme } = useTheme()

    // Escuchar cambios de tamaño del contenedor
    const { width } = useResizeObserver(chartRef)

    // Inicializar gráfico y manejar cambios de tema - combinado en un useEffect
    useEffect(() => {
        if (!chartRef.current) return

        const isDarkMode = theme === 'dark'

        // Si la instancia no existe, crearla; si existe, no recrearla, solo aplicar el tema
        if (!chartInstanceRef.current) {
            chartInstanceRef.current = echarts.init(chartRef.current, isDarkMode ? 'dark' : undefined)
        } else {
            // Si el tema cambia, aplicar el nuevo tema
            chartInstanceRef.current.dispose()
            chartInstanceRef.current = echarts.init(chartRef.current, isDarkMode ? 'dark' : undefined)
        }

        // Establecer el estado de carga
        if (loading) {
            chartInstanceRef.current.showLoading({
                text: '',
                color: isDarkMode ? '#ffffff' : '#1f2937',
                textColor: isDarkMode ? '#ffffff' : '#1f2937',
                maskColor: isDarkMode ? 'rgba(0, 0, 0, 0.1)' : 'rgba(255, 255, 255, 0.8)',
            })
        } else {
            chartInstanceRef.current.hideLoading()
        }

        // Actualizar el gráfico
        chartInstanceRef.current.setOption(options)

        // Función de limpieza
        return () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.dispose()
                chartInstanceRef.current = null
            }
        }
    }, [options, theme, loading])

    // Responder a cambios de tamaño del contenedor
    useEffect(() => {
        if (chartInstanceRef.current && width) {
            chartInstanceRef.current.resize()
        }
    }, [width])

    return (
        <div
            ref={chartRef}
            style={{ height: typeof height === 'number' ? `${height}px` : height }}
            className={`w-full ${className}`}
        />
    )
}