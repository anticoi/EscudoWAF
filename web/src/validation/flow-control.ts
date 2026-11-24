import { z } from 'zod'

export const limitConfigSchema = z.object({
    enabled: z.boolean().default(false),
    threshold: z.number()
        .min(1, "El umbral no puede ser menor que 1")
        .max(10000, "El umbral no puede superar 10000")
        .default(100),
    statDuration: z.number()
        .min(10, "La ventana de tiempo no puede ser menor de 10 segundos")
        .max(3600, "La ventana de tiempo no puede superar 3600 segundos")
        .default(60),
    blockDuration: z.number()
        .min(60, "La duración del bloqueo no puede ser menor de 60 segundos")
        .max(86400, "La duración del bloqueo no puede superar 86400 segundos")
        .default(600),
    burstCount: z.number()
        .min(1, "El número de peticiones en ráfaga no puede ser menor que 1")
        .max(1000, "El número de peticiones en ráfaga no puede superar 1000")
        .default(10),
    paramsCapacity: z.number()
        .min(1000, "La capacidad de caché no puede ser menor que 1000")
        .max(100000, "La capacidad de caché no puede superar 100000")
        .default(10000),
})

export const flowControlConfigSchema = z.object({
    visitLimit: limitConfigSchema,
    attackLimit: limitConfigSchema,
    errorLimit: limitConfigSchema,
})

export const blockedIPListRequestSchema = z.object({
    page: z.number().min(1).default(1),
    size: z.number().min(1).max(100).default(10),
    ip: z.string().optional(),
    reason: z.string().optional(),
    status: z.enum(['active', 'expired', 'all']).default('all'),
    sortBy: z.enum(['blocked_at', 'blocked_until', 'ip']).default('blocked_at'),
    sortDir: z.enum(['asc', 'desc']).default('desc'),
})

export type LimitConfigFormValues = z.infer<typeof limitConfigSchema>
export type FlowControlConfigFormValues = z.infer<typeof flowControlConfigSchema>
export type BlockedIPListRequestFormValues = z.infer<typeof blockedIPListRequestSchema> 