import { get } from './index'
import { 
  AttackEventResponse,
  AttackLogResponse,
  AttackEventQuery,
  AttackLogQuery
} from '@/types/log'

// Ruta base del WAF
const BASE_URL = '/log'

/**
 * Servicios de API relacionados con el WAF
 */
export const logApi = {
  /**
   * Obtener eventos de ataque (agregados por IP de cliente y dominio)
   * @param query parámetros de consulta
   * @returns datos de respuesta de eventos de ataque
   */
  getAttackEvents: (query: AttackEventQuery): Promise<AttackEventResponse> => {
    return get<AttackEventResponse>(`${BASE_URL}/event`, { params: query })
  },

  /**
   * Obtener el detalle del log de ataque
   * @param query parámetros de consulta
   * @returns datos de respuesta de logs de ataque
   */
  getAttackLogs: (query: AttackLogQuery): Promise<AttackLogResponse> => {
    return get<AttackLogResponse>(`${BASE_URL}`, { params: query })
  }
}
