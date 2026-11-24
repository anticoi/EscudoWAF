import { useState, useRef, useEffect, useMemo } from 'react'
import {
    useReactTable,
    getCoreRowModel,
    ColumnDef,
} from '@tanstack/react-table'
import { useInfiniteQuery } from '@tanstack/react-query'
import { certificatesApi } from '@/api/certificate'
import { Certificate } from '@/types/certificate'
import { Button } from '@/components/ui/button'
import {
    MoreHorizontal, Plus, Trash2, RefreshCcw, Pencil
} from 'lucide-react'
import {
    DropdownMenu, DropdownMenuContent,
    DropdownMenuItem, DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Card } from '@/components/ui/card'
import { CertificateDialog } from './CertificateDialog'
import { Loader2 } from 'lucide-react'
import { DataTable } from '@/components/table/motion-data-table'
import { DeleteCertificateDialog } from './DeleteCertificateDialog'
import { useTranslation } from 'react-i18next'
import { AnimatedIcon } from '@/components/ui/animation/components/animated-icon'
import { AnimatedButton } from '@/components/ui/animation/components/animated-button'

export function CertificateTable() {
    const { t } = useTranslation()

    // Gestión de estado
    const [certificateDialogOpen, setCertificateDialogOpen] = useState(false)
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
    const [selectedCertId, setSelectedCertId] = useState<string | null>(null)
    const sentinelRef = useRef<HTMLDivElement>(null)
    const [dialogMode, setDialogMode] = useState<'create' | 'update'>('create')
    const [selectedCertificate, setSelectedCertificate] = useState<Certificate | null>(null)

    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)

    // Obtener lista de certificados
    const {
        data,
        isLoading,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        refetch
    } = useInfiniteQuery({
        queryKey: ['certificates'],
        queryFn: ({ pageParam }) => certificatesApi.getCertificates(pageParam as number, 20),
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            if (!lastPage || typeof lastPage.total === 'undefined') {
                return undefined
            }

            // Comprobar si allPages existe
            if (!allPages) {
                return undefined
            }

            // Calcular el total de elementos cargados (acumular la longitud real de items de cada página)
            const loadedItemsCount = allPages.reduce((count, page) => {
                return count + (page.items?.length || 0)
            }, 0)

            // Si el total devuelto por el servidor es mayor que lo cargado, hay otra página
            return lastPage.total > loadedItemsCount ? allPages.length + 1 : undefined
        },
        enabled: true,
    })

    // Datos de paginación aplanados
    const flatData = useMemo(() =>
        data?.pages.flatMap(page => page.items) || [],
        [data]
    )

    // Implementación optimizada de scroll infinito
    useEffect(() => {
        // Solo crear el observador cuando hay más páginas por cargar
        if (!hasNextPage) return

        const options = {
            // Reducir el umbral para que se dispare más fácilmente
            threshold: 0.1,
            // Reducir rootMargin para evitar activación prematura, manteniendo algo de precarga
            rootMargin: '100px 0px'
        }

        const handleObserver = (entries: IntersectionObserverEntry[]) => {
            const [entry] = entries
            if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
                fetchNextPage()
            }
        }

        const observer = new IntersectionObserver(handleObserver, options)

        const sentinel = sentinelRef.current
        if (sentinel) {
            observer.observe(sentinel)
        }

        return () => {
            if (sentinel) {
                observer.unobserve(sentinel)
            }
            observer.disconnect()
        }
    }, [hasNextPage, isFetchingNextPage, fetchNextPage])

    // Abrir el diálogo de creación de certificado
    const openCreateDialog = () => {
        setDialogMode('create')
        setSelectedCertificate(null)
        setCertificateDialogOpen(true)
    }

    // Abrir el diálogo de actualización de certificado
    const openUpdateDialog = (certificate: Certificate) => {
        setDialogMode('update')
        setSelectedCertificate(certificate)
        setCertificateDialogOpen(true)
    }

    // Abrir el diálogo de eliminación
    const openDeleteDialog = (id: string) => {
        setSelectedCertId(id)
        setDeleteDialogOpen(true)
    }

    // Función auxiliar: comprobar si el certificado expira pronto (30 días)
    const isExpiringSoon = (expireDate: string): boolean => {
        const expiryDate = new Date(expireDate)
        const now = new Date()
        const thirtyDaysInMs = 30 * 24 * 60 * 60 * 1000
        return expiryDate.getTime() - now.getTime() < thirtyDaysInMs && expiryDate > now
    }

    // Función auxiliar: comprobar si el certificado ya expiró
    const isExpired = (expireDate: string): boolean => {
        return new Date(expireDate) < new Date()
    }

    const refreshSites = () => {
        setIsRefreshAnimating(true)
        refetch()

        // Detener la animación con 1 segundo de retraso para coincidir con el efecto
        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }

    // Definición de columnas de tabla
    const columns: ColumnDef<Certificate>[] = [
        {
            accessorKey: 'name',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("certificate.name")}</div>,
            cell: ({ row }) => <div className="font-medium dark:text-shadow-glow-white">{row.original.name}</div>,
        },
        {
            accessorKey: 'domains',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("certificate.domains")}</div>,
            cell: ({ row }) => (
                <div className="flex flex-wrap gap-1 max-w-xs">
                    {row.original.domains.map((domain, index) => (
                        <span key={index} className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs dark:text-shadow-glow-white">
                            {domain}
                        </span>
                    ))}
                </div>
            ),
        },
        {
            accessorKey: 'issuerAndExpiry',
            header: () => <div className="font-medium py-3.5 whitespace-nowrap dark:text-shadow-glow-white dark:text-white">{t("certificate.issuerAndExpiry")}</div>,
            cell: ({ row }) => (
                <div className="flex flex-col">
                    <span className="text-sm dark:text-shadow-glow-white">{row.original.issuerName}</span>
                    <span className="text-xs text-muted-foreground dark:text-shadow-glow-white">
                        {t("certificate.expiryDate")}{new Date(row.original.expireDate).toLocaleDateString()}
                        {isExpiringSoon(row.original.expireDate) && (
                            <span className="ml-2 px-1.5 py-0.5 bg-yellow-100 text-yellow-800 rounded text-xs dark:text-shadow-glow-white">
                                {t("certificate.expiringSoon")}
                            </span>
                        )}
                        {isExpired(row.original.expireDate) && (
                            <span className="ml-2 px-1.5 py-0.5 bg-red-100 text-red-800 rounded text-xs dark:text-shadow-glow-white">
                                {t("certificate.expired")}
                            </span>
                        )}
                    </span>
                </div>
            ),
        },
        {
            id: 'actions',
            cell: ({ row }) => (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="dark:text-shadow-glow-white">
                            <MoreHorizontal className="h-4 w-4 dark:text-shadow-glow-white" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuItem
                            onClick={() => openUpdateDialog(row.original)}
                        >
                            <Pencil className="mr-2 h-4 w-4 dark:text-shadow-glow-white" />
                            {t("certificate.edit")}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                            onClick={() => openDeleteDialog(row.original.id)}
                            className="text-red-600 dark:text-red-400 dark:text-shadow-glow-white"
                        >
                            <Trash2 className="mr-2 h-4 w-4" />
                            {t("certificate.delete")}
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            ),
        },
    ]

    // Inicializar la tabla
    const table = useReactTable({
        data: flatData,
        columns,
        getCoreRowModel: getCoreRowModel(),
    })

    return (
        <>
            <Card className="border-none shadow-none p-6 flex flex-col h-full">
                {/* Título y botones de acción - fijados arriba */}
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-semibold text-primary dark:text-white">{t("certificate.management")}</h2>
                    <div className="flex gap-2">
                        <AnimatedButton >
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={refreshSites}
                                className="flex items-center gap-2 justify-center dark:text-shadow-glow-white"
                            >
                                <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                    <RefreshCcw className="h-4 w-4" />
                                </AnimatedIcon>
                                {t("certificate.refresh")}
                            </Button>
                        </AnimatedButton>
                        <AnimatedButton>
                            <Button
                                size="sm"
                                onClick={openCreateDialog}
                                className="flex items-center gap-1 dark:text-shadow-glow-white"
                            >
                                <Plus className="h-3.5 w-3.5 dark:text-shadow-glow-white" />
                                {t("certificate.add")}
                            </Button>
                        </AnimatedButton>
                    </div>
                </div>

                {/* Contenedor de tabla - altura fija y scroll */}
                <div className="flex-1 overflow-hidden flex flex-col">
                    <div className="overflow-auto h-full">
                        <DataTable table={table}
                            loadingStyle="skeleton"
                            columns={columns}
                            isLoading={isLoading}
                            fixedHeader={true}
                            animatedRows={true}
                            showScrollShadows={true}
                        />

                        {/* Elemento de detección de scroll infinito - dentro del área de scroll */}
                        {hasNextPage && <div
                            ref={sentinelRef}
                            className="h-5 flex justify-center items-center mt-4"
                        >
                            {isFetchingNextPage && (
                                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                            )}
                        </div>}
                    </div>
                </div>
            </Card>

            {/* Diálogo unificado de certificado */}
            <CertificateDialog
                open={certificateDialogOpen}
                onOpenChange={setCertificateDialogOpen}
                mode={dialogMode}
                certificate={selectedCertificate}
            />

            {/* Usar el componente de diálogo de eliminación extraído */}
            <DeleteCertificateDialog
                open={deleteDialogOpen}
                onOpenChange={setDeleteDialogOpen}
                certificateId={selectedCertId}
                onDeleted={() => setSelectedCertId(null)}
            />
        </>
    )
}