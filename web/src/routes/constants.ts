export const BASE_PATH = '/' as const

export const ROUTES = {
    LOGS: "/logs",
    MONITOR: "/monitor",
    RULES: "/rules",
    SETTINGS: "/settings",
} as const

export type RouteKey = keyof typeof ROUTES
export type RoutePath = typeof ROUTES[RouteKey]

// Función auxiliar para obtener rutas de forma segura en tipos
export const getRoute = (key: RouteKey): RoutePath => ROUTES[key] 