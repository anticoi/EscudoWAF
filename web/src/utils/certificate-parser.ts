import * as forge from 'node-forge'

/**
 * Interfaz de información del certificado analizado
 */
export interface ParsedCertificate {
    /** Hora de expiración del certificado */
    expireDate: string
    /** Huella del certificado (SHA-256) */
    fingerPrint: string
    /** Nombre del emisor del certificado */
    issuerName: string
    /** Lista de dominios incluidos en el certificado */
    domains: string[]
}

/**
 * Extraer la información clave de un certificado en formato PEM
 * @param certPem certificado en formato PEM
 * @returns información del certificado analizada
 */
export function parseCertificate(certPem: string): ParsedCertificate {
    try {
        // Analizar certificado desde formato PEM
        const cert = forge.pki.certificateFromPem(certPem)

        // Analizar la información del emisor
        const issuerCN = cert.issuer.getField('CN')
        const issuerName = issuerCN ? issuerCN.value : 'Desconocido'

        // Analizar la fecha de expiración
        const expireDate = cert.validity.notAfter.toISOString()

        // Calcular la huella del certificado (SHA-256)
        const md = forge.md.sha256.create()
        const derBytes = forge.asn1.toDer(forge.pki.certificateToAsn1(cert)).getBytes()
        md.update(derBytes)
        const fingerPrint = md.digest().toHex().match(/.{2}/g)?.join(':') || ''

        // Analizar el dominio
        const domains: string[] = []

        // Intentar obtener el dominio del nombre alternativo (extensión SAN)
        try {
            const altNamesExt = cert.getExtension('subjectAltName')
            if (altNamesExt && 'altNames' in altNamesExt) {
                const altNames = altNamesExt.altNames as Array<{ type: number; value: string }>
                if (Array.isArray(altNames)) {
                    altNames.forEach((altName) => {
                        if (altName.type === 2) { // Tipo DNS
                            domains.push(altName.value)
                        }
                    })
                }
            }
        } catch (e) {
            console.error('Error al analizar la extensión del dominio', e instanceof Error ? e.message : String(e))
        }

        // Si no se obtuvo el dominio de la extensión, intentar obtenerlo del CN
        if (domains.length === 0) {
            const commonName = cert.subject.getField('CN')
            if (commonName) {
                domains.push(commonName.value)
            }
        }

        return {
            expireDate,
            fingerPrint,
            issuerName,
            domains
        }
    } catch (error) {
        console.error('Error al analizar el certificado:', error instanceof Error ? error.message : String(error))
        return {
            expireDate: '',
            fingerPrint: '',
            issuerName: 'Error de análisis',
            domains: []
        }
    }
}

/**
 * Leer el contenido de texto del archivo
 * @param file objeto de archivo
 * @returns Promise que resuelve al contenido del archivo
 */
export function readFileAsText(file: File): Promise<string> {
    return new Promise<string>((resolve, reject) => {
        const reader = new FileReader()

        reader.onload = (event: ProgressEvent<FileReader>) => {
            if (event.target && typeof event.target.result === 'string') {
                resolve(event.target.result)
            } else {
                reject(new Error('Error al leer el archivo: contenido no válido'))
            }
        }

        reader.onerror = (event: ProgressEvent<FileReader>) => {
            reject(new Error(`Error al leer el archivo: ${event.target?.error?.message || 'Error desconocido'}`))
        }

        reader.readAsText(file)
    })
}