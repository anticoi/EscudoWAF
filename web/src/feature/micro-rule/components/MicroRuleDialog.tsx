import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from '@/components/ui/dialog'
import { MicroRuleForm } from './MicroRuleForm'
import { MicroRule } from '@/types/rule'
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'

interface MicroRuleDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    mode: 'create' | 'update'
    rule?: MicroRule | null
}

export function MicroRuleDialog({
    open,
    onOpenChange,
    mode = 'create',
    rule = null
}: MicroRuleDialogProps) {
    const { t } = useTranslation()

    // Determinar el título y la descripción según el modo
    const title = mode === 'create' ? t("microRule.dialog.createTitle") : t("microRule.dialog.updateTitle")
    const description = mode === 'create'
        ? t("microRule.dialog.createDescription")
        : t("microRule.dialog.updateDescription")

    // Preparar valores por defecto del formulario según el modo
    const defaultValues = mode === 'update' && rule
        ? rule
        : undefined // Usar los valores por defecto de MicroRuleForm

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence mode="wait">
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto scrollbar-neon p-0">
                            <motion.div {...dialogContentAnimation}>
                                <motion.div {...dialogHeaderAnimation}>
                                    <DialogHeader className="p-6 pb-3">
                                        <DialogTitle className="text-xl">{title}</DialogTitle>
                                        <DialogDescription className="dark:text-shadow-glow-white">{description}</DialogDescription>
                                    </DialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <MicroRuleForm
                                        mode={mode}
                                        ruleId={rule?.id}
                                        defaultValues={defaultValues}
                                        onSuccess={() => onOpenChange(false)}
                                    />
                                </motion.div>
                            </motion.div>
                        </DialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </Dialog>
    )
}