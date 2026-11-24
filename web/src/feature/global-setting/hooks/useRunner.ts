// src/feature/global-setting/hooks/useRunner.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { runnerApi } from '@/api/runner'
import { RunnerAction } from '@/types/runner'
import { useToast } from '@/hooks/use-toast'
import { ApiError } from '@/api/index'
import { getConstant } from '@/constant'
import { ConstantCategory } from '@/constant'
import { useTranslation } from 'react-i18next'

// Hook de consulta del estado del runner
export const useRunnerStatusQuery = () => {
    const query = useQuery({
        queryKey: ['runner-status'],
        queryFn: runnerApi.getStatus,
        // refetchInterval: 5000 // recarga automática cada 5 segundos
    })

    return {
        status: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// Hook de mutación de control del runner
export const useRunnerControl = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (action: RunnerAction) => runnerApi.control({ action }),
        onSuccess: (data) => {
            toast({
                title: t("globalSetting.engine.operationSuccess", "Operación exitosa"),
                description: data.message || t("globalSetting.engine.operationCompleted", "La operación de control del runner se ejecutó con éxito"),
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
            queryClient.invalidateQueries({ queryKey: ['runner-status'] })
        },
        onError: (error: ApiError) => {
            console.error('Error al controlar el runner:', error)
            setError(error.message || "Se produjo un error al controlar el runner")
            toast({
                title: t("globalSetting.engine.operationFailed", "Operación fallida"),
                description: error.message || t("globalSetting.engine.operationError", "Se produjo un error al controlar el runner"),
                variant: "destructive",
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
        }
    })

    return {
        controlRunner: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}