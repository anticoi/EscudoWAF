import * as z from 'zod'
import { WAFMode } from '@/types/site'

// Modo de validación de servidor
const serverSchema = z.object({
    host: z.string().min(1, "La dirección del host no puede estar vacía"),
    port: z.number().min(1, "El puerto debe ser mayor que 0").max(65535, "El puerto no puede superar 65535"),
    isSSL: z.boolean().default(false)
})

// Modo de validación de la configuración del backend
const backendSchema = z.object({
    servers: z.array(serverSchema).min(1, "Se requiere al menos un servidor")
})

// Esquema de validación de certificado (opcional)
const certificateSchema = z.object({
    certName: z.string().optional(),
    expireDate: z.string().optional(),
    fingerPrint: z.string().optional(),
    issuerName: z.string().optional(),
    privateKey: z.string().optional(),
    publicKey: z.string().optional()
}).optional()

// Esquema de validación del formulario de sitio
export const siteFormSchema = z.object({
    name: z.string().min(1, "El nombre del sitio no puede estar vacío"),
    domain: z.string().min(1, "El dominio no puede estar vacío"),
    listenPort: z.number().min(1, "El puerto de escucha debe ser mayor que 0").max(65535, "El puerto de escucha no puede superar 65535"),
    enableHTTPS: z.boolean().default(false),
    activeStatus: z.boolean().default(true),
    wafEnabled: z.boolean().default(false),
    wafMode: z.enum([WAFMode.Protection, WAFMode.Observation]).default(WAFMode.Observation),
    backend: backendSchema,
    certificate: certificateSchema,
    certificateId: z.string().optional() // Para seleccionar un certificado existente
})

// Esquema de validación del formulario de actualización de sitio
export const siteUpdateFormSchema = siteFormSchema.partial() 