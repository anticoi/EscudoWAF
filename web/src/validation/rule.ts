import { z } from 'zod'
import { Condition, MatchType, TARGET_MATCH_TYPES, TargetType } from '@/types/rule'

const allMatchTypesArray = Object.values(TARGET_MATCH_TYPES).flat() as MatchType[]
const matchTypeEnum = z.enum(allMatchTypesArray as [MatchType, ...MatchType[]])


// Validación de condición simple - usa superRefine para validaciones dependientes
// TODO: los distintos tipos de match_value requieren diferentes métodos de validación
const simpleConditionSchema = z.object({
    type: z.literal('simple'),
    target: z.enum(['source_ip', 'url', 'path']),
    match_type: matchTypeEnum,
    match_value: z.string().min(1, { message: 'Match value is required' })
}).superRefine((data, ctx) => {
    // Validar la compatibilidad entre el tipo de objetivo y el modo de coincidencia
    const validMatchTypes = TARGET_MATCH_TYPES[data.target as TargetType]
    if (!validMatchTypes.includes(data.match_type as MatchType)) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: `Match type '${data.match_type}' is not valid for target '${data.target}'`,
            path: ['match_type']
        })
    }
})

// Definir recursivamente la validación de condiciones compuestas

const conditionSchema: z.ZodType<Condition> = z.lazy(() =>
    z.union([
        simpleConditionSchema,
        z.object({
            type: z.literal('composite'),
            operator: z.enum(['AND', 'OR']),
            conditions: z.array(conditionSchema).min(1)
        })
    ])
)

// Validación de la petición de creación de regla
export const ruleCreateSchema = z.object({
    name: z.string().min(1, { message: 'Name is required' }),
    type: z.enum(['whitelist', 'blacklist']),
    status: z.enum(['enabled', 'disabled']),
    priority: z.number().int().min(1).max(10000),
    condition: conditionSchema,
})

// Validación de la petición de actualización de regla
export const ruleUpdateSchema = ruleCreateSchema.partial()