package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/anticoi/EscudoWAF/server/model"
	"github.com/golang-jwt/jwt/v5"
)

// Definición de errores JWT
var (
	ErrInvalidToken = errors.New("Token no válido")
	ErrExpiredToken = errors.New("Token expirado")
)

// Definición de la clave; en producción debe leerse de la configuración
var jwtSecret []byte

// Inicializar la clave JWT
func InitJWTSecret(secret string) error {
	if secret == "" {
		return fmt.Errorf("Clave JWT no configurada")
	}
	jwtSecret = []byte(secret)
	return nil
}

// Claims estructura de claims JWT personalizados
type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken genera un token JWT
func GenerateToken(user model.User, expiration time.Duration) (string, error) {
	// Establecer los claims del JWT
	claims := Claims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "EscudoWAF",
			Subject:   user.ID.Hex(),
		},
	}

	// Crear el token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token
	return token.SignedString(jwtSecret)
}

// ParseToken analiza un token JWT
func ParseToken(tokenString string) (*Claims, error) {
	// Analizar el token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validar el método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		// Manejar errores específicos
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Extraer los claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
