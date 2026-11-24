package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mvrilo/go-redoc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	mongodb "github.com/anticoi/EscudoWAF/pkg/database/mongo"
	"github.com/anticoi/EscudoWAF/server/config"
	_ "github.com/anticoi/EscudoWAF/server/docs" // Importar la documentación swagger
	"github.com/anticoi/EscudoWAF/server/router"
	haproxyStats "github.com/anticoi/EscudoWAF/server/service/cornjob/haproxy"
	"github.com/anticoi/EscudoWAF/server/service/daemon"
	"github.com/anticoi/EscudoWAF/server/validator"
)

//	@title			EscudoWAF-WAF API
//	@version		1.0
//	@description	EscudoWAF - API del sistema de gestión de firewall de aplicaciones
//	@termsOfService	https://github.com/anticoi/EscudoWAF

//	@contact.name	API Support
//	@contact.url	https://github.com/anticoi/EscudoWAF
//	@contact.email	huahua1319873800@outlook.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:2333
//	@BasePath	/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Autenticación con el formato Bearer {token}
func main() {
	// Load configuration
	err := config.InitConfig()
	if err != nil {
		config.GlobalLogger.Error().Err(err).Msg("Failed to load configuration")
		return
	}

	// Conectar a la base de datos
	client, err := mongodb.Connect(config.Global.DBConfig.URI)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to connect to database")
		return
	}

	// Obtener la base de datos
	db := client.Database(config.Global.DBConfig.Database)

	err = config.InitDB(db)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to initialize database")
	}

	// Create service runner and start background services
	runner, err := daemon.GetRunnerService()

	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to create service runner")
		return
	}

	err = runner.StartServices()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to start daemon services")
		return
	}

	// Start HAProxy stats aggregation cornjob service
	haproxyStatsCleanup, err := haproxyStats.Start(runner, config.Logger)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to start HAProxy stats service")
		return
	}
	// Register cleanup function to be called during shutdown
	defer haproxyStatsCleanup()

	// Set Gin mode based on configuration
	if config.Global.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize the Gin route
	route := gin.New()

	// Setup the router
	router.Setup(route, db)

	// Inicializar los validadores
	validator.InitValidators()
	// validators.InitStructValidators()

	// Configurar la documentación de Swagger
	route.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(2),
		ginSwagger.DocExpansion("list"),
		ginSwagger.DeepLinking(true),
		ginSwagger.PersistAuthorization(true),
	))

	route.GET("/scalar", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		scheme := "http://"

		if c.GetHeader("X-Forwarded-Proto") == "https" || c.Request.TLS != nil {
			scheme = "https://"
		}

		// Priorizar X-Forwarded-Host; si no existe, usar Request.Host
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}

		swaggerJsonUrl := scheme + host + "/swagger/doc.json"

		content := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>EscudoWAF-WAF API Reference</title>
			<meta charset="utf-8" />
			<meta name="viewport" content="width=device-width, initial-scale=1" />
			<style>
				body {
					margin: 0;
				}
			</style>
		</head>
		<body>
			<script
			id="api-reference"
			type="application/json"
			data-url="%s"
			data-theme="light"
			data-layout="modern"
			></script>
			<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
		</body>
		</html>
		`, swaggerJsonUrl)

		c.String(http.StatusOK, content)
	})

	// Obtener el handler de Redoc
	doc := redoc.Redoc{
		Title:       "EscudoWAF-WAF API",
		Description: "EscudoWAF - API del sistema de gestión de firewall de aplicaciones",
		SpecFile:    "./docs/swagger.json",
		SpecPath:    "/swagger.json",
		DocsPath:    "/redoc",
	}
	redocHandler := doc.Handler()

	// Definir explícitamente la ruta de Redoc
	route.GET("/redoc", func(c *gin.Context) {
		redocHandler(c.Writer, c.Request)
	})
	route.GET("/swagger.json", func(c *gin.Context) {
		redocHandler(c.Writer, c.Request)
	})

	// Crear servidor HTTP
	srv := &http.Server{
		Addr:    config.Global.Bind,
		Handler: route,
	}

	// Crear un canal de errores
	serverError := make(chan error, 1)

	// Iniciar el servidor en una goroutine
	go func() {
		config.Logger.Info().Msgf("Starting server on %s", config.Global.Bind)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			config.Logger.Error().Err(err).Msg("Server error")
			serverError <- err
		}
	}()

	// Esperar señal de interrupción o error del servidor
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Usar select para esperar mensajes de cualquier canal
	select {
	case <-quit:
		config.Logger.Info().Msg("Received shutdown signal, shutting down services...")
	case err := <-serverError:
		config.Logger.Error().Err(err).Msg("Server failed, initiating shutdown...")
	}

	// Establecer el tiempo de espera de cierre
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Intentar cerrar el servidor de forma elegante
	if err := srv.Shutdown(ctx); err != nil {
		config.Logger.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		config.Logger.Info().Msg("Server shutdown gracefully")
	}

	// Detener servicios en segundo plano
	err = runner.StopServices()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to stop daemon services")
	}
	config.Logger.Info().Msg("Background services have been shut down, exiting...")

	// Si se sale por error de servidor, usar código de estado distinto de cero
	if len(serverError) > 0 {
		os.Exit(1)
	}
}
