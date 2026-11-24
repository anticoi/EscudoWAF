import { ReactNode, useRef, useEffect, useState } from "react"
import { motion, useScroll, useTransform } from "motion/react"
import { containerAnimation } from "../container-animation"

interface TableScrollContainerProps {
    children: ReactNode
    className?: string
    showShadows?: boolean
    shadowOpacity?: number
}

export function TableScrollContainer({
    children,
    className = "",
    showShadows = true,
    shadowOpacity = 0.15
}: TableScrollContainerProps) {
    const containerRef = useRef<HTMLDivElement>(null)
    const { scrollYProgress } = useScroll({ container: containerRef })
    const [canScroll, setCanScroll] = useState(false)

    // Opacidad de sombras superior e inferior
    const topShadowOpacity = useTransform(scrollYProgress, [0, 0.1], [0, shadowOpacity])
    const bottomShadowOpacity = useTransform(scrollYProgress, [0.9, 1], [shadowOpacity, 0])

    // Comprobar si el contenido es desplazable
    useEffect(() => {
        const checkScrollability = () => {
            if (containerRef.current) {
                const { scrollHeight, clientHeight } = containerRef.current
                setCanScroll(scrollHeight > clientHeight)
            }
        }

        checkScrollability()

        // Añadir listener de cambios de tamaño de ventana
        window.addEventListener('resize', checkScrollability)
        return () => window.removeEventListener('resize', checkScrollability)
    }, [])

    return (
        <div className={`relative w-full h-full ${className}`}>
            {/* Sombra de scroll superior */}
            {showShadows && canScroll && (
                <motion.div
                    className="absolute top-0 left-0 right-0 h-4 pointer-events-none z-10"
                    style={{
                        opacity: topShadowOpacity,
                        background: `linear-gradient(to bottom, rgba(0,0,0,${shadowOpacity}), transparent)`
                    }}
                />
            )}

            {/* Contenedor de scroll */}
            <motion.div
                ref={containerRef}
                className="overflow-auto h-full w-full scroll-smooth"
                style={{
                    // Añadir efecto de scroll suave
                    scrollBehavior: 'smooth',
                    // Scroll en móvil optimizado
                    WebkitOverflowScrolling: 'touch'
                }}
                {...containerAnimation}
            >
                {children}
            </motion.div>

            {/* Sombra de scroll inferior */}
            {showShadows && canScroll && (
                <motion.div
                    className="absolute bottom-0 left-0 right-0 h-4 pointer-events-none z-10"
                    style={{
                        opacity: bottomShadowOpacity,
                        background: `linear-gradient(to top, rgba(0,0,0,${shadowOpacity}), transparent)`
                    }}
                />
            )}
        </div>
    )
} 