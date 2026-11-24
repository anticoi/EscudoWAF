// src/feature/global-setting/components/ConfigForm.tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useState } from 'react'
import { ConfigResponse } from '@/types/config'
import { configSchema, ConfigFormValues } from '@/validation/config'
import { Card, CardContent, CardFooter } from '@/components/ui/card'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Slider } from '@/components/ui/slider'
import {
    AlertCircle,
    Save,
    Loader2,
    Cpu,
    Shield,
    Bug,
    Code,
    Folder,
    Terminal,
    Globe,
    Palette
} from 'lucide-react'
import { useUpdateConfig, getEngineNameConstant } from '../hooks/useConfig'
import { AdvancedErrorDisplay } from '@/components/common/error/errorDisplay'
import { LanguageSelector } from '@/components/common/language-selector'
import { ThemeToggle } from '@/components/common/theme-toggle'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'

interface ConfigFormProps {
    config?: ConfigResponse
    isLoading: boolean
}

export function ConfigForm({ config, isLoading }: ConfigFormProps) {
    const { updateConfig, isLoading: isUpdating, error, clearError } = useUpdateConfig()
    const [showThreadWarning, setShowThreadWarning] = useState(false)
    const { t } = useTranslation()

    // Obtener el nombre del motor
    const engineName = getEngineNameConstant()

    // Crear formulario
    const form = useForm<ConfigFormValues>({
        resolver: zodResolver(configSchema),
        defaultValues: {
            isDebug: false,
            isResponseCheck: false,
            haproxy: {
                thread: 0,
                configBaseDir: '',
                haproxyBin: '',
            },
            engine: {
                appConfig: [
                    {
                        name: engineName,
                        directives: ''
                    }
                ]
            }
        }
    })

    // Al cargar los datos de configuración, actualizar los valores del formulario
    useEffect(() => {
        if (config) {
            const engineAppConfig = config.engine.appConfig.find(app => app.name === engineName)

            form.reset({
                isDebug: config.isDebug,
                isResponseCheck: config.isResponseCheck,
                haproxy: {
                    thread: config.haproxy.thread,
                    configBaseDir: config.haproxy.configBaseDir,
                    haproxyBin: config.haproxy.haproxyBin,
                },
                engine: {
                    appConfig: [
                        {
                            name: engineName,
                            directives: engineAppConfig?.directives || ''
                        }
                    ]
                }
            })

            // Comprobar en la carga inicial si los hilos son 0
            setShowThreadWarning(config.haproxy.thread === 0)
        }
    }, [config, engineName, form])

    // Escuchar cambios en el número de hilos
    useEffect(() => {
        const subscription = form.watch((value, { name }) => {
            if (name === 'haproxy.thread') {
                setShowThreadWarning(Number(value.haproxy?.thread) === 0)
            }
        })

        return () => subscription.unsubscribe()
    }, [form])

    // Enviar formulario
    const onSubmit = (values: ConfigFormValues) => {
        clearError()

        const engineDirectives = values.engine.appConfig[0]?.directives || ''

        updateConfig({
            isDebug: values.isDebug,
            isResponseCheck: values.isResponseCheck,
            haproxy: {
                thread: values.haproxy.thread,
                configBaseDir: values.haproxy.configBaseDir,
                haproxyBin: values.haproxy.haproxyBin,
            },
            engine: {
                appConfig: [
                    {
                        name: engineName,
                        directives: engineDirectives
                    }
                ]
            }
        })
    }

    return (
        <Card className="h-full border-none shadow-none flex flex-col gap-8">
            <CardContent className="p-0 flex flex-col gap-8">
                <div className="flex items-center gap-2">
                    <Shield className="h-5 w-5 text-primary dark:text-shadow-primary" />
                    <h3 className="text-lg font-medium dark:text-shadow-glow-white">{t("globalSetting.engine.setting")}</h3>
                </div>

                <Form {...form}>
                    <form id="config-form" onSubmit={form.handleSubmit(onSubmit)} className="flex flex-col gap-8">
                        {/* Número de hilos del motor */}
                        <FormField
                            control={form.control}
                            name="haproxy.thread"
                            render={({ field }) => (
                                <FormItem>
                                    <div className="flex items-center gap-2 mb-2">
                                        <Cpu className="h-4 w-4 text-muted-foreground" />
                                        <FormLabel className="dark:text-shadow-glow-white">{t("globalSetting.config.engineThreads")}</FormLabel>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <FormControl className="flex-1">
                                            <Slider
                                                min={0}
                                                max={256}
                                                step={1}
                                                value={[field.value]}
                                                onValueChange={(value) => field.onChange(value[0])}
                                            />
                                        </FormControl>
                                        <div className="w-12 text-center font-medium dark:text-shadow-primary">
                                            {field.value}
                                        </div>
                                    </div>
                                    <FormDescription className="dark:text-shadow-primary-bold">
                                        {t("globalSetting.config.threadsDescription")}
                                    </FormDescription>
                                    {showThreadWarning && (
                                        <Alert variant="default" className="mt-2 bg-zinc-50 dark:bg-muted border-none shadow-none">
                                            <AlertCircle className="h-4 w-4 dark:text-shadow-primary" />
                                            <AlertDescription className="dark:text-shadow-glow-white">
                                                {t("globalSetting.config.threadsWarning")}
                                            </AlertDescription>
                                        </Alert>
                                    )}
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Inspección de respuestas */}
                        <FormField
                            control={form.control}
                            name="isResponseCheck"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between border p-0 border-none shadow-none">
                                    <div className="space-y-0.5">
                                        <div className="flex items-center gap-2">
                                            <Shield className="h-4 w-4 text-muted-foreground" />
                                            <FormLabel className="text-base dark:text-shadow-glow-white">{t("globalSetting.config.responseCheck")}</FormLabel>
                                        </div>
                                        <FormDescription className="dark:text-shadow-primary-bold">
                                            {t("globalSetting.config.responseCheckDescription")}
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

                        {/* Modo de depuración */}
                        <FormField
                            control={form.control}
                            name="isDebug"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-center justify-between border p-0 border-none shadow-none">
                                    <div className="space-y-0.5">
                                        <div className="flex items-center gap-2">
                                            <Bug className="h-4 w-4 text-muted-foreground" />
                                            <FormLabel className="text-base dark:text-shadow-glow-white">{t("globalSetting.config.debugMode")}</FormLabel>
                                        </div>
                                        <FormDescription className="dark:text-shadow-primary-bold">
                                            {t("globalSetting.config.debugModeDescription")}
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

                        {/* Configuración de directivas del motor */}
                        <FormField
                            control={form.control}
                            name="engine.appConfig.0.directives"
                            render={({ field }) => (
                                <FormItem>
                                    <div className="flex items-center gap-2 mb-2">
                                        <Code className="h-4 w-4 text-muted-foreground dark:text-shadow-primary" />
                                        <FormLabel className="dark:text-shadow-glow-white">{t("globalSetting.config.engineDirectives")}</FormLabel>
                                    </div>
                                    <FormControl>
                                        <Textarea
                                            placeholder={t("globalSetting.config.engineDirectivesPlaceholder")}
                                            className="min-h-[200px] font-mono text-sm scrollbar-neon dark:text-shadow-glow-white"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormDescription className="dark:text-shadow-primary-bold">
                                        {t("globalSetting.config.engineDirectivesDescription")}
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Directorio raíz de configuración de HAProxy */}
                        <FormField
                            control={form.control}
                            name="haproxy.configBaseDir"
                            render={({ field }) => (
                                <FormItem>
                                    <div className="flex items-center gap-2 mb-2">
                                        <Folder className="h-4 w-4 text-muted-foreground dark:text-shadow-primary" />
                                        <FormLabel className="dark:text-shadow-glow-white">{t("globalSetting.config.haproxyConfigDir")}</FormLabel>
                                    </div>
                                    <FormControl>
                                        <Input
                                            {...field}
                                            className="border-0 border-b border-gray-300 rounded-none focus:ring-0 focus-visible:ring-0 focus-visible:border-primary px-0 dark:text-shadow-glow-white"
                                        />
                                    </FormControl>
                                    <FormDescription className="dark:text-shadow-primary-bold">
                                        {t("globalSetting.config.haproxyConfigDirDescription")}
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Ruta del ejecutable de HAProxy */}
                        <FormField
                            control={form.control}
                            name="haproxy.haproxyBin"
                            render={({ field }) => (
                                <FormItem>
                                    <div className="flex items-center gap-2 mb-2">
                                        <Terminal className="h-4 w-4 text-muted-foreground dark:text-shadow-primary" />
                                        <FormLabel className="dark:text-shadow-glow-white">{t("globalSetting.config.haproxyBinPath")}</FormLabel>
                                    </div>
                                    <FormControl>
                                        <Input
                                            {...field}
                                            className="border-0 border-b border-gray-300 rounded-none focus:ring-0 focus-visible:ring-0 focus-visible:border-primary px-0 dark:text-shadow-glow-white"
                                        />
                                    </FormControl>
                                    <FormDescription className="dark:text-shadow-primary-bold">
                                        {t("globalSetting.config.haproxyBinPathDescription")}
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Internacionalización */}
                        <div className="flex flex-row items-center justify-between p-0 mt-6">
                            <div className="space-y-0.5">
                                <div className="flex items-center gap-2">
                                    <Globe className="h-4 w-4 text-muted-foreground dark:text-shadow-glow-white" />
                                    <div className="text-base font-medium dark:text-shadow-glow-white">{t("globalSetting.config.internationalization")}</div>
                                </div>
                                <div className="text-sm text-muted-foreground dark:text-shadow-glow-white">
                                    {t("globalSetting.config.internationalizationDescription")}
                                </div>
                            </div>
                            <LanguageSelector />
                        </div>

                        {/* Cambio de tema */}
                        <div className="flex flex-row items-center justify-between p-0 mt-6">
                            <div className="space-y-0.5">
                                <div className="flex items-center gap-2">
                                    <Palette className="h-4 w-4 text-muted-foreground dark:text-shadow-primary" />
                                    <div className="text-base font-medium dark:text-shadow-glow-white">{t("globalSetting.config.theme") || "Configuración del tema"}</div>
                                </div>
                                <div className="text-sm text-muted-foreground dark:text-shadow-primary-bold">
                                    {t("globalSetting.config.themeDescription") || "Alternar entre modo claro y oscuro"}
                                </div>
                            </div>
                            <ThemeToggle />
                        </div>

                        {error && (
                            <AdvancedErrorDisplay error={error} />
                        )}
                    </form>
                </Form>
            </CardContent>
            <CardFooter className="flex justify-end  px-0 py-4">
                <AnimatedButton >
                    <Button
                        type="submit"
                        form="config-form"
                        disabled={isLoading || isUpdating}
                        className="gap-1 dark:text-shadow-glow-white"
                    >
                        {isUpdating ? (
                            <>
                                <Loader2 className="h-4 w-4 animate-spin dark:text-shadow-primary" />
                                {t("globalSetting.config.saving")}
                            </>
                        ) : (
                            <>
                                <Save className="h-4 w-4 dark:text-shadow-primary" />
                                {t("globalSetting.config.saveSettings")}
                            </>
                        )}
                    </Button>
                </AnimatedButton>
            </CardFooter>
        </Card>
    )
}