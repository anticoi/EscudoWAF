import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import { flexRender, Table as TableType, ColumnDef } from "@tanstack/react-table"
import { Loader2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import { cn } from "@/lib/utils"
import { TableScrollContainer } from "@/components/ui/animation/components/table-scroll"
import { useEffect, useRef, useState } from "react"

interface DataTableProps<TData, TValue> {
    table: TableType<TData>
    columns: ColumnDef<TData, TValue>[]
    style?: 'default' | 'border' | 'simple'
    isLoading?: boolean
    loadingRows?: number
    loadingStyle?: 'centered' | 'skeleton'
    fixedHeader?: boolean
    animatedRows?: boolean
    showScrollShadows?: boolean
}

// Componente skeleton de estado de carga
const TableSkeleton = <TData, TValue>({
    columns,
    rows = 5
}: {
    columns: ColumnDef<TData, TValue>[],
    rows?: number
}) => (
    <>
        {Array.from({ length: rows }).map((_, index) => (
            <TableRow key={`skeleton-row-${index}`} className="animate-pulse">
                {Array.from({ length: columns.length }).map((_, cellIndex) => (
                    <TableCell key={`skeleton-cell-${index}-${cellIndex}`}>
                        <div className="h-4 bg-gray-200 rounded w-3/4 dark:bg-gray-700"></div>
                    </TableCell>
                ))}
            </TableRow>
        ))}
    </>
)

// Componente de animación de carga centrado
const CenteredLoader = <TData, TValue>({
    columns
}: {
    columns: ColumnDef<TData, TValue>[]
}) => (
    <TableRow>
        <TableCell colSpan={columns.length} className="h-24">
            <div className="flex items-center justify-center space-x-2">
                <Loader2 className="h-6 w-6 animate-spin text-primary" />
                <span className="text-sm text-muted-foreground">Cargando...</span>
            </div>
        </TableCell>
    </TableRow>
)

// Componente de estado sin datos
const NoResults = <TData, TValue>({
    columns
}: {
    columns: ColumnDef<TData, TValue>[]
}) => {
    const { t } = useTranslation()
    return (
        <TableRow>
            <TableCell colSpan={columns.length} className="h-24 text-center dark:text-shadow-glow-white">
                {t('noResult')}
            </TableCell>
        </TableRow>
    )
}

export function DataTable<TData, TValue>({
    table,
    columns,
    style = 'default',
    isLoading = false,
    loadingRows = 5,
    loadingStyle = 'centered',
    fixedHeader = false,
    animatedRows = false,
    showScrollShadows = true,
}: DataTableProps<TData, TValue>) {
    // Ref para rastrear las filas renderizadas
    const rowsRef = useRef<HTMLElement[]>([])
    const [inViewRows, setInViewRows] = useState<Set<string>>(new Set())

    // Extraer expresiones complejas a variables
    const tableRows = table.getRowModel().rows

    // Escuchar el scroll para detectar qué filas están en el viewport
    useEffect(() => {
        if (!animatedRows) return

        // Vaciar el array de referencias solo cuando cambian los datos de la tabla, sin reasignar
        if (rowsRef.current.length > tableRows.length) {
            rowsRef.current.length = tableRows.length
        }

        const observer = new IntersectionObserver(
            (entries) => {
                entries.forEach(entry => {
                    const rowId = entry.target.getAttribute('data-row-id')
                    if (rowId) {
                        setInViewRows(prev => {
                            const updated = new Set(prev)
                            if (entry.isIntersecting) {
                                updated.add(rowId)
                            }
                            return updated
                        })
                    }
                })
            },
            { threshold: 0.1 }
        )

        // Observar directamente las filas existentes
        rowsRef.current.forEach(row => {
            if (row) observer.observe(row)
        })

        return () => {
            observer.disconnect()
        }
    }, [tableRows, animatedRows])

    // Renderizar el contenido del cuerpo de la tabla
    const renderTableBody = () => {
        if (isLoading) {
            // Según la opción loadingStyle, decidir qué animación de carga usar
            return loadingStyle === 'centered'
                ? <CenteredLoader<TData, TValue> columns={columns} />
                : <TableSkeleton<TData, TValue> columns={columns} rows={loadingRows} />
        }

        if (!table.getRowModel().rows?.length) {
            return <NoResults<TData, TValue> columns={columns} />
        }

        return table.getRowModel().rows.map((row, rowIndex) => {
            const isInView = inViewRows.has(row.id) || !animatedRows

            return (
                <TableRow
                    key={row.id}
                    data-row-id={row.id}
                    data-state={row.getIsSelected() && "selected"}
                    ref={el => {
                        if (el && animatedRows) {
                            rowsRef.current[rowIndex] = el
                        }
                    }}
                    className={cn(
                        animatedRows && "transition-opacity duration-300",
                        animatedRows && !isInView ? "opacity-0" : "opacity-100"
                    )}
                    style={{
                        transitionDelay: animatedRows ? `${rowIndex * 30}ms` : '0ms'
                    }}
                >
                    {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                    ))}
                </TableRow>
            )
        })
    }

    // Función de renderizado de cabecera
    const renderTableHeader = () => (
        <TableHeader className={fixedHeader ? "sticky top-0 z-10 bg-background border-b" : ""}>
            {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                        <TableHead key={header.id}>
                            {header.isPlaceholder
                                ? null
                                : flexRender(
                                    header.column.columnDef.header,
                                    header.getContext()
                                )}
                        </TableHead>
                    ))}
                </TableRow>
            ))}
        </TableHeader>
    )

    // Usar el contenedor de scroll
    const renderScrollableTable = () => (
        <TableScrollContainer showShadows={showScrollShadows} className="scrollbar-neon">
            <table className="w-full caption-bottom text-sm">
                {renderTableHeader()}
                <tbody className={cn(
                    // Solo quitar el borde de la última fila cuando isLoading es true y no hay datos
                    (isLoading || !table.getRowModel().rows?.length) ? "[&_tr:last-child]:border-0" : ""
                )}>
                    {renderTableBody()}
                </tbody>
            </table>
        </TableScrollContainer>
    )

    // Construir la tabla según el estilo elegido y la opción de cabecera fija
    if (fixedHeader) {
        // Usar estructura de layout con cabecera fija
        return (
            <div className={cn(
                "w-full h-full relative",
                style === 'border' && "rounded-md border"
            )}>
                {renderScrollableTable()}
            </div>
        )
    }

    // Layout de tabla original (sin cabecera fija)
    switch (style) {
        case 'simple':
            return (
                <div className="w-full h-full">
                    <TableScrollContainer showShadows={showScrollShadows} className="scrollbar-neon">
                        <Table>
                            {renderTableHeader()}
                            <TableBody>
                                {renderTableBody()}
                            </TableBody>
                        </Table>
                    </TableScrollContainer>
                </div>
            )

        case 'border':
            return (
                <div className="rounded-md border h-full w-full">
                    <TableScrollContainer showShadows={showScrollShadows} className="scrollbar-neon">
                        <Table>
                            {renderTableHeader()}
                            <TableBody>
                                {renderTableBody()}
                            </TableBody>
                        </Table>
                    </TableScrollContainer>
                </div>
            )

        default:
            return (
                <div className="w-full h-full">
                    <TableScrollContainer showShadows={showScrollShadows} className="scrollbar-neon">
                        <Table>
                            {renderTableHeader()}
                            <TableBody className={isLoading || !table.getRowModel().rows?.length ? "[&_tr:last-child]:!border-b-0" : ""}>
                                {renderTableBody()}
                            </TableBody>
                        </Table>
                    </TableScrollContainer>
                </div>
            )
    }
}