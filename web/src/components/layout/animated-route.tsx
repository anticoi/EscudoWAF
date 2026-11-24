import { ReactNode } from "react"
import { AnimatePresence, motion, useReducedMotion } from "motion/react"
import { useLocation } from "react-router"
import { 
    pageSlideTransition, 
    pageFadeTransition, 
    pageScaleTransition, 
    pageFlipTransition 
} from "@/components/ui/animation/route-animation"

interface AnimatedRouteProps {
    children: ReactNode
    transitionType?: "slide" | "fade" | "scale" | "flip"
}

export function AnimatedRoute({ 
    children, 
    transitionType = "slide"
}: AnimatedRouteProps) {
    const location = useLocation()
    const prefersReducedMotion = useReducedMotion()
    
    // Si el usuario activó reducir animaciones, usar fundido simple
    if (prefersReducedMotion) {
        return (
            <AnimatePresence mode="wait">
                <motion.div
                    key={location.pathname}
                    {...pageFadeTransition}
                >
                    {children}
                </motion.div>
            </AnimatePresence>
        )
    }
    
    // Elegir distintos efectos de animación según el tipo recibido
    const getTransitionProps = () => {
        switch (transitionType) {
            case "fade":
                return pageFadeTransition
            case "scale":
                return pageScaleTransition
            case "flip":
                return pageFlipTransition
            case "slide":
            default:
                return pageSlideTransition
        }
    }

    return (
        <AnimatePresence mode="wait">
            <motion.div
                key={location.pathname}
                {...getTransitionProps()}
            >
                {children}
            </motion.div>
        </AnimatePresence>
    )
} 