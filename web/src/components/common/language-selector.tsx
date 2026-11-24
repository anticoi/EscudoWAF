import { useEffect, useState } from "react"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
} from "@/components/ui/select"
import { useTranslation } from "react-i18next"
import { Languages } from "lucide-react"
import { AnimatedIcon } from "../ui/animation/components/animated-icon"

export function LanguageSelector() {
    const { i18n } = useTranslation()
    const [language, setLanguage] = useState(i18n.language || 'es')

    const handleLanguageChange = (value: string) => {
        setLanguage(value)
        i18n.changeLanguage(value)
        localStorage.setItem('i18nextLng', value)
    }

    // Obtener la configuración de idioma del almacenamiento local al inicializar
    useEffect(() => {
        const savedLanguage = localStorage.getItem('i18nextLng')
        if (savedLanguage && savedLanguage !== language) {
            setLanguage(savedLanguage)
            i18n.changeLanguage(savedLanguage) // Asegurar que se aplique el idioma actual
        }
    }, []) // Ejecutar solo una vez al montar el componente

    return (
        <Select value={language} onValueChange={handleLanguageChange}>
            <SelectTrigger className="w-11 h-9 p-1 border-0 bg-transparent shadow-none hover:bg-muted focus:ring-0  transition-colors">
                <AnimatedIcon animationVariant="pulse" className="h-5 w-5">
                    <Languages className="h-5 w-5" />
                </AnimatedIcon>
            </SelectTrigger>
            <SelectContent>
                <SelectItem value="es">Español</SelectItem>
                <SelectItem value="en">English</SelectItem>
                <SelectItem value="zh">简体中文</SelectItem>
            </SelectContent>
        </Select>
    )
}