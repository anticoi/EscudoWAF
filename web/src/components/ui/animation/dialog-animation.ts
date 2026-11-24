import { HTMLMotionProps } from "motion/react"

// Animación del contenedor de entrada y salida del diálogo
export const dialogEnterExitAnimation: HTMLMotionProps<"div"> = {
  className: "fixed left-[50%] top-[50%] z-50 grid w-full translate-x-[-50%] translate-y-[-50%]",
  initial: "hidden",
  animate: "visible",
  exit: "exit",
  variants: {
    hidden: { 
      opacity: 0, 
      scale: 0.96,
      y: 8
    },
    visible: { 
      opacity: 1, 
      scale: 1,
      y: 0,
      transition: {
        duration: 0.3,
        ease: [0.16, 1, 0.3, 1], // custom ease curve for natural feel
        when: "beforeChildren"
      }
    },
    exit: {
      opacity: 0,
      scale: 0.96,
      y: -8,
      transition: {
        duration: 0.25,
        ease: [0.32, 0, 0.67, 0], // easeInCubic for natural exit
        when: "afterChildren"
      }
    }
  }
}

// DialogContent animación del área de contenido - efecto de aparición más natural
export const dialogContentAnimation: HTMLMotionProps<"div"> = {
  variants: {
    hidden: { 
      opacity: 0,
    },
    visible: { 
      opacity: 1,
      transition: { 
        duration: 0.2,
        ease: "easeOut",
        staggerChildren: 0.06,
        delayChildren: 0.05
      }
    },
    exit: { 
      opacity: 0,
      transition: { 
        duration: 0.15,
        ease: "easeIn",
        staggerChildren: 0.03,
        staggerDirection: -1
      } 
    }
  }
}

// Animación de elementos hijos con efecto de flujo más natural
export const dialogContentItemAnimation: HTMLMotionProps<"div"> = {
  variants: {
    hidden: { 
      opacity: 0,
      y: 12,
      scale: 0.98
    },
    visible: { 
      opacity: 1,
      y: 0,
      scale: 1,
      transition: {
        type: "spring",
        damping: 20, // Mayor amortiguación para una animación más natural sin rebote excesivo
        stiffness: 260, // La rigidez adecuada equilibra velocidad y fluidez
        mass: 0.4 // Menor masa hace la animación más fluida
      }
    },
    exit: { 
      opacity: 0,
      y: -8,
      transition: { 
        duration: 0.18,
        ease: [0.32, 0, 0.67, 0] // easeInCubic for natural exit
      }
    }
  }
}

// DialogHeader animación del área de título
export const dialogHeaderAnimation: HTMLMotionProps<"div"> = {
  variants: {
    hidden: { 
      opacity: 0,
      y: -10 
    },
    visible: { 
      opacity: 1,
      y: 0,
      transition: { 
        duration: 0.3,
        ease: "easeOut" 
      }
    },
    exit: { 
      opacity: 0,
      y: -5,
      transition: { 
        duration: 0.2,
        ease: "easeIn" 
      }
    }
  }
} 