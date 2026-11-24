import { HTMLMotionProps } from "motion/react"

// Animación del contenedor base - con efecto elástico
export const containerAnimation: HTMLMotionProps<"div"> = {
    layout: true,
    transition: {
        layout: {
            type: "spring",
            stiffness: 300,
            damping: 30
        }
    }
}

// Animación de contenedor suave - usa tweening para transiciones más uniformes
export const smoothContainerAnimation: HTMLMotionProps<"div"> = {
    layout: true,
    transition: {
        layout: {
            duration: 0.4,
            ease: [0.4, 0, 0.2, 1],
            type: "tween"
        }
    }
}

// Animación del contenedor de scroll optimizada - menos tiempo de transición, más respuesta
export const scrollContainerAnimation: HTMLMotionProps<"div"> = {
    layout: true,
    layoutRoot: true,
    transition: {
        layout: {
            duration: 0.25,
            ease: [0.25, 0.1, 0.25, 1.0],
            type: "tween"
        }
    }
}

// Animación de raíz de layout de alto rendimiento - para listas y tablas complejas
export const layoutRootAnimation: HTMLMotionProps<"div"> = {
    layout: true,
    layoutRoot: true,
    transition: {
        layout: {
            type: "tween",
            duration: 0.2
        }
    }
} 