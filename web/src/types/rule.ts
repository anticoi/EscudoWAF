// Tipo de objetivo de coincidencia
export type TargetType = 'source_ip' | 'url' | 'path'

// Tipo de modo de coincidencia
export type MatchType =
    // Modo de coincidencia de IP
    | 'equal'
    | 'not_equal'
    | 'fuzzy'
    | 'in_cidr'
    | 'not_in_cidr'
    | 'in_ipgroup'
    | 'not_in_ipgroup'
    // Modo de coincidencia de URL y Path
    | 'include'
    | 'contains'
    | 'not_contains'
    | 'prefix_keyword'
    | 'regex'

// Operador lógico
export type LogicalOperator = 'AND' | 'OR'

// Tipo de condición
export type ConditionType = 'simple' | 'composite'

// Tipo de regla
export type RuleType = 'whitelist' | 'blacklist'

// Estado de la regla
export type RuleStatus = 'enabled' | 'disabled'

// Condición simple
export interface SimpleCondition {
    type: 'simple'
    target: TargetType
    match_type: MatchType
    match_value: string
}

// Condición compuesta
export interface CompositeCondition {
    type: 'composite'
    operator: LogicalOperator
    conditions: (SimpleCondition | CompositeCondition)[]
}

// Tipo de condición (tipo unión)
export type Condition = SimpleCondition | CompositeCondition

// Modelo de microrregla
export interface MicroRule {
    id: string
    name: string
    type: RuleType
    status: RuleStatus
    priority: number
    condition: Condition
    createdAt?: string
    updatedAt?: string
}

// Petición de creación de regla
export interface MicroRuleCreateRequest {
    name: string
    type: RuleType
    status: RuleStatus
    priority: number
    condition: Condition
}

// Petición de actualización de regla
export interface MicroRuleUpdateRequest {
    name?: string
    type?: RuleType
    status?: RuleStatus
    priority?: number
    condition?: Condition
}

// Respuesta de lista de reglas
export interface MicroRuleListResponse {
    total: number
    items: MicroRule[]
}

// Relación de mapeo entre tipos de objetivo y modos de coincidencia
export const TARGET_MATCH_TYPES: Record<TargetType, MatchType[]> = {
    'source_ip': ['equal', 'not_equal', 'fuzzy', 'in_cidr', 'not_in_cidr', 'in_ipgroup', 'not_in_ipgroup'],
    'url': ['equal', 'not_equal', 'contains', 'not_contains', 'prefix_keyword', 'regex'],
    'path': ['equal', 'not_equal', 'contains', 'not_contains', 'prefix_keyword', 'regex']
}