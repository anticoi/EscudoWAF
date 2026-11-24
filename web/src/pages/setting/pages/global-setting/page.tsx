import { useEffect } from "react"
import { useConfigQuery } from "@/feature/global-setting/hooks/useConfig"
import { useRunnerStatusQuery, useRunnerControl } from "@/feature/global-setting/hooks/useRunner"
import { EngineStatus } from "@/feature/global-setting/components/EngineStatus"
import { ConfigForm } from "@/feature/global-setting/components/ConfigForm"
import { Settings, Info } from "lucide-react"
import { AdvancedErrorDisplay } from "@/components/common/error/errorDisplay"
import { AnimatedContainer } from "@/components/ui/animation/components/animated-container"
import { useTranslation } from "react-i18next"

export default function GlobalSettingPage() {
    const { t } = useTranslation()

    // Obtener los datos de configuración
    const { config, isLoading: isConfigLoading, error: configError, refetch: refetchConfig } = useConfigQuery()

    // Obtener el estado del runner
    const { status, isLoading: isStatusLoading, error: statusError, refetch: refetchStatus } = useRunnerStatusQuery()

    // Control del runner
    const {
        controlRunner,
        isLoading: isControlLoading,
        error: controlError,
        clearError: clearControlError,
    } = useRunnerControl()

    // Obtener la configuración más reciente al cargar la página
    useEffect(() => {
        // Al cargar la página, obtener la configuración y el estado más recientes
        refetchConfig()
        refetchStatus()
    }, [refetchConfig, refetchStatus])

    // Función de manejo del control del runner
    const handleStart = () => {
        clearControlError()
        controlRunner("start")
    }
    const handleStop = () => {
        clearControlError()
        controlRunner("stop")
    }
    const handleRestart = () => {
        clearControlError()
        controlRunner("restart")
    }
    const handleForceStop = () => {
        clearControlError()
        controlRunner("force_stop")
    }
    const handleReload = () => {
        clearControlError()
        controlRunner("reload")
    }

    // Elegir la función de reintento adecuada según el tipo de error
    const handleRetry = () => {
        if (configError) refetchConfig()
        if (statusError) refetchStatus()
    }

    return (
        <AnimatedContainer variant="smooth" className="h-full overflow-y-auto scrollbar-none p-0">
            <div className="max-w-5xl p-6 center mx-auto">
                {/* Manejo de errores: mostrar primero errores de configuración, luego de estado */}
                {(configError || statusError) && (
                    <AdvancedErrorDisplay error={configError || statusError} onRetry={handleRetry} />
                )}
                {controlError && <AdvancedErrorDisplay error={controlError} />}

                {/* Configuración general */}
                <div className="bg-background rounded-xl border-none p-6 mb-6 animate-fade-in-up transition-colors duration-200">
                    <div className="space-y-3 mb-8">
                        <div className="flex items-center gap-2 pb-2 border-b border-border">
                            <Settings className="w-5 h-5 text-iconStroke dark:text-primary dark:text-shadow-glow-white" />
                            <h3 className="text-lg font-medium text-foreground text-shadow-glow-white dark:text-shadow-glow-white">{t("globalSetting.config.generalConfig")}</h3>
                        </div>
                        <div className="pl-7">
                            <p className="text-sm text-muted-foreground text-shadow-glow-white dark:text-shadow-primary-bold">{t("globalSetting.description")}</p>
                        </div>
                    </div>

                    {/* Estado del motor */}
                    <div className="space-y-3 mb-8">
                        <div className="flex items-center gap-2 pb-2 border-b border-border">
                            <Info className="w-5 h-5 text-iconStroke dark:text-primary dark:text-shadow-glow-white" />
                            <h3 className="text-base font-medium text-foreground text-shadow-glow-white dark:text-shadow-glow-white">{t("globalSetting.engine.management")}</h3>
                        </div>
                        <div className="pl-7">
                            <EngineStatus
                                status={status}
                                isLoading={isStatusLoading}
                                onStart={handleStart}
                                onStop={handleStop}
                                onRestart={handleRestart}
                                onForceStop={handleForceStop}
                                onReload={handleReload}
                                isControlLoading={isControlLoading}
                            />
                        </div>
                    </div>
                </div>

                {/* Formulario de configuración */}
                <div
                    className="bg-background rounded-xl border-none p-6 mb-6 animate-fade-in-up"
                    style={{ animationDelay: "0.1s" }}
                >
                    <ConfigForm config={config} isLoading={isConfigLoading} />
                </div>
            </div>
        </AnimatedContainer>
    )
}
