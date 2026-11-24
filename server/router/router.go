package router

import (
	"github.com/anticoi/EscudoWAF/server/controller"
	"github.com/anticoi/EscudoWAF/server/middleware"
	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/anticoi/EscudoWAF/server/repository"
	"github.com/anticoi/EscudoWAF/server/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Setup configures all the routes for the application
func Setup(route *gin.Engine, db *mongo.Database) {
	// Middleware base
	route.Use(middleware.RequestID())
	route.Use(middleware.Logger())
	route.Use(middleware.Cors())
	route.Use(gin.CustomRecovery(middleware.CustomErrorHandler))

	// Crear los repositorios
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	siteRepo := repository.NewSiteRepository(db)
	wafLogRepo := repository.NewWAFLogRepository(db)
	certRepo := repository.NewCertificateRepository(db)
	configRepo := repository.NewConfigRepository(db)
	ipGroupRepo := repository.NewIPGroupRepository(db)
	ruleRepo := repository.NewMicroRuleRepository(db)
	blockedIPRepo := repository.NewBlockedIPRepository(db)

	// Crear los servicios
	authService := service.NewAuthService(userRepo, roleRepo)
	siteService := service.NewSiteService(siteRepo)
	wafLogService := service.NewWAFLogService(wafLogRepo)
	certService := service.NewCertificateService(certRepo)
	runnerService, _ := service.NewRunnerService()
	configService := service.NewConfigService(configRepo)
	ipGroupService := service.NewIPGroupService(ipGroupRepo)
	ruleService := service.NewMicroRuleService(ruleRepo)
	statsService := service.NewStatsService(wafLogRepo)
	blockedIPService := service.NewBlockedIPService(blockedIPRepo)
	// Crear los controladores
	authController := controller.NewAuthController(authService)
	siteController := controller.NewSiteController(siteService)
	wafLogController := controller.NewWAFLogController(wafLogService)
	certController := controller.NewCertificateController(certService)
	runnerController := controller.NewRunnerController(runnerService)
	configController := controller.NewConfigController(configService)
	ipGroupController := controller.NewIPGroupController(ipGroupService)
	ruleController := controller.NewMicroRuleController(ruleService)
	statsController := controller.NewStatsController(runnerService, statsService)
	blockedIPController := controller.NewBlockedIPController(blockedIPService)
	// Añadir los repositorios al contexto para uso del middleware
	route.Use(func(c *gin.Context) {
		c.Set("userRepo", userRepo)
		c.Set("roleRepo", roleRepo)
		c.Next()
	})

	// Endpoint de health check
	route.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Rutas de API v1
	api := route.Group("/api/v1")

	// Rutas de autenticación - no requieren verificación de permisos
	auth := api.Group("/auth")
	{
		auth.POST("/login", authController.Login)

		// Rutas que requieren autenticación
		authRequired := auth.Group("")
		authRequired.Use(middleware.JWTAuth())
		{
			// Interfaz de restablecimiento de contraseña - accesible a cualquier usuario autenticado
			authRequired.POST("/reset-password", authController.ResetPassword)

			// Rutas que requieren comprobación de restablecimiento de contraseña
			passwordChecked := authRequired.Group("")
			passwordChecked.Use(middleware.PasswordResetRequired())
			{
				// Obtener información personal - accesible para cualquier usuario autenticado
				passwordChecked.GET("/me", authController.GetUserInfo)
			}
		}
	}

	// Rutas de API que requieren autenticación y comprobación de restablecimiento de contraseña
	authenticated := api.Group("")
	authenticated.Use(middleware.JWTAuth())
	authenticated.Use(middleware.PasswordResetRequired())

	// Módulo de gestión de usuarios
	userRoutes := authenticated.Group("/users")
	{
		// Crear usuario - requiere permiso user:create
		userRoutes.POST("", middleware.HasPermission(model.PermUserCreate), authController.CreateUser)
		// Obtener la lista de usuarios - requiere permiso user:read
		userRoutes.GET("", middleware.HasPermission(model.PermUserRead), authController.GetUsers)
		// Actualizar usuario - requiere permiso user:update
		userRoutes.PUT("/:id", middleware.HasPermission(model.PermUserUpdate), authController.UpdateUser)
		// Eliminar usuario - requiere permiso user:delete
		userRoutes.DELETE("/:id", middleware.HasPermission(model.PermUserDelete), authController.DeleteUser)
	}

	// Módulo de gestión de sitios
	siteRoutes := authenticated.Group("/site")
	{
		// Crear sitio - requiere permiso site:create
		siteRoutes.POST("", middleware.HasPermission(model.PermSiteCreate), siteController.CreateSite)
		// Obtener lista de sitios - requiere permiso site:read
		siteRoutes.GET("", middleware.HasPermission(model.PermSiteRead), siteController.GetSites)
		// Obtener un sitio - requiere permiso site:read
		siteRoutes.GET("/:id", middleware.HasPermission(model.PermSiteRead), siteController.GetSiteByID)
		// Actualizar sitio - requiere permiso site:update
		siteRoutes.PUT("/:id", middleware.HasPermission(model.PermSiteUpdate), siteController.UpdateSite)
		// Eliminar sitio - requiere permiso site:delete
		siteRoutes.DELETE("/:id", middleware.HasPermission(model.PermSiteDelete), siteController.DeleteSite)
	}

	// Rutas de gestión de certificados
	certRoutes := authenticated.Group("/certificate")
	{
		certRoutes.POST("", middleware.HasPermission(model.PermCertCreate), certController.CreateCertificate)
		certRoutes.GET("", middleware.HasPermission(model.PermCertRead), certController.GetCertificates)
		certRoutes.GET("/:id", middleware.HasPermission(model.PermCertRead), certController.GetCertificateByID)
		certRoutes.PUT("/:id", middleware.HasPermission(model.PermCertUpdate), certController.UpdateCertificate)
		certRoutes.DELETE("/:id", middleware.HasPermission(model.PermCertDelete), certController.DeleteCertificate)
	}

	// Rutas de gestión de grupos de IPs
	ipGroupRoutes := authenticated.Group("/ip-groups")
	{
		ipGroupRoutes.POST("", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.CreateIPGroup)
		ipGroupRoutes.GET("", middleware.HasPermission(model.PermConfigRead), ipGroupController.GetIPGroups)
		ipGroupRoutes.GET("/:id", middleware.HasPermission(model.PermConfigRead), ipGroupController.GetIPGroupByID)
		ipGroupRoutes.PUT("/:id", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.UpdateIPGroup)
		ipGroupRoutes.DELETE("/:id", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.DeleteIPGroup)
		// Añadir IP a la lista negra por defecto del sistema
		ipGroupRoutes.POST("/blacklist/add", middleware.HasPermission(model.PermConfigUpdate), ipGroupController.AddIPToBlacklist)
	}

	// Rutas de gestión de reglas
	ruleRoutes := authenticated.Group("/micro-rules")
	{
		ruleRoutes.POST("", middleware.HasPermission(model.PermConfigUpdate), ruleController.CreateMicroRule)
		ruleRoutes.GET("", middleware.HasPermission(model.PermConfigRead), ruleController.GetMicroRules)
		ruleRoutes.GET("/:id", middleware.HasPermission(model.PermConfigRead), ruleController.GetMicroRuleByID)
		ruleRoutes.PUT("/:id", middleware.HasPermission(model.PermConfigUpdate), ruleController.UpdateMicroRule)
		ruleRoutes.DELETE("/:id", middleware.HasPermission(model.PermConfigUpdate), ruleController.DeleteMicroRule)
	}

	// Log
	wafLogRoutes := authenticated.Group("/log")
	{
		// Obtener eventos de ataque - requiere permiso logs:read
		wafLogRoutes.GET("/event", middleware.HasPermission(model.PermWAFLogRead), wafLogController.GetAttackEvents)
		// Obtener logs de ataque - requiere permiso logs:read
		wafLogRoutes.GET("", middleware.HasPermission(model.PermWAFLogRead), wafLogController.GetAttackLogs)
	}

	// Rutas de información estadística
	statsRoutes := authenticated.Group("/stats")
	{
		// Obtener estadísticas de resumen - requiere permiso config:read
		statsRoutes.GET("/overview", middleware.HasPermission(model.PermWAFLogRead), statsController.GetOverviewStats)
		// Obtener QPS en tiempo real - requiere permiso config:read
		statsRoutes.GET("/realtime-qps", middleware.HasPermission(model.PermWAFLogRead), statsController.GetRealtimeQPS)
		// Obtener datos de series temporales - requiere permiso config:read
		statsRoutes.GET("/time-series", middleware.HasPermission(model.PermWAFLogRead), statsController.GetTimeSeriesData)
		// Obtener datos de series temporales combinadas - requiere permiso config:read
		statsRoutes.GET("/combined-time-series", middleware.HasPermission(model.PermWAFLogRead), statsController.GetCombinedTimeSeriesData)
		// Obtener datos de series temporales de tráfico - requiere permiso config:read
		statsRoutes.GET("/traffic-time-series", middleware.HasPermission(model.PermWAFLogRead), statsController.GetTrafficTimeSeriesData)
	}

	// Módulo de gestión de configuración
	runnerRoutes := authenticated.Group("/runner")
	{
		// Obtener la configuración - requiere permiso config:read
		runnerRoutes.GET("/status", middleware.HasPermission(model.PermConfigRead), runnerController.GetStatus)
		// Actualizar configuración - requiere permiso config:update
		runnerRoutes.POST("/control", middleware.HasPermission(model.PermConfigUpdate), runnerController.Control)
	}
	configRoutes := authenticated.Group("/config")
	{
		// Obtener la configuración - requiere permiso config:read
		configRoutes.GET("", middleware.HasPermission(model.PermConfigRead), configController.GetConfig)
		// Actualizar configuración - requiere permiso config:update
		configRoutes.PATCH("", middleware.HasPermission(model.PermConfigUpdate), configController.PatchConfig)
	}

	// Módulo de gestión de IPs bloqueadas
	blockedIPRoutes := authenticated.Group("/blocked-ips")
	{
		blockedIPRoutes.GET("", middleware.HasPermission(model.PermConfigRead), blockedIPController.GetBlockedIPs)
		blockedIPRoutes.GET("/stats", middleware.HasPermission(model.PermConfigRead), blockedIPController.GetBlockedIPStats)
		blockedIPRoutes.DELETE("/cleanup", middleware.HasPermission(model.PermConfigUpdate), blockedIPController.CleanupExpiredBlockedIPs)
	}

	// Módulo de logs de auditoría
	auditRoutes := authenticated.Group("/audit")
	{
		// Obtener logs de auditoría - requiere permiso audit:read
		auditRoutes.GET("", middleware.HasPermission(model.PermAuditRead), nil)
	}

	// Módulo de administración del sistema
	systemRoutes := authenticated.Group("/system")
	{
		// Obtener el estado del sistema - requiere permiso system:status
		systemRoutes.GET("/status", middleware.HasPermission(model.PermSystemStatus), nil)
		// Reiniciar el sistema - requiere permiso system:restart
		systemRoutes.POST("/restart", middleware.HasPermission(model.PermSystemRestart), nil)
	}

	// ===== Alojamiento de recursos estáticos del frontend =====
	SetStaticFileRouter(route)
}
