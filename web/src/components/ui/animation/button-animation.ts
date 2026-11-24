// src/components/ui/animation/button-animation.ts
import { HTMLMotionProps } from "motion/react"

// Animación de botón estándar - efecto fluido de escala en secuencia
export const buttonAnimation: HTMLMotionProps<"div"> = {
    whileHover: {
        scale: [null, 1.05, 1.02],
        transition: {
            duration: 0.4,
            times: [0, 0.6, 1],
            ease: ["easeInOut", "easeOut"],
        },
    },
    whileTap: {
        scale: 0.95,
        transition: {
            duration: 0.1,
            ease: "easeIn",
        }
    },
    // Añadir respuesta táctil del trackpad
    transition: {
        duration: 0.3,
        ease: "easeOut",
        type: "spring", // Usar animación de resorte para mejorar la respuesta táctil
        stiffness: 400,
        damping: 17
    }
}

// Animación de botón principal - con efecto de sombra mejorado
export const primaryButtonAnimation: HTMLMotionProps<"div"> = {
    whileHover: {
        scale: [null, 1.08, 1.03],
        boxShadow: "0 4px 12px rgba(0, 0, 0, 0.15)",
        transition: {
            duration: 0.5,
            times: [0, 0.6, 1],
            ease: ["easeInOut", "easeOut"],
        },
    },
    whileTap: {
        scale: 0.95,
        boxShadow: "0 2px 6px rgba(0, 0, 0, 0.1)",
        transition: {
            duration: 0.1,
            ease: "easeIn",
            // Parámetros de optimización del trackpad
            type: "spring",
            stiffness: 600,
            damping: 25
        }
    },
    transition: {
        duration: 0.3,
        ease: "easeOut",
        type: "spring",
        stiffness: 400,
        damping: 17
    }
}

// Animación de botón secundario - efecto más sutil
export const secondaryButtonAnimation: HTMLMotionProps<"div"> = {
    whileHover: {
        scale: [null, 1.04, 1.02],
        transition: {
            duration: 0.3,
            times: [0, 0.5, 1],
            ease: ["easeInOut", "easeOut"],
        },
    },
    whileTap: {
        scale: 0.97,
        transition: {
            duration: 0.1,
            ease: "easeIn",
            // Optimización para respuesta del trackpad
            type: "spring",
            stiffness: 500,
            damping: 20
        }
    },
    transition: {
        duration: 0.3,
        ease: "easeOut",
        type: "spring",
        stiffness: 350,
        damping: 15
    }
}

// Animación de botón ghost - efecto sutil
export const ghostButtonAnimation: HTMLMotionProps<"div"> = {
    whileHover: {
        scale: 1.03,
        transition: {
            duration: 0.3,
            ease: "easeOut",
        },
    },
    whileTap: {
        scale: 0.98,
        transition: {
            duration: 0.1,
            ease: "easeIn",
            // Respuesta táctil más sensible
            type: "spring",
            stiffness: 450,
            damping: 15
        }
    },
    transition: {
        duration: 0.3,
        ease: "easeOut",
        type: "spring",
        stiffness: 300,
        damping: 12
    }
}

// Animación de botón de operación destructiva - efecto de advertencia más enfatizado
export const destructiveButtonAnimation: HTMLMotionProps<"div"> = {
    whileHover: {
        scale: [null, 1.06, 1.03],
        transition: {
            duration: 0.4,
            times: [0, 0.5, 1],
            ease: ["easeInOut", "easeOut"],
        },
    },
    whileTap: {
        // Efecto de desvanecimiento fragmentado
        scale: [0.95, 0.75, 0],
        opacity: [1, 0.7, 0],
        x: [0, 5, -5], // Ligero temblor lateral
        transition: {
            duration: 0.4,
            times: [0, 0.5, 1],
            ease: "easeOut",
        }
    },
    transition: {
        duration: 0.3,
        ease: "easeOut",
        type: "spring",
        stiffness: 450,
        damping: 18
    }
}
