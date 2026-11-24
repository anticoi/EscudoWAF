import { z } from 'zod'

// Reglas de validación del formulario de consulta de eventos de ataque
export const attackEventQuerySchema = z.object({
  srcIp: z.string().optional(),
  dstIp: z.string().optional(),
  domain: z.string().optional(),
  srcPort: z.coerce.number().optional(),
  dstPort: z.coerce.number().optional(),
  startTime: z.string().optional(),
  endTime: z.string().optional(),
  page: z.coerce.number().default(1),
  pageSize: z.coerce.number().default(10)
})

export type AttackEventQueryFormValues = z.infer<typeof attackEventQuerySchema>

// Reglas de validación del formulario de consulta de logs de ataque
export const attackLogQuerySchema = z.object({
  ruleId: z.coerce.number().optional(),
  srcIp: z.string().optional(),
  dstIp: z.string().optional(),
  domain: z.string().optional(),
  srcPort: z.coerce.number().optional(),
  dstPort: z.coerce.number().optional(),
  requestId: z.string().optional(),
  startTime: z.string().optional(),
  endTime: z.string().optional(),
  page: z.coerce.number().default(1),
  pageSize: z.coerce.number().default(10)
})

export type AttackLogQueryFormValues = z.infer<typeof attackLogQuerySchema> 