// Tipo de datos de trayectoria de ataques WAF - para visualización del dashboard de seguridad
export type WAFAttackTrajectory = {
    type: string
    order: number
    from: string
    to: string
    flightCode: string
    date: string
    status: boolean
    startLat: number
    startLng: number
    endLat: number
    endLng: number
    arcAlt: number
    colorIndex: number // Índice de colores por tipo de ataque
}

// Configuración de colores de trayectorias de ataque WAF - cada color representa un tipo de ataque
export const WAF_ATTACK_TRAJECTORY_COLORS = [
    "#8ed4ff", // Azul claro - ataque general
    "#a071da", // Púrpura - ataque del color del tema
    "#ff6b6b", // Rojo - ataque de alto riesgo
    "#4ecdc4", // Cian - ataque medio
    "#45b7d1", // Azul - ataque de sondeo
    "#96ceb4", // Verde - ataque de bajo nivel
    "#ffeaa7", // Amarillo - ataques de escaneo
    "#fd79a8"  // Rosa - ataque especial
]

// Coordenadas del centro de protección WAF (Hangzhou)
export const WAF_DEFENSE_CENTER_COORDS = {
    lat: 30.274084,
    lng: 120.155070
} 