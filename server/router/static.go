package router

import (
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/public"
	"github.com/gin-gonic/gin"
)

// SetStaticFileRouter configura las rutas de archivos estáticos
func SetStaticFileRouter(router *gin.Engine) {
	logger := config.Logger

	// Cargar la plantilla HTML
	router.SetHTMLTemplate(template.Must(template.New("").ParseFS(public.Templates, "templates/*")))

	// Comprobar si la función web está deshabilitada
	if config.Global.DisableWeb {
		logger.Info().Msg("Función web deshabilitada")
		router.GET("/", func(ctx *gin.Context) {
			ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
				"URL":               "https://github.com/anticoi/EscudoWAF",
				"INITIAL_COUNTDOWN": 15,
			})
		})
		return
	}

	// Elegir el sistema de archivos según la configuración WebPath
	if config.Global.WebPath == "" {
		// Usar sistema de archivos embebido
		logger.Info().Msg("Usar recursos de frontend embebidos")
		err := initFSRouter(router, public.Public.(fs.ReadDirFS), ".")
		if err != nil {
			panic(err)
		}
		fs := http.FS(public.Public)
		router.NoRoute(newIndexNoRouteHandler(fs))
	} else {
		// Usar sistema de archivos externo
		logger.Info().Msg("Usar ruta de recursos de frontend externa")
		absPath, err := filepath.Abs(config.Global.WebPath)
		if err != nil {
			panic(err)
		}
		logger.Info().Str("path", absPath).Msg("Usar ruta de recursos de frontend externa")
		err = initFSRouter(router, os.DirFS(absPath).(fs.ReadDirFS), ".")
		if err != nil {
			panic(err)
		}
		router.NoRoute(newDynamicNoRouteHandler(http.Dir(absPath)))
	}
}

// checkNoRouteNotFound comprueba si la ruta es una ruta de API
func checkNoRouteNotFound(path string) bool {
	if strings.HasPrefix(path, "/api") ||
		strings.HasPrefix(path, "/health") {
		return true
	}
	return false
}

// newIndexNoRouteHandler crea el manejador NoRoute para rutas SPA (sistema de archivos embebido)
func newIndexNoRouteHandler(fs http.FileSystem) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if checkNoRouteNotFound(ctx.Request.URL.Path) {
			ctx.String(http.StatusNotFound, "404 page not found")
			return
		}
		// Para rutas del frontend, usar rutas de directorio para evitar problemas de redirección de index.html
		// Según la solución de https://github.com/gin-gonic/gin/issues/2654
		// Usar "" o "/" en lugar de "index.html" para evitar bucles de redirección
		ctx.FileFromFS("", fs)
	}
}

// newDynamicNoRouteHandler crea el manejador NoRoute para rutas SPA (sistema de archivos externo)
func newDynamicNoRouteHandler(fs http.FileSystem) func(ctx *gin.Context) {
	fileServer := http.StripPrefix("/", http.FileServer(fs))
	return func(c *gin.Context) {
		if checkNoRouteNotFound(c.Request.URL.Path) {
			c.String(http.StatusNotFound, "404 page not found")
			return
		}

		// Intentar abrir el archivo solicitado
		f, err := fs.Open(c.Request.URL.Path)
		if err != nil {
			// El archivo no existe, devolver index.html (manejo de rutas SPA)
			c.FileFromFS("", fs)
			return
		}
		f.Close()

		// El archivo existe, servirlo directamente
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// Definición de la interfaz staticFileFS
type staticFileFS interface {
	StaticFileFS(relativePath string, filepath string, fs http.FileSystem) gin.IRoutes
}

// initFSRouter recorre el sistema de archivos recursivamente y registra todos los archivos estáticos
func initFSRouter(e staticFileFS, f fs.ReadDirFS, path string) error {
	dirs, err := f.ReadDir(path)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		u, err := url.JoinPath(path, dir.Name())
		if err != nil {
			return err
		}

		if dir.IsDir() {
			// Procesar subdirectorios recursivamente
			err = initFSRouter(e, f, u)
			if err != nil {
				return err
			}
		} else {
			// Registrar archivo individual, pero omitir index.html de la raíz para evitar conflictos
			routePath := "/" + u
			if path == "." {
				routePath = "/" + dir.Name()
				// Omitir index.html de la raíz, lo maneja el handler NoRoute
				if dir.Name() == "index.html" {
					continue
				}
			}
			e.StaticFileFS(routePath, u, http.FS(f))
		}
	}
	return nil
}
