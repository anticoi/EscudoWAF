// src/feature/global-setting/hooks/useConfig.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { configApi } from '@/api/config'
import { ConfigPatchRequest } from '@/types/config'
import { useToast } from '@/hooks/use-toast'
import { ApiError } from '@/api/index'
import { ConstantCategory, getConstant } from '@/constant'
import { useTranslation } from 'react-i18next'

// Hook de consulta de configuración
export const useConfigQuery = () => {
    const query = useQuery({
        queryKey: ['config'],
        queryFn: configApi.getConfig
    })

    return {
        config: query.data,
        isLoading: query.isPending,
        error: query.error,
        refetch: query.refetch
    }
}

// Hook de mutación de actualización de configuración
export const useUpdateConfig = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<ApiError | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (data: ConfigPatchRequest) => configApi.updateConfig(data),
        onSuccess: () => {
            toast({
                title: t("globalSetting.config.updateSuccess", "Actualizado con éxito"),
                description: t("globalSetting.config.systemConfigUpdated", "Configuración del sistema actualizada con éxito"),
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
            queryClient.invalidateQueries({ queryKey: ['config'] })
        },
        onError: (error: ApiError) => {
            console.error('Error al actualizar la configuración:', error)
            setError(error)
            toast({
                title: t("globalSetting.config.updateFailed", "Error de actualización"),
                description: error.message || t("globalSetting.config.updateError", "Se produjo un error al actualizar la configuración"),
                variant: "destructive",
                duration: getConstant(ConstantCategory.FEATURE, 'TOAST_DURATION', 3000),
            })
        }
    })

    return {
        updateConfig: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

// Constante para obtener el nombre del motor
export const getEngineNameConstant = (): string => {
    return getConstant(ConstantCategory.CONFIG, 'ENGINE_NAME', 'coraza')
}

// Actualizar las directivas del motor especificado en la configuración
export const updateEngineDirectives = (config: ConfigPatchRequest, directives: string): ConfigPatchRequest => {
    const engineName = getEngineNameConstant()

    // Si no hay configuración de motor ni appConfig, devolver la configuración original
    if (!config.engine || !config.engine.appConfig) {
        return config
    }

    // Buscar los elementos de configuración de un motor específico
    const engineAppConfig = config.engine.appConfig.find(app => app.name === engineName)

    // Solo actualizar si se encuentra, si no, no hacer nada
    if (engineAppConfig) {
        engineAppConfig.directives = directives
    }

    return config
}