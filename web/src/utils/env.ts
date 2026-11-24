// src/utils/env.ts
// Definir un método seguro de acceso al entorno, evitando usar process.env directamente
export const ENV = {
    NODE_ENV: import.meta.env.MODE || 'development',
    isDevelopment: import.meta.env.DEV,
    isProduction: import.meta.env.PROD,
    // Añadir otras variables de entorno necesarias
    API_BASE_URL: import.meta.env.VITE_API_BASE_URL,
    API_TIMEOUT: import.meta.env.VITE_API_TIMEOUT,
    // Si hay otras variables de entorno, pueden añadirse aquí
}