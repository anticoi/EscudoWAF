import { get, post, put, del } from './index'
import {
    Certificate,
    CertificateListResponse,
    CertificateCreateRequest,
    CertificateUpdateRequest,
} from '@/types/certificate'

// Ruta base de la API de certificados
const BASE_URL = '/certificate'

/**
 * Servicios de API relacionados con certificados
 */
export const certificatesApi = {
    /**
     * Obtener lista de certificados
     * @param page número de página
     * @param size cantidad por página
     * @returns datos de respuesta de lista de certificados
     */
    getCertificates: (page: number = 1, size: number = 10): Promise<CertificateListResponse> => {
        return get<CertificateListResponse>(BASE_URL, {
            params: { page, size }
        })
    },

    /**
     * Crear nuevo certificado
     * @param certificate datos de la petición de creación de certificado
     * @returns detalles del certificado creado
     */
    createCertificate: (certificate: CertificateCreateRequest): Promise<Certificate> => {
        return post<Certificate>(BASE_URL, certificate)
    },

    /**
     * Obtener detalles de un certificado
     * @param id ID del certificado
     * @returns detalles del certificado
     */
    getCertificate: (id: string): Promise<Certificate> => {
        return get<Certificate>(`${BASE_URL}/${id}`)
    },

    /**
     * Actualizar certificado
     * @param id ID del certificado
     * @param certificate datos de la petición de actualización de certificado
     * @returns detalles del certificado actualizado
     */
    updateCertificate: (id: string, certificate: CertificateUpdateRequest): Promise<Certificate> => {
        return put<Certificate>(`${BASE_URL}/${id}`, certificate)
    },

    /**
     * Eliminar certificado
     * @param id ID del certificado
     * @returns void
     */
    deleteCertificate: (id: string): Promise<void> => {
        return del<void>(`${BASE_URL}/${id}`)
    }
}