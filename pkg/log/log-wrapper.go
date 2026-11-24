package log

import (
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
)

// LoggerWrapper es un wrapper de zerolog.Logger
type LoggerWrapper struct {
	logger zerolog.Logger
}

// NewLoggerWrapper crea un nuevo LoggerWrapper
func NewLoggerWrapper(logger zerolog.Logger) *LoggerWrapper {
	return &LoggerWrapper{
		logger: logger,
	}
}

// Debug crea un log de nivel Debug con información de la ubicación de la llamada
func (lw *LoggerWrapper) Debug() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Debug().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Debug().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Info crea un log de nivel Info con información de la ubicación de la llamada
func (lw *LoggerWrapper) Info() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Info().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Info().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Warn crea un log de nivel Warn con información de la ubicación de la llamada
func (lw *LoggerWrapper) Warn() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Warn().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Warn().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Error crea un log de nivel Error con información de la ubicación de la llamada
func (lw *LoggerWrapper) Error() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Error().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Error().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Fatal crea un log de nivel Fatal con información de la ubicación de la llamada
func (lw *LoggerWrapper) Fatal() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return lw.logger.Fatal().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return lw.logger.Fatal().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// example
// func example() {
// 	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
// 	log := NewLoggerWrapper(logger)
// 	log.Debug().Msg("debug message")
// 	log.Info().Msg("info message")
// }
