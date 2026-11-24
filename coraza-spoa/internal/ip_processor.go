package internal

import (
	"context"
	"net"
	"sync"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog"
)

// IPProcessor es la interfaz del procesador de geolocalización de IPs
type IPProcessor interface {
	// GetIPInfo obtiene información de geolocalización a partir de una cadena de dirección IP
	GetIPInfo(ipStr string) *model.IPInfo

	// Close cierra el procesador y libera recursos
	Close()
}

// GeoIP2Processor es una implementación de procesador de IP basada en la base de datos MaxMind GeoIP2
type GeoIP2Processor struct {
	cityDB *geoip2.Reader  // Base de datos de ciudades
	asnDB  *geoip2.Reader  // Base de datos ASN
	logger zerolog.Logger  // Logger
	mutex  sync.RWMutex    // Bloqueo de lectura-escritura
	ctx    context.Context // Contexto
	closed bool            // Si está cerrado
}

// GeoIP2Options contiene las opciones de configuración del procesador GeoIP2
type GeoIP2Options struct {
	CityDBPath string // Ruta de la base de datos de ciudades
	ASNDBPath  string // Ruta de la base de datos ASN
}

// NewGeoIP2Processor crea una nueva instancia del procesador GeoIP2
func NewGeoIP2Processor(ctx context.Context, options GeoIP2Options, logger zerolog.Logger) (IPProcessor, error) {
	processor := &GeoIP2Processor{
		logger: logger,
		ctx:    ctx,
	}

	// Intentar abrir la base de datos de ciudades
	if options.CityDBPath != "" {
		cityDB, err := geoip2.Open(options.CityDBPath)
		if err != nil {
			logger.Error().Err(err).Str("path", options.CityDBPath).Msg("Error al abrir la base de datos de ciudades")
		} else {
			processor.cityDB = cityDB
		}
	} else {
		logger.Warn().Msg("No se proporcionó la ruta de la base de datos de ciudades, la geolocalización no estará disponible")
	}

	// Intentar abrir la base de datos ASN
	if options.ASNDBPath != "" {
		asnDB, err := geoip2.Open(options.ASNDBPath)
		if err != nil {
			logger.Warn().Err(err).Str("path", options.ASNDBPath).Msg("Error al abrir la base de datos ASN, la información ASN no estará disponible")
		} else {
			processor.asnDB = asnDB
		}
	} else {
		logger.Warn().Msg("No se proporcionó la ruta de la base de datos ASN, la información ASN no estará disponible")
	}

	// Si no se puede abrir ninguna base de datos, devolver implementación vacía
	if processor.cityDB == nil && processor.asnDB == nil {
		return NewNullIPProcessor(), nil
	}

	// Iniciar en segundo plano la goroutine que escucha la señal de cancelación del contexto
	go processor.watchContext()

	return processor, nil
}

// watchContext escucha la señal de cancelación del contexto y libera recursos automáticamente
func (p *GeoIP2Processor) watchContext() {
	<-p.ctx.Done()
	p.closeResources()
}

// closeResources cierra la base de datos y marca el procesador como cerrado
func (p *GeoIP2Processor) closeResources() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return
	}

	// Cerrar la base de datos
	if p.cityDB != nil {
		p.cityDB.Close()
		p.cityDB = nil
	}

	if p.asnDB != nil {
		p.asnDB.Close()
		p.asnDB = nil
	}

	p.closed = true
	p.logger.Debug().Msg("Recursos del procesador de IP liberados")
}

// Close cierra manualmente el procesador de IP
func (p *GeoIP2Processor) Close() {
	p.closeResources()
}

// GetIPInfo obtiene información de geolocalización a partir de una cadena de dirección IP
func (p *GeoIP2Processor) GetIPInfo(ipStr string) *model.IPInfo {
	// Comprobación rápida de validez del formato de IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		p.logger.Warn().Str("ip", ipStr).Msg("Dirección IP no válida")
		return nil
	}

	// Comprobar si el contexto fue cancelado
	select {
	case <-p.ctx.Done():
		return nil
	default:
		// Continuar el procesamiento
	}

	// Si el procesador está cerrado, devolver nil directamente
	if p.closed || (p.cityDB == nil && p.asnDB == nil) {
		return nil
	}

	ipInfo := &model.IPInfo{}
	hasInfo := false

	// Consultar información de la ciudad
	if p.cityDB != nil {
		cityRecord, err := p.cityDB.City(ip)
		if err != nil {
			p.logger.Warn().Str("ip", ipStr).Err(err).Msg("Error al consultar la información de la ciudad")
		} else {
			hasInfo = true

			// Rellenar la información de la ciudad
			if len(cityRecord.City.Names) > 0 {
				if name, exists := cityRecord.City.Names["zh-CN"]; exists {
					ipInfo.City.NameZH = name
				}
				if name, exists := cityRecord.City.Names["en"]; exists {
					ipInfo.City.NameEN = name
				}
			}

			// Rellenar la información de provincia/estado
			if len(cityRecord.Subdivisions) > 0 {
				subdivision := cityRecord.Subdivisions[0]
				if name, exists := subdivision.Names["zh-CN"]; exists {
					ipInfo.Subdivision.NameZH = name
				}
				if name, exists := subdivision.Names["en"]; exists {
					ipInfo.Subdivision.NameEN = name
				}
				ipInfo.Subdivision.IsoCode = subdivision.IsoCode
			}

			// Rellenar la información del país
			if name, exists := cityRecord.Country.Names["zh-CN"]; exists {
				ipInfo.Country.NameZH = name
			}
			if name, exists := cityRecord.Country.Names["en"]; exists {
				ipInfo.Country.NameEN = name
			}
			ipInfo.Country.IsoCode = cityRecord.Country.IsoCode

			// Rellenar la información del continente
			if len(cityRecord.Continent.Names) > 0 {
				if name, exists := cityRecord.Continent.Names["zh-CN"]; exists {
					ipInfo.Continent.NameZH = name
				}
				if name, exists := cityRecord.Continent.Names["en"]; exists {
					ipInfo.Continent.NameEN = name
				}
			}

			// Rellenar la información de ubicación
			ipInfo.Location.Longitude = cityRecord.Location.Longitude
			ipInfo.Location.Latitude = cityRecord.Location.Latitude
			ipInfo.Location.TimeZone = cityRecord.Location.TimeZone
		}
	}

	// Consultar información ASN
	if p.asnDB != nil {
		asnRecord, err := p.asnDB.ASN(ip)
		if err != nil {
			p.logger.Warn().Str("ip", ipStr).Err(err).Msg("Error al consultar la información ASN")
		} else {
			hasInfo = true
			ipInfo.ASN.Number = asnRecord.AutonomousSystemNumber
			ipInfo.ASN.Organization = asnRecord.AutonomousSystemOrganization
		}
	}

	// Si no se obtuvo información, devolver nil
	if !hasInfo {
		return nil
	}

	return ipInfo
}

// NewIPProcessor método fábrica para crear instancias del procesador de IP
func NewIPProcessor(ctx context.Context, cityDBPath, asnDBPath string, logger zerolog.Logger) (IPProcessor, error) {
	if cityDBPath == "" && asnDBPath == "" {
		// Si no se proporcionó ninguna ruta de base de datos, devolver implementación vacía
		return NewNullIPProcessor(), nil
	}

	// Crear el procesador
	return NewGeoIP2Processor(ctx, GeoIP2Options{
		CityDBPath: cityDBPath,
		ASNDBPath:  asnDBPath,
	}, logger)
}

// NullIPProcessor es la implementación vacía de la interfaz IPProcessor
type NullIPProcessor struct{}

// NewNullIPProcessor crea un procesador de IP con implementación vacía
func NewNullIPProcessor() IPProcessor {
	return &NullIPProcessor{}
}

// GetIPInfo devuelve nil; la implementación vacía siempre devuelve nil
func (p *NullIPProcessor) GetIPInfo(ipStr string) *model.IPInfo {
	return nil
}

// Close implementación vacía, no hace nada
func (p *NullIPProcessor) Close() {
	// No requiere ninguna acción
}
