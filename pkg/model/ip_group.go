package model

import "go.mongodb.org/mongo-driver/v2/bson"

// IPGroup representa la información de un grupo de IPs
// @Description Información del grupo de IPs, incluye nombre del grupo y lista de IPs
type IPGroup struct {
	ID    bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"60d21b4367d0d8992e89e964"` // Identificador único del grupo
	Name  string        `bson:"name" json:"name" example:"Servidor interno"`                                     // Nombre del grupo
	Items []string      `bson:"items" json:"items" example:"['192.168.1.1', '10.0.0.1/24']"`          // Lista de direcciones IP o CIDR
}

func (i *IPGroup) GetCollectionName() string {
	return "ip_group"
}
