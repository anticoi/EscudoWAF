// src/components/ui/animation/components/animated-icon.tsx
import { motion } from "motion/react"
import React, { forwardRef } from "react"
import {
    spinIconAnimation,
    continuousSpinAnimation,
    continuousShakeAnimation,
    continuousPulseAnimation,
    shakeIconAnimation,
    bounceIconAnimation,
    pulseIconAnimation,
    glowIconAnimation
} from "../icon-animation"

interface AnimatedIconProps {
    children: React.ReactNode
    className?: string
    animationVariant?: "spin" | "continuous-spin" | "continuous-shake" | "continuous-pulse" | "shake" | "bounce" | "pulse" | "glow"
    isAnimating?: boolean // Para controlar animaciones continuas
    onClick?: () => void
}

export const AnimatedIcon = forwardRef<HTMLDivElement, AnimatedIconProps>(
    ({
        children,
        className = "",
        animationVariant = "spin",
        isAnimating = false,
        onClick,
        ...props
    }, ref) => {
        // Elegir las propiedades de animación según la variante
        const getAnimationProps = () => {
            // Para animaciones continuas, aplicar animación según isAnimating
            if ((animationVariant === "continuous-spin" || animationVariant === "continuous-shake" || animationVariant === "continuous-pulse") && isAnimating) {
                switch (animationVariant) {
                    case "continuous-spin":
                        return continuousSpinAnimation
                    case "continuous-shake":
                        return continuousShakeAnimation
                    case "continuous-pulse":
                        return continuousPulseAnimation
                    default:
                        return continuousSpinAnimation
                }
            }

            // Para otras animaciones interactivas, seleccionar según la variante
            switch (animationVariant) {
                case "continuous-spin":
                case "spin":
                    return spinIconAnimation
                case "continuous-shake":
                case "shake":
                    return shakeIconAnimation
                case "continuous-pulse":
                case "bounce":
                    return bounceIconAnimation
                case "pulse":
                    return pulseIconAnimation
                case "glow":
                    return glowIconAnimation
                default:
                    return spinIconAnimation
            }
        }

        return (
            <motion.div
                ref={ref}
                className={className}
                {...getAnimationProps()}
                onClick={onClick}
                {...props}
            >
                {children}
            </motion.div>
        )
    }
)

AnimatedIcon.displayName = "AnimatedIcon"