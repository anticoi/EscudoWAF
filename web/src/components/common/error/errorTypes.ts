import { ApiError } from '@/api/index'

// Enumeración de tipos de error
export const ErrorType = {
    NETWORK: 'network',
    TIMEOUT: 'timeout',
    FORBIDDEN: 'forbidden',
    UNAUTHORIZED: 'unauthorized',
    SERVER: 'server',
    CLIENT: 'client',
    VALIDATION: 'validation',
    UNKNOWN: 'unknown'
} as const

// Definición de tipos de error
export type ErrorTypeValue = typeof ErrorType[keyof typeof ErrorType]

/**
 * Determinar el tipo de error según el objeto de error
 * @param error - objeto de error
 * @returns tipo de error
 */
export const getErrorType = (error: unknown): ErrorTypeValue => {
    // Comprobar si es un ApiError
    const isApiError = error instanceof Error && error.name === 'ApiError'

    // Si es ApiError, usar code para determinar el tipo de error
    if (isApiError && 'code' in error) {
        const apiError = error as ApiError
        const code = apiError.code

        if (code === 401) {
            return ErrorType.UNAUTHORIZED
        } else if (code === 403) {
            return ErrorType.FORBIDDEN
        } else if (code === 408 || code === 504) {
            return ErrorType.TIMEOUT
        } else if (code >= 400 && code < 500) {
            // Comprobar si es un error de validación
            if ('errorDetail' in apiError && apiError.errorDetail && typeof apiError.errorDetail === 'object') {
                return ErrorType.VALIDATION
            }
            return ErrorType.CLIENT
        } else if (code >= 500) {
            return ErrorType.SERVER
        }
    } else if (error instanceof Error) {
        // En casos que no son ApiError, intentar deducirlo del mensaje
        const message = error.message.toLowerCase()
        if (message.includes('network') || message.includes('Conexión') || message.includes('connection')) {
            return ErrorType.NETWORK
        } else if (message.includes('timeout') || message.includes('Tiempo de espera agotado')) {
            return ErrorType.TIMEOUT
        }
    }

    return ErrorType.UNKNOWN
}