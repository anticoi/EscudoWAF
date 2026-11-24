package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/anticoi/EscudoWAF/server/constant"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitDB(db *mongo.Database) error {
	if err := initConfig(db); err != nil {
		return err
	}

	if err := initWAFLog(db); err != nil {
		return err
	}

	return nil
}

func initConfig(db *mongo.Database) error {
	// Comprobar si existe la colección de configuración
	var cfg model.Config
	configCollection := db.Collection(cfg.GetCollectionName())

	// Comprobar si hay registro de configuración - usa sintaxis v2
	filter := bson.D{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := configCollection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to count documents: %w", err)
	}

	// Solo crear la configuración por defecto cuando no hay registro de configuración
	if count == 0 {
		defaultConfig := createDefaultConfig()
		_, err = configCollection.InsertOne(ctx, defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to insert default config: %w", err)
		}
		Logger.Info().Msg("Created default configuration")
	} else {
		Logger.Info().Int64("count", count).Msg("Found existing configuration documents in database, skip initialization")
	}

	return nil
}

// Crear la configuración por defecto
func createDefaultConfig() model.Config {
	now := time.Now()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Logger.Error().Err(err).Msg("No se pudo obtener el directorio home del usuario")
	}

	return model.Config{
		Name: constant.GetString("APP_CONFIG_NAME", "AppConfig"),
		Engine: model.EngineConfig{
			Bind:            "127.0.0.1:2342",
			UseBuiltinRules: true,
			ASNDBPath:       filepath.Join(homeDir, "escudo-waf", "geo-ip", "GeoLite2-ASN.mmdb"),
			CityDBPath:      filepath.Join(homeDir, "escudo-waf", "geo-ip", "GeoLite2-City.mmdb"),
			FlowController:  model.GetDefaultFlowControlConfig(),
			AppConfig: []model.AppConfig{
				{
					Name: constant.GetString("Default_ENGINE_NAME", "coraza"),
					Directives: `SecAction \
    "id:20001,\
    phase:1,\
    nolog,\
    pass,\
    t:none,\
    setvar:'tx.allowed_methods=GET HEAD POST OPTIONS PUT DELETE PATCH'"

Include @coraza.conf-recommended
Include @crs-setup.conf.example
Include @owasp_crs/*.conf
SecRuleEngine On

SecRuleUpdateTargetById 933120 !ARGS:json.engine.appConfig.0.directives`,
					// The transaction cache lifetime in milliseconds (60000ms = 60s)
					TransactionTTL: 60000,
					LogLevel:       "info",
					LogFile:        "/dev/stdout",
					LogFormat:      "console",
				},
			},
		},
		Haproxy: model.HaproxyConfig{
			ConfigBaseDir: filepath.Join(homeDir, "escudo-waf"),
			HaproxyBin:    "haproxy",
			BackupsNumber: 5,
			SpoeAgentAddr: "127.0.0.1",
			SpoeAgentPort: 2342,
			Thread:        0,
		},
		CreatedAt:       now,
		UpdatedAt:       now,
		IsResponseCheck: false,
		IsDebug:         !Global.IsProduction,
		IsK8s:           Global.IsK8s,
	}
}

func initWAFLog(db *mongo.Database) error {
	// Obtener el nombre de la colección de logs del WAF
	var wafLog model.WAFLog
	collectionName := wafLog.GetCollectionName()

	// Crear el contexto
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtener los nombres de todas las colecciones
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if slices.Contains(collections, collectionName) {
		return nil
	}

	// La colección no existe, hay que crear la colección y los índices
	// Crear contexto; la creación de índices puede tardar más
	indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer indexCancel()

	// Acceder a la colección (se crea vacía si no existe)
	wafLogCollection := db.Collection(collectionName)

	// Crear modelo de índices - optimiza consultas de rangos temporales grandes
	indexModels := []mongo.IndexModel{
		{
			// Índice principal de series temporales - optimiza consultas por rango de tiempo
			Keys: bson.D{
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().SetName("idx_createdAt"),
		},
		{
			// Índice compuesto de IP de origen y tiempo - optimiza consultas de estadísticas por IP
			Keys: bson.D{
				{Key: "srcIp", Value: 1},
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().SetName("idx_srcIp_createdAt"),
		},
		{
			// Índice de IP de destino
			Keys: bson.D{
				{Key: "dstIp", Value: 1},
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().SetName("idx_dstIp_createdAt"),
		},
		{
			// Índice de dominio
			Keys: bson.D{
				{Key: "domain", Value: 1},
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().SetName("idx_domain_createdAt"),
		},
		{
			// Índice de grupos de 6 horas - optimiza consultas agregadas de rangos medios
			Keys: bson.D{
				{Key: "createdAt", Value: 1},
				{Key: "date", Value: 1},
				{Key: "hour", Value: 1},
				{Key: "hourGroupSix", Value: 1},
			},
			Options: options.Index().SetName("idx_time_series_6hour"),
		},
	}

	// Crear índices
	_, err = wafLogCollection.Indexes().CreateMany(indexCtx, indexModels)
	if err != nil {
		return fmt.Errorf("failed to create indexes for waf_log collection: %w", err)
	}

	return nil
}
