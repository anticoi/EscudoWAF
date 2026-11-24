import axios, { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios'
import { useAuthStore } from '@/store/auth'
import { ENV } from '@/utils/env'

// ======================
// Definiciones de tipos de respuesta de la API
// ======================

/**
 * Interfaz de respuesta de API genérica
 */
export interface APIResponse<T = unknown> {
    code: number
    message: string
    success: boolean
    requestId: string
    timestamp: string
    data?: T
    error?: string
}

/**
 * Clase de error de API personalizada
 */
export class ApiError extends Error {
    code: number
    requestId?: string
    errorDetail?: string

    constructor(message: string, code: number, requestId?: string, errorDetail?: string) {
        super(message)
        this.name = 'ApiError'
        this.code = code
        this.requestId = requestId
        this.errorDetail = errorDetail
    }
}

// Definición de interfaces base de petición y respuesta
export type ApiRequestData = Record<string, unknown>
export type ApiResponseData = Record<string, unknown>

// ======================
// Configuración del cliente de API
// ======================

const API_BASE_URL = ENV.API_BASE_URL || '/api/v1'
const API_TIMEOUT = Number(ENV.API_TIMEOUT || 10000)

// Crear instancia de axios
const apiClient = axios.create({
    baseURL: API_BASE_URL,
    timeout: API_TIMEOUT,
    headers: {
        'Content-Type': 'application/json',
    },
})

// Interceptor de peticiones
apiClient.interceptors.request.use(
    (config) => {
        const token = useAuthStore.getState().token

        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`
        }

        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// Interceptor de respuestas - manejo unificado de errores y formato de respuesta
apiClient.interceptors.response.use(
    (response) => {
        // Comprobar si el formato de respuesta cumple el formato estándar de la API
        const data = response.data as APIResponse

        // Si la respuesta tiene success=false, tratarlo como error de lógica de negocio
        if (data && data.success === false) {
            throw new ApiError(
                data.message || 'Petición fallida',
                data.code || response.status,
                data.requestId,
                data.error
            )
        }

        return response
    },
    (error: AxiosError<APIResponse>) => {
        const status = error.response?.status
        const errorData = error.response?.data

        // Manejar el error 401 no autorizado
        if (status === 401) {
            useAuthStore.getState().logout()
            window.location.href = '/login'
        }

        // Convertir a objeto de error de API personalizado
        if (errorData) {
            throw new ApiError(
                errorData.message || 'Petición fallida',
                errorData.code || status || 500,
                errorData.requestId,
                errorData.error
            )
        } else {
            // Error de red u otro error no estándar
            throw new ApiError(
                error.message || 'Falló la petición de red',
                status || 500
            )
        }
    }
)

// Método envolvente que añade reintentos a las peticiones
const withRetry = async <T>(
    requestFn: () => Promise<T>,
    maxRetries = 3,
    delay = 1000
): Promise<T> => {
    let retries = 0

    while (retries < maxRetries) {
        try {
            return await requestFn()
        } catch (error) {
            // Usar aserción de tipos en lugar de any
            const err = error as Error
            if (error instanceof ApiError && error.code >= 500 && retries < maxRetries - 1) {
                // Solo reintentar en errores de servidor y si no es el último intento
                retries++
                await new Promise(resolve => setTimeout(resolve, delay * retries))
                continue
            }
            throw err
        }
    }

    throw new Error('Max retries reached')
}

// ======================
// Métodos de petición de la API
// ======================

/**
 * Petición GET
 * @param url URL de la petición
 * @param config configuración de la petición
 * @returns datos de la respuesta
 */
export const get = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.get<APIResponse<T>>(url, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            return response.data.data as T
        })
}

/**
 * Petición POST
 * @param url URL de la petición
 * @param data datos de la petición
 * @param config configuración de la petición
 * @returns datos de la respuesta
 */
export const post = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.post<APIResponse<T>>(url, data, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            return response.data.data as T
        })
}

/**
 * Petición PUT
 * @param url URL de la petición
 * @param data datos de la petición
 * @param config configuración de la petición
 * @returns datos de la respuesta
 */
export const put = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.put<APIResponse<T>>(url, data, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            return response.data.data as T
        })
}

/**
 * Petición PATCH
 * @param url URL de la petición
 * @param data datos de la petición
 * @param config configuración de la petición
 * @returns datos de la respuesta
 */
export const patch = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.patch<APIResponse<T>>(url, data, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            return response.data.data as T
        })
}

/**
 * Petición DELETE
 * @param url URL de la petición
 * @param config configuración de la petición
 * @returns datos de la respuesta
 */
export const del = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return apiClient.delete<APIResponse<T>>(url, config)
        .then((response: AxiosResponse<APIResponse<T>>) => {
            return response.data.data as T
        })
}

/**
 * Petición GET con reintentos
 * @param url URL de la petición
 * @param config configuración de la petición
 * @param retries número de reintentos
 * @returns datos de la respuesta
 */
export const getWithRetry = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => get<T>(url, config), retries)
}

/**
 * Petición POST con reintentos
 * @param url URL de la petición
 * @param data datos de la petición
 * @param config configuración de la petición
 * @param retries número de reintentos
 * @returns datos de la respuesta
 */
export const postWithRetry = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => post<T>(url, data, config), retries)
}

/**
 * Petición PUT con reintentos
 * @param url URL de la petición
 * @param data datos de la petición
 * @param config configuración de la petición
 * @param retries número de reintentos
 * @returns datos de la respuesta
 */
export const putWithRetry = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => put<T>(url, data, config), retries)
}

/**
 * Petición DELETE con reintentos
 * @param url URL de la petición
 * @param config configuración de la petición
 * @param retries número de reintentos
 * @returns datos de la respuesta
 */
export const delWithRetry = <T = ApiResponseData>(url: string, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => del<T>(url, config), retries)
}

/**
 * Petición PATCH con reintentos
 * @param url URL de la petición
 * @param data datos de la petición
 * @param config configuración de la petición
 * @param retries número de reintentos
 * @returns datos de la respuesta
 */
export const patchWithRetry = <T = ApiResponseData>(url: string, data?: unknown, config?: AxiosRequestConfig, retries = 3): Promise<T> => {
    return withRetry(() => patch<T>(url, data, config), retries)
}


export default apiClient