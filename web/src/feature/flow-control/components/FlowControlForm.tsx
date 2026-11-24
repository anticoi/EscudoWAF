import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    FormDescription,
} from '@/components/ui/form'
import { flowControlConfigSchema, FlowControlConfigFormValues } from '@/validation/flow-control'
import { useUpdateFlowControlConfig, useFlowControlConfig } from '../hooks/useFlowControl'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
import { useTranslation } from 'react-i18next'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle } from 'lucide-react'
import { useEffect } from 'react'

interface FlowControlFormProps {
    onSuccess?: () => void
}

// Componente skeleton del formulario de configuración de control de flujo
const FlowControlFormSkeleton = () => {
    return (
        <AnimatedContainer>
            <div className="space-y-8">
                {/* Skeleton del límite de frecuencia de acceso */}
                <div className="bg-background/50 rounded-xl p-6 space-y-6">
                    <div className="space-y-2">
                        <Skeleton className="h-7 w-40" />
                        <Skeleton className="h-4 w-96" />
                    </div>
                    
                    {/* Skeleton del interruptor de activación */}
                    <div className="flex flex-row items-center justify-between py-4">
                        <div className="space-y-1">
                            <Skeleton className="h-5 w-24" />
                            <Skeleton className="h-4 w-32" />
                        </div>
                        <Skeleton className="h-6 w-11 rounded-full" />
                    </div>

                    {/* Skeleton de la cuadrícula de elementos de configuración */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {Array.from({ length: 5 }).map((_, index) => (
                            <div key={index} className="space-y-2">
                                <Skeleton className="h-4 w-20" />
                                <Skeleton className="h-10 w-full" />
                                <Skeleton className="h-3 w-32" />
                            </div>
                        ))}
                    </div>
                </div>

                {/* Skeleton del límite de frecuencia de ataques */}
                <div className="bg-background/50 rounded-xl p-6 space-y-6">
                    <div className="space-y-2">
                        <Skeleton className="h-7 w-40" />
                        <Skeleton className="h-4 w-96" />
                    </div>
                    
                    <div className="flex flex-row items-center justify-between py-4">
                        <div className="space-y-1">
                            <Skeleton className="h-5 w-24" />
                            <Skeleton className="h-4 w-32" />
                        </div>
                        <Skeleton className="h-6 w-11 rounded-full" />
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {Array.from({ length: 5 }).map((_, index) => (
                            <div key={index} className="space-y-2">
                                <Skeleton className="h-4 w-20" />
                                <Skeleton className="h-10 w-full" />
                                <Skeleton className="h-3 w-32" />
                            </div>
                        ))}
                    </div>
                </div>

                {/* Skeleton del límite de frecuencia de errores */}
                <div className="bg-background/50 rounded-xl p-6 space-y-6">
                    <div className="space-y-2">
                        <Skeleton className="h-7 w-40" />
                        <Skeleton className="h-4 w-96" />
                    </div>
                    
                    <div className="flex flex-row items-center justify-between py-4">
                        <div className="space-y-1">
                            <Skeleton className="h-5 w-24" />
                            <Skeleton className="h-4 w-32" />
                        </div>
                        <Skeleton className="h-6 w-11 rounded-full" />
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {Array.from({ length: 5 }).map((_, index) => (
                            <div key={index} className="space-y-2">
                                <Skeleton className="h-4 w-20" />
                                <Skeleton className="h-10 w-full" />
                                <Skeleton className="h-3 w-32" />
                            </div>
                        ))}
                    </div>
                </div>

                {/* Skeleton del botón de envío */}
                <div className="flex justify-end pt-4">
                    <Skeleton className="h-11 w-24" />
                </div>
            </div>
        </AnimatedContainer>
    )
}

export function FlowControlForm({ onSuccess }: FlowControlFormProps) {
    const { t } = useTranslation()

    // Obtener la configuración actual
    const { flowControlConfig, isLoading: isConfigLoading } = useFlowControlConfig()

    // Actualizar configuración
    const {
        updateFlowControlConfig,
        isLoading: isUpdating,
        error,
        clearError
    } = useUpdateFlowControlConfig()

    // Configuración del formulario
    const form = useForm<FlowControlConfigFormValues>({
        resolver: zodResolver(flowControlConfigSchema),
        defaultValues: flowControlConfig || {
            visitLimit: {
                enabled: false,
                threshold: 100,
                statDuration: 60,
                blockDuration: 600,
                burstCount: 10,
                paramsCapacity: 10000,
            },
            attackLimit: {
                enabled: false,
                threshold: 5,
                statDuration: 60,
                blockDuration: 3600,
                burstCount: 2,
                paramsCapacity: 10000,
            },
            errorLimit: {
                enabled: false,
                threshold: 20,
                statDuration: 60,
                blockDuration: 1800,
                burstCount: 5,
                paramsCapacity: 10000,
            },
        },
    })

    // Restablecer el formulario cuando termine la carga de configuración
    useEffect(() => {
        if (flowControlConfig) {
            form.reset(flowControlConfig)
        }
    }, [flowControlConfig, form])

    // Manejo del envío del formulario
    const handleSubmit = (data: FlowControlConfigFormValues) => {
        if (clearError) clearError()

        updateFlowControlConfig({
            engine: {
                flowController: data
            }
        }, {
            onSuccess: () => {
                if (onSuccess) onSuccess()
            }
        })
    }

    if (isConfigLoading) {
        return <FlowControlFormSkeleton />
    }

    return (
        <AnimatedContainer>
            <Form {...form}>
                <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-8">
                    {/* Aviso de error de API */}
                    {error && (
                        <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* Configuración del límite de frecuencia de acceso */}
                    <div className="bg-background/50 rounded-xl p-6 space-y-6">
                        <div className="space-y-2">
                            <h3 className="text-xl font-semibold dark:text-shadow-glow-white">
                                {t('flowControl.visitLimit.title', 'Límite de frecuencia de acceso')}
                            </h3>
                            <p className="text-muted-foreground dark:text-shadow-glow-white">
                                {t('flowControl.visitLimit.description', 'Limitar el número de accesos de una IP en una ventana de tiempo, para prevenir acceso de alta frecuencia')}
                            </p>
                        </div>

                        <FormField
                            control={form.control}
                            name="visitLimit.enabled"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between py-4">
                                    <div className="space-y-1">
                                        <FormLabel className="text-base font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.enabled', 'Habilitar restricción')}
                                        </FormLabel>
                                        <FormDescription className="text-sm dark:text-shadow-glow-white">
                                            {t('flowControl.enabledDescription', 'Activar esta regla de restricción')}
                                        </FormDescription>
                                    </div>
                                    <FormControl>
                                        <Switch
                                            checked={field.value}
                                            onCheckedChange={field.onChange}
                                        />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <FormField
                                control={form.control}
                                name="visitLimit.threshold"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.threshold', 'Umbral de activación')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="100"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.thresholdDescription', 'Número máximo de peticiones dentro de la ventana de tiempo')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.statDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.statDuration', 'Ventana de tiempo estadístico')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="60"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.statDurationDescription', 'Duración de la ventana de tiempo (segundos)')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.blockDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.blockDuration', 'Duración del bloqueo')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="600"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.blockDurationDescription', 'Duración del bloqueo tras activar la restricción (segundos)')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.burstCount"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.burstCount', 'Número de peticiones en ráfaga')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.burstCountDescription', 'Cantidad de peticiones en ráfaga permitidas')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="visitLimit.paramsCapacity"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacity', 'Capacidad de caché')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10000"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacityDescription', 'Cantidad máxima de IPs en caché')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    </div>

                    {/* Configuración del límite de frecuencia de ataques */}
                    <div className="bg-background/50 rounded-xl p-6 space-y-6">
                        <div className="space-y-2">
                            <h3 className="text-xl font-semibold dark:text-shadow-glow-white">
                                {t('flowControl.attackLimit.title', 'Límite de frecuencia de ataques')}
                            </h3>
                            <p className="text-muted-foreground dark:text-shadow-glow-white">
                                {t('flowControl.attackLimit.description', 'Limitar el número de ataques de una IP en una ventana de tiempo, para prevenir ataques maliciosos')}
                            </p>
                        </div>

                        <FormField
                            control={form.control}
                            name="attackLimit.enabled"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between py-4">
                                    <div className="space-y-1">
                                        <FormLabel className="text-base font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.enabled', 'Habilitar restricción')}
                                        </FormLabel>
                                        <FormDescription className="text-sm dark:text-shadow-glow-white">
                                            {t('flowControl.enabledDescription', 'Activar esta regla de restricción')}
                                        </FormDescription>
                                    </div>
                                    <FormControl>
                                        <Switch
                                            checked={field.value}
                                            onCheckedChange={field.onChange}
                                        />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <FormField
                                control={form.control}
                                name="attackLimit.threshold"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.threshold', 'Umbral de activación')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="5"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.attackThresholdDescription', 'Número máximo de ataques dentro de la ventana de tiempo')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.statDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.statDuration', 'Ventana de tiempo estadístico')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="60"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.statDurationDescription', 'Duración de la ventana de tiempo (segundos)')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.blockDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.blockDuration', 'Duración del bloqueo')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="3600"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.blockDurationDescription', 'Duración del bloqueo tras activar la restricción (segundos)')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.burstCount"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.burstCount', 'Número de peticiones en ráfaga')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="2"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.burstCountDescription', 'Cantidad de peticiones en ráfaga permitidas')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="attackLimit.paramsCapacity"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacity', 'Capacidad de caché')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10000"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacityDescription', 'Cantidad máxima de IPs en caché')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    </div>

                    {/* Configuración del límite de frecuencia de errores */}
                    <div className="bg-background/50 rounded-xl p-6 space-y-6">
                        <div className="space-y-2">
                            <h3 className="text-xl font-semibold dark:text-shadow-glow-white">
                                {t('flowControl.errorLimit.title', 'Límite de frecuencia de errores')}
                            </h3>
                            <p className="text-muted-foreground dark:text-shadow-glow-white">
                                {t('flowControl.errorLimit.description', 'Limitar el número de respuestas de error de una IP en una ventana de tiempo, para prevenir sondeos maliciosos')}
                            </p>
                        </div>

                        <FormField
                            control={form.control}
                            name="errorLimit.enabled"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between py-4">
                                    <div className="space-y-1">
                                        <FormLabel className="text-base font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.enabled', 'Habilitar restricción')}
                                        </FormLabel>
                                        <FormDescription className="text-sm dark:text-shadow-glow-white">
                                            {t('flowControl.enabledDescription', 'Activar esta regla de restricción')}
                                        </FormDescription>
                                    </div>
                                    <FormControl>
                                        <Switch
                                            checked={field.value}
                                            onCheckedChange={field.onChange}
                                        />
                                    </FormControl>
                                </FormItem>
                            )}
                        />

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <FormField
                                control={form.control}
                                name="errorLimit.threshold"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.threshold', 'Umbral de activación')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="20"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.errorThresholdDescription', 'Número máximo de errores dentro de la ventana de tiempo')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.statDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.statDuration', 'Ventana de tiempo estadístico')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="60"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.statDurationDescription', 'Duración de la ventana de tiempo (segundos)')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.blockDuration"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.blockDuration', 'Duración del bloqueo')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="1800"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.blockDurationDescription', 'Duración del bloqueo tras activar la restricción (segundos)')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.burstCount"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.burstCount', 'Número de peticiones en ráfaga')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="5"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.burstCountDescription', 'Cantidad de peticiones en ráfaga permitidas')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="errorLimit.paramsCapacity"
                                render={({ field }) => (
                                    <FormItem className="space-y-2">
                                        <FormLabel className="font-medium dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacity', 'Capacidad de caché')}
                                        </FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="10000"
                                                className="dark:text-shadow-glow-white"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription className="text-xs dark:text-shadow-glow-white">
                                            {t('flowControl.paramsCapacityDescription', 'Cantidad máxima de IPs en caché')}
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    </div>

                    {/* Botón de envío */}
                    <div className="flex justify-end pt-4">
                        <AnimatedButton>
                            <Button type="submit" disabled={isUpdating} size="lg">
                                {isUpdating
                                    ? t('flowControl.updating', 'Actualizando...')
                                    : t('flowControl.updateConfig', 'Actualizar configuración')
                                }
                            </Button>
                        </AnimatedButton>
                    </div>
                </form>
            </Form>
        </AnimatedContainer>
    )
} 