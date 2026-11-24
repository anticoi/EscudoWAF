package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
)

// GlobalLogger se usa cuando Logger no está inicializado
var GlobalLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()

// Logger instancia global del logger
var Logger zerolog.Logger

type LogConfig struct {
	Level  string
	File   string
	Format string
}

func (lc LogConfig) outputWriter() (io.Writer, error) {
	var out io.Writer
	if lc.File == "" || lc.File == "/dev/stdout" {
		out = os.Stdout
	} else if lc.File == "/dev/stderr" {
		out = os.Stderr
	} else if lc.File == "/dev/null" {
		out = io.Discard
	} else {
		f, err := os.OpenFile(lc.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		out = f
	}
	return out, nil
}

func (lc LogConfig) newLogger() (zerolog.Logger, error) {
	out, err := lc.outputWriter()
	if err != nil {
		return GlobalLogger, err
	}

	switch lc.Format {
	case "console":
		out = zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: "2006-01-02 15:04:05",
			FormatLevel: func(i interface{}) string {
				if ll, ok := i.(string); ok {
					switch ll {
					case "debug":
						return "\x1b[36m[DEBUG]\x1b[0m" // Cian
					case "info":
						return "\x1b[32m[ INFO]\x1b[0m" // Verde
					case "warn":
						return "\x1b[33m[ WARN]\x1b[0m" // Amarillo
					case "error":
						return "\x1b[31m[ERROR]\x1b[0m" // Rojo
					case "fatal":
						return "\x1b[35m[FATAL]\x1b[0m" // Púrpura
					}
				}
				return fmt.Sprintf("[%-5s]", i)
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("\x1b[1m%s\x1b[0m", i) // Mensaje en negrita
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("\x1b[34m%s:\x1b[0m", i) // Nombres de campo en azul
			},
			FormatFieldValue: func(i interface{}) string {
				return fmt.Sprintf("\x1b[37m%s\x1b[0m", i) // Valores de campo en gris claro
			},
		}
	case "json":
		// Usar el formato JSON por defecto
	default:
		return GlobalLogger, fmt.Errorf("unknown log format: %v", lc.Format)
	}

	if lc.Level == "" {
		lc.Level = "info"
	}
	lvl, err := zerolog.ParseLevel(lc.Level)
	if err != nil {
		return GlobalLogger, err
	}

	return zerolog.New(out).Level(lvl).With().Timestamp().Logger(), nil
}

// Crear un logger con información del llamante basado en el Logger global
// GetLogger obtiene un logger con información del llamante
func GetLogger() zerolog.Logger {
	// Obtener la información del llamador
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.With().Str("source", "unknown").Logger()
	}

	// Extraer el nombre de archivo (sin la ruta)
	fileName := filepath.Base(file)

	return Logger.With().
		Str("file", fileName).
		Int("line", line).
		Logger()
}

// GetServiceLogger obtiene el logger de la capa de servicios
func GetServiceLogger(serviceName string) zerolog.Logger {
	_, file, line, _ := runtime.Caller(1)
	fileName := filepath.Base(file)

	return Logger.With().
		Str("layer", "service").
		Str("service", serviceName).
		Str("file", fileName).
		Int("line", line).
		Logger()
}

// GetControllerLogger obtiene el logger de la capa de controladores
func GetControllerLogger(controllerName string) zerolog.Logger {
	_, file, line, _ := runtime.Caller(1)
	fileName := filepath.Base(file)

	return Logger.With().
		Str("layer", "controller").
		Str("controller", controllerName).
		Str("file", fileName).
		Int("line", line).
		Logger()
}

// GetRepositoryLogger obtiene el logger de la capa de acceso a datos
func GetRepositoryLogger(repoName string) zerolog.Logger {
	_, file, line, _ := runtime.Caller(1)
	fileName := filepath.Base(file)

	return Logger.With().
		Str("layer", "repository").
		Str("repository", repoName).
		Str("file", fileName).
		Int("line", line).
		Logger()
}
