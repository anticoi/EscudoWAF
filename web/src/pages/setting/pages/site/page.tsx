import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
    Plus,
    LayoutGrid,
    AlignJustify,
    RefreshCw
} from "lucide-react"
import { useQueryClient } from "@tanstack/react-query"
import { SiteGrid } from "@/feature/site/components/SiteGrid"
import { SiteTable } from "@/feature/site/components/SiteTable"
import { SiteDialog } from "@/feature/site/components/SiteDialog"
import { DeleteSiteDialog } from "@/feature/site/components/DeleteSiteDialog"
import { Site } from "@/types/site"
import { TabsAnimationProvider } from "@/components/ui/animation/components/tab-animation"
import { useTranslation } from "react-i18next"
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"
import { AnimatedIcon } from "@/components/ui/animation/components/animated-icon"

export default function SiteManagerPage() {
    const { t } = useTranslation()
    const [view, setView] = useState<'grid' | 'table'>('grid')
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
    const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
    const [selectedSite, setSelectedSite] = useState<Site | null>(null)
    const [selectedSiteId, setSelectedSiteId] = useState<string | null>(null)
    const [isRefreshAnimating, setIsRefreshAnimating] = useState(false)

    const queryClient = useQueryClient()


    // Manejar la adición de sitio
    const handleAddSite = () => {
        setIsAddDialogOpen(true)
    }

    // Manejar la edición de sitio
    const handleEditSite = (site: Site) => {
        setSelectedSite(site)
        setIsEditDialogOpen(true)
    }

    // Manejar la eliminación del sitio
    const handleDeleteSite = (id: string) => {
        setSelectedSiteId(id)
        setIsDeleteDialogOpen(true)
    }

    // Actualizar la lista de sitios
    const refreshSites = () => {
        setIsRefreshAnimating(true)
        queryClient.invalidateQueries({ queryKey: ['sites'] })

        // Detener la animación con 1 segundo de retraso para coincidir con el efecto
        setTimeout(() => {
            setIsRefreshAnimating(false)
        }, 1000)
    }


    return (
        <Card className="p-6 w-full h-full border-none shadow-none rounded-none">
            <div className="flex justify-between items-center mb-6 bg-zinc-50 dark:bg-muted/30 rounded-md p-4 transition-colors duration-200">
                <h2 className="text-xl font-semibold text-primary dark:text-white">{t('site.management')}</h2>
                <div className="flex gap-2">
                    <AnimatedButton >
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={refreshSites}
                            className="flex items-center gap-2 dark:text-shadow-glow-white"
                        >
                            <AnimatedIcon animationVariant="continuous-spin" isAnimating={isRefreshAnimating} className="h-4 w-4">
                                <RefreshCw className="h-4 w-4" />
                            </AnimatedIcon>
                            {t('site.refresh')}
                        </Button>
                    </AnimatedButton>
                    <AnimatedButton>
                        <Button
                            size="sm"
                            onClick={handleAddSite}
                            className="flex items-center gap-1 dark:text-shadow-glow-white"
                        >
                            <Plus className="h-4 w-4" />
                            {t('site.add')}
                        </Button>
                    </AnimatedButton>
                </div>
            </div>

            <Tabs value={view} onValueChange={(v) => setView(v as 'grid' | 'table')}>
                <TabsList className="mb-4 dark:bg-muted/30">
                    <TabsTrigger value="grid" className="flex items-center gap-1 group">
                        <LayoutGrid className="h-4 w-4 group-data-[state=active]:text-primary group-data-[state=active]:fill-primary group-hover:text-primary group-hover:fill-primary transition-colors" />
                    </TabsTrigger>
                    <TabsTrigger value="table" className="flex items-center gap-1 group">
                        <AlignJustify className="h-4 w-4 group-data-[state=active]:text-primary group-data-[state=active]:fill-primary group-hover:text-primary group-hover:fill-primary transition-colors" />
                    </TabsTrigger>
                </TabsList>

                <TabsAnimationProvider currentView={view} animationVariant="slide">
                    {view === "grid" ? (
                        <TabsContent value="grid" forceMount>
                            <SiteGrid
                                onEdit={handleEditSite}
                                onDelete={handleDeleteSite}
                            />
                        </TabsContent>
                    ) : (
                        <TabsContent value="table" forceMount>
                            <SiteTable
                                onEdit={handleEditSite}
                                onDelete={handleDeleteSite}
                            />
                        </TabsContent>
                    )}
                </TabsAnimationProvider>
            </Tabs>

            {/* Diálogo de añadir sitio */}
            <SiteDialog
                open={isAddDialogOpen}
                onOpenChange={setIsAddDialogOpen}
                mode="create"
            />

            {/* Diálogo de editar sitio */}
            <SiteDialog
                open={isEditDialogOpen}
                onOpenChange={setIsEditDialogOpen}
                mode="update"
                site={selectedSite}
            />

            {/* Diálogo de confirmación de eliminación de sitio */}
            <DeleteSiteDialog
                open={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
                siteId={selectedSiteId}
            />
        </Card >
    )
} 