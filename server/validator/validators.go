// validators/validators.go
package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Mapa global de validadores
var customValidators = map[string]validator.Func{}

// Registrar un validador personalizado
func Register(tag string, fn validator.Func) {
	customValidators[tag] = fn
}

// Inicializar todos los validadores personalizados
func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Registrar todos los validadores
		for tag, fn := range customValidators {
			_ = v.RegisterValidation(tag, fn)
		}
	}
}
