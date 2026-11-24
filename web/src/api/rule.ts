import { get, post, put, del } from './index'
import {
    MicroRule,
    MicroRuleListResponse,
    MicroRuleCreateRequest,
    MicroRuleUpdateRequest,
} from '@/types/rule'

// Ruta base de la API de reglas
const BASE_URL = '/micro-rules'

/**
 * Servicios de API relacionados con reglas
 */
export const ruleApi = {
    /**
     * Obtener la lista de reglas
     * @param page número de página
     * @param size cantidad por página
     * @returns datos de respuesta de lista de reglas
     */
    getMicroRules: (page: number = 1, size: number = 10): Promise<MicroRuleListResponse> => {
        return get<MicroRuleListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * Crear nueva regla
     * @param rule datos de la petición de creación de regla
     * @returns detalles de la regla creada
     */
    createMicroRule: (rule: MicroRuleCreateRequest): Promise<MicroRule> => {
        return post<MicroRule>(BASE_URL, rule)
    },

    /**
     * Obtener el detalle de una regla
     * @param id ID de la regla
     * @returns detalles de la regla
     */
    getMicroRule: (id: string): Promise<MicroRule> => {
        return get<MicroRule>(`${BASE_URL}/${id}`)
    },

    /**
     * Actualizar regla
     * @param id ID de la regla
     * @param rule datos de la petición de actualización de regla
     * @returns detalles de la regla actualizada
     */
    updateMicroRule: (id: string, rule: MicroRuleUpdateRequest): Promise<MicroRule> => {
        return put<MicroRule>(`${BASE_URL}/${id}`, rule)
    },

    /**
     * Eliminar regla
     * @param id ID de la regla
     * @returns void
     */
    deleteMicroRule: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    }
}