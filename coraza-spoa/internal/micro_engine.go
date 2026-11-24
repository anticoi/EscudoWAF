package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/anticoi/EscudoWAF/pkg/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Modo de coincidencia
type MatchType string

const (
	// Modo de coincidencia de IP
	MatchEqual        MatchType = "equal"
	MatchNotEqual     MatchType = "not_equal"
	MatchFuzzy        MatchType = "fuzzy"
	MatchInCIDR       MatchType = "in_cidr"
	MatchNotInCIDR    MatchType = "not_in_cidr"
	MatchInIPGroup    MatchType = "in_ipgroup"
	MatchNotInIPGroup MatchType = "not_in_ipgroup"

	// Modo de coincidencia de URL y Path
	MatchInclude       MatchType = "include"
	MatchContains      MatchType = "contains"
	MatchNotContains   MatchType = "not_contains"
	MatchPrefixKeyword MatchType = "prefix_keyword"
	MatchRegex         MatchType = "regex"
)

// Tipo de objetivo de coincidencia
type TargetType string

const (
	SourceIP   TargetType = "source_ip"
	TargetURL  TargetType = "url"
	TargetPath TargetType = "path"
)

// Operador lógico
type LogicalOperator string

const (
	LogicalAND LogicalOperator = "AND"
	LogicalOR  LogicalOperator = "OR"
)

// La interfaz Matcher define los métodos de coincidencia de condiciones
type Matcher interface {
	Match(eng *RuleEngine, ip, url, path string) (bool, error)
}

// Tipo de condición
type ConditionType string

const (
	SimpleConditionType    ConditionType = "simple"
	CompositeConditionType ConditionType = "composite"
)

// SimpleCondition condición simple
type SimpleCondition struct {
	Type       ConditionType `json:"type" bson:"type"`
	Target     TargetType    `json:"target" bson:"target"`
	MatchType  MatchType     `json:"match_type" bson:"match_type"`
	MatchValue string        `json:"match_value" bson:"match_value"`
}

// Match implementa la interfaz Matcher
func (c *SimpleCondition) Match(eng *RuleEngine, ip, url, path string) (bool, error) {
	switch c.Target {
	case SourceIP:
		return eng.matchIP(c, ip)
	case TargetURL:
		return eng.matchURL(c, url)
	case TargetPath:
		return eng.matchPath(c, path)
	default:
		return false, fmt.Errorf("Tipo de objetivo no soportado: %s", c.Target)
	}
}

// CompositeCondition condición compuesta
type CompositeCondition struct {
	Type       ConditionType   `json:"type" bson:"type"`
	Operator   LogicalOperator `json:"operator" bson:"operator"`
	Conditions []bson.Raw      `json:"conditions" bson:"conditions"`

	// Campo de tiempo de ejecución, no usado en JSON/BSON
	parsedConditions []Matcher
}

// Match implementa la interfaz Matcher
func (c *CompositeCondition) Match(eng *RuleEngine, ip, url, path string) (bool, error) {
	if len(c.parsedConditions) == 0 {
		return false, fmt.Errorf("Condición compuesta no inicializada")
	}

	var result bool
	if c.Operator == LogicalAND {
		result = true
	} else {
		result = false
	}

	for _, condition := range c.parsedConditions {
		match, err := condition.Match(eng, ip, url, path)
		if err != nil {
			return false, err
		}

		if c.Operator == LogicalAND {
			result = result && match
			if !result {
				return false, nil // Cortocircuito
			}
		} else {
			result = result || match
			if result {
				return true, nil // Cortocircuito
			}
		}
	}

	return result, nil
}

// ConditionFactory fábrica de condiciones
type ConditionFactory struct{}

// ParseCondition analiza una condición
func (f *ConditionFactory) ParseCondition(data bson.Raw) (Matcher, error) {
	var baseCondition struct {
		Type ConditionType `json:"type" bson:"type"`
	}

	if err := bson.Unmarshal(data, &baseCondition); err != nil {
		return nil, fmt.Errorf("Error al analizar el tipo de condición: %v", err)
	}

	switch baseCondition.Type {
	case SimpleConditionType:
		var condition SimpleCondition
		if err := bson.Unmarshal(data, &condition); err != nil {
			return nil, fmt.Errorf("Error al analizar la condición simple: %v", err)
		}
		return &condition, nil

	case CompositeConditionType:
		var condition CompositeCondition
		if err := bson.Unmarshal(data, &condition); err != nil {
			return nil, fmt.Errorf("Error al analizar la condición compuesta: %v", err)
		}

		condition.parsedConditions = make([]Matcher, 0, len(condition.Conditions))
		for _, rawCondition := range condition.Conditions {
			parsedCondition, err := f.ParseCondition(rawCondition)
			if err != nil {
				return nil, err
			}
			condition.parsedConditions = append(condition.parsedConditions, parsedCondition)
		}

		return &condition, nil

	default:
		return nil, fmt.Errorf("Tipo de condición no soportado: %s", baseCondition.Type)
	}
}

// Rule regla - con campos de prioridad y número de serie añadidos
type Rule struct {
	// MicroRule embebida
	model.MicroRule `bson:",inline" json:",inline"`

	// Campo de tiempo de ejecución, no usado en JSON/BSON
	parsedCondition Matcher `bson:"-" json:"-"`
	sequence        int     `bson:"-" json:"-"`
}

// Configuración de MongoDB
type MongoDBConfig struct {
	MongoClient       *mongo.Client
	Database          string // Nombre de la base de datos
	RuleCollection    string // Nombre de la colección de reglas
	IPGroupCollection string // Nombre de la colección de grupos de IPs
}

// RuleEngine motor de reglas
type RuleEngine struct {
	Rules       []Rule                    `json:"rules"`     // Lista de todas las reglas
	IPGroups    map[string]*model.IPGroup `json:"ip_groups"` // Tabla de mapeo de grupos de IPs
	regexCache  map[string]*regexp.Regexp // Caché de expresiones regulares
	factory     ConditionFactory          // Fábrica de condiciones
	mongoConfig *MongoDBConfig            // Configuración de MongoDB
}

// NewRuleEngine crea el motor de reglas
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		Rules:    make([]Rule, 0),
		IPGroups: make(map[string]*model.IPGroup),
		// TODO: optimizar con LRU, establecer expiración de caché para evitar que crezca demasiado
		regexCache: make(map[string]*regexp.Regexp),
		factory:    ConditionFactory{},
	}
}

func (e *RuleEngine) InitMongoConfig(config *MongoDBConfig) error {
	e.mongoConfig = config
	return nil
}

// LoadIPGroupsFromMongoDB carga los grupos de IPs desde MongoDB
func (e *RuleEngine) LoadIPGroupsFromMongoDB() error {
	if e.mongoConfig.MongoClient == nil {
		return fmt.Errorf("Cliente de MongoDB no inicializado")
	}

	// Obtener la colección
	collection := e.mongoConfig.MongoClient.
		Database(e.mongoConfig.Database).
		Collection(e.mongoConfig.IPGroupCollection)

	// Crear el contexto
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Lógica para comprobar y crear el grupo de IPs por defecto
	defaultBlacklistCount, err := collection.CountDocuments(ctx, bson.D{{Key: "name", Value: "system_default_blacklist"}})
	if err != nil {
		return fmt.Errorf("Error al comprobar si existe la lista negra por defecto: %v", err)
	}

	// Si la lista negra por defecto no existe, crearla
	if defaultBlacklistCount == 0 {
		// Crear el grupo de lista negra por defecto
		defaultBlacklist := model.IPGroup{
			Name:  "system_default_blacklist",
			Items: []string{}, // Inicialmente vacío
		}

		// Insertar en la base de datos
		_, err := collection.InsertOne(ctx, defaultBlacklist)
		if err != nil {
			return fmt.Errorf("Error al crear el grupo de lista negra de IPs por defecto: %v", err)
		}
	}

	// Consultar todos los grupos de IPs
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("Error al consultar el grupo de IPs: %v", err)
	}
	defer cursor.Close(ctx)

	// Decodificar el grupo de IPs
	var ipGroups []model.IPGroup
	if err = cursor.All(ctx, &ipGroups); err != nil {
		return fmt.Errorf("Error al decodificar el grupo de IPs: %v", err)
	}

	// Inicializar la tabla de mapeo
	e.IPGroups = make(map[string]*model.IPGroup)

	// Rellenar el mapa de grupos de IPs
	for _, group := range ipGroups {
		for _, item := range group.Items {
			if !isValidIPOrCIDR(item) {
				return fmt.Errorf("El grupo de IPs %s contiene una IP o CIDR no válida: %s", group.Name, item)
			}
		}
		e.IPGroups[group.Name] = &group
	}

	return nil
}

// LoadRulesFromMongoDB carga las reglas desde MongoDB
func (e *RuleEngine) LoadRulesFromMongoDB() error {
	if e.mongoConfig.MongoClient == nil {
		return fmt.Errorf("Cliente de MongoDB no inicializado")
	}

	// Obtener la colección
	collection := e.mongoConfig.MongoClient.
		Database(e.mongoConfig.Database).
		Collection(e.mongoConfig.RuleCollection)

	// Crear el contexto
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Lógica para comprobar y crear la regla por defecto
	defaultRuleCount, err := collection.CountDocuments(ctx, bson.D{{Key: "name", Value: "system_default_ip_block"}})
	if err != nil {
		return fmt.Errorf("Error al comprobar si existe la regla por defecto: %v", err)
	}

	// Si la regla por defecto no existe, crearla
	if defaultRuleCount == 0 {
		// Crear las condiciones de la regla por defecto
		defaultCondition := SimpleCondition{
			Type:       "simple",
			Target:     "source_ip",
			MatchType:  "in_ipgroup",
			MatchValue: "system_default_blacklist",
		}

		// Serializar las condiciones a BSON
		conditionBytes, err := bson.Marshal(defaultCondition)
		if err != nil {
			return fmt.Errorf("Error al serializar las condiciones de la regla por defecto: %v", err)
		}

		// Crear la regla por defecto
		defaultRule := model.MicroRule{
			Name:      "system_default_ip_block",
			Type:      "blacklist",
			Status:    "enabled",
			Priority:  9999, // Prioridad más alta
			Condition: conditionBytes,
		}

		// Insertar en la base de datos
		_, err = collection.InsertOne(ctx, defaultRule)
		if err != nil {
			return fmt.Errorf("Error al crear la regla por defecto: %v", err)
		}
	}

	// Consultar todas las reglas
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("Error al consultar la regla: %v", err)
	}
	defer cursor.Close(ctx)

	// Decodificar la regla
	var rules []Rule
	if err = cursor.All(ctx, &rules); err != nil {
		return fmt.Errorf("Error al decodificar la regla: %v", err)
	}

	// Establecer el número de secuencia
	for i := range rules {
		rules[i].sequence = i
	}

	// Analizar las condiciones de cada regla
	for i := range rules {
		rule := &rules[i]
		parsedCondition, err := e.factory.ParseCondition(rule.Condition)
		if err != nil {
			return fmt.Errorf("Error al analizar las condiciones de la regla %s: %v", rule.ID, err)
		}
		rule.parsedCondition = parsedCondition
	}

	// Ordenar por prioridad; con igual prioridad mantener el orden original
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority // Las de mayor prioridad van primero
		}
		return rules[i].sequence < rules[j].sequence // Con igual prioridad, mantener el orden original
	})

	e.Rules = rules
	return nil
}

// LoadAllFromMongoDB carga todas las reglas y grupos de IPs desde MongoDB
func (e *RuleEngine) LoadAllFromMongoDB() error {
	if err := e.LoadIPGroupsFromMongoDB(); err != nil {
		return err
	}

	return e.LoadRulesFromMongoDB()
}

// AddIPGroup añade un grupo de IPs
func (e *RuleEngine) AddIPGroup(group model.IPGroup) error {
	if _, exists := e.IPGroups[group.Name]; exists {
		return fmt.Errorf("El grupo de IPs %s ya existe", group.Name)
	}

	for _, item := range group.Items {
		if !isValidIPOrCIDR(item) {
			return fmt.Errorf("El grupo de IPs %s contiene una IP o CIDR no válida: %s", group.Name, item)
		}
	}

	e.IPGroups[group.Name] = &group
	return nil
}

// LoadRulesFromJSON carga reglas desde JSON - lógica de carga modificada, con manejo de número de serie
// BUG type transform error bson raw and json raw
func (e *RuleEngine) LoadRulesFromJSON(data []byte) error {
	var rules []Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}

	// Establecer el número de secuencia - registra el orden de la regla en la configuración original
	for i := range rules {
		rules[i].sequence = i
	}

	// Analizar las condiciones de cada regla
	for i := range rules {
		rule := &rules[i]
		parsedCondition, err := e.factory.ParseCondition(rule.Condition)
		if err != nil {
			return fmt.Errorf("Error al analizar las condiciones de la regla %s: %v", rule.ID, err)
		}
		rule.parsedCondition = parsedCondition
	}

	// Ordenar por prioridad; con igual prioridad mantener el orden original
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority // Las de mayor prioridad van primero
		}
		return rules[i].sequence < rules[j].sequence // Con igual prioridad, mantener el orden original
	})

	e.Rules = rules
	return nil
}

// AddRule añade una regla
func (e *RuleEngine) AddRule(rule Rule) error {
	// Analizar las condiciones de la regla
	parsedCondition, err := e.factory.ParseCondition(rule.Condition)
	if err != nil {
		return fmt.Errorf("Error al analizar las condiciones de la regla %s: %v", rule.ID, err)
	}
	rule.parsedCondition = parsedCondition

	// Establecer el número de secuencia de la regla como la longitud actual de la lista de reglas
	rule.sequence = len(e.Rules)

	// Añadir la regla a la lista
	e.Rules = append(e.Rules, rule)

	// Reordenar reglas
	sort.Slice(e.Rules, func(i, j int) bool {
		if e.Rules[i].Priority != e.Rules[j].Priority {
			return e.Rules[i].Priority > e.Rules[j].Priority
		}
		return e.Rules[i].sequence < e.Rules[j].sequence
	})

	return nil
}

// MatchRequest petición de coincidencia
// Parámetros:
// - ip: dirección IP de origen
// - url: URL de la petición
// - path: ruta de la petición
// Valor de retorno:
// - shouldBlock: si se debe bloquear la petición (true bloquea, false permite)
// - ruleType: tipo de la regla que coincidió
// - rule: regla que coincidió
// - error: información del error
func (e *RuleEngine) MatchRequest(ip string, url string, path string) (shouldBlock bool, ruleType model.RuleType, rule *Rule, err error) {
	// Validar el formato de la dirección IP
	if !isValidIP(ip) {
		return false, "", nil, fmt.Errorf("Dirección IP no válida: %s", ip)
	}

	// Marca si existen reglas de lista blanca habilitadas
	hasWhitelistRule := false

	// Iterar sobre todas las reglas (ya ordenadas por prioridad y número de secuencia)
	for _, r := range e.Rules {
		// Comprobar si existen reglas de lista blanca habilitadas
		if r.Status == model.RuleEnabled && r.Type == model.WhitelistRule {
			hasWhitelistRule = true
		}

		// Omitir reglas deshabilitadas
		if r.Status == model.RuleDisabled {
			continue
		}

		// Condiciones de coincidencia de la regla
		match, err := r.parsedCondition.Match(e, ip, url, path)
		if err != nil {
			return false, "", nil, err
		}

		// Si las condiciones de la regla coinciden
		if match {
			// Determinar si se debe bloquear según el tipo de regla
			switch r.Type {
			case model.BlacklistRule:
				// Regla de lista negra coincidente -> devolver true (bloquear)
				return true, r.Type, &r, nil
			case model.WhitelistRule:
				// Regla de lista blanca coincidente -> devolver false (permitir)
				return false, r.Type, &r, nil
			default:
				return false, "", nil, fmt.Errorf("Tipo de regla desconocido: %s", r.Type)
			}
		}
	}

	// Si no coincide ninguna regla y existen reglas de lista blanca, bloquear
	if hasWhitelistRule {
		// Hay reglas de lista blanca pero ninguna coincidió -> bloquear la petición (valor seguro por defecto)
		return true, "", nil, nil
	}

	// Si no coincide ninguna regla y no hay reglas de lista blanca, permitir por defecto
	return false, "", nil, nil
}

// GetRules obtiene la lista de reglas actual
func (e *RuleEngine) GetRules() []Rule {
	return e.Rules
}

// matchIP evalúa la condición de IP
func (e *RuleEngine) matchIP(cond *SimpleCondition, ip string) (bool, error) {
	switch cond.MatchType {
	case MatchEqual:
		return ip == cond.MatchValue, nil
	case MatchNotEqual:
		return ip != cond.MatchValue, nil
	case MatchFuzzy:
		return matchIPFuzzy(ip, cond.MatchValue)
	case MatchInCIDR:
		return isIPInCIDR(ip, cond.MatchValue)
	case MatchNotInCIDR:
		inCIDR, err := isIPInCIDR(ip, cond.MatchValue)
		return !inCIDR, err
	case MatchInIPGroup:
		return e.isIPInGroup(ip, cond.MatchValue)
	case MatchNotInIPGroup:
		inGroup, err := e.isIPInGroup(ip, cond.MatchValue)
		return !inGroup, err
	default:
		return false, fmt.Errorf("Modo de coincidencia no soportado para IP: %s", cond.MatchType)
	}
}

// matchURL evalúa la condición de URL
func (e *RuleEngine) matchURL(cond *SimpleCondition, url string) (bool, error) {
	switch cond.MatchType {
	case MatchEqual:
		return url == cond.MatchValue, nil
	case MatchNotEqual:
		return url != cond.MatchValue, nil
	case MatchInclude, MatchContains:
		return strings.Contains(url, cond.MatchValue), nil
	case MatchNotContains:
		return !strings.Contains(url, cond.MatchValue), nil
	case MatchPrefixKeyword:
		return strings.HasPrefix(url, cond.MatchValue), nil
	case MatchRegex:
		return e.matchRegex(url, cond.MatchValue)
	default:
		return false, fmt.Errorf("Modo de coincidencia no soportado para URL: %s", cond.MatchType)
	}
}

// matchPath evalúa la condición de Path
func (e *RuleEngine) matchPath(cond *SimpleCondition, path string) (bool, error) {
	return e.matchURL(cond, path)
}

// A continuación funciones auxiliares

// isValidIP comprueba si la IP es válida
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isValidCIDR comprueba si el CIDR es válido
func isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// isValidIPOrCIDR comprueba si es una IP o CIDR válida
func isValidIPOrCIDR(s string) bool {
	return isValidIP(s) || isValidCIDR(s)
}

// isValidIPPattern comprueba si el patrón de IP es válido
func isValidIPPattern(pattern string) bool {
	parts := strings.Split(pattern, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part != "*" {
			if part == "" {
				return false
			}

			var num int
			if _, err := fmt.Sscanf(part, "%d", &num); err != nil {
				return false
			}

			if num < 0 || num > 255 {
				return false
			}
		}
	}

	return true
}

// isIPInCIDR comprueba si la IP está dentro del rango CIDR
func isIPInCIDR(ipStr, cidrStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("Dirección IP no válida: %s", ipStr)
	}

	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("CIDR no válido: %s", cidrStr)
	}

	return ipNet.Contains(ip), nil
}

// matchIPFuzzy coincidencia difusa de IP
func matchIPFuzzy(ip, pattern string) (bool, error) {
	ipParts := strings.Split(ip, ".")
	patternParts := strings.Split(pattern, ".")

	if len(ipParts) != 4 || len(patternParts) != 4 {
		return false, fmt.Errorf("Formato de IP incorrecto")
	}

	for i := 0; i < 4; i++ {
		if patternParts[i] != "*" && ipParts[i] != patternParts[i] {
			return false, nil
		}
	}

	return true, nil
}

// isIPInGroup comprueba si la IP está en el grupo de IPs
// TODO: evitar el recorrido lineal O(N), optimizar con árbol radix (Radix Tree/Patricia Trie)
func (e *RuleEngine) isIPInGroup(ip, groupName string) (bool, error) {
	group, exists := e.IPGroups[groupName]
	if !exists {
		return false, fmt.Errorf("El grupo de IPs no existe: %s", groupName)
	}

	for _, item := range group.Items {
		if isValidIP(item) {
			if ip == item {
				return true, nil
			}
		} else if isValidCIDR(item) {
			inCIDR, err := isIPInCIDR(ip, item)
			if err != nil {
				return false, err
			}
			if inCIDR {
				return true, nil
			}
		} else {
			return false, fmt.Errorf("El grupo de IPs %s contiene elementos no válidos: %s", groupName, item)
		}
	}

	return false, nil
}

// matchRegex coincidencia por expresión regular
func (e *RuleEngine) matchRegex(s, pattern string) (bool, error) {
	re, exists := e.regexCache[pattern]
	if !exists {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return false, fmt.Errorf("Expresión regular no válida: %s", pattern)
		}
		e.regexCache[pattern] = re
	}

	return re.MatchString(s), nil
}
