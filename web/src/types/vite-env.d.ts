// src/vite-env.d.ts
/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly MODE: string
    readonly BASE_URL: string
    readonly DEV: boolean
    readonly PROD: boolean
    readonly SSR: boolean

    // Variables de entorno personalizadas
    readonly VITE_API_BASE_URL?: string
    readonly VITE_API_TIMEOUT?: string
    // Añadir más variables de entorno personalizadas
}

interface ImportMeta {
    readonly env: ImportMetaEnv
}