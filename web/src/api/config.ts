import { get, patch } from './index'
import { ConfigResponse, ConfigPatchRequest } from '@/types/config'

const BASE_URL = '/config'
/**
 * Servicios de API relacionados con configuración
 */
export const configApi = {
    /**
     * Obtener la configuración del sistema
     * @returns información de configuración del sistema
     */
    getConfig: (): Promise<ConfigResponse> => {
        return get<ConfigResponse>(BASE_URL)
    },

    /**
     * Actualizar la configuración del sistema
     * @param config información de configuración a actualizar
     * @returns configuración del sistema actualizada
     */
    updateConfig: (config: ConfigPatchRequest): Promise<ConfigResponse> => {
        return patch<ConfigResponse>(BASE_URL, config)
    }
}

