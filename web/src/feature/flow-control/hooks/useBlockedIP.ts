import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { blockedIPApi } from '@/api/blocked-ip'
import { BlockedIPListRequest } from '@/types/blocked-ip'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

// Hook de consulta de lista de IPs bloqueadas
export const useBlockedIPsQuery = (params: BlockedIPListRequest) => {
    const query = useQuery({
        queryKey: ['blocked-ips', params],
        queryFn: () => blockedIPApi.getBlockedIPs(params)
    })

    return {
        blockedIPs: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// Hook de consulta de estadísticas de IPs bloqueadas
export const useBlockedIPStatsQuery = () => {
    const query = useQuery({
        queryKey: ['blocked-ip-stats'],
        queryFn: blockedIPApi.getBlockedIPStats
    })

    return {
        stats: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// Hook de mutación de limpieza de IPs bloqueadas expiradas
export const useCleanupExpiredBlockedIPs = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: blockedIPApi.cleanupExpiredBlockedIPs,
        onSuccess: (data) => {
            toast({
                title: t('flowControl.toast.cleanupSuccess', 'Limpieza exitosa'),
                description: t('flowControl.toast.cleanupMessage', `Se limpiaron ${data.deletedCount} registros expirados`),
            })
            queryClient.invalidateQueries({ queryKey: ['blocked-ips'] })
            queryClient.invalidateQueries({ queryKey: ['blocked-ip-stats'] })
        },
        onError: (error: ApiError) => {
            console.error('Error al limpiar IPs bloqueadas expiradas:', error)
            setError(error.message || t('flowControl.toast.cleanupError', 'Se produjo un error al limpiar los registros expirados'))
            toast({
                title: t('flowControl.toast.cleanupFailed', 'Error de limpieza'),
                description: error.message || t('flowControl.toast.cleanupError', 'Se produjo un error al limpiar los registros expirados'),
                variant: "destructive",
            })
        }
    })

    return {
        cleanupExpiredBlockedIPs: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
} 