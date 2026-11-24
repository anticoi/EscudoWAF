package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	mongodb "github.com/anticoi/EscudoWAF/pkg/database/mongo"
	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/constant"
	"github.com/anticoi/EscudoWAF/server/utils/jwt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Global instancia de configuración global
var Global Config

// Config guarda la configuración de la aplicación
type Config struct {
	Bind         string
	IsProduction bool
	IsK8s        bool
	DisableWeb   bool
	WebPath      string
	Log          LogConfig
	DBConfig     DBConfig
	JWT          JWTConfig
}

// DBConfig configuración de base de datos
type DBConfig struct {
	URI      string
	Database string
}

// JWTConfig configuración JWT
type JWTConfig struct {
	Secret        string
	ExpirationHrs int
}

// InitConfig inicializa la configuración desde variables de entorno
func InitConfig() error {
	// Cargar el archivo .env
	err := godotenv.Load()
	if err != nil {
		// Si el archivo .env no existe, solo registrar información, sin devolver error
		GlobalLogger.Info().Msg(".env file not found, using default environment variables")
	}

	// Establecer valores por defecto
	Global = Config{
		Bind:         "0.0.0.0:2333",
		IsProduction: false,
		IsK8s:        false,
		DisableWeb:   false, // La función web está habilitada por defecto
		WebPath:      "",    // Usar recursos de frontend embebidos
		Log: LogConfig{
			Level:  "info",
			File:   "/dev/stdout",
			Format: "console",
		},
		DBConfig: DBConfig{
			URI:      "",
			Database: "waf",
		},
		JWT: JWTConfig{
			Secret:        "default-jwt-secret-key",
			ExpirationHrs: 24,
		},
	}

	// Cargar configuración desde variables de entorno
	if env := os.Getenv("BIND"); env != "" {
		Global.Bind = env
	}

	if env := os.Getenv("IS_PRODUCTION"); env != "" {
		Global.IsProduction = env == "true"
	}

	if env := os.Getenv("IS_K8S"); env != "" {
		Global.IsK8s = env == "true"
	}

	// Configuración de logs
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		Global.Log.Level = env
	}
	if env := os.Getenv("LOG_FILE"); env != "" {
		Global.Log.File = env
	}
	if env := os.Getenv("LOG_FORMAT"); env != "" {
		Global.Log.Format = env
	}

	// Configuración de base de datos
	if env := os.Getenv("DB_URI"); env != "" {
		Global.DBConfig.URI = env
	}
	if env := os.Getenv("DB_NAME"); env != "" {
		Global.DBConfig.Database = env
	}

	// Configuración JWT
	if env := os.Getenv("JWT_SECRET"); env != "" {
		Global.JWT.Secret = env
	}
	if env := os.Getenv("JWT_EXPIRATION_HRS"); env != "" {
		if hrs, err := strconv.Atoi(env); err == nil {
			Global.JWT.ExpirationHrs = hrs
		}
	}

	// Configuración del frontend
	if env := os.Getenv("DISABLE_WEB"); env != "" {
		Global.DisableWeb = env == "true"
	}
	if env := os.Getenv("WEB_PATH"); env != "" {
		Global.WebPath = env
	}

	// Inicializar JWT
	err = jwt.InitJWTSecret(Global.JWT.Secret)
	if err != nil {
		return fmt.Errorf("failed to initialize JWT: %w", err)
	}

	// Inicializar el logger
	Logger, err = Global.Log.newLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	Logger.Info().Msg("✨ Application configure loaded successfully")
	return nil
}

func GetAppConfig() (*model.Config, error) {
	// Conectar a la base de datos
	client, err := mongodb.Connect(Global.DBConfig.URI)
	if err != nil {
		return nil, err
	}

	var cfg model.Config
	// Obtener la colección de configuración
	db := client.Database(Global.DBConfig.Database)
	collection := db.Collection(cfg.GetCollectionName())

	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Asegurar que los recursos se liberen

	// Usar constante para el nombre de configuración; si no existe usar "AppConfig" por defecto
	configName := constant.GetString("APP_CONFIG_NAME", "AppConfig")

	// Consultar la configuración con el nombre indicado
	err = collection.FindOne(
		ctx,
		bson.D{{Key: "name", Value: configName}},
	).Decode(&cfg)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("Registro de configuración no encontrado")
		}
		return nil, fmt.Errorf("Error al obtener la configuración: %w", err)
	}
	return &cfg, nil
}
