import { HTMLMotionProps } from "motion/react"

// Configuración de animación de layout - para animar adición, eliminación y reordenación de elementos
export const layoutAnimationProps = {
    layout: true,
    initial: { opacity: 0, scale: 0.8 },
    animate: { opacity: 1, scale: 1 },
    exit: { opacity: 0, scale: 0.8 },
    transition: {
        layout: { 
            type: "spring", 
            damping: 25, 
            stiffness: 300,
            mass: 0.8
        },
        opacity: { duration: 0.3 },
        scale: { 
            type: "spring", 
            damping: 15, 
            stiffness: 200 
        }
    }
}

// Conservar la configuración de animación de los elementos de la cuadrícula, usada en SiteCard
export const gridItemAnimation: HTMLMotionProps<"div"> = {
    variants: {
        initial: {
            opacity: 0,
            y: 20,
            scale: 0.95
        },
        animate: {
            opacity: 1,
            y: 0,
            scale: 1,
            transition: {
                type: "spring",
                damping: 15,
                stiffness: 200
            }
        }
    }
}