// src/api/runner.ts
import { get, post } from './index'
import { RunnerStatusResponse, RunnerControlRequest, RunnerControlResponse } from '@/types/runner'

const BASE_URL = '/runner'
/**
 * Servicios de API relacionados con el runner
 */
export const runnerApi = {
    /**
     * Obtener el estado del runner
     * @returns información del estado del runner
     */
    getStatus: (): Promise<RunnerStatusResponse> => {
        return get<RunnerStatusResponse>(`${BASE_URL}/status`)
    },

    /**
     * Controlar el runner (iniciar, detener, reiniciar, etc.)
     * @param request petición de control
     * @returns resultado de la operación de control
     */
    control: (request: RunnerControlRequest): Promise<RunnerControlResponse> => {
        return post<RunnerControlResponse>(`${BASE_URL}/control`, request)
    }
}