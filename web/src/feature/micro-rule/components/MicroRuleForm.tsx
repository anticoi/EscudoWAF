import { useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { AlertCircle } from 'lucide-react'
import { ruleCreateSchema } from '@/validation/rule'
import { MicroRuleCreateRequest, Condition, MicroRule } from '@/types/rule'
import { useCreateMicroRule, useUpdateMicroRule } from '../hooks/useMicroRule'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { ConditionBuilder } from './ConditionBuilder'

// Condición compuesta por defecto
const DEFAULT_CONDITION: Condition = {
    type: 'composite',
    operator: 'AND',
    conditions: [
        {
            type: 'simple',
            target: 'source_ip',
            match_type: 'equal',
            match_value: '',
        }
    ]
}

interface MicroRuleFormProps {
    mode?: 'create' | 'update'
    ruleId?: string
    onSuccess?: () => void
    defaultValues?: MicroRule
}

export function MicroRuleForm({
    mode = 'create',
    ruleId,
    onSuccess,
    defaultValues,
}: MicroRuleFormProps) {
    const { t } = useTranslation()

    // Preparar valores por defecto
    const initialValues: MicroRuleCreateRequest = defaultValues || {
        name: '',
        type: 'blacklist',
        status: 'enabled',
        priority: 100,
        condition: DEFAULT_CONDITION,
    }

    // Hooks de la API
    const {
        createMicroRule,
        isLoading: isCreating,
        error: createError,
        clearError: clearCreateError
    } = useCreateMicroRule()

    const {
        updateMicroRule,
        isLoading: isUpdating,
        error: updateError,
        clearError: clearUpdateError
    } = useUpdateMicroRule()

    // Estado dinámico
    const isLoading = mode === 'create' ? isCreating : isUpdating
    const error = mode === 'create' ? createError : updateError
    const clearError = mode === 'create' ? clearCreateError : clearUpdateError

    // Configuración del formulario
    const form = useForm<MicroRuleCreateRequest>({
        resolver: zodResolver(ruleCreateSchema),
        defaultValues: initialValues,
    })


    // Manejo del envío del formulario
    const handleFormSubmit = useCallback((data: MicroRuleCreateRequest) => {
        // Limpiar los errores anteriores
        if (clearError) clearError()


        // Ejecutar creación o actualización según el modo
        if (mode === 'create') {
            createMicroRule(data, {
                onSuccess: () => {
                    // Restablecer el formulario
                    form.reset({
                        name: '',
                        type: 'blacklist',
                        status: 'enabled',
                        priority: 100,
                        condition: DEFAULT_CONDITION,
                    })
                    // Notificar éxito al componente padre
                    if (onSuccess) onSuccess()
                }
            })
        } else if (mode === 'update' && ruleId) {
            updateMicroRule({ id: ruleId, data }, {
                onSuccess: () => {
                    // Notificar éxito al componente padre
                    if (onSuccess) onSuccess()
                }
            })
        }
    }, [mode, ruleId, clearError, createMicroRule, updateMicroRule, form, onSuccess])

    return (
        <AnimatedContainer>
            <Form {...form}>
                <form onSubmit={form.handleSubmit(handleFormSubmit)} className="space-y-6">
                    {/* Aviso de error de API */}
                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* Campos de información básica */}
                    <div className="grid grid-cols-2 gap-4">
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.name")}</FormLabel>
                                    <FormControl>
                                        <Input className="dark:text-shadow-glow-white" placeholder={t("microRule.form.namePlaceholder")} {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="priority"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.priority")}</FormLabel>
                                    <FormControl>
                                        <Input
                                            className="dark:text-shadow-glow-white"
                                            placeholder={t("microRule.form.priorityPlaceholder")}
                                            {...field}
                                            onChange={(e) => field.onChange(Number(e.target.value))}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Tipo de regla */}
                        <FormField
                            control={form.control}
                            name="type"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.type")}</FormLabel>
                                    <Select
                                        onValueChange={field.onChange}
                                        defaultValue={field.value}
                                        value={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger className="dark:text-shadow-glow-white">
                                                <SelectValue placeholder={t("microRule.form.selectType")} />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="whitelist">{t("microRule.form.whitelist")}</SelectItem>
                                            <SelectItem value="blacklist">{t("microRule.form.blacklist")}</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Estado de la regla */}
                        <FormField
                            control={form.control}
                            name="status"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel className="dark:text-shadow-glow-white">{t("microRule.form.status")}</FormLabel>
                                    <Select
                                        onValueChange={field.onChange}
                                        defaultValue={field.value}
                                        value={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger className="dark:text-shadow-glow-white">
                                                <SelectValue placeholder={t("microRule.form.selectStatus")} />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="enabled">{t("microRule.form.enabled")}</SelectItem>
                                            <SelectItem value="disabled">{t("microRule.form.disabled")}</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>

                    {/* Constructor de condiciones */}

                    <div className="space-y-4 mt-6">
                        <FormLabel className="dark:text-shadow-glow-white text-base">{t("microRule.form.condition")}</FormLabel>
                        <div className="border p-6 rounded-md dark:border-gray-700">
                            <ConditionBuilder
                                form={form}
                                path="condition"
                                isRoot={true}
                            />
                        </div>
                    </div>


                    {/* Botón de envío */}
                    <div className="flex justify-end mt-6">
                        <AnimatedButton>
                            <Button type="submit" disabled={isLoading}>
                                {isLoading ? t("common.submitting") : mode === 'create' ? t("common.create") : t("common.save")}
                            </Button>
                        </AnimatedButton>
                    </div>
                </form>
            </Form>
        </AnimatedContainer>
    )
}