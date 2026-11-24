// src/components/ui/animations/collapse-animations.tsx
import { HTMLMotionProps } from "motion/react"

// Animación base de colapsar/expandir optimizada, evita saltos de layout
export const collapseAnimation: HTMLMotionProps<"div"> = {
    initial: "collapsed",
    animate: "open",
    exit: "collapsed",
    variants: {
        open: {
            height: "auto",
            opacity: 1,
            marginBottom: 16, // Aumentar el margen inferior en estado expandido para reservar espacio
            transition: {
                height: {
                    duration: 0.35, // Aumentar ligeramente la duración para una transición más suave
                    ease: [0.33, 1, 0.68, 1]
                },
                opacity: {
                    duration: 0.25,
                    delay: 0.1 // Aumentar el retardo del cambio de opacidad
                },
                marginBottom: {
                    duration: 0.35,
                    ease: [0.33, 1, 0.68, 1]
                }
            }
        },
        collapsed: {
            height: 0,
            opacity: 0,
            marginBottom: 0,
            transition: {
                height: {
                    duration: 0.3,
                    ease: [0.33, 1, 0.68, 1]
                },
                opacity: {
                    duration: 0.2,
                    ease: "easeIn"
                },
                marginBottom: {
                    duration: 0.3,
                    ease: [0.33, 1, 0.68, 1]
                }
            }
        }
    },
    style: {
        overflow: "hidden",
        position: "relative",
        willChange: "height, opacity, margin",
        transformOrigin: "top center", // Asegurar que cualquier transformación empiece desde arriba
        minHeight: 0 // Asegurar que pueda colapsarse completamente
    }
}

// Animación de colapso con efecto de escala, optimizada para reducir saltos visuales
export const collapseScaleAnimation: HTMLMotionProps<"div"> = {
    initial: "collapsed",
    animate: "open",
    exit: "collapsed",
    variants: {
        open: {
            height: "auto",
            opacity: 1,
            scale: 1,
            marginBottom: 16, // Añadir margen para el estado expandido
            transformOrigin: "top center",
            y: 0, // Mantener la posición vertical
            transition: {
                height: {
                    duration: 0.35,
                    ease: [0.33, 1, 0.68, 1]
                },
                scale: {
                    duration: 0.25,
                    delay: 0.05
                },
                opacity: {
                    duration: 0.25,
                    delay: 0.1
                },
                marginBottom: {
                    duration: 0.35
                },
                y: {
                    duration: 0.25
                }
            }
        },
        collapsed: {
            height: 0,
            opacity: 0,
            scale: 0.95, // Ajuste leve para reducir saltos visuales
            marginBottom: 0,
            transformOrigin: "top center",
            y: -5, // Ligero desplazamiento hacia arriba para desaparecer más naturalmente
            transition: {
                height: {
                    duration: 0.3,
                    ease: [0.33, 1, 0.68, 1]
                },
                scale: {
                    duration: 0.25
                },
                opacity: {
                    duration: 0.2
                },
                marginBottom: {
                    duration: 0.3
                },
                y: {
                    duration: 0.2
                }
            }
        }
    },
    style: {
        overflow: "hidden",
        position: "relative",
        willChange: "height, opacity, margin, transform",
        minHeight: 0,
        isolation: "isolate" // Crear nuevo contexto de apilamiento para reducir el área de repintado
    }
}

// Animación de expansión ligera, optimizada para reducir saltos visuales
export const collapseLightAnimation: HTMLMotionProps<"div"> = {
    initial: "collapsed",
    animate: "open",
    exit: "collapsed",
    variants: {
        open: {
            height: "auto",
            opacity: 1,
            marginBottom: 12, // Proporcionar márgenes adecuados
            y: 0,
            transition: {
                height: {
                    duration: 0.25,
                    ease: [0.33, 1, 0.68, 1]
                },
                opacity: {
                    duration: 0.2,
                    delay: 0.05
                },
                marginBottom: {
                    duration: 0.25
                },
                y: {
                    duration: 0.2
                }
            }
        },
        collapsed: {
            height: 0,
            opacity: 0,
            marginBottom: 0,
            y: -3, // Ligero desplazamiento hacia arriba
            transition: {
                height: {
                    duration: 0.25,
                    ease: [0.33, 1, 0.68, 1]
                },
                opacity: {
                    duration: 0.15
                },
                marginBottom: {
                    duration: 0.25
                },
                y: {
                    duration: 0.15
                }
            }
        }
    },
    style: {
        overflow: "hidden",
        position: "relative",
        willChange: "height, opacity, margin, transform",
        minHeight: 0,
        transformOrigin: "top center"
    }
}

// Nueva animación para elementos de lista, ideal para añadir/eliminar servidores
export const listItemAnimation: HTMLMotionProps<"div"> = {
    initial: "collapsed",
    animate: "open",
    exit: "collapsed",
    layout: true, // Habilitar animaciones de layout
    variants: {
        open: {
            height: "auto",
            opacity: 1,
            scale: 1,
            y: 0,
            marginBottom: 16,
            transition: {
                height: {
                    duration: 0.35,
                    ease: [0.33, 1, 0.68, 1]
                },
                opacity: {
                    duration: 0.25,
                    delay: 0.05
                },
                scale: {
                    duration: 0.25,
                    delay: 0.05
                },
                y: {
                    duration: 0.25
                },
                marginBottom: {
                    duration: 0.35
                }
            }
        },
        collapsed: {
            height: 0,
            opacity: 0,
            scale: 0.98,
            y: -10,
            marginBottom: 0,
            transition: {
                height: {
                    duration: 0.3,
                    ease: [0.33, 1, 0.68, 1]
                },
                opacity: {
                    duration: 0.2
                },
                scale: {
                    duration: 0.2
                },
                y: {
                    duration: 0.2
                },
                marginBottom: {
                    duration: 0.3
                }
            }
        }
    },
    style: {
        overflow: "hidden",
        position: "relative",
        willChange: "height, opacity, margin, transform",
        minHeight: 0,
        transformOrigin: "top center",
        zIndex: 1
    }
}