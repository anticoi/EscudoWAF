import { HTMLMotionProps } from "motion/react"

// Animación de fundido
export const fadeAnimation: HTMLMotionProps<"div"> = {
    initial: "hidden",
    animate: "visible",
    exit: "hidden",
    variants: {
        visible: {
            opacity: 1,
            scale: 1,
            transition: {
                opacity: {
                    duration: 0.3,
                    ease: "easeOut"
                },
                scale: {
                    duration: 0.25,
                    ease: "easeOut"
                }
            }
        },
        hidden: {
            opacity: 0,
            scale: 0.98,
            transition: {
                opacity: {
                    duration: 0.25,
                    ease: "easeIn"
                },
                scale: {
                    duration: 0.2,
                    ease: "easeIn"
                }
            }
        }
    }
}

// Animación de deslizamiento
export const slideAnimation: HTMLMotionProps<"div"> = {
    initial: "hidden",
    animate: "visible",
    exit: "hidden",
    variants: {
        visible: {
            opacity: 1,
            x: 0,
            transition: {
                opacity: {
                    duration: 0.3,
                    ease: "easeOut"
                },
                x: {
                    duration: 0.3,
                    ease: [0.33, 1, 0.68, 1]
                }
            }
        },
        hidden: {
            opacity: 0,
            x: -20,
            transition: {
                opacity: {
                    duration: 0.25,
                    ease: "easeIn"
                },
                x: {
                    duration: 0.25,
                    ease: [0.33, 1, 0.68, 1]
                }
            }
        }
    }
}

// Animación de escala
export const scaleAnimation: HTMLMotionProps<"div"> = {
    initial: "hidden",
    animate: "visible",
    exit: "hidden",
    variants: {
        visible: {
            opacity: 1,
            scale: 1,
            y: 0,
            transition: {
                opacity: {
                    duration: 0.3,
                    ease: "easeOut"
                },
                scale: {
                    duration: 0.3,
                    ease: [0.33, 1, 0.68, 1]
                },
                y: {
                    duration: 0.3,
                    ease: [0.33, 1, 0.68, 1]
                }
            }
        },
        hidden: {
            opacity: 0,
            scale: 0.95,
            y: 10,
            transition: {
                opacity: {
                    duration: 0.2,
                    ease: "easeIn"
                },
                scale: {
                    duration: 0.25,
                    ease: [0.33, 1, 0.68, 1]
                },
                y: {
                    duration: 0.25,
                    ease: [0.33, 1, 0.68, 1]
                }
            }
        }
    }
} 