/**
 * Estructura de datos del grupo de IPs
 */
export interface IPGroup {
    id: string             // Identificador único del grupo de IPs
    name: string           // Nombre del grupo de IPs
    items: string[]        // Lista de direcciones IP o CIDR
}

/**
 * Petición de creación de grupo de IPs
 */
export interface IPGroupCreateRequest {
    name: string           // Nombre del grupo de IPs
    items: string[]        // Lista de direcciones IP o CIDR
}

/**
 * Petición de actualización de grupo de IPs
 */
export interface IPGroupUpdateRequest {
    name?: string          // Nombre del grupo de IPs
    items?: string[]       // Lista de direcciones IP o CIDR
}

/**
 * Respuesta de lista de grupos de IPs
 */
export interface IPGroupListResponse {
    total: number          // Total
    items: IPGroup[]       // Lista de grupos de IPs
}

/**
 * Datos del grupo de IPs para el formulario
 */
export interface IPGroupFormValues {
    name: string
    items: string[]
}