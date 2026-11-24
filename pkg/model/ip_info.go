// model/ip_info.go
package model

// IPInfo representa la información de geolocalización de una dirección IP
// @Description Información detallada de geolocalización de la IP, incluyendo ciudad, región, país y datos ASN
type IPInfo struct {
	City struct {
		NameZH string `json:"nameZh" bson:"nameZh" example:"Hangzhou"`       // Nombre de la ciudad en chino
		NameEN string `json:"nameEn" bson:"nameEn" example:"Hangzhou"` // Nombre de la ciudad en inglés
	} `json:"city" bson:"city"`

	Subdivision struct {
		NameZH  string `json:"nameZh" bson:"nameZh" example:"Zhejiang"`       // Nombre de la provincia/estado en chino
		NameEN  string `json:"nameEn" bson:"nameEn" example:"Zhejiang"` // Nombre de la provincia/estado en inglés
		IsoCode string `json:"isoCode" bson:"isoCode" example:"ZJ"`     // Código de la provincia/estado
	} `json:"subdivision" bson:"subdivision"`

	Country struct {
		NameZH  string `json:"nameZh" bson:"nameZh" example:"China"`    // Nombre del país en chino
		NameEN  string `json:"nameEn" bson:"nameEn" example:"China"` // Nombre del país en inglés
		IsoCode string `json:"isoCode" bson:"isoCode" example:"CN"`  // Código ISO del país
	} `json:"country" bson:"country"`

	Continent struct {
		NameZH string `json:"nameZh" bson:"nameZh" example:"Asia"`   // Nombre del continente en chino
		NameEN string `json:"nameEn" bson:"nameEn" example:"Asia"` // Nombre del continente en inglés
	} `json:"continent" bson:"continent"`

	Location struct {
		Longitude float64 `json:"longitude" bson:"longitude" example:"120.1663"`    // Longitud
		Latitude  float64 `json:"latitude" bson:"latitude" example:"30.2943"`       // Latitud
		TimeZone  string  `json:"timeZone" bson:"timeZone" example:"Asia/Shanghai"` // Zona horaria
	} `json:"location" bson:"location"`

	ASN struct {
		Number       uint   `json:"number" bson:"number" example:"4134"`                      // Número ASN
		Organization string `json:"organization" bson:"organization" example:"China Telecom"` // Nombre de la organización
	} `json:"asn" bson:"asn"`
}
