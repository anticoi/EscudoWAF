import { get, post, put, del } from './index'
import {
    IPGroup,
    IPGroupListResponse,
    IPGroupCreateRequest,
    IPGroupUpdateRequest,
} from '@/types/ip-group'

// Ruta base de la API de grupos de IPs
const BASE_URL = '/ip-groups'

/**
 * Servicios de API relacionados con grupos de IPs
 */
export const ipGroupApi = {
    /**
     * Obtener lista de grupos de IPs
     * @param page número de página
     * @param size cantidad por página
     * @returns datos de respuesta de lista de grupos de IPs
     */
    getIPGroups: (page: number = 1, size: number = 10): Promise<IPGroupListResponse> => {
        return get<IPGroupListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * Crear nuevo grupo de IPs
     * @param ipGroup datos de la petición de creación de grupo de IPs
     * @returns detalles del grupo de IPs creado
     */
    createIPGroup: (ipGroup: IPGroupCreateRequest): Promise<IPGroup> => {
        return post<IPGroup>(BASE_URL, ipGroup)
    },

    /**
     * Obtener detalles de un grupo de IPs
     * @param id ID del grupo de IPs
     * @returns detalles del grupo de IPs
     */
    getIPGroup: (id: string): Promise<IPGroup> => {
        return get<IPGroup>(`${BASE_URL}/${id}`)
    },

    /**
     * Actualizar grupo de IPs
     * @param id ID del grupo de IPs
     * @param ipGroup datos de la petición de actualización de grupo de IPs
     * @returns detalles del grupo de IPs actualizado
     */
    updateIPGroup: (id: string, ipGroup: IPGroupUpdateRequest): Promise<IPGroup> => {
        return put<IPGroup>(`${BASE_URL}/${id}`, ipGroup)
    },

    /**
     * Eliminar grupo de IPs
     * @param id ID del grupo de IPs
     * @returns void
     */
    deleteIPGroup: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    },

    /**
     * Añadir la IP a la lista negra del sistema
     * @param ip dirección IP o CIDR
     * @returns void
     */
    blockIP: (ip: string): Promise<void> => {
        return post<void>(`${BASE_URL}/blacklist/add`, { ip })
    }
}