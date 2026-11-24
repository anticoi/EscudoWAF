package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Mapeo de tipos de estructura a funciones de validación
var structValidators = map[interface{}]validator.StructLevelFunc{}

// Registrar validadores a nivel de estructura
func RegisterStructValidator(structType interface{}, fn validator.StructLevelFunc) {
	// Guardar el validador en el mapa
	structValidators[structType] = fn
}

// Inicializar todos los validadores a nivel de estructura
func InitStructValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Registrar todos los validadores de estructuras desde el mapa
		for structType, fn := range structValidators {
			v.RegisterStructValidation(fn, structType)
		}
	}
}
