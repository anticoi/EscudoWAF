// src/components/error-boundary.tsx
import { Component, ReactNode } from 'react'
import { isRouteErrorResponse, useRouteError } from 'react-router'
import { AlertCircle, Home, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { motion } from 'motion/react'

interface Props {
    children?: ReactNode
}

interface State {
    hasError: boolean
    error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = { hasError: false }
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error }
    }

    render() {
        if (this.state.hasError) {
            return <ErrorDisplay error={this.state.error} />
        }

        return this.props.children
    }
}

interface ErrorDisplayProps {
    error?: Error
}

export function ErrorDisplay({ error }: ErrorDisplayProps) {
    const { t } = useTranslation()

    // Si se usa en rutas, intentar obtener el error de ruta
    try {
        const routeError = useRouteError()
        if (isRouteErrorResponse(routeError)) {
            return (
                <div className="flex min-h-screen w-full flex-col items-center justify-center p-6 text-center bg-gradient-to-b from-background to-muted/30">
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.5 }}
                        className="w-full max-w-md p-8 rounded-lg bg-card shadow-lg border border-border"
                    >
                        <motion.div
                            initial={{ scale: 0.8 }}
                            animate={{ scale: 1 }}
                            transition={{ delay: 0.2, type: "spring" }}
                            className="mx-auto bg-red-100 dark:bg-red-900/30 p-3 rounded-full w-16 h-16 flex items-center justify-center mb-6"
                        >
                            <AlertCircle className="h-10 w-10 text-red-500 dark:text-red-400" />
                        </motion.div>
                        <h1 className="text-3xl font-bold mb-3">{t("error.code", "Error")} {routeError.status}</h1>
                        <p className="mb-4 text-muted-foreground text-lg">{routeError.statusText}</p>
                        <p className="text-sm text-muted-foreground mb-6">{routeError.data?.message || t("error.unexpected", "Se produjo un error inesperado")}</p>

                        <div className="flex flex-col sm:flex-row gap-3 justify-center">
                            <Button
                                onClick={() => window.location.reload()}
                                variant="outline"
                                className="gap-2"
                            >
                                <RefreshCw className="h-4 w-4" />
                                {t("error.refresh", "Recargar página")}
                            </Button>
                            <Button
                                onClick={() => window.location.href = '/'}
                                className="gap-2"
                            >
                                <Home className="h-4 w-4" />
                                {t("error.returnHome", "Volver al inicio")}
                            </Button>
                        </div>
                    </motion.div>
                </div>
            )
        }

        if (routeError instanceof Error) {
            error = routeError
        }
    } catch {
        // useRouteError solo puede usarse dentro del contexto de rutas
    }

    return (
        <div className="flex min-h-screen w-full flex-col items-center justify-center p-6 text-center bg-gradient-to-b from-background to-muted/30 dark:bg-background/30">
            <motion.div
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="w-full max-w-md p-8 rounded-lg bg-card shadow-lg border border-border dark:bg-background/10 dark:border-gray-700"
            >
                <motion.div
                    initial={{ scale: 0.8 }}
                    animate={{ scale: 1 }}
                    transition={{ delay: 0.2, type: "spring" }}
                    className="mx-auto bg-red-100 dark:bg-red-900/30 p-3 rounded-full w-16 h-16 flex items-center justify-center mb-6"
                >
                    <AlertCircle className="h-10 w-10 text-red-500 dark:text-red-400" />
                </motion.div>
                <h1 className="text-3xl font-bold mb-3">{t("error.title", "Se produjo un error")}</h1>
                <p className="mb-4 text-muted-foreground text-lg">{error?.message || t("error.appProblem", "La aplicación encontró un problema inesperado")}</p>

                {error?.stack && (
                    <div className="mb-6 mx-auto max-w-md">
                        <div className="bg-muted dark:bg-gray-800/30 p-3 rounded-md text-left overflow-x-auto text-xs font-mono text-muted-foreground dark:text-gray-300 dark:text-shadow-glow-white mb-4">
                            {error.stack.split('\n').slice(0, 3).join('\n')}
                        </div>
                    </div>
                )}

                <div className="flex flex-col sm:flex-row gap-3 justify-center">
                    <Button
                        onClick={() => window.location.reload()}
                        variant="outline"
                        className="gap-2"
                    >
                        <RefreshCw className="h-4 w-4" />
                        {t("error.refresh", "Recargar página")}
                    </Button>
                    <Button
                        onClick={() => window.location.href = '/'}
                        className="gap-2"
                    >
                        <Home className="h-4 w-4" />
                        {t("error.returnHome", "Volver al inicio")}
                    </Button>
                </div>
            </motion.div>
        </div>
    )
}

// Wrapper simple usado como elemento de error de rutas
export function RouteErrorBoundary() {
    return <ErrorDisplay />
}