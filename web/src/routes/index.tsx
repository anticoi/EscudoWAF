// src/routes/AppRouter.tsx
import { createBrowserRouter, RouteObject, RouterProvider } from "react-router"
import { useRoutes } from "./config"
import { RouteErrorBoundary } from "@/handler/error-boundary"

export function AppRouter() {
    // Usar la configuración de rutas existente
    const routes = useRoutes()

    // Iterar sobre las rutas y añadir errorElement
    const routesWithErrorHandling = addErrorElementToRoutes(routes)

    // Crear el router
    const router = createBrowserRouter(routesWithErrorHandling)

    return <RouterProvider router={router} />
}

// Añadir manejo de errores recursivamente
function addErrorElementToRoutes(routes: RouteObject[]) {
    return routes.map(route => {
        // Añadir elemento de error a cada ruta
        const updatedRoute = {
            ...route,
            errorElement: <RouteErrorBoundary />
        }

        // Procesar subrutas recursivamente
        if (route.children) {
            updatedRoute.children = addErrorElementToRoutes(route.children)
        }

        return updatedRoute
    })
}