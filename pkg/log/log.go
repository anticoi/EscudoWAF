package log

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
)

var Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

// LoggerWithCaller devuelve un hook de log con la ubicación del llamante
func LoggerWithCaller() zerolog.Logger {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.With().Str("source", "unknown").Logger()
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.With().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName).
		Logger()
}

// Debug crea un log de nivel Debug con ubicación de llamada
func Debug() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Debug().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Debug().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Info crea un log de nivel Info con ubicación de llamada
func Info() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Info().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Info().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Warn crea un log de nivel Warn con ubicación de llamada
func Warn() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Warn().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Warn().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Error crea un log de nivel Error con ubicación de llamada
func Error() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Error().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Error().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}

// Fatal crea un log de nivel Fatal con ubicación de llamada
func Fatal() *zerolog.Event {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return Logger.Fatal().Str("source", "unknown")
	}
	fileName := filepath.Base(file)
	funcName := runtime.FuncForPC(pc).Name()
	return Logger.Fatal().
		Str("file", fileName).
		Int("line", line).
		Str("func", funcName)
}
