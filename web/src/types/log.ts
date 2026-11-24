// Parámetros de consulta de eventos de ataque
export interface AttackEventQuery {
    srcIp?: string          // Dirección IP de origen, dirección del atacante
    dstIp?: string          // Dirección IP de destino, dirección del servidor atacado
    domain?: string         // Dominio del sitio atacado
    srcPort?: number        // Puerto de origen desde el que se inició el ataque
    dstPort?: number        // Puerto de destino, puerto del servicio atacado
    startTime?: string      // Hora de inicio de la consulta
    endTime?: string        // Hora de fin de la consulta
    page?: number           // Página actual, comenzando en 1
    pageSize?: number       // Registros por página, máximo 100
}

// Parámetros de consulta de logs de ataque
export interface AttackLogQuery {
    ruleId?: number         // ID de la regla, identificador de la regla WAF que detectó el ataque
    srcIp?: string          // Dirección IP de origen, dirección del atacante
    dstIp?: string          // Dirección IP de destino, dirección del servidor atacado
    domain?: string         // Dominio del sitio atacado
    srcPort?: number        // Puerto de origen desde el que se inició el ataque
    dstPort?: number        // Puerto de destino, puerto del servicio atacado
    requestId?: string      // ID de la petición, identificador único de la petición HTTP
    startTime?: string      // Hora de inicio de la consulta
    endTime?: string        // Hora de fin de la consulta
    page?: number           // Página actual, comenzando en 1
    pageSize?: number       // Registros por página, máximo 100
}

// Interfaz de información de geolocalización de IPs
export interface IPInfo {
    city: {
        nameZh: string       // Nombre de la ciudad en chino
        nameEn: string       // Nombre de la ciudad en inglés
    }
    subdivision: {
        nameZh: string       // Nombre de la provincia/estado en chino
        nameEn: string       // Nombre de la provincia/estado en inglés
        isoCode: string      // Código de la provincia/estado
    }
    country: {
        nameZh: string       // Nombre del país en chino
        nameEn: string       // Nombre del país en inglés
        isoCode: string      // Código ISO del país
    }
    continent: {
        nameZh: string       // Nombre del continente en chino
        nameEn: string       // Nombre del continente en inglés
    }
    location: {
        longitude: number    // Longitud
        latitude: number     // Latitud
        timeZone: string     // Zona horaria
    }
    asn: {
        number: number       // Número ASN
        organization: string // Nombre de la organización
    }
}
// Resultado de agregación de eventos de ataque
export interface AttackEventAggregateResult {
    srcIp: string            // Dirección IP de origen, dirección del atacante
    srcIpInfo: IPInfo        // Información de geolocalización de la IP de origen
    count: number            // Número total de ataques desde el mismo origen
    domain: string           // Dominio del sitio atacado
    dstPort: number          // Puerto de destino, puerto del servicio atacado
    durationInMinutes: number // Duración del ataque (minutos), lapso desde el primer hasta el último ataque
    firstAttackTime: string   // Hora del primer ataque desde esta IP
    isOngoing: boolean        // Si está en curso, indica si el ataque continúa
    lastAttackTime: string    // Hora del último ataque desde esta IP
}

// Resultado paginado de eventos de ataque
export interface AttackEventResponse {
    currentPage: number
    pageSize: number
    results: AttackEventAggregateResult[]
    totalCount: number
    totalPages: number
}

// Entrada de log del WAF
export interface Log {
    accuracy: number           // Precisión de coincidencia de la regla (0-10)
    logRaw: string             // Datos de log originales
    message: string            // Mensaje de log
    payload: string            // Carga del ataque
    phase: number              // Fase de procesamiento de la petición
    ruleId: number             // ID de la regla
    secLangRaw: string         // Definición original de la regla de seguridad
    secMark: string            // Marca de seguridad
    severity: number           // Nivel de severidad (0-5)
}

// Registro de eventos de seguridad del WAF
export interface WAFLog {
    id: string                // Identificador único del log
    requestId: string         // Identificador único de la petición
    ruleId: number            // ID de la regla que se disparó
    secLangRaw: string        // Definición original de la regla de seguridad
    severity: number          // Nivel de severidad del evento (0-5)
    phase: number             // Fase de procesamiento de la petición
    secMark: string           // Marca de seguridad
    accuracy: number          // Precisión de coincidencia de la regla (0-10)
    payload: string           // Carga del ataque
    uri: string               // Ruta del URI de la petición
    srcIp: string             // Dirección IP de origen
    srcIpInfo?: IPInfo        // Información de geolocalización de la IP de origen
    dstIp: string             // Dirección IP de destino
    clientIp: string          // Dirección IP del cliente
    serverIp: string          // Dirección IP del servidor
    srcPort: number           // Puerto de origen
    dstPort: number           // Puerto de destino
    domain: string            // Dominio de destino
    logs: Log[]               // Entradas de log asociadas
    message: string           // Mensaje descriptivo del evento
    request: string           // Petición HTTP original
    response: string          // Respuesta HTTP original
    createdAt: string         // Marca de tiempo del evento
}
// Resultado paginado de logs de ataque
export interface AttackLogResponse {
    currentPage: number
    pageSize: number
    results: WAFLog[]
    totalCount: number
    totalPages: number
}

// Datos del diálogo de detalles del log
export interface AttackDetailData {
    target: string           // Objetivo (domain + dstPort + url)
    srcIp: string            // Dirección IP de origen
    srcIpInfo?: IPInfo       // Información de geolocalización de la IP de origen
    srcPort: number          // Puerto de origen
    dstIp: string            // Dirección IP de destino 
    dstPort: number          // Puerto de destino
    payload: string          // Carga del ataque
    message: string          // Mensaje
    ruleId: number           // ID de la regla
    requestId: string        // ID de la petición
    createdAt: string        // Fecha de creación
    request: string          // Contenido de la petición
    response: string         // Contenido de la respuesta
    logs: string             // Contenido del log
} 