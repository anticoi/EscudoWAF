import { useState, useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { Upload, FileText, X, AlertCircle, Info } from 'lucide-react'
import { certificateFormSchema } from '@/validation/certificate'
import { parseCertificate, readFileAsText } from '@/utils/certificate-parser'
import { CertificateCreateRequest, ParsedCertificate } from '@/types/certificate'
import { useCreateCertificate, useUpdateCertificate } from '../hooks/useCertificate'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'
interface CertificateFormProps {
    mode?: 'create' | 'update'
    certificateId?: string
    onSuccess?: () => void
    defaultValues?: Partial<CertificateCreateRequest>
}

export function CertificateForm({
    mode = 'create',
    certificateId,
    onSuccess,
    defaultValues = {
        name: '',
        description: '',
        publicKey: '',
        privateKey: '',
    },
}: CertificateFormProps) {
    const { t } = useTranslation()

    // Gestión de estado
    const [parsedInfo, setParsedInfo] = useState<ParsedCertificate | null>(null)
    const [publicKeyFile, setPublicKeyFile] = useState<string | null>(null)
    const [privateKeyFile, setPrivateKeyFile] = useState<string | null>(null)
    const [parseError, setParseError] = useState<string | null>(null)

    // Hooks de la API
    const {
        createCertificate,
        isLoading: isCreating,
        error: createError,
        clearError: clearCreateError
    } = useCreateCertificate()

    const {
        updateCertificate,
        isLoading: isUpdating,
        error: updateError,
        clearError: clearUpdateError
    } = useUpdateCertificate()

    // Estado dinámico
    const isLoading = mode === 'create' ? isCreating : isUpdating
    const error = mode === 'create' ? createError : updateError
    const clearError = mode === 'create' ? clearCreateError : clearUpdateError

    // Configuración del formulario
    const form = useForm<CertificateCreateRequest>({
        resolver: zodResolver(certificateFormSchema),
        defaultValues,
    })

    // Intentar analizar el contenido del certificado
    const tryParseCertificate = useCallback((content: string) => {
        if (!content) {
            setParsedInfo(null)
            setParseError(null)
            return
        }

        try {
            const parsed = parseCertificate(content)
            setParsedInfo(parsed)
            setParseError(null)

            // Si hubo errores de análisis previos, limpiar los errores del formulario
            form.clearErrors('publicKey')
        } catch (error) {
            console.error('Error de análisis del certificado:', error)

            // Limpiar la información ya analizada
            setParsedInfo(null)

            // Establecer el mensaje de error sin bloquear el envío del formulario
            if (error instanceof Error) {
                setParseError(`${t("certificate.dialog.parseFailed")}${error.message}`)
            } else {
                setParseError(`${t("certificate.dialog.parseFailed")}${t("certificate.dialog.unknownError")}`)
            }
        }
    }, [form, t])

    // Manejar la subida del archivo de clave pública
    const handlePublicKeyFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files.length > 0) {
            const file = e.target.files[0]

            try {
                const content = await readFileAsText(file)

                // Actualizar el formulario y el estado de la UI
                form.setValue('publicKey', content)
                setPublicKeyFile(file.name)

                // Intentar analizar el certificado
                tryParseCertificate(content)
            } catch (error) {
                console.error('Error de lectura de archivo:', error)

                // Los errores de lectura del archivo se muestran en el formulario
                const errorMessage = error instanceof Error ? error.message : t("certificate.dialog.unknownError")
                form.setError('publicKey', {
                    type: 'manual',
                    message: `${t("certificate.dialog.fileReadFailed")}${errorMessage}`
                })
            }
        }
    }, [form, tryParseCertificate, t])

    // Manejar la subida del archivo de clave privada
    const handlePrivateKeyFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files.length > 0) {
            const file = e.target.files[0]

            try {
                const content = await readFileAsText(file)

                // Actualizar el formulario y el estado de la UI
                form.setValue('privateKey', content)
                setPrivateKeyFile(file.name)
            } catch (error) {
                console.error('Error de lectura de archivo:', error)

                // Los errores de lectura del archivo se muestran en el formulario
                const errorMessage = error instanceof Error ? error.message : t("certificate.dialog.unknownError")
                form.setError('privateKey', {
                    type: 'manual',
                    message: `${t("certificate.dialog.fileReadFailed")}${errorMessage}`
                })
            }
        }
    }, [form, t])

    // Borrar el archivo de clave pública
    const clearPublicKeyFile = useCallback(() => {
        setPublicKeyFile(null)
        form.setValue('publicKey', '')
        setParsedInfo(null)
        setParseError(null)
    }, [form])

    // Borrar el archivo de clave privada
    const clearPrivateKeyFile = useCallback(() => {
        setPrivateKeyFile(null)
        form.setValue('privateKey', '')
    }, [form])

    // Manejar el cambio del cuadro de texto de la clave pública
    const handlePublicKeyTextChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const value = e.target.value
        tryParseCertificate(value)
    }, [tryParseCertificate])

    // Manejo del envío del formulario
    const handleFormSubmit = useCallback((data: CertificateCreateRequest) => {
        // Limpiar los errores anteriores
        if (clearError) clearError()

        // Fusionar la información del certificado analizada
        const finalData = {
            ...data,
            ...(parsedInfo || {}),
        }

        // Ejecutar creación o actualización según el modo
        if (mode === 'create') {
            createCertificate(finalData, {
                onSuccess: () => {
                    // Restablecer el formulario
                    form.reset()
                    setPublicKeyFile(null)
                    setPrivateKeyFile(null)
                    setParsedInfo(null)
                    setParseError(null)
                    // Notificar éxito al componente padre
                    if (onSuccess) onSuccess()
                }
            })
        } else if (mode === 'update' && certificateId) {
            updateCertificate({ id: certificateId, data: finalData }, {
                onSuccess: () => {
                    // Notificar éxito al componente padre
                    if (onSuccess) onSuccess()
                }
            })
        }
    }, [mode, certificateId, clearError, parsedInfo, createCertificate, updateCertificate, form, onSuccess])

    // Renderizar la información del certificado analizada
    const renderParsedInfo = useCallback(() => {
        if (!parsedInfo) return null

        return (
            <div className="p-4 border rounded-md bg-zinc-50 dark:bg-gray-800/10 dark:border-gray-700">
                <h3 className="text-sm font-medium mb-2">{t("certificate.dialog.parsedInfo")}</h3>
                <div className="space-y-2 text-sm">
                    <InfoRow label={t("certificate.dialog.issuer")} value={parsedInfo.issuerName} />
                    <InfoRow
                        label={t("certificate.dialog.expiryDate")}
                        value={new Date(parsedInfo.expireDate).toLocaleDateString()}
                    />
                    <InfoRow label={t("certificate.dialog.fingerprint")} value={parsedInfo.fingerPrint} className="font-mono text-xs" />
                    <div className="flex">
                        <span className="w-24 text-muted-foreground">{t("certificate.dialog.domains")}:</span>
                        <div className="flex flex-wrap gap-1">
                            {parsedInfo.domains.map((domain, index) => (
                                <span key={index} className="px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs dark:text-shadow-glow-white">
                                    {domain}
                                </span>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        )
    }, [parsedInfo, t])

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

                    {/* Aviso de error de análisis de certificado */}
                    {parseError && (
                        <Alert variant="default" className="bg-yellow-50 border-yellow-200">
                            <Info className="h-4 w-4 text-yellow-800" />
                            <AlertDescription className="text-yellow-800">{parseError}</AlertDescription>
                        </Alert>
                    )}

                    {/* Campos de información básica */}
                    <FormField
                        control={form.control}
                        name="name"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.certName")}</FormLabel>
                                <FormControl>
                                    <Input className="dark:text-shadow-glow-white" placeholder={t("certificate.dialog.certName")} {...field} />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="description"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.description")}</FormLabel>
                                <FormControl>
                                    <Textarea
                                        placeholder={t("certificate.dialog.descriptionPlaceholder")}
                                        className="resize-none dark:text-shadow-glow-white"
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    {/* Subida del archivo de clave pública */}
                    <div className="space-y-2">
                        <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.publicKeyFile")}</FormLabel>
                        {publicKeyFile ? (
                            <FilePreview
                                filename={publicKeyFile}
                                onClear={clearPublicKeyFile}
                            />
                        ) : (
                            <FileUpload
                                label={t("certificate.dialog.uploadPublicKey")}
                                accept=".pem,.crt,.cert,.key"
                                onChange={handlePublicKeyFileChange}
                            />
                        )}
                    </div>

                    {/* Contenido de la clave pública */}
                    <FormField
                        control={form.control}
                        name="publicKey"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.publicKeyContent")}</FormLabel>
                                <FormControl>
                                    <Textarea
                                        placeholder={t("certificate.dialog.publicKeyPlaceholder")}
                                        className="font-mono text-xs h-32 dark:text-shadow-glow-white"
                                        {...field}
                                        onChange={(e) => {
                                            field.onChange(e)
                                            handlePublicKeyTextChange(e)
                                        }}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    {/* Subida del archivo de clave privada */}
                    <div className="space-y-2">
                        <FormLabel className='dark:text-shadow-glow-white'>{t("certificate.dialog.privateKeyFile")}</FormLabel>
                        {privateKeyFile ? (
                            <FilePreview
                                filename={privateKeyFile}
                                onClear={clearPrivateKeyFile}
                            />
                        ) : (
                            <FileUpload
                                label={t("certificate.dialog.uploadPrivateKey")}
                                accept=".pem,.key"
                                onChange={handlePrivateKeyFileChange}
                            />
                        )}
                    </div>

                    {/* Contenido de la clave privada */}
                    <FormField
                        control={form.control}
                        name="privateKey"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel className="dark:text-shadow-glow-white">{t("certificate.dialog.privateKeyContent")}</FormLabel>
                                <FormControl>
                                    <Textarea
                                        placeholder={t("certificate.dialog.privateKeyPlaceholder")}
                                        className="font-mono text-xs h-32 dark:text-shadow-glow-white"
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    {/* Información del certificado analizado */}
                    {renderParsedInfo()}

                    {/* Botón de envío */}
                    <div className="flex justify-end">
                        <AnimatedButton>
                            <Button type="submit" disabled={isLoading}>
                                {isLoading ? t("certificate.dialog.submitting") : mode === 'create' ? t("certificate.dialog.create") : t("certificate.dialog.update")}
                            </Button>
                        </AnimatedButton>
                    </div>
                </form>
            </Form>
        </AnimatedContainer >
    )
}

// Componente auxiliar: fila de información
interface InfoRowProps {
    label: string
    value: string
    className?: string
}

export function InfoRow({ label, value, className = '' }: InfoRowProps) {
    return (
        <div className="flex">
            <span className="w-24 flex-shrink-0 text-muted-foreground">{label}:</span>
            <span className={`break-all ${className}`}>{value}</span>
        </div>
    )
}

// Componente auxiliar: vista previa de archivos
interface FilePreviewProps {
    filename: string
    onClear: () => void
}

function FilePreview({ filename, onClear }: FilePreviewProps) {
    return (
        <div className="flex items-center gap-2 p-2 border rounded">
            <FileText className="h-4 w-4" />
            <span className="text-sm flex-1 truncate">{filename}</span>
            <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={onClear}
            >
                <X className="h-4 w-4" />
            </Button>
        </div>
    )
}

// Componente auxiliar: subida de archivos
interface FileUploadProps {
    label: string
    accept: string
    onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
}

function FileUpload({ label, accept, onChange }: FileUploadProps) {
    const { t } = useTranslation()

    return (
        <div className="flex items-center gap-2">
            <Button
                type="button"
                variant="outline"
                size="sm"
                asChild
            >
                <label className="cursor-pointer flex items-center gap-2">
                    <Upload className="h-4 w-4" />
                    <span>{label}</span>
                    <input
                        type="file"
                        className="hidden"
                        accept={accept}
                        onChange={onChange}
                    />
                </label>
            </Button>
            <span className="text-sm text-muted-foreground">{t("certificate.dialog.orEnterContent")}</span>
        </div>
    )
}