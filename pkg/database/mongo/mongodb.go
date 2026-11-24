// mongodb/client.go
package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	client     *mongo.Client
	clientOnce sync.Once
	clientErr  error
)

// Connect conecta a MongoDB según la cadena de conexión, con las mejores prácticas recomendadas
// El cliente devuelto es singleton; llamadas posteriores devuelven la misma instancia
func Connect(uri string) (*mongo.Client, error) {
	// Modo de carga diferida
	clientOnce.Do(func() {
		// Establecer las opciones por defecto
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().
			ApplyURI(uri).
			SetServerAPIOptions(serverAPI).
			SetRetryWrites(true).
			SetRetryReads(true)

		// Crear el cliente y conectar
		c, err := mongo.Connect(opts)
		if err != nil {
			clientErr = fmt.Errorf("Error al conectar con MongoDB: %w", err)
			return
		}

		// Validar la conexión
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.Database("admin").RunCommand(
			ctx,
			bson.D{{Key: "ping", Value: 1}},
		).Err(); err != nil {
			// Desconectar para liberar recursos
			_ = c.Disconnect(context.Background())
			clientErr = fmt.Errorf("Error al verificar la conexión con MongoDB: %w", err)
			return
		}

		client = c
	})

	if clientErr != nil {
		return nil, clientErr
	}

	return client, nil
}

// Disconnect desconecta de MongoDB
func Disconnect() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("Error al desconectar de MongoDB: %w", err)
	}

	// Restablecer el estado del singleton, permite reconectar
	client = nil
	clientOnce = sync.Once{}

	return nil
}

// GetDatabase obtiene la instancia de base de datos con el nombre indicado
func GetDatabase(dbName string) (*mongo.Database, error) {
	if client == nil {
		return nil, fmt.Errorf("MongoDB no conectado, llame primero a Connect")
	}
	return client.Database(dbName), nil
}

// GetCollection obtiene una colección de la base de datos especificada
func GetCollection(dbName, collName string) (*mongo.Collection, error) {
	if client == nil {
		return nil, fmt.Errorf("MongoDB no conectado, llame primero a Connect")
	}
	return client.Database(dbName).Collection(collName), nil
}

// Ping comprueba si la conexión a MongoDB está activa
func Ping() error {
	if client == nil {
		return fmt.Errorf("MongoDB no conectado, llame primero a Connect")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Database("admin").RunCommand(
		ctx,
		bson.D{{Key: "ping", Value: 1}},
	).Err()
}
