import { get, del } from './index'
import {
    BlockedIPListRequest,
    BlockedIPListResponse,
    BlockedIPStatsResponse,
    BlockedIPCleanupResponse,
} from '@/types/blocked-ip'

// Ruta base de la API de IPs bloqueadas
const BASE_URL = '/blocked-ips'

/**
 * Servicios de API relacionados con IPs bloqueadas
 */
export const blockedIPApi = {
    /**
     * Obtener lista de IPs bloqueadas
     * @param params parámetros de consulta
     * @returns datos de respuesta de lista de IPs bloqueadas
     */
    getBlockedIPs: (params: BlockedIPListRequest = {}): Promise<BlockedIPListResponse> => {
        // Mapear parámetros del frontend a los nombres que espera el backend, asegurando valores por defecto
        const queryParams: Record<string, string | number> = {
            page: params.page || 1,
            size: params.size || 20,
            status: params.status || 'all',
            sortBy: params.sortBy || 'blocked_at',
            sortDir: params.sortDir || 'desc'
        }
        
        // Solo añadir parámetros opcionales cuando tienen valor
        if (params.ip) {
            queryParams.ip = params.ip
        }
        if (params.reason) {
            queryParams.reason = params.reason
        }

        return get<BlockedIPListResponse>(BASE_URL, {
            params: queryParams
        })
    },

    /**
     * Obtener estadísticas de IPs bloqueadas
     * @returns Información estadística
     */
    getBlockedIPStats: (): Promise<BlockedIPStatsResponse> => {
        return get<BlockedIPStatsResponse>(`${BASE_URL}/stats`)
    },

    /**
     * Limpiar registros de IPs bloqueadas expirados
     * @returns resultado de la limpieza
     */
    cleanupExpiredBlockedIPs: (): Promise<BlockedIPCleanupResponse> => {
        return del<BlockedIPCleanupResponse>(`${BASE_URL}/cleanup`)
    }
} 