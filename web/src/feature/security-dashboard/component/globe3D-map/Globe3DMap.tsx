import React, { useEffect, useRef } from "react"
import { WebGLRenderer, PerspectiveCamera, Scene, AmbientLight, DirectionalLight, Color, PointLight, MeshPhysicalMaterial, DoubleSide } from "three"
import { OrbitControls } from "three/addons/controls/OrbitControls.js"
import ThreeGlobe from "three-globe"
import countries from "./globe-data-min.json"
import type { FeatureCollection, GeoJsonProperties, Geometry } from "./geojson"
import { WAFAttackTrajectory, WAF_ATTACK_TRAJECTORY_COLORS } from "./types"

// Gestión de la instancia global del Globe del dashboard WAF - evita el problema de doble creación de StrictMode
const wafSecurityGlobeInstance: {
    renderer: WebGLRenderer | null
    camera: PerspectiveCamera | null
    scene: Scene | null
    globe: ThreeGlobe | null
    controls: OrbitControls | null
    animationId: number | null
    isInitialized: boolean
    isCreating: boolean  // Evitar creación concurrente
    elementRef: HTMLDivElement | null
    componentCount: number  // Rastrear el número de componentes que usan esta instancia
    resizeHandler: (() => void) | null // Almacenar el manejador de resize
} = {
    renderer: null,
    camera: null,
    scene: null,
    globe: null,
    controls: null,
    animationId: null,
    isInitialized: false,
    isCreating: false,
    elementRef: null,
    componentCount: 0,
    resizeHandler: null
}

// Componente Globe de seguridad WAF - usa gestión global del contexto WebGL
const Globe3DMap = React.memo(({ wafAttackTrajectoryData }: { wafAttackTrajectoryData: WAFAttackTrajectory[] }) => {
    const globeRef = useRef<HTMLDivElement>(null)
    const isMountedRef = useRef(false)

    // Inicializar o reutilizar la instancia global de Three.js
    useEffect(() => {
        if (!globeRef.current) return

        const el = globeRef.current
        isMountedRef.current = true
        wafSecurityGlobeInstance.componentCount++

        console.log(`Globe component mounted, count: ${wafSecurityGlobeInstance.componentCount}`)

        // Si se está creando, esperar a que termine la creación
        if (wafSecurityGlobeInstance.isCreating) {
            console.log('WebGL context is being created, waiting...')
            const checkInterval = setInterval(() => {
                if (!wafSecurityGlobeInstance.isCreating && wafSecurityGlobeInstance.isInitialized) {
                    clearInterval(checkInterval)
                    attachToElement(el)
                }
            }, 50)
            return () => clearInterval(checkInterval)
        }

        // Si la instancia global existe y es válida, reutilizarla
        if (wafSecurityGlobeInstance.isInitialized && wafSecurityGlobeInstance.renderer) {
            console.log('Reusing existing WebGL context...')
            attachToElement(el)
            return
        }

        // Crear nueva instancia
        createGlobeInstance(el)

        // Función de limpieza
        return () => {
            console.log(`Globe component unmounting, count: ${wafSecurityGlobeInstance.componentCount}`)
            isMountedRef.current = false
            wafSecurityGlobeInstance.componentCount--

            // Solo limpiar la instancia global cuando ningún componente la usa
            if (wafSecurityGlobeInstance.componentCount <= 0) {
                setTimeout(() => {
                    // Verificar de nuevo que ningún componente lo está usando
                    if (wafSecurityGlobeInstance.componentCount <= 0) {
                        console.log('No components using global instance, cleaning up...')
                        cleanupGlobalInstance()
                    }
                }, 1000) // Limpieza diferida, dar tiempo al cambio de rutas
            }
        }
    }, [])

    // Adjuntar el canvas existente al nuevo elemento
    const attachToElement = (el: HTMLDivElement) => {
        if (!wafSecurityGlobeInstance.renderer) return

        const rect = el.getBoundingClientRect()

        // Si el canvas no está en el contenedor actual, moverlo
        if (!el.contains(wafSecurityGlobeInstance.renderer.domElement)) {
            el.innerHTML = ''
            el.appendChild(wafSecurityGlobeInstance.renderer.domElement)
            wafSecurityGlobeInstance.elementRef = el
        }

        // Establecer el estilo del contenedor
        el.style.setProperty("width", "100%")
        el.style.setProperty("height", "100%")
        el.style.setProperty("overflow", "hidden")

        // Actualizar las dimensiones
        if (wafSecurityGlobeInstance.camera) {
            wafSecurityGlobeInstance.camera.aspect = rect.width / rect.height
            wafSecurityGlobeInstance.camera.updateProjectionMatrix()
        }
        wafSecurityGlobeInstance.renderer.setSize(rect.width, rect.height)

        // Asegurar el ángulo de rotación correcto también al reutilizar la instancia
        // if (wafSecurityGlobeInstance.globe) {
        //     console.log('Setting rotation for reused globe instance...')
        //     wafSecurityGlobeInstance.globe.rotation.y = -Math.PI * (95 / 180)  // -95 grados
        //     wafSecurityGlobeInstance.globe.rotation.z = -Math.PI / 8  // ligera inclinación
        // }
    }

    // Crear nueva instancia de Globe
    const createGlobeInstance = (el: HTMLDivElement) => {
        if (wafSecurityGlobeInstance.isCreating || wafSecurityGlobeInstance.isInitialized) {
            return
        }

        console.log('Creating new global WebGL context...')
        wafSecurityGlobeInstance.isCreating = true

        const rect = el.getBoundingClientRect()
        el.innerHTML = ''

        try {
            // Crear nueva instancia de Three.js
            const renderer = new WebGLRenderer({
                antialias: true,
                logarithmicDepthBuffer: true,
                alpha: true,
                preserveDrawingBuffer: true  // Mantener el buffer de dibujo
            })
            const camera = new PerspectiveCamera(50, rect.width / rect.height, 0.001, 1000)
            const scene = new Scene()

            renderer.setPixelRatio(window.devicePixelRatio)
            renderer.setSize(rect.width, rect.height)
            el.appendChild(renderer.domElement)
            el.style.setProperty("width", "100%")
            el.style.setProperty("height", "100%")
            el.style.setProperty("overflow", "hidden")

            const controls = new OrbitControls(camera, renderer.domElement)

            // Configurar las luces
            const ambientLight = new AmbientLight(new Color(0xFFFFFF), 1)
            const dLight = new DirectionalLight(new Color(0x000000), 0.6)
            dLight.position.set(-400, 100, 400)
            const dLight1 = new DirectionalLight(new Color(0xFFFFFF), 1)
            dLight1.position.set(-200, 500, 200)
            const dLight2 = new PointLight(new Color(0xFFFFFF), 0.8)
            dLight2.position.set(-200, 500, 200)
            const dirLight = new DirectionalLight(0x000000, 1)
            dirLight.position.set(5, 3, 4)
            scene.add(ambientLight, dLight, dLight1, dLight2, dirLight)

            // Configurar la cámara
            camera.aspect = rect.width / rect.height
            camera.position.z = 320  // Reducido de 400 a 280, para que el globo se vea más grande
            camera.position.x = 0
            camera.position.y = 0
            camera.updateProjectionMatrix()

            // Establecer el controlador
            controls.enableDamping = true
            controls.enablePan = false
            controls.minDistance = 150  // Reducido de 200 a 150
            controls.maxDistance = 400  // Reducido de 500 a 400
            controls.rotateSpeed = 0.8
            controls.zoomSpeed = 1
            controls.autoRotate = false
            controls.minPolarAngle = Math.PI / 5
            controls.maxPolarAngle = Math.PI - Math.PI / 5

            // Crear el globo
            console.log('Creating globe instance...')
            const globe = new ThreeGlobe({ waitForGlobeReady: true, animateIn: true })
            scene.add(globe)

            const globeMaterial = new MeshPhysicalMaterial({
                color: new Color(0x0d0c27),
                transparent: true,       // Habilitar transparencia
                transmission: 1,         // Transmitancia
                thickness: 0.5,          // Grosor
                roughness: 0.75,         // Rugosidad
                metalness: 0,            // Apariencia metálica
                ior: 1.5,                // Índice de refracción
                envMapIntensity: 1.5,    // Intensidad del mapa de entorno
                reflectivity: 0.1,       // Reflectividad
                opacity: 0.55,           // Opacidad
                side: DoubleSide,        // Renderizado de doble cara
            })

            globe
                .showGlobe(true)
                .globeMaterial(globeMaterial)
                .showAtmosphere(true)
                .atmosphereColor("#0d0c27")
                .atmosphereAltitude(0.25)

            globe
                .hexPolygonsData(countries.features)
                .hexPolygonResolution(3)
                .hexPolygonAltitude(0.001)
                .hexPolygonMargin(0.4)
                .hexPolygonColor((e) => {
                    const countryCode = (e as FeatureCollection<Geometry, GeoJsonProperties>["features"][number]).properties!.ISO_A3
                    return (countryCode === "CHN" || countryCode === "TWN") ? "rgba(255, 255, 255, 1)" : "rgba(241, 230, 255, 1)"
                })

            // Establecer el callback onGlobeReady (solo una vez)
            globe.onGlobeReady(() => {
                console.log('Globe is ready!')
                // Establecer el ángulo de rotación cuando el globo esté listo
                // globe.rotation.y = -Math.PI * (95 / 180)  // -95 grados
                // globe.rotation.z = -Math.PI / 8  // ligera inclinación para mejor ángulo de visión
            })

            // Establecer el ángulo de rotación inmediatamente (sin esperar a onGlobeReady)
            // console.log('Setting globe rotation immediately...')
            // globe.rotation.y = -Math.PI * (95 / 180)  // -95 grados
            // globe.rotation.z = -Math.PI / 8  // ligera inclinación para mejor ángulo de visión

            // Actualizar la instancia global
            wafSecurityGlobeInstance.renderer = renderer
            wafSecurityGlobeInstance.camera = camera
            wafSecurityGlobeInstance.scene = scene
            wafSecurityGlobeInstance.globe = globe
            wafSecurityGlobeInstance.controls = controls
            wafSecurityGlobeInstance.elementRef = el
            wafSecurityGlobeInstance.isInitialized = true
            wafSecurityGlobeInstance.isCreating = false

            // Bucle de animación
            const animate = () => {
                if (!wafSecurityGlobeInstance.isInitialized || !isMountedRef.current) return

                camera.lookAt(scene.position)
                controls.update()
                renderer.render(scene, camera)
                wafSecurityGlobeInstance.animationId = requestAnimationFrame(animate)
            }

            console.log('Starting animation loop...')
            animate()

            // Manejar cambios de tamaño de ventana
            const handleResize = () => {
                if (!wafSecurityGlobeInstance.camera || !wafSecurityGlobeInstance.renderer) return
                if (!wafSecurityGlobeInstance.elementRef) return

                const newRect = wafSecurityGlobeInstance.elementRef.getBoundingClientRect()
                wafSecurityGlobeInstance.camera.aspect = newRect.width / newRect.height
                wafSecurityGlobeInstance.camera.updateProjectionMatrix()
                wafSecurityGlobeInstance.renderer.setSize(newRect.width, newRect.height)
            }

            // Guardar la referencia del manejador de resize
            wafSecurityGlobeInstance.resizeHandler = handleResize
            window.addEventListener('resize', handleResize)

        } catch (error) {
            console.error('Failed to create WebGL context:', error)
            wafSecurityGlobeInstance.isCreating = false
        }
    }

    // Limpiar la instancia global
    const cleanupGlobalInstance = () => {
        console.log('Cleaning up global WebGL context...')

        if (wafSecurityGlobeInstance.animationId) {
            cancelAnimationFrame(wafSecurityGlobeInstance.animationId)
        }

        if (wafSecurityGlobeInstance.renderer) {
            wafSecurityGlobeInstance.renderer.dispose()
            wafSecurityGlobeInstance.renderer.forceContextLoss()
        }

        if (wafSecurityGlobeInstance.controls) {
            wafSecurityGlobeInstance.controls.dispose()
        }

        if (wafSecurityGlobeInstance.scene) {
            wafSecurityGlobeInstance.scene.clear()
        }

        // Eliminar correctamente el listener de eventos
        if (wafSecurityGlobeInstance.resizeHandler) {
            window.removeEventListener('resize', wafSecurityGlobeInstance.resizeHandler)
            wafSecurityGlobeInstance.resizeHandler = null
        }

        // Restablecer la instancia global
        wafSecurityGlobeInstance.renderer = null
        wafSecurityGlobeInstance.camera = null
        wafSecurityGlobeInstance.scene = null
        wafSecurityGlobeInstance.globe = null
        wafSecurityGlobeInstance.controls = null
        wafSecurityGlobeInstance.animationId = null
        wafSecurityGlobeInstance.isInitialized = false
        wafSecurityGlobeInstance.isCreating = false
        wafSecurityGlobeInstance.elementRef = null
    }

    // Actualizar los datos de trayectorias
    useEffect(() => {
        if (!wafSecurityGlobeInstance.globe || !wafSecurityGlobeInstance.isInitialized || !wafAttackTrajectoryData || wafAttackTrajectoryData.length === 0) {
            return
        }

        console.log('Updating globe data, trajectoryData length:', wafAttackTrajectoryData.length)

        const globe = wafSecurityGlobeInstance.globe

        // // Asegurar el ángulo de rotación correcto en cada actualización de datos
        // console.log('Setting rotation in data update...')
        // // Primero restablecer la rotación
        // globe.rotation.set(0, 0, 0)
        // // Luego aplicar la nueva rotación
        // globe.rotation.y = -Math.PI * (95 / 180)  // -95 grados
        // globe.rotation.z = -Math.PI / 8  // ligera inclinación
        // console.log('Current globe rotation:', globe.rotation.y, globe.rotation.z)

        // Generar datos de puntos de origen de ataques y del centro de protección del WAF
        const attackPointsData = Array.from(wafAttackTrajectoryData
            .reduce((p, c) => {
                // Puntos de origen de ataques
                p.set(`${c.startLat},${c.startLng}`, {
                    lat: c.startLat,
                    lng: c.startLng,
                })
                // Centro de protección WAF (Hangzhou)
                p.set(`${c.endLat},${c.endLng}`, {
                    lat: c.endLat,
                    lng: c.endLng,
                })
                return p
            }, new Map<string, { lat: number; lng: number }>()).values()
        )

        // Actualizar datos de trayectorias de ataque del WAF - estilos de línea optimizados
        globe
            .arcsData(wafAttackTrajectoryData)
            .arcColor((e: unknown) => {
                const attackTrajectory = e as WAFAttackTrajectory
                // Asignar colores según el tipo de ataque
                return WAF_ATTACK_TRAJECTORY_COLORS[attackTrajectory.colorIndex]
            })
            .arcAltitude((e: unknown) => Math.pow((e as WAFAttackTrajectory).arcAlt * 0.22, 1.225)) // Altura de la trayectoria de ataque (arcAlt es distancia geográfica; hay que adaptar la altura de la curva Bézier cuártica, a mayor distancia mayor compensación, por eso se usa multiplicador con potencia)
            .arcStroke(0.3) // Grosor de las líneas de trayectoria de ataques
            .arcDashLength(0.9)
            .arcDashGap(4)
            .arcDashAnimateTime(1000)
            .arcsTransitionDuration(1000)
            .arcDashInitialGap((e: unknown) => (e as WAFAttackTrajectory).order * 1)

        // Marcar los puntos de ataque con efecto de anillo animado
        globe
            .ringsData(attackPointsData)
            .pointColor("#ffffaa")
            .pointsMerge(true)
            .pointRadius(0.25)

        console.log('Globe data updated successfully!')
    }, [wafAttackTrajectoryData])

    return <div className="w-full h-full" ref={globeRef}></div>
})

Globe3DMap.displayName = 'Globe3DMap'

export default Globe3DMap 