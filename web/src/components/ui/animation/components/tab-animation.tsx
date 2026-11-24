import { tabFadeAnimation, tabScaleAnimation, tabContentAnimation } from "../tab-animation"
import React from "react"
import { AnimatePresence, motion } from "motion/react"

interface TabsAnimationProviderProps {
    children: React.ReactNode
    currentView: string
    animationVariant?: "slide" | "fade" | "scale"
}

export function TabsAnimationProvider({
    children,
    currentView,
    animationVariant = "slide"
}: TabsAnimationProviderProps) {
    // Elegir la animación correspondiente según la variante seleccionada
    const getAnimationProps = () => {
        switch (animationVariant) {
            case "fade":
                return tabFadeAnimation
            case "scale":
                return tabScaleAnimation
            case "slide":
            default:
                return tabContentAnimation
        }
    }

    return (
        <AnimatePresence mode="wait">
            <motion.div
                key={currentView}
                {...getAnimationProps()}
            >
                {children}
            </motion.div>
        </AnimatePresence>
    )
}