// validators/string_validators.go
package validator

import (
	"net"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Inicializar validadores de cadenas
func init() {
	Register("domain", DomainOrIPValidator)
}

// DomainOrIPValidator valida si una cadena es un dominio o dirección IP válida
var DomainOrIPValidator validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 1. Comprobar si es una dirección IP válida
	if ip := net.ParseIP(value); ip != nil {
		return true
	}

	// 2. Comprobar si es un dominio válido
	// Expresión regular de dominio
	// Reglas:
	// 1. Compuesto por letras, números y guiones; el guion no puede ir al inicio ni al final
	// 2. Cada etiqueta (parte entre puntos) no puede superar 63 caracteres
	// 3. Al menos un punto; la última parte debe tener al menos 2 caracteres (TLD)
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

	return domainRegex.MatchString(value)
}
