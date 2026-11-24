import { ENV } from "@/utils/env"

/**
 * Definición de tipos de constantes
 */
export type ConstantValue = string | number | boolean | null | undefined

/**
 * Enumeración de categorías de constantes
 */
export enum ConstantCategory {
    SYSTEM = 'system',
    UI = 'ui',
    CONFIG = 'config',
    FEATURE = 'feature'
}

/**
 * Interfaz de almacenamiento de constantes
 */
interface ConstantStore {
    [category: string]: {
        [key: string]: ConstantValue
    }
}

/**
 * Objeto de almacenamiento de constantes
 * Usar categorías para almacenar distintos tipos de constantes
 */
const constantStore: ConstantStore = {
    [ConstantCategory.SYSTEM]: {
        VERSION: '1.0.0',
        API_TIMEOUT: Number(ENV.API_TIMEOUT),
        DEBUG_MODE: ENV.isDevelopment,
    },
    [ConstantCategory.UI]: {
        DEFAULT_THEME: 'light',
        // MOBILE_BREAKPOINT: 768,
    },
    [ConstantCategory.CONFIG]: {
        DEFAULT_PAGE_SIZE: 10,
        MAX_PASSWORD_LENGTH: 20,
        MIN_PASSWORD_LENGTH: 8,
        ENGINE_NAME: 'coraza',
    },
    [ConstantCategory.FEATURE]: {
        QUERY_STALE_TIME: 5 * 60 * 1000,
        DEFAULT_QUERY_RETRY: 1,
        TOAST_DURATION: 2000,
        // LOG_RETENTION_DAYS: 90,
        // MAX_CERTIFICATES: 100,
    }
}

/**
 * Función de obtención de constantes
 * @param category categoría de la constante
 * @param key nombre de la clave de la constante
 * @param defaultValue valor por defecto (devuelto si la constante no existe)
 * @returns valor de la constante o valor por defecto
 */
export function getConstant<T extends ConstantValue>(
    category: ConstantCategory,
    key: string,
    defaultValue?: T
): T {
    if (
        constantStore[category] &&
        constantStore[category][key] !== undefined
    ) {
        return constantStore[category][key] as T
    }
    return defaultValue as T
}

/**
 * Función de establecimiento de constantes
 * @param category categoría de la constante
 * @param key nombre de la clave de la constante
 * @param value valor a establecer
 * @returns si se estableció con éxito
 */
export function setConstant(
    category: ConstantCategory,
    key: string,
    value: ConstantValue
): boolean {
    try {
        // Asegurar que la categoría exista
        if (!constantStore[category]) {
            constantStore[category] = {}
        }

        // Establecer el valor de la constante
        constantStore[category][key] = value
        return true
    } catch (error) {
        console.error(`Failed to set constant [${category}.${key}]:`, error)
        return false
    }
}

/**
 * Comprobar si la constante existe
 * @param category categoría de la constante
 * @param key nombre de la clave de la constante
 * @returns si existe
 */
export function hasConstant(
    category: ConstantCategory,
    key: string
): boolean {
    return (
        constantStore[category] !== undefined &&
        constantStore[category][key] !== undefined
    )
}

/**
 * Obtener todas las constantes de la categoría
 * @param category categoría de la constante
 * @returns objeto de la constante
 */
export function getCategoryConstants(
    category: ConstantCategory
): Record<string, ConstantValue> {
    return constantStore[category] || {}
}

/**
 * Establecer constantes por lotes
 * @param category categoría de la constante
 * @param constants pares clave-valor de constantes
 * @returns si todo se estableció con éxito
 */
export function setBatchConstants(
    category: ConstantCategory,
    constants: Record<string, ConstantValue>
): boolean {
    try {
        // Asegurar que la categoría exista
        if (!constantStore[category]) {
            constantStore[category] = {}
        }

        // Establecer constantes por lotes
        Object.entries(constants).forEach(([key, value]) => {
            constantStore[category][key] = value
        })

        return true
    } catch (error) {
        console.error(`Failed to set batch constants for category [${category}]:`, error)
        return false
    }
}