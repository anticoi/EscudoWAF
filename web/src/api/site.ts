import { get, post, put, del } from './index'
import {
    Site,
    SiteListResponse,
    CreateSiteRequest,
    UpdateSiteRequest
} from '@/types/site'

// Ruta base de la API de sitios
const BASE_URL = '/site'

/**
 * Servicios de API relacionados con sitios
 */
export const siteApi = {
    /**
     * Obtener lista de sitios
     * @param page número de página
     * @param size cantidad por página
     * @returns datos de respuesta de lista de sitios
     */
    getSites: (page: number = 1, size: number = 10): Promise<SiteListResponse> => {
        return get<SiteListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * Crear nuevo sitio
     * @param site datos de la petición de creación de sitio
     * @returns detalles del sitio creado
     */
    createSite: (site: CreateSiteRequest): Promise<Site> => {
        return post<Site>(BASE_URL, site)
    },

    /**
     * Obtener detalles de un sitio
     * @param id ID del sitio
     * @returns detalles del sitio
     */
    getSite: (id: string): Promise<Site> => {
        return get<Site>(`${BASE_URL}/${id}`)
    },

    /**
     * Actualizar sitio
     * @param id ID del sitio
     * @param site datos de la petición de actualización de sitio
     * @returns detalles del sitio actualizado
     */
    updateSite: (id: string, site: UpdateSiteRequest): Promise<Site> => {
        return put<Site>(`${BASE_URL}/${id}`, site)
    },

    /**
     * Eliminar sitio
     * @param id ID del sitio
     * @returns void
     */
    deleteSite: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    }
}