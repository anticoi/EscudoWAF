package internal

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	flowcontroller "github.com/anticoi/EscudoWAF/coraza-spoa/internal/flow-controller"
	"github.com/anticoi/EscudoWAF/pkg/model"
	coreruleset "github.com/corazawaf/coraza-coreruleset"
	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/debuglog"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/dropmorepackets/haproxy-go/pkg/encoding"
	"github.com/jcchavezs/mergefs"
	"github.com/jcchavezs/mergefs/io"
	"github.com/rs/zerolog"
	"istio.io/istio/pkg/cache"
)

// Configuración de MongoDB
type MongoConfig struct {
	Client     *mongo.Client
	Database   string
	Collection string
}

type AppConfig struct {
	Directives     string
	ResponseCheck  bool
	Logger         zerolog.Logger
	TransactionTTL time.Duration
}

// ApplicationOptions opciones de configuración de la aplicación: activar resolución de IP, registro de logs
type ApplicationOptions struct {
	MongoConfig          *MongoConfig          // Configuración de MongoDB, para almacenamiento de logs
	GeoIPConfig          *GeoIP2Options        // Configuración GeoIP, para el procesamiento de geolocalización de IPs
	RuleEngineDbConfig   *MongoDBConfig        // Configuración de base de datos del motor de reglas
	FlowControllerConfig *FlowControllerConfig // Configuración del controlador de flujo
}

// FlowControllerConfig configuración del controlador de flujo
type FlowControllerConfig struct {
	Client   *mongo.Client // Cliente de MongoDB
	Database string        // Nombre de la base de datos
}

type Application struct {
	waf            coraza.WAF
	cache          cache.ExpiringCache
	logStore       LogStore
	ipProcessor    IPProcessor
	ruleEngine     *RuleEngine
	flowController *flowcontroller.FlowController
	ipRecorder     flowcontroller.IPRecorder

	AppConfig
}

// Estructura transaction extendida, con información de la petición añadida
type transaction struct {
	tx      types.Transaction
	m       sync.Mutex
	request *applicationRequest // Almacenar la información de la petición
}

type applicationRequest struct {
	SrcIp   netip.Addr
	SrcPort int64
	DstIp   netip.Addr
	DstPort int64
	Method  string
	ID      string
	Path    []byte
	Query   []byte
	Version string
	Headers []byte
	Body    []byte
}

func (a *Application) HandleRequest(ctx context.Context, writer *encoding.ActionWriter, message *encoding.Message) (err error) {
	k := encoding.AcquireKVEntry()
	// run defer via anonymous function to not directly evaluate the arguments.
	defer func() {
		encoding.ReleaseKVEntry(k)
	}()

	// parse request
	var req applicationRequest
	for message.KV.Next(k) {
		switch name := string(k.NameBytes()); name {
		case "src-ip":
			req.SrcIp = k.ValueAddr()
		case "src-port":
			req.SrcPort = k.ValueInt()
		case "dst-ip":
			req.DstIp = k.ValueAddr()
		case "dst-port":
			req.DstPort = k.ValueInt()
		case "method":
			req.Method = string(k.ValueBytes())
		case "path":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Path = currK.ValueBytes()

			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "query":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Query = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "version":
			req.Version = string(k.ValueBytes())
		case "headers":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Headers = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "body":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			req.Body = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "id":
			req.ID = string(k.ValueBytes())
		default:
			a.Logger.Debug().Str("name", name).Msg("unknown kv entry")
		}
	}

	if len(req.ID) == 0 {
		const idLength = 16
		var sb strings.Builder
		sb.Grow(idLength)
		for i := 0; i < idLength; i++ {
			sb.WriteRune(rune('A' + rand.Intn(26)))
		}
		req.ID = sb.String()
	}

	realIP := getRealClientIP(&req)
	// Comprobar si la IP ya está restringida
	if a.ipRecorder != nil {
		if blocked, record := a.ipRecorder.IsIPBlocked(realIP); blocked {
			a.Logger.Info().
				Str("ip", realIP).
				Str("reason", record.Reason).
				Time("blocked_until", record.BlockedUntil).
				Msg("Petición rechazada: la IP está restringida")

			return ErrInterrupted{
				Interruption: &types.Interruption{
					Action: "deny",
					Status: 403,
					Data:   fmt.Sprintf("IP has been blocked until %s due to %s", record.BlockedUntil.Format(time.RFC3339), record.Reason),
				},
			}
		}
	}

	host := getHostFromRequest(&req)
	// Realizar la comprobación de acceso de alta frecuencia
	if a.flowController != nil {
		allowed, err := a.flowController.CheckVisit(realIP, buildFullURL(host, req.Path, req.Query))
		if err != nil {
			a.Logger.Error().Err(err).Str("ip", realIP).Msg("Falló la comprobación de control de flujo")
		} else if !allowed {
			return ErrInterrupted{
				Interruption: &types.Interruption{
					Action: "deny",
					Status: 429,
					Data:   "Too many requests",
				},
			}
		}
	}

	// micro engine detection
	if a.ruleEngine != nil {
		realIP := getRealClientIP(&req)
		// Obtener la parte de la ruta
		path := string(req.Path)

		url := buildURLFromBytes(req.Path, req.Query)

		shouldBlock, _, rule, err := a.ruleEngine.MatchRequest(realIP, url, path)

		if err != nil {
			a.Logger.Error().Err(err).
				Str("url", url).
				Str("clientIP", realIP).
				Msg("failed to match request")
		}

		ruleName := "whitelist block"
		ruleId := "none"
		if rule != nil {
			ruleName = rule.Name
			ruleId = rule.ID.String()
		}

		if shouldBlock && err == nil {
			// Registrar el ataque
			if a.flowController != nil {
				_, _ = a.flowController.RecordAttack(realIP, buildFullURL(host, req.Path, req.Query))
			}

			a.Logger.Info().
				Str("ruleName", ruleName).
				Str("ruleId", ruleId).
				Str("url", url).
				Str("clientIP", realIP).
				Msg("request blocked by micro engine")

			err := a.saveMicroEngineLog(rule, &req, req.Headers)
			if err != nil {
				a.Logger.Error().Err(err).
					Str("ruleName", ruleName).
					Str("ruleId", ruleId).
					Str("url", url).
					Str("clientIP", realIP).
					Msg("failed to save micro engine log")
			}

			return ErrInterrupted{
				Interruption: &types.Interruption{
					Action: "deny",
					Status: 403,
				},
			}
		}
	}

	tx := a.waf.NewTransactionWithID(req.ID)
	defer func() {
		if err == nil && a.ResponseCheck {
			// Almacenar la transacción y la información de la petición en caché
			txCache := &transaction{
				tx:      tx,
				request: &req, // Almacenar la información de la petición
			}
			a.cache.SetWithExpiration(tx.ID(), txCache, a.TransactionTTL)
			return
		}

		// Manejar interrupciones y registro de logs
		if tx.IsInterrupted() && a.logStore != nil {
			// Registrar el ataque
			if a.flowController != nil {
				_, _ = a.flowController.RecordAttack(realIP, buildFullURL(host, req.Path, req.Query))
			}

			interruption := tx.Interruption()
			if matchedRules := tx.MatchedRules(); len(matchedRules) > 0 {
				err := a.saveFirewallLog(matchedRules, interruption, &req, req.Headers)
				if err != nil {
					a.Logger.Error().Err(err).Msg("failed to save firewall log")
				}
			}
		}

		tx.ProcessLogging()
		if err := tx.Close(); err != nil {
			a.Logger.Error().Str("tx", tx.ID()).Err(err).Msg("failed to close transaction")
		}
	}()

	// Establecer el response id como id de transacción, para soportar la inspección de respuestas
	if err := writer.SetString(encoding.VarScopeTransaction, "id", tx.ID()); err != nil {
		return err
	}

	if tx.IsRuleEngineOff() {
		a.Logger.Warn().Msg("Rule engine is Off, Coraza is not going to process any rule")
		return nil
	}

	tx.ProcessConnection(req.SrcIp.String(), int(req.SrcPort), req.DstIp.String(), int(req.DstPort))

	tx.ProcessURI(buildURLFromBytes(req.Path, req.Query), req.Method, "HTTP/"+req.Version)

	if err := readHeaders(req.Headers, tx.AddRequestHeader); err != nil {
		return fmt.Errorf("reading headers: %v", err)
	}

	if it := tx.ProcessRequestHeaders(); it != nil {
		return ErrInterrupted{it}
	}

	switch it, _, err := tx.WriteRequestBody(req.Body); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

	switch it, err := tx.ProcessRequestBody(); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

	return nil
}

// readHeaders parses HTTP headers with optimized performance while maintaining correctness
func readHeaders(headers []byte, callback func(key string, value string)) error {
	if len(headers) == 0 {
		return nil
	}

	start := 0
	length := len(headers)

	for start < length {
		// Buscar el fin de línea
		end := start
		for end < length && headers[end] != '\n' {
			end++
		}

		// Obtener la línea actual
		line := headers[start:end]

		// Manejar \r\n (eliminar \r)
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		// Omitir líneas vacías
		if len(line) == 0 {
			start = end + 1
			continue
		}

		// Buscar los dos puntos
		colonPos := -1
		for i, b := range line {
			if b == ':' {
				colonPos = i
				break
			}
		}

		if colonPos == -1 {
			return fmt.Errorf("invalid header: %q", string(line))
		}

		// Extraer y recortar clave y valor
		keyBytes := line[:colonPos]
		valueBytes := line[colonPos+1:]

		keyStart, keyEnd := trimSpaceIndices(keyBytes)
		valueStart, valueEnd := trimSpaceIndices(valueBytes)

		// Comprobar si la clave está vacía (consistente con la versión original)
		if keyStart >= keyEnd {
			// Para clave vacía, el bytes.SplitN original no da error, crea una cadena vacía
			// Mantenemos también este comportamiento
			key := ""
			value := string(valueBytes[valueStart:valueEnd])
			callback(key, value)
		} else {
			key := string(keyBytes[keyStart:keyEnd])
			value := string(valueBytes[valueStart:valueEnd])
			callback(key, value)
		}

		// Pasar a la siguiente línea
		start = end + 1
	}

	return nil
}

// trimSpaceIndices devuelve los índices de inicio y fin tras eliminar espacios en blanco
// Evitar asignación de memoria, devolver solo el índice
func trimSpaceIndices(data []byte) (start, end int) {
	// Omitir espacios iniciales
	start = 0
	for start < len(data) && isWhitespace(data[start]) {
		start++
	}

	// Omitir espacios finales
	end = len(data)
	for end > start && isWhitespace(data[end-1]) {
		end--
	}

	return start, end
}

// isWhitespace comprueba si el carácter es un espacio en blanco (igual que bytes.TrimSpace)
func isWhitespace(b byte) bool {
	// Incluye todos los caracteres en blanco que maneja bytes.TrimSpace
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
}

type applicationResponse struct {
	ID      string
	Version string
	Status  int64
	Headers []byte
	Body    []byte
}

func (a *Application) HandleResponse(ctx context.Context, writer *encoding.ActionWriter, message *encoding.Message) (err error) {
	if !a.ResponseCheck {
		return fmt.Errorf("got response but response check is disabled")
	}

	k := encoding.AcquireKVEntry()
	// run defer via anonymous function to not directly evaluate the arguments.
	defer func() {
		encoding.ReleaseKVEntry(k)
	}()

	var res applicationResponse
	for message.KV.Next(k) {
		switch name := string(k.NameBytes()); name {
		case "id":
			res.ID = string(k.ValueBytes())
		case "version":
			res.Version = string(k.ValueBytes())
		case "status":
			res.Status = k.ValueInt()
		case "headers":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			res.Headers = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		case "body":
			// make a copy of the pointer and add a defer in case there is another entry
			currK := k
			// run defer via anonymous function to not directly evaluate the arguments.
			defer func() {
				encoding.ReleaseKVEntry(currK)
			}()

			res.Body = currK.ValueBytes()
			// acquire a new kv entry to continue reading other message values.
			k = encoding.AcquireKVEntry()
		default:
			a.Logger.Debug().Str("name", name).Msg("unknown kv entry")
		}
	}

	if res.ID == "" {
		return fmt.Errorf("response id is empty")
	}

	cv, ok := a.cache.Get(res.ID)
	if !ok {
		a.Logger.Error().Str("id", res.ID).Msg("transaction not found")
		return nil
		// TODO: decidir si reportar error o solo registrar; al reiniciar el detector esto falla porque application se reemplaza y a.cache.Get no encuentra la transacción de res.ID
		// return fmt.Errorf("transaction not found: %s", res.ID)
	}
	a.cache.Remove(res.ID)

	t := cv.(*transaction)
	if !t.m.TryLock() {
		return fmt.Errorf("transaction is already being deleted: %s", res.ID)
	}
	/*
		En efecto no hace falta defer t.m.Unlock(): llegar a TryLock implica que a.cache.Remove(res.ID) se ejecutó,
		tx siempre se elimina; TryLock falla en dos casos: uno cuando el caché ya lo recogió y tx fue reciclado,
		El otro caso es que otra goroutine lo obtuvo; quien no lo obtiene termina directamente, dejando que la que sí lo tiene lo procese, garantizando que la respuesta se procese una sola vez
	*/
	tx := t.tx

	// Obtener la IP real del cliente
	realIP := getRealClientIP(t.request)
	host := getHostFromRequest(t.request)
	if res.Status >= 400 {
		// Comprobar las respuestas de error y registrarlas
		// Registrar el error
		if a.flowController != nil {
			_, _ = a.flowController.RecordError(realIP, buildFullURL(host, t.request.Path, t.request.Query))
		}
	}

	defer func() {
		// Manejar interrupciones y registro de logs
		if tx.IsInterrupted() && a.logStore != nil {
			// Registrar el ataque
			if a.flowController != nil {
				_, _ = a.flowController.RecordAttack(realIP, buildFullURL(host, t.request.Path, t.request.Query))
			}

			interruption := tx.Interruption()
			if matchedRules := tx.MatchedRules(); len(matchedRules) > 0 && t.request != nil {
				err := a.saveFirewallLog(matchedRules, interruption, t.request, t.request.Headers)
				if err != nil {
					a.Logger.Error().Err(err).Msg("failed to save firewall log")
				}
			}
		}

		tx.ProcessLogging()
		if err := tx.Close(); err != nil {
			a.Logger.Error().Str("tx", tx.ID()).Err(err).Msg("failed to close transaction")
		}
	}()

	if tx.IsRuleEngineOff() {
		goto exit
	}

	if err := readHeaders(res.Headers, tx.AddResponseHeader); err != nil {
		return fmt.Errorf("reading headers: %v", err)
	}

	if it := tx.ProcessResponseHeaders(int(res.Status), "HTTP/"+res.Version); it != nil {
		return ErrInterrupted{it}
	}

	switch it, _, err := tx.WriteResponseBody(res.Body); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

	switch it, err := tx.ProcessResponseBody(); {
	case err != nil:
		return err
	case it != nil:
		return ErrInterrupted{it}
	}

exit:
	return nil
}

// Construir la cadena de petición HTTP
func buildRequestString(req *applicationRequest, headers []byte) string {
	// Precalcular la capacidad total
	capacity := len(req.Method) + 1 + // Method + space
		len(req.Path) + // Path
		len(req.Version) + 6 + // " HTTP/" + version
		1 + // \n
		len(headers) // headers

	if req.Query != nil {
		capacity += 1 + len(req.Query) // ? + query
	}

	if len(req.Body) > 0 {
		capacity += 1 + len(req.Body) // \n + body
	}

	// Inicializar el Builder con capacidad precalculada
	var sb strings.Builder
	sb.Grow(capacity)

	// Construir la cadena de la petición
	sb.WriteString(req.Method)
	sb.WriteByte(' ')
	sb.Write(req.Path)
	if req.Query != nil {
		sb.WriteByte('?')
		sb.Write(req.Query)
	}
	sb.WriteString(" HTTP/")
	sb.WriteString(req.Version)
	sb.WriteByte('\n')
	sb.Write(headers)

	if len(req.Body) > 0 {
		sb.WriteByte('\n')
		sb.Write(req.Body)
	}

	return sb.String()
}

func (a *Application) saveMicroEngineLog(rule *Rule, req *applicationRequest, headers []byte) error {
	// Definición de constantes para evitar cadenas repetidas
	const defaultRuleName = "whitelist block"
	const defaultRuleID = "none"
	const blockMessage = "request blocked by micro engine"

	// Obtener la IP real del cliente
	realIP := getRealClientIP(req)

	// Determinar la información de la regla
	ruleName := defaultRuleName
	ruleID := defaultRuleID
	if rule != nil {
		ruleName = rule.Name
		ruleID = rule.ID.String()
	}

	// Construir el mensaje de log - usar fmt.Sprintf en lugar de concatenaciones múltiples
	logMessage := fmt.Sprintf("%s, ruleId: %s, ruleName: %s", blockMessage, ruleID, ruleName)

	// Crear directamente un slice de logs con un solo elemento
	logs := []model.Log{
		{
			Message: logMessage,
			LogRaw:  logMessage,
		},
	}

	now := time.Now()
	// Inicializar los logs del firewall
	firewallLog := model.WAFLog{
		CreatedAt:    now,
		Request:      buildRequestString(req, headers),
		Response:     "", // No procesar respuestas por ahora
		Domain:       getHostFromRequest(req),
		SrcIP:        realIP,
		DstIP:        req.DstIp.String(),
		SrcPort:      int(req.SrcPort),
		DstPort:      int(req.DstPort),
		RequestID:    req.ID,
		Logs:         logs, // Configurar el log directamente en la inicialización
		Payload:      logMessage,
		Date:         now.Format("2006-01-02"),
		Hour:         now.Hour(),
		HourGroupSix: now.Hour() / 6,
		Minute:       now.Minute(),
	}

	// Obtener y añadir la información de geolocalización de la IP de origen
	if a.ipProcessor != nil && realIP != "" {
		srcIPInfo := a.ipProcessor.GetIPInfo(realIP)
		if srcIPInfo != nil {
			firewallLog.SrcIPInfo = srcIPInfo
		}
	}

	// Usar el almacenamiento de logs de forma asíncrona
	return a.logStore.Store(firewallLog)
}

func (a *Application) saveFirewallLog(matchedRules []types.MatchedRule, interruption *types.Interruption, req *applicationRequest, headers []byte) error {
	// Construir la entrada de log
	logs := make([]model.Log, 0)

	realIP := getRealClientIP(req)
	now := time.Now()

	// Inicializar los logs del firewall
	firewallLog := model.WAFLog{
		CreatedAt:    now,
		Request:      buildRequestString(req, headers),
		Response:     "", // No procesar respuestas por ahora
		Domain:       getHostFromRequest(req),
		SrcIP:        realIP,
		DstIP:        req.DstIp.String(),
		SrcPort:      int(req.SrcPort),
		DstPort:      int(req.DstPort),
		RequestID:    req.ID,
		Date:         now.Format("2006-01-02"),
		Hour:         now.Hour(),
		HourGroupSix: now.Hour() / 6,
		Minute:       now.Minute(),
	}

	// Obtener y añadir la información de geolocalización de la IP de origen
	if a.ipProcessor != nil && realIP != "" {
		srcIPInfo := a.ipProcessor.GetIPInfo(realIP)
		if srcIPInfo != nil {
			firewallLog.SrcIPInfo = srcIPInfo
		}
	}

	// Iterar sobre todas las reglas coincidentes
	for _, matchedRule := range matchedRules {
		if data := matchedRule.Data(); matchedRule.Rule().ID() == interruption.RuleID || len(data) > 0 {
			// Añadir entrada de log
			log := model.Log{
				Message:    matchedRule.Message(),
				Payload:    matchedRule.Data(),
				RuleID:     matchedRule.Rule().ID(),
				Severity:   int(matchedRule.Rule().Severity()),
				Phase:      int(matchedRule.Rule().Phase()),
				SecMark:    matchedRule.Rule().SecMark(),
				Accuracy:   matchedRule.Rule().Accuracy(),
				SecLangRaw: matchedRule.Rule().Raw(),
				LogRaw:     matchedRule.ErrorLog(),
			}
			logs = append(logs, log)

			// Actualizar campos del log del firewall (solo sobrescribir si el nuevo valor no está vacío)
			if id := matchedRule.Rule().ID(); id != 0 {
				firewallLog.RuleID = id
			}
			if raw := matchedRule.Rule().Raw(); raw != "" {
				firewallLog.SecLangRaw = raw
			}
			if severity := matchedRule.Rule().Severity(); severity != 0 {
				firewallLog.Severity = int(severity)
			}
			if phase := matchedRule.Rule().Phase(); phase != 0 {
				firewallLog.Phase = int(phase)
			}
			if secMark := matchedRule.Rule().SecMark(); secMark != "" {
				firewallLog.SecMark = secMark
			}
			if accuracy := matchedRule.Rule().Accuracy(); accuracy != 0 {
				firewallLog.Accuracy = accuracy
			}
			if payload := matchedRule.Data(); payload != "" {
				firewallLog.Payload = payload
			}
			if msg := matchedRule.Message(); msg != "" {
				firewallLog.Message = msg
			}
			if uri := matchedRule.URI(); uri != "" {
				firewallLog.URI = uri
			}
			if clientIP := matchedRule.ClientIPAddress(); clientIP != "" {
				firewallLog.ClientIP = clientIP
			}
			if serverIP := matchedRule.ServerIPAddress(); serverIP != "" {
				firewallLog.ServerIP = serverIP
			}
		}
	}

	// Añadir todos los logs recopilados
	firewallLog.Logs = logs

	// Usar el almacenamiento de logs de forma asíncrona
	return a.logStore.Store(firewallLog)
}

// NewApplication creates a new Application with a custom context
func (a AppConfig) NewApplicationWithContext(ctx context.Context, options ApplicationOptions, isDebug bool) (*Application, error) {
	// If no context is provided, use background context
	isDev := os.Getenv("IS_DEV") == "true"
	app := &Application{
		AppConfig: a,
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if options.MongoConfig != nil && options.MongoConfig.Client != nil {
		logStore := NewMongoLogStore(
			options.MongoConfig.Client,
			options.MongoConfig.Database,
			options.MongoConfig.Collection,
			a.Logger,
		)
		logStore.Start()
		app.logStore = logStore
	}

	// Inicializar el motor de reglas según la configuración de base de datos del motor
	if options.RuleEngineDbConfig != nil && options.RuleEngineDbConfig.MongoClient != nil {
		ruleEngine := NewRuleEngine()
		ruleEngine.InitMongoConfig(options.RuleEngineDbConfig)
		ruleEngine.LoadAllFromMongoDB()
		app.ruleEngine = ruleEngine
	}

	// Inicializar el procesador de IP según la configuración GeoIP
	if options.GeoIPConfig != nil {
		processor, err := NewIPProcessor(
			ctx,
			options.GeoIPConfig.CityDBPath,
			options.GeoIPConfig.ASNDBPath,
			a.Logger,
		)
		if err != nil {
			a.Logger.Warn().Err(err).Msg("Error al inicializar el procesador de IP, se usará una implementación vacía")
			app.ipProcessor = NewNullIPProcessor()
		} else {
			app.ipProcessor = processor
		}
	} else {
		// Si no se proporciona configuración GeoIP, usar implementación vacía
		app.ipProcessor = NewNullIPProcessor()
	}

	// Inicializar el controlador de flujo
	if options.FlowControllerConfig != nil && options.FlowControllerConfig.Client != nil {
		// Crear primero el registrador de IPs
		ipRecorder := flowcontroller.NewMongoIPRecorder(
			options.FlowControllerConfig.Client,
			options.FlowControllerConfig.Database,
			10000, // Capacidad por defecto
			a.Logger,
		)
		app.ipRecorder = ipRecorder

		// Crear el controlador de flujo
		flowController, err := flowcontroller.NewFlowControllerFromMongoConfig(
			options.FlowControllerConfig.Client,
			options.FlowControllerConfig.Database,
			a.Logger,
			ipRecorder,
		)
		if err != nil {
			a.Logger.Warn().Err(err).Msg("Error al inicializar el controlador de flujo")
		} else {
			app.flowController = flowController
			if err := app.flowController.Initialize(); err != nil {
				a.Logger.Warn().Err(err).Msg("Error al inicializar el controlador de flujo")
			}
		}
	}

	debugLogger := debuglog.Default().
		WithLevel(debuglog.LevelDebug).
		WithOutput(os.Stdout)

	var config coraza.WAFConfig
	switch {
	case isDev && isDebug:
		config = coraza.NewWAFConfig().
			WithDirectives(a.Directives).
			WithErrorCallback(app.logCallback).
			WithDebugLogger(debugLogger).
			WithRootFS(mergefs.Merge(coreruleset.FS, io.OSFS))
	case isDebug:
		config = coraza.NewWAFConfig().
			WithDirectives(a.Directives).
			WithErrorCallback(app.logCallback).
			WithRootFS(mergefs.Merge(coreruleset.FS, io.OSFS))
	default:
		config = coraza.NewWAFConfig().
			WithDirectives(a.Directives).
			WithRootFS(mergefs.Merge(coreruleset.FS, io.OSFS))
	}

	waf, err := coraza.NewWAF(config)
	if err != nil {
		return nil, err
	}
	app.waf = waf

	const defaultExpire = time.Second * 10
	const defaultEvictionInterval = time.Second * 1

	app.cache = cache.NewTTLWithCallback(defaultExpire, defaultEvictionInterval, func(key, value any) {
		// Cerrar la transacción cuando expira
		t := value.(*transaction)
		if !t.m.TryLock() {
			// Perdimos la competencia, la transacción ya está en uso en otro lugar
			a.Logger.Info().Str("tx", t.tx.ID()).Msg("eviction called on currently used transaction")
			return
		}

		// El callback de timeout solo libera recursos, ya no comprueba interrupciones ni registra logs
		// Porque si la transacción se interrumpe, el log ya debería haberse registrado en la fase de petición o respuesta

		// Process Logging won't do anything if TX was already logged.
		t.tx.ProcessLogging()
		if err := t.tx.Close(); err != nil {
			a.Logger.Error().Err(err).Str("tx", t.tx.ID()).Msg("error closing transaction")
		}
	})

	return app, nil
}

// NewDefaultApplication creates a new Application with background context
func (a AppConfig) NewApplication(options ApplicationOptions) (*Application, error) {
	return a.NewApplicationWithContext(context.Background(), options, false)
}

func (a *Application) logCallback(mr types.MatchedRule) {
	var l *zerolog.Event

	switch mr.Rule().Severity() {
	case types.RuleSeverityWarning:
		l = a.Logger.Warn()
	case types.RuleSeverityNotice,
		types.RuleSeverityInfo:
		l = a.Logger.Info()
	case types.RuleSeverityDebug:
		l = a.Logger.Debug()
	default:
		l = a.Logger.Error()
	}
	l.Msg(mr.ErrorLog())
}

type ErrInterrupted struct {
	Interruption *types.Interruption
}

func (e ErrInterrupted) Error() string {
	return fmt.Sprintf("interrupted with status %d and action %s", e.Interruption.Status, e.Interruption.Action)
}

func (e ErrInterrupted) Is(target error) bool {
	t, ok := target.(*ErrInterrupted)
	if !ok {
		return false
	}

	// Comprobar primero si ambos punteros son nil
	if e.Interruption == nil || t.Interruption == nil {
		return e.Interruption == t.Interruption
	}

	// Comparar los valores de los campos de la estructura Interruption
	return e.Interruption.RuleID == t.Interruption.RuleID &&
		e.Interruption.Action == t.Interruption.Action &&
		e.Interruption.Status == t.Interruption.Status &&
		e.Interruption.Data == t.Interruption.Data
}

// Función getHeaderValue optimizada - versión de alto rendimiento
func getHeaderValue(headers []byte, targetHeader string) (string, error) {
	if len(headers) == 0 || targetHeader == "" {
		return "", nil
	}

	// Preconvertir el header objetivo a minúsculas, evita conversiones repetidas en el bucle
	targetHeaderLower := strings.ToLower(targetHeader)
	targetLen := len(targetHeaderLower)

	start := 0
	for start < len(headers) {
		// Buscar el fin de línea
		lineEnd := start
		for lineEnd < len(headers) && headers[lineEnd] != '\n' {
			lineEnd++
		}

		// Obtener la línea actual y quitar posibles \r
		line := headers[start:lineEnd]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		// Omitir líneas vacías
		if len(line) == 0 {
			start = lineEnd + 1
			continue
		}

		// Buscar los dos puntos
		colonIdx := bytes.IndexByte(line, ':')
		if colonIdx <= 0 {
			start = lineEnd + 1
			continue
		}

		// Extraer la clave y comprobar su longitud
		key := bytes.TrimSpace(line[:colonIdx])
		if len(key) != targetLen {
			start = lineEnd + 1
			continue
		}

		// Comparación rápida sin distinción de mayúsculas (solo ASCII)
		isMatch := true
		for i := 0; i < targetLen; i++ {
			a := key[i]
			if a >= 'A' && a <= 'Z' {
				a += 32 // Convertir a minúsculas
			}
			if a != targetHeaderLower[i] {
				isMatch = false
				break
			}
		}

		if isMatch {
			value := bytes.TrimSpace(line[colonIdx+1:])
			return string(value), nil
		}

		// Pasar a la siguiente línea
		start = lineEnd + 1
	}

	return "", nil
}

func getHostFromRequest(req *applicationRequest) string {
	if host, err := getHeaderValue(req.Headers, "host"); err == nil && host != "" {
		// Separar el nombre de host y el número de puerto
		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			return host[:colonIndex]
		}
		return host
	}
	// Si la IP de destino puede incluir puerto, separarla también
	dstIpStr := req.DstIp.String()
	if colonIndex := strings.Index(dstIpStr, ":"); colonIndex != -1 {
		return dstIpStr[:colonIndex]
	}
	return dstIpStr
}

// getRealClientIP obtiene la IP real del cliente desde varias cabeceras HTTP (versión optimizada)
func getRealClientIP(req *applicationRequest) string {
	if req == nil {
		return ""
	}

	headers := req.Headers
	if len(headers) == 0 {
		// Si no hay cabecera, devolver la IP de origen directamente
		if req.SrcIp.IsValid() {
			return req.SrcIp.String()
		}
		return ""
	}

	// Ruta rápida: buscar primero la cabecera X-Forwarded-For más común
	if ip := getXForwardedForIP(headers); ip != "" {
		return ip
	}

	// Probar otras cabeceras por prioridad
	priorityHeaders := []string{
		"x-real-ip",        // Común en Nginx
		"true-client-ip",   // Akamai
		"cf-connecting-ip", // Cloudflare
		"fastly-client-ip", // Fastly
		"x-client-ip",      // General
	}

	// Búsqueda por lotes de headers simples (que devuelven el valor directamente)
	for _, header := range priorityHeaders {
		if value, err := getHeaderValue(headers, header); err == nil && value != "" {
			if ip := strings.TrimSpace(value); ip != "" {
				return ip
			}
		}
	}

	// Buscar headers complejos
	if value, err := getHeaderValue(headers, "x-original-forwarded-for"); err == nil && value != "" {
		if ips := strings.Split(value, ","); len(ips) > 0 {
			if ip := strings.TrimSpace(ips[0]); ip != "" {
				return ip
			}
		}
	}

	// La cabecera Forwarded requiere un parseo especial
	if value, err := getHeaderValue(headers, "forwarded"); err == nil && value != "" {
		if ip := parseForwardedHeaderFast(value); ip != "" {
			return ip
		}
	}

	// X-Cluster-Client-IP (última comprobación)
	if value, err := getHeaderValue(headers, "x-cluster-client-ip"); err == nil && value != "" {
		if ip := strings.TrimSpace(value); ip != "" {
			return ip
		}
	}

	// Si ninguna cabecera está presente, devolver la IP de origen
	if req.SrcIp.IsValid() {
		return req.SrcIp.String()
	}

	return ""
}

// getXForwardedForIP analiza rápidamente la cabecera X-Forwarded-For
func getXForwardedForIP(headers []byte) string {
	// Búsqueda rápida de "x-forwarded-for:" o "X-Forwarded-For:"
	target := []byte("x-forwarded-for:")
	targetUpper := []byte("X-Forwarded-For:")

	start := 0
	for start < len(headers) {
		// Buscar el fin de línea
		lineEnd := start
		for lineEnd < len(headers) && headers[lineEnd] != '\n' {
			lineEnd++
		}

		line := headers[start:lineEnd]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		// Comprobar si coincide con X-Forwarded-For
		if len(line) > 16 { // "x-forwarded-for:" tiene longitud mínima 16
			// Comprobación rápida de prefijo
			if (bytes.HasPrefix(line, target) || bytes.HasPrefix(line, targetUpper)) ||
				(len(line) > 15 && isXForwardedForHeader(line)) {

				// Encontrar el valor después de los dos puntos
				colonIdx := bytes.IndexByte(line, ':')
				if colonIdx > 0 && colonIdx < len(line)-1 {
					value := bytes.TrimSpace(line[colonIdx+1:])
					if len(value) > 0 {
						// Extraer la primera IP (separadas por comas)
						valueStr := string(value)
						if commaIdx := strings.Index(valueStr, ","); commaIdx > 0 {
							ip := strings.TrimSpace(valueStr[:commaIdx])
							if ip != "" {
								return ip
							}
						} else {
							ip := strings.TrimSpace(valueStr)
							if ip != "" {
								return ip
							}
						}
					}
				}
			}
		}

		start = lineEnd + 1
	}
	return ""
}

// isXForwardedForHeader comprueba si es la cabecera X-Forwarded-For (sin distinción de mayúsculas)
func isXForwardedForHeader(line []byte) bool {
	if len(line) < 16 { // "x-forwarded-for:" tiene longitud 16
		return false
	}

	target := "x-forwarded-for:"
	for i := 0; i < 16; i++ {
		c := line[i]
		if c >= 'A' && c <= 'Z' {
			c += 32 // Convertir a minúsculas
		}
		if c != target[i] {
			return false
		}
	}
	return true
}

// parseForwardedHeaderFast análisis rápido de la cabecera Forwarded
func parseForwardedHeaderFast(forwarded string) string {
	// Búsqueda rápida del patrón "for="
	forPrefix := "for="
	start := 0

	for {
		idx := strings.Index(forwarded[start:], forPrefix)
		if idx == -1 {
			break
		}

		start += idx + 4 // len("for=")
		if start >= len(forwarded) {
			break
		}

		// Buscar la posición final del valor (punto y coma o fin de cadena)
		end := start
		for end < len(forwarded) && forwarded[end] != ';' {
			end++
		}

		if end > start {
			ip := strings.TrimSpace(forwarded[start:end])
			// Eliminar comillas
			ip = strings.Trim(ip, "\"")

			// Manejar formatos de dirección IPv6 [ip]:port o [ip]
			if strings.HasPrefix(ip, "[") {
				if closeBracket := strings.Index(ip, "]"); closeBracket > 1 {
					ip = ip[1:closeBracket]
				}
			}

			if ip != "" {
				return ip
			}
		}

		start = end
	}

	return ""
}

// buildURLFromBytes función de construcción de URL de alto rendimiento
func buildURLFromBytes(path, query []byte) string {
	if len(query) == 0 {
		return string(path)
	}

	// Asignar la memoria necesaria de una vez
	result := make([]byte, 0, len(path)+1+len(query))
	result = append(result, path...)
	result = append(result, '?')
	result = append(result, query...)
	return string(result)
}

// buildFullURL construye la URL completa (incluye host)
func buildFullURL(host string, path, query []byte) string {
	// Calcular la longitud total
	totalLen := len(host) + len(path)
	if len(query) > 0 {
		totalLen += 1 + len(query) // +1 for '?'
	}

	// Asignación única
	result := make([]byte, 0, totalLen)
	result = append(result, host...)
	result = append(result, path...)
	if len(query) > 0 {
		result = append(result, '?')
		result = append(result, query...)
	}
	return string(result)
}
