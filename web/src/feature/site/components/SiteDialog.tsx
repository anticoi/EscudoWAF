import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from '@/components/ui/dialog'
import { SiteForm } from './SiteForm'
import { Site } from '@/types/site'
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'

interface SiteDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    mode: 'create' | 'update'
    site?: Site | null // Solo necesario en modo edición
}

export function SiteDialog({
    open,
    onOpenChange,
    mode = 'create',
    site = null
}: SiteDialogProps) {
    const { t } = useTranslation()

    // Determinar el título y la descripción según el modo
    const title = mode === 'create' ? t('site.dialog.createTitle') : t('site.dialog.updateTitle')
    const description = mode === 'create'
        ? t('site.dialog.createDescription')
        : t('site.dialog.updateDescription')

    // Preparar valores por defecto del formulario según el modo
    const defaultValues = mode === 'update' && site
        ? {
            name: site.name,
            domain: site.domain,
            listenPort: site.listenPort,
            enableHTTPS: site.enableHTTPS,
            activeStatus: site.activeStatus,
            wafEnabled: site.wafEnabled,
            wafMode: site.wafMode,
            backend: site.backend,
            certificate: site.certificate,
        }
        : undefined

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
                                        <DialogDescription className='dark:text-shadow-glow-white'>{description}</DialogDescription>
                                    </DialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <SiteForm
                                        mode={mode}
                                        siteId={site?.id}
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