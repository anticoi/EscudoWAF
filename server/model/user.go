package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

// Modelo de usuario extendido, con campo de ID de rol añadido
type User struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username    string        `bson:"username" json:"username"`
	Password    string        `bson:"password" json:"-"`                        // La contraseña no se devuelve al cliente
	Role        string        `bson:"role" json:"role"`                         // Rol del usuario
	Permissions []string      `bson:"permissions" json:"permissions,omitempty"` // Permisos adicionales
	NeedReset   bool          `bson:"needReset" json:"needReset"`               // Si se requiere restablecer la contraseña
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
	LastLogin   time.Time     `bson:"lastLogin" json:"lastLogin"`
}

// HashPassword genera el hash de la contraseña
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifica si la contraseña es correcta
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// GetCollectionName devuelve el nombre de la colección
func (u *User) GetCollectionName() string {
	return "user"
}
