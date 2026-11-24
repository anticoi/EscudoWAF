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
import { cn } from "@/lib/utils" // Asegurar que las funciones de utilidad estén importadas; si no, añadirlas manualmente

interface DataTableProps<TData, TValue> {
    table: TableType<TData>
    columns: ColumnDef<TData, TValue>[]
    style?: 'default' | 'border' | 'simple'
    isLoading?: boolean
    loadingRows?: number
    loadingStyle?: 'centered' | 'skeleton'
    fixedHeader?: boolean
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
            <TableCell colSpan={columns.length} className="h-24 text-center">
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
}: DataTableProps<TData, TValue>) {
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

        return table.getRowModel().rows.map((row) => (
            <TableRow
                key={row.id}
                data-state={row.getIsSelected() && "selected"}
            >
                {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                ))}
            </TableRow>
        ))
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

    // Construir la tabla según el estilo elegido y la opción de cabecera fija
    if (fixedHeader) {
        // Usar estructura de layout con cabecera fija
        return (
            <div className={cn(
                "w-full h-full relative",
                style === 'border' && "rounded-md border"
            )}>
                <div className="overflow-auto h-full scrollbar-neon">
                    <table className="w-full caption-bottom text-sm">
                        {renderTableHeader()}
                        <tbody className={cn(
                            // Solo quitar el borde de la última fila cuando isLoading es true y no hay datos
                            (isLoading || !table.getRowModel().rows?.length) ? "[&_tr:last-child]:border-0" : ""
                        )}>
                            {renderTableBody()}
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }

    // Layout de tabla original (sin cabecera fija)
    switch (style) {
        case 'simple':
            return (
                <div className="w-full h-full scrollbar-neon">
                    <Table>
                        {renderTableHeader()}
                        <TableBody>
                            {renderTableBody()}
                        </TableBody>
                    </Table>
                </div>
            )

        case 'border':
            return (
                <div className="rounded-md border h-full w-full scrollbar-neon">
                    <Table>
                        {renderTableHeader()}
                        <TableBody>
                            {renderTableBody()}
                        </TableBody>
                    </Table>
                </div>
            )

        default:
            return (
                <div className="w-full h-full scrollbar-neon">
                    <Table>
                        {renderTableHeader()}
                        <TableBody className={isLoading || !table.getRowModel().rows?.length ? "[&_tr:last-child]:!border-b-0" : ""}>
                            {renderTableBody()}
                        </TableBody>
                    </Table>
                </div>
            )
    }
}