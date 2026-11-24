package constant

import (
	"sync"
)

// Estructura de almacenamiento de constantes, usa sync.Map para seguridad en concurrencia
var (
	constants sync.Map
)

// Set establece una constante; si ya existe se sobrescribe
// Parámetros:
//   - key: clave de la constante
//   - value: valor de la constante
func Set(key string, value interface{}) {
	constants.Store(key, value)
}

// Get obtiene el valor de una constante; si no existe devuelve nil y false
// Parámetros:
//   - key: clave de la constante
//
// Devuelve:
//   - value: valor de la constante
//   - exists: si la constante existe
func Get(key string) (value interface{}, exists bool) {
	return constants.Load(key)
}

// GetString obtiene una constante de cadena; si no existe o el tipo no coincide devuelve el valor por defecto
// Parámetros:
//   - key: clave de la constante
//   - defaultValue: valor por defecto
//
// Devuelve:
//   - valor de cadena de la constante o valor por defecto
func GetString(key string, defaultValue string) string {
	if value, exists := constants.Load(key); exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// GetInt obtiene una constante entera; si no existe o el tipo no coincide devuelve el valor por defecto
// Parámetros:
//   - key: clave de la constante
//   - defaultValue: valor por defecto
//
// Devuelve:
//   - valor entero de la constante o valor por defecto
func GetInt(key string, defaultValue int) int {
	if value, exists := constants.Load(key); exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}

// GetBool obtiene una constante booleana; si no existe o el tipo no coincide devuelve el valor por defecto
// Parámetros:
//   - key: clave de la constante
//   - defaultValue: valor por defecto
//
// Devuelve:
//   - valor booleano de la constante o valor por defecto
func GetBool(key string, defaultValue bool) bool {
	if value, exists := constants.Load(key); exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// GetFloat64 obtiene una constante float64; si no existe o el tipo no coincide devuelve el valor por defecto
// Parámetros:
//   - key: clave de la constante
//   - defaultValue: valor por defecto
//
// Devuelve:
//   - valor de punto flotante de la constante o valor por defecto
func GetFloat64(key string, defaultValue float64) float64 {
	if value, exists := constants.Load(key); exists {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
	}
	return defaultValue
}

// Delete elimina una constante
// Parámetros:
//   - key: clave de la constante a eliminar
func Delete(key string) {
	constants.Delete(key)
}

// InitSystemConstants inicializa las constantes por defecto del sistema
func InitSystemConstants() {
	// Aquí se establecen las constantes por defecto del sistema
	Set("APP_CONFIG_NAME", "AppConfig")
	Set("Default_ENGINE_NAME", "coraza")
}
