// src/validation/config.ts
import { z } from 'zod'

export const configSchema = z.object({
    isDebug: z.boolean().default(false),
    isResponseCheck: z.boolean().default(false),
    haproxy: z.object({
        thread: z.number()
            .min(0, "El número de hilos no puede ser menor que 0")
            .max(256, "El número de hilos no puede superar 256")
            .default(0),
        configBaseDir: z.string().min(1, "El directorio raíz de configuración no puede estar vacío"),
        haproxyBin: z.string().min(1, "La ruta del ejecutable no puede estar vacía"),
    }),
    engine: z.object({
        appConfig: z.array(z.object({
            name: z.string(),
            directives: z.string()
        }))
    })
})

export type ConfigFormValues = z.infer<typeof configSchema>