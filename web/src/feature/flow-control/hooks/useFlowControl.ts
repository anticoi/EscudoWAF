import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { configApi } from '@/api/config'
import { ConfigPatchRequest } from '@/types/config'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'
import { useConfigQuery } from '@/feature/global-setting/hooks/useConfig'

// Hook de mutación de actualización de configuración de control de flujo
export const useUpdateFlowControlConfig = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const { t } = useTranslation()
    const [error, setError] = useState<string | null>(null)

    const mutation = useMutation({
        mutationFn: (data: ConfigPatchRequest) => configApi.updateConfig(data),
        onSuccess: () => {
            toast({
                title: t('flowControl.toast.updateSuccess', 'Actualizado con éxito'),
                description: t('flowControl.toast.configUpdated', 'Configuración de control de flujo actualizada con éxito'),
            })
            queryClient.invalidateQueries({ queryKey: ['config'] })
        },
        onError: (error: ApiError) => {
            console.error('Error al actualizar la configuración de control de flujo:', error)
            setError(error.message || t('flowControl.toast.updateError', 'Se produjo un error al actualizar la configuración'))
            toast({
                title: t('flowControl.toast.updateFailed', 'Error de actualización'),
                description: error.message || t('flowControl.toast.updateError', 'Se produjo un error al actualizar la configuración'),
                variant: "destructive",
            })
        }
    })

    return {
        updateFlowControlConfig: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

// Obtener la configuración de control de flujo
export const useFlowControlConfig = () => {
    const { config, isLoading, error, refetch } = useConfigQuery()
    
    return {
        flowControlConfig: config?.engine?.flowController,
        isLoading,
        error,
        refetch
    }
} 