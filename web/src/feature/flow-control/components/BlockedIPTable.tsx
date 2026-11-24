import { useState, useMemo, useEffect, useCallback, memo } from 'react'
import {
    useReactTable,
    getCoreRowModel,
    getPaginationRowModel,
    ColumnDef,
    PaginationState,
    Updater,
} from '@tanstack/react-table'
import { BlockedIPRecord, BlockedIPListRequest } from '@/types/blocked-ip'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { RefreshCcw, Trash2, Filter, X, ChevronUp, ChevronDown } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { DataTable } from '@/components/table/motion-data-table'
import { DataTablePagination } from '@/components/table/pagination'
import { useBlockedIPsQuery, useCleanupExpiredBlockedIPs } from '../hooks/useBlockedIP'
import { useTranslation } from 'react-i18next'
import { AnimatedIcon } from '@/components/ui/animation/components/animated-icon'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
import { Badge } from '@/components/ui/badge'
import { produce } from 'immer'
import React from 'react'
import { TFunction } from 'i18next'

// Hook de cuenta atrás independiente - gestiona su propio estado dentro del componente
const useCountdown = (initialSeconds: number, isActive: boolean) => {
    const [seconds, setSeconds] = useState(initialSeconds)

    useEffect(() => {
        if (!isActive || seconds <= 0) return

        const timer = setInterval(() => {
            setSeconds(prev => {
                const newValue = prev - 1
                if (newValue <= 0) {
                    clearInterval(timer)
                    return 0
                }
                return newValue
            })
        }, 1000)

        return () => clearInterval(timer)
    }, [isActive, seconds])

    // Reiniciar cuando cambia el valor inicial (p. ej. al refrescar datos)
    useEffect(() => {
        setSeconds(initialSeconds)
    }, [initialSeconds])

    return seconds
}

// Componente de cuenta atrás independiente - solo este componente se actualiza cada segundo
const CountdownDisplay = React.memo(({
    remainingTTL,
    lastDataUpdateTime,
    t
}: {
    remainingTTL: number
    lastDataUpdateTime: number
    t: TFunction
}) => {
    const timeSinceDataUpdate = Math.floor((Date.now() - lastDataUpdateTime) / 1000)
    const initialRemaining = Math.max(0, remainingTTL - timeSinceDataUpdate)
    const isActive = initialRemaining > 0

    const currentRemaining = useCountdown(initialRemaining, isActive)

    const formatRemainingTime = useCallback((seconds: number) => {
        if (seconds <= 0) return t('flowControl.expired', 'Expirado')

        const hours = Math.floor(seconds / 3600)
        const minutes = Math.floor((seconds % 3600) / 60)
        const remainingSeconds = seconds % 60

        if (hours > 0) {
            return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')} ${t('flowControl.untilUnblock', 'hasta el desbloqueo')}`
        } else if (minutes > 0) {
            return `${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')} ${t('flowControl.untilUnblock', 'hasta el desbloqueo')}`
        } else {
            return `00:${remainingSeconds.toString().padStart(2, '0')} ${t('flowControl.untilUnblock', 'hasta el desbloqueo')}`
        }
    }, [t])

    if (!isActive) return null

    return (
        <div className="text-xs text-muted-foreground mt-1 dark:text-shadow-glow-white font-mono w-20 text-center">
            {formatRemainingTime(currentRemaining)}
        </div>
    )
})

CountdownDisplay.displayName = 'CountdownDisplay'

// Celda de estado optimizada - ya no depende del temporizador global
const StatusCell = memo(({
    remainingTTL,
    lastDataUpdateTime,
    t
}: {
    remainingTTL: number
    lastDataUpdateTime: number
    t: TFunction
}) => {
    // Cálculo inicial de si está activo
    const timeSinceDataUpdate = Math.floor((Date.now() - lastDataUpdateTime) / 1000)
    const isInitiallyActive = remainingTTL - timeSinceDataUpdate > 0

    return (
        <div className="flex flex-col">
            <Badge variant={isInitiallyActive ? "destructive" : "secondary"}>
                {isInitiallyActive
                    ? t('flowControl.status.active', 'Activo')
                    : t('flowControl.status.expired', 'Expirado')
                }
            </Badge>
            {isInitiallyActive && (
                <CountdownDisplay
                    remainingTTL={remainingTTL}
                    lastDataUpdateTime={lastDataUpdateTime}
                    t={t}
                />
            )}
        </div>
    )
})

StatusCell.displayName = 'StatusCell'

// Componente principal - se eliminó el temporizador global
export function BlockedIPTable() {
    const { t } = useTranslation()

    // Estado de búsqueda y filtrado
    const [filters, setFilters] = useState<BlockedIPListRequest>({
        page: 1,
        size: 20,
        status: 'all',
        sortBy: 'blocked_at',
        sortDir: 'desc'
    })

    // Estado temporal de la entrada de IP
    const [ipInputValue, setIpInputValue] = useState('')
    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)
    const [isFilterAnimating, setIsFilterAnimating] = useState(false)

    // Registrar la última actualización de datos
    const [lastDataUpdateTime, setLastDataUpdateTime] = useState(Date.now())

    // Obtener datos
    const { blockedIPs, isLoading, refetch } = useBlockedIPsQuery(filters)
    const { cleanupExpiredBlockedIPs, isLoading: isCleanupLoading } = useCleanupExpiredBlockedIPs()

    // Al actualizarse los datos, restablecer la hora de actualización
    useEffect(() => {
        if (blockedIPs?.items) {
            setLastDataUpdateTime(Date.now())
        }
    }, [blockedIPs?.items])

    // Cachear los datos de la tabla
    const tableData = useMemo(() => {
        return blockedIPs?.items || []
    }, [blockedIPs?.items])

    // Disparar la animación del filtro
    const triggerFilterAnimation = useCallback(() => {
        setIsFilterAnimating(true)
        setTimeout(() => {
            setIsFilterAnimating(false)
        }, 1000)
    }, [])

    // Actualizar datos
    const handleRefresh = useCallback(() => {
        setIsRefreshAnimating(true)
        refetch()
        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }, [refetch])

    // Actualizar el filtro
    const updateFilters = useCallback((newFilters: Partial<BlockedIPListRequest>) => {
        setFilters(produce(draft => {
            Object.assign(draft, newFilters)
            // Si no es un cambio de paginación, volver a la primera página
            if (!('page' in newFilters)) {
                draft.page = 1
            }
        }))
        // Si se actualizó el filtro de IP, sincronizar el valor del cuadro de texto
        if ('ip' in newFilters) {
            setIpInputValue(newFilters.ip || '')
        }
        // Disparar la animación del filtro
        triggerFilterAnimation()
    }, [triggerFilterAnimation])

    // Limpiar el filtro
    const clearFilters = useCallback(() => {
        setFilters({
            page: 1,
            size: 20,
            status: 'all',
            sortBy: 'blocked_at',
            sortDir: 'desc'
        })
        setIpInputValue('')
        // Disparar la animación del filtro
        triggerFilterAnimation()
    }, [triggerFilterAnimation])

    // Formatear la hora
    const formatTime = useCallback((timeStr: string) => {
        return new Date(timeStr).toLocaleString()
    }, [])

    // Obtener la descripción del motivo
    const getReasonDescription = useCallback((reason: string) => {
        switch (reason) {
            case 'high_frequency_visit':
                return t('flowControl.reason.highFrequencyVisit', 'Acceso de alta frecuencia')
            case 'high_frequency_attack':
                return t('flowControl.reason.highFrequencyAttack', 'Ataque de alta frecuencia')
            case 'high_frequency_error':
                return t('flowControl.reason.highFrequencyError', 'Errores de alta frecuencia')
            default:
                return reason
        }
    }, [t])

    // Manejar el clic de ordenación
    const handleSortClick = useCallback(() => {
        const newSortDir = filters.sortDir === 'desc' ? 'asc' : 'desc'
        updateFilters({ sortBy: 'blocked_at', sortDir: newSortDir })
    }, [filters.sortDir, updateFilters])

    // Definición de columnas de tabla - usar useCallback para optimizar las funciones de renderizado
    const columns: ColumnDef<BlockedIPRecord>[] = useMemo(() => [
        {
            accessorKey: 'ip',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('flowControl.table.ip', 'Dirección IP')}</div>,
            cell: ({ row }) => (
                <div className="font-mono dark:text-shadow-glow-white">{row.original.ip}</div>
            ),
        },
        {
            accessorKey: 'reason',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('flowControl.table.reason', 'Motivo del bloqueo')}</div>,
            cell: ({ row }) => (
                <Badge variant="secondary" className="dark:text-shadow-glow-white">
                    {getReasonDescription(row.original.reason)}
                </Badge>
            ),
        },
        {
            accessorKey: 'requestUri',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('flowControl.table.requestUri', 'Ruta de la petición')}</div>,
            cell: ({ row }) => (
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger className="font-mono text-sm dark:text-shadow-glow-white max-w-xs truncate cursor-pointer text-left block">
                            {row.original.requestUri}
                        </TooltipTrigger>
                        <TooltipContent>
                            <p>{row.original.requestUri}</p>
                        </TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            ),
        },
        {
            accessorKey: 'blockedAt',
            header: () => (
                <div
                    className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700/50 px-2 -mx-2 rounded transition-colors flex items-center gap-1"
                    onClick={handleSortClick}
                    title="Pulsar para ordenar"
                >
                    {t('flowControl.table.blockedAt', 'Hora del bloqueo')}
                    {filters.sortBy === 'blocked_at' && (
                        filters.sortDir === 'desc' ?
                            <ChevronDown className="h-4 w-4" /> :
                            <ChevronUp className="h-4 w-4" />
                    )}
                </div>
            ),
            cell: ({ row }) => (
                <div className="text-sm dark:text-shadow-glow-white">
                    {formatTime(row.original.blockedAt)}
                </div>
            ),
        },
        {
            accessorKey: 'status',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t('flowControl.table.status', 'Estado')}</div>,
            cell: ({ row }) => (
                <StatusCell
                    remainingTTL={row.original.remainingTTL}
                    lastDataUpdateTime={lastDataUpdateTime}
                    t={t}
                />
            ),
        },
    ], [t, filters.sortBy, filters.sortDir, lastDataUpdateTime, formatTime, getReasonDescription, handleSortClick])

    // Manejar cambios de paginación
    const handlePaginationChange = useCallback((updater: Updater<PaginationState>) => {
        if (typeof updater === 'function') {
            const oldPagination = {
                pageIndex: (filters.page || 1) - 1,
                pageSize: filters.size || 20
            }
            const newPagination = updater(oldPagination)

            setFilters(produce(draft => {
                if (newPagination.pageIndex !== oldPagination.pageIndex) {
                    draft.page = newPagination.pageIndex + 1
                }

                if (newPagination.pageSize !== oldPagination.pageSize) {
                    draft.size = newPagination.pageSize
                    draft.page = 1
                }
            }))
        }
    }, [filters.page, filters.size])

    // Inicializar la tabla
    const table = useReactTable({
        data: tableData,
        columns,
        pageCount: blockedIPs?.pages || 0,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        manualPagination: true,
        state: {
            pagination: {
                pageIndex: (filters.page || 1) - 1,
                pageSize: filters.size || 20
            }
        },
        onPaginationChange: handlePaginationChange
    })

    // Manejar el evento Enter en la entrada de IP
    const handleIpInputKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter') {
            updateFilters({ ip: e.currentTarget.value || undefined })
        }
    }, [updateFilters])

    return (
        <Card className="flex flex-col h-full p-0 border-none shadow-none">
            {/* Filtros de cabecera y botones de acción - altura fija */}
            <div className="p-6 flex-shrink-0">
                {/* Título y botones de acción */}
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-semibold text-primary dark:text-white">
                        {t('flowControl.blockIpList', 'Lista de IPs interceptadas')}
                    </h2>
                    <div className="flex gap-2">
                        <AnimatedButton>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                className="flex items-center gap-2 dark:text-shadow-glow-white"
                            >
                                <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                    <RefreshCcw className="h-4 w-4" />
                                </AnimatedIcon>
                                {t('flowControl.refresh', 'Actualizar')}
                            </Button>
                        </AnimatedButton>
                        <AnimatedButton>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => cleanupExpiredBlockedIPs()}
                                disabled={isCleanupLoading}
                                className="flex items-center gap-2 dark:text-shadow-glow-white"
                            >
                                <Trash2 className="h-4 w-4" />
                                {isCleanupLoading
                                    ? t('flowControl.cleaning', 'Limpiando...')
                                    : t('flowControl.cleanup', 'Limpiar registros expirados')
                                }
                            </Button>
                        </AnimatedButton>
                    </div>
                </div>

                {/* Filtro */}
                <div className="flex flex-wrap gap-3 p-4 bg-gray-50 dark:bg-gray-800/20 rounded-lg">
                    <div className="flex items-center gap-2 mr-2">
                        <AnimatedIcon animationVariant="continuous-pulse" isAnimating={isFilterAnimating} className="h-4 w-4">
                            <Filter className="h-4 w-4 text-muted-foreground" />
                        </AnimatedIcon>
                        <span className="text-sm font-medium dark:text-shadow-glow-white">
                            {t('flowControl.filters', 'Condiciones de filtrado')}
                        </span>
                    </div>

                    <div className="flex items-center gap-3 flex-wrap">
                        <Input
                            placeholder={t('flowControl.filterByIP', 'Filtrar por dirección IP')}
                            value={ipInputValue}
                            onChange={(e) => setIpInputValue(e.target.value)}
                            onKeyDown={handleIpInputKeyDown}
                            className="w-48 dark:text-shadow-glow-white"
                        />

                        <div className="flex items-center gap-2">
                            <span className="text-sm font-medium text-muted-foreground whitespace-nowrap dark:text-shadow-glow-white">
                                {t('flowControl.table.reason', 'Motivo del bloqueo')}:
                            </span>
                            <Select
                                value={filters.reason || 'all'}
                                onValueChange={(value: string) => updateFilters({ reason: value === 'all' ? undefined : value })}
                            >
                                <SelectTrigger className="w-auto max-w-[200px] dark:text-shadow-glow-white gap-2">
                                    <SelectValue placeholder={t('flowControl.filterByReason', 'Filter by Reason')} />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">{t('flowControl.reason.all', 'All Reasons')}</SelectItem>
                                    <SelectItem value="high_frequency_visit">{t('flowControl.reason.highFrequencyVisit', 'High Frequency Visit')}</SelectItem>
                                    <SelectItem value="high_frequency_attack">{t('flowControl.reason.highFrequencyAttack', 'High Frequency Attack')}</SelectItem>
                                    <SelectItem value="high_frequency_error">{t('flowControl.reason.highFrequencyError', 'High Frequency Error')}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="flex items-center gap-2">
                            <span className="text-sm font-medium text-muted-foreground whitespace-nowrap dark:text-shadow-glow-white">
                                {t('flowControl.table.status', 'Estado')}:
                            </span>
                            <Select
                                value={filters.status}
                                onValueChange={(value: 'all' | 'active' | 'expired') => updateFilters({ status: value })}
                            >
                                <SelectTrigger className="w-auto max-w-[200px] dark:text-shadow-glow-white gap-2">
                                    <SelectValue placeholder={t('flowControl.status.placeholder', 'Status')} />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">{t('flowControl.status.all', 'All')}</SelectItem>
                                    <SelectItem value="active">{t('flowControl.status.active', 'Active')}</SelectItem>
                                    <SelectItem value="expired">{t('flowControl.status.expired', 'Expired')}</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={clearFilters}
                            className="flex items-center gap-2 dark:text-shadow-glow-white ml-2"
                        >
                            <X className="h-4 w-4" />
                            {t('flowControl.clearFilters', 'Borrar')}
                        </Button>
                    </div>
                </div>
            </div>

            {/* Área de tabla - altura flexible, con scroll */}
            <div className="px-6 flex-1 overflow-auto">
                <DataTable
                    loadingStyle="skeleton"
                    table={table}
                    columns={columns}
                    isLoading={isLoading}
                    fixedHeader={true}
                    animatedRows={true}
                    showScrollShadows={true}
                />
            </div>

            {/* Paginación inferior - altura fija */}
            <div className="py-6 px-4 flex-shrink-0">
                <DataTablePagination table={table} />
            </div>
        </Card>
    )
}