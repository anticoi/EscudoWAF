import { useEffect, useState, useRef } from 'react'

interface ResizeObserverEntry {
    contentRect: DOMRectReadOnly
}

interface Size {
    width: number | undefined
    height: number | undefined
}

export function useResizeObserver(
    ref: React.RefObject<HTMLElement>
): Size {
    const [size, setSize] = useState<Size>({
        width: undefined,
        height: undefined,
    })

    // Referencia a la instancia de ResizeObserver
    const observerRef = useRef<ResizeObserver | null>(null)

    useEffect(() => {
        // Si el elemento DOM no existe, retornar directamente
        if (!ref.current) return

        // Limpiar el observer anterior
        if (observerRef.current) {
            observerRef.current.disconnect()
        }

        // Crear un nuevo ResizeObserver
        const observer = new ResizeObserver((entries: ResizeObserverEntry[]) => {
            const { width, height } = entries[0].contentRect
            setSize({ width, height })
        })

        // Empezar a observar el elemento objetivo
        observer.observe(ref.current)
        observerRef.current = observer

        // Función de limpieza
        return () => {
            if (observerRef.current) {
                observerRef.current.disconnect()
                observerRef.current = null
            }
        }
    }, [ref])

    return size
}