import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { siteApi } from '@/api/site'
import { CreateSiteRequest, UpdateSiteRequest } from '@/types/site'
import { useToast } from '@/hooks/use-toast'
import { useTranslation } from 'react-i18next'

type ApiError = {
    message: string
}

/**
 * Hook de creación de sitio
 */
export const useCreateSite = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (data: CreateSiteRequest) => siteApi.createSite(data),
        onSuccess: () => {
            toast({
                title: t("site.toast.createSuccess", "Creado con éxito"),
                description: t("site.toast.siteCreated", "El sitio se creó con éxito"),
            })
            queryClient.invalidateQueries({ queryKey: ['sites'] })
        },
        onError: (error: ApiError) => {
            console.error('Error al crear el sitio:', error)
            setError(error.message || "Se produjo un error al crear el sitio")
            toast({
                title: t("site.toast.createFailed", "Error de creación"),
                description: error.message || t("site.toast.createError", "Se produjo un error al crear el sitio"),
                variant: "destructive",
            })
        }
    })

    return {
        createSite: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

/**
 * Eliminar sitioHook
 */
export const useDeleteSite = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: (id: string) => siteApi.deleteSite(id),
        onSuccess: () => {
            toast({
                title: t("site.toast.deleteSuccess", "Eliminado con éxito"),
                description: t("site.toast.siteDeleted", "El sitio se eliminó con éxito"),
            })
            queryClient.invalidateQueries({ queryKey: ['sites'] })
        },
        onError: (error: ApiError) => {
            console.error('Error al eliminar el sitio:', error)
            setError(error.message || "Se produjo un error al eliminar el sitio")
            toast({
                title: t("site.toast.deleteFailed", "Error de eliminación"),
                description: error.message || t("site.toast.deleteError", "Se produjo un error al eliminar el sitio"),
                variant: "destructive",
            })
        }
    })

    return {
        deleteSite: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}

/**
 * Actualizar sitioHook
 */
export const useUpdateSite = () => {
    const queryClient = useQueryClient()
    const { toast } = useToast()
    const [error, setError] = useState<string | null>(null)
    const { t } = useTranslation()

    const mutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: UpdateSiteRequest }) =>
            siteApi.updateSite(id, data),
        onSuccess: () => {
            toast({
                title: t("site.toast.updateSuccess", "Actualizado con éxito"),
                description: t("site.toast.siteUpdated", "El sitio se actualizó con éxito"),
            })
            queryClient.invalidateQueries({ queryKey: ['sites'] })
        },
        onError: (error: ApiError) => {
            console.error('Error al actualizar el sitio:', error)
            setError(error.message || "Se produjo un error al actualizar el sitio")
            toast({
                title: t("site.toast.updateFailed", "Error de actualización"),
                description: error.message || t("site.toast.updateError", "Se produjo un error al actualizar el sitio"),
                variant: "destructive",
            })
        }
    })

    return {
        updateSite: mutation.mutate,
        isLoading: mutation.isPending,
        error,
        clearError: () => setError(null),
    }
}