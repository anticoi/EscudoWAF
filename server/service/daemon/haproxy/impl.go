package haproxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"text/template"
	"time"

	"github.com/anticoi/EscudoWAF/server/config"
	"github.com/anticoi/EscudoWAF/server/model"
	client_native "github.com/haproxytech/client-native/v6"
	"github.com/haproxytech/client-native/v6/configuration"
	cfg_opt "github.com/haproxytech/client-native/v6/configuration/options"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/haproxytech/client-native/v6/options"
	runtime_api "github.com/haproxytech/client-native/v6/runtime"
	runtime_options "github.com/haproxytech/client-native/v6/runtime/options"
	spoe "github.com/haproxytech/client-native/v6/spoe"
	"github.com/rs/zerolog"
)

type HAProxyStatus int32

const (
	StatusStopped HAProxyStatus = iota
	StatusRunning
	StatusError
)

type HAProxyServiceImpl struct {
	ConfigBaseDir      string
	HAProxyConfigFile  string // Ruta del archivo de configuración
	HaproxyBin         string // Ruta del binario de HAProxy
	BackupsNumber      int
	CertDir            string // Directorio de certificados
	TransactionDir     string // Directorio de transacciones
	SpoeDir            string // Directorio SPOE
	SpoeTransactionDir string // Directorio de transacciones SPOE
	SocketFile         string // Ruta del archivo de socket
	PidFile            string // Ruta del archivo PID
	SpoeConfigFile     string // Ruta del archivo de configuración SPOE
	SpoeAgentAddress   string // Dirección del agente SPOE
	SpoeAgentPort      int64  // Puerto del agente SPOE

	// internal field
	haproxyCmd      *exec.Cmd                   // Comandos del proceso HAProxy
	confClient      configuration.Configuration // Cliente de configuración
	runtimeClient   runtime_api.Runtime         // Cliente de tiempo de ejecución
	spoeClient      spoe.Spoe                   // Cliente SPOE
	clientNative    client_native.HAProxyClient // Cliente completo
	isResponseCheck bool                        // Si el procesamiento de respuestas está habilitado
	status          atomic.Int32                // Estado usando operaciones atómicas
	isDebug         bool                        // Si es entorno de producción
	isK8s           bool                        // Si está en entorno K8s
	thread          int                         // Número de hilos

	logger zerolog.Logger
	ctx    context.Context
	mutex  sync.Mutex
}

func (s *HAProxyServiceImpl) GetStatus() HAProxyStatus {
	return HAProxyStatus(s.status.Load())
}

func (s *HAProxyServiceImpl) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, err := os.Stat(s.HAProxyConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("No hay archivo de configuración de haproxy")
	}
	if _, err := os.Stat(s.SpoeConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("No hay archivo de configuración spoe")
	}

	// Comprobar si HAProxy ya está en ejecución
	running, _ := s.isHAProxyRunning()
	if running {
		return fmt.Errorf("HAProxy ya está en ejecución")
	}

	// Iniciar el proceso HAProxy
	args := []string{
		"-f", s.HAProxyConfigFile,
		"-p", s.PidFile,
		"-W",
		"-S", fmt.Sprintf("unix@%s", s.SocketFile),
	}

	// Si es producción, añadir el parámetro de modo silencioso
	if !s.isDebug {
		args = append([]string{"-q"}, args...)
	}

	// Iniciar el proceso HAProxy
	cmd := exec.Command(s.HaproxyBin, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Error al iniciar HAProxy: %v", err)
	}

	s.haproxyCmd = cmd

	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		if _, err := os.Stat(s.SocketFile); err == nil {
			// Socket creado, esperar un tiempo para asegurar que HAProxy esté listo
			time.Sleep(500 * time.Millisecond)
			break
		}

		if i == maxAttempts-1 {
			return fmt.Errorf("El archivo de socket no fue creado: %s", s.SocketFile)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Inicializar el cliente
	if err := s.initClients(); err != nil {
		// Si falla la inicialización del cliente, intentar terminar el proceso HAProxy
		s.stopHAProxy()
		return fmt.Errorf("Error al inicializar el cliente: %v", err)
	}

	s.status.Store(int32(StatusRunning))

	return nil
}

func (s *HAProxyServiceImpl) AddSiteConfig(site model.Site) error {
	if err := model.ValidateSite(&site); err != nil {
		return fmt.Errorf("site config invalid: %v", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !site.ActiveStatus {
		return nil
	}

	s.logger.Info().Msgf("Añadir configuración de sitio %s", site.Domain)

	// Asegurar que el cliente de configuración se inicialice
	if err := s.ensureConfClient(); err != nil {
		return err
	}
	if _, err := s.getFeCombined(site.ListenPort); err != nil {
		err = s.createFeCombined(site.ListenPort, site.EnableHTTPS)
		if err != nil {
			return fmt.Errorf("Error al crear la combinación de frontends: %v", err)
		}
	}

	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("Error al obtener la versión: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("Error al iniciar la transacción: %v", err)
	}

	// handle http
	if isIPAddress(site.Domain) {
		// IP address handling
		err = s.confClient.DeleteServer("loopback-for-default", "backend", fmt.Sprintf("p%d_backend", site.ListenPort), transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al eliminar el servidor backend: %v", err)
		}

		for index, server := range site.Backend.Servers {
			err = s.createBackendServer(fmt.Sprintf("s%s_%d", getDashDomain(site.Domain), index), server.Host, server.Port, transaction.ID, fmt.Sprintf("p%d_backend", site.ListenPort), server.IsSSL)
			if err != nil {
				return fmt.Errorf("Error al crear el servidor backend: %v", err)
			}
		}

	} else {
		_, aclList, err := s.confClient.GetACLs("frontend", fmt.Sprintf("fe_%d_http", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("Error al obtener ACL: %v", err)
		}
		aclIndex := len(aclList)
		acl_http := &models.ACL{
			ACLName:   fmt.Sprintf("host_%s", getDashDomain(site.Domain)), // Usar el campo ACLName
			Criterion: "hdr(host) -i -m end",
			Value:     site.Domain, // Usar el campo Value
			// Criterion: "hdr(host) -i",                                     // usar el campo Criterion
			// Value:     site.Domain,                                        // usar el campo Value
		}
		err = s.confClient.CreateACL(int64(aclIndex), "frontend", fmt.Sprintf("fe_%d_http", site.ListenPort), acl_http, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al crear ACL: %v", err)
		}

		backend_http := &models.Backend{
			BackendBase: models.BackendBase{
				Name:    fmt.Sprintf("be_%s", getDashDomain(site.Domain)),
				Mode:    "http",
				Enabled: true,
				From:    "http",
				// Añadir la opción forwarded
				Forwardfor: &models.Forwardfor{
					Enabled: StringP("enabled"),
					Ifnone:  true,
				},
			},
		}
		err = s.confClient.CreateBackend(backend_http, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al crear el backend: %v", err)
		}

		if s.isK8s {
			/*
				Host Header Rewriting "Origin Host Forwarding" (reenvío del host original)
					Asegurar que el servidor backend reciba el nombre de host original correcto
					Implementar servicio de virtual hosting basado en nombre de host
					Resolver problemas de enrutamiento en entornos con múltiples proxies
					Satisfacer los requisitos de la cabecera Host de servicios backend específicos
					Implementar proxy transparente
			*/
			site_backend_request_rule := []struct {
				index int64
				rule  *models.HTTPRequestRule
			}{
				{
					index: 0,
					rule: &models.HTTPRequestRule{
						Type:      "set-header",
						HdrName:   "X-Original-Host",
						HdrFormat: "%[req.hdr(host)]",
					},
				},
				{
					index: 1,
					rule: &models.HTTPRequestRule{
						Type:    "set-header",
						HdrName: "Host",
						// TODO: en k8s, si hay varios dominios backend, aquí solo se pasa el primero; al encontrar otros backends, la cabecera Host reenviada es incorrecta,
						// TODO: pendiente de resolver: la cabecera Host reenviada difiere según el dominio backend; debe pasarse según el dominio correspondiente
						HdrFormat: site.Backend.Servers[0].Host,
					},
				},
			}

			for _, item := range site_backend_request_rule {
				err = s.confClient.CreateHTTPRequestRule(item.index, "backend", backend_http.Name, item.rule, transaction.ID, 0)
				if err != nil {
					return fmt.Errorf("Sitio %s backend %s, error al añadir la regla de petición HTTP #%d: %v", site.Domain, backend_http.Name, item.index, err)
				}
			}
		}

		_, switchingRules, err := s.confClient.GetBackendSwitchingRules(fmt.Sprintf("fe_%d_http", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("Error al obtener la regla de cambio de backend: %v", err)
		}
		switchingRuleIndex := len(switchingRules)

		httpUseBackendRule := &models.BackendSwitchingRule{
			Name:     backend_http.Name,
			Cond:     "if",
			CondTest: acl_http.ACLName,
		}
		err = s.confClient.CreateBackendSwitchingRule(int64(switchingRuleIndex), fmt.Sprintf("fe_%d_http", site.ListenPort), httpUseBackendRule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al crear la regla de cambio de backend: %v", err)
		}

		for index, server := range site.Backend.Servers {
			err = s.createBackendServer(fmt.Sprintf("%s_%d", getDashDomain(site.Domain), index), server.Host, server.Port, transaction.ID, backend_http.Name, server.IsSSL)
			if err != nil {
				return fmt.Errorf("Error al crear el servidor backend: %v", err)
			}
		}
	}

	// handle https
	if site.EnableHTTPS {
		// add cert
		err = s.addSiteCert(site)
		if err != nil {
			return fmt.Errorf("Error al añadir el certificado: %v", err)
		}
		// cert load
		crtLoad := &models.CrtLoad{
			Certificate: site.Domain + ".crt",
			Key:         site.Domain + ".key",
			Alias:       fmt.Sprintf("%s_cert", getDashDomain(site.Domain)),
		}
		err = s.confClient.CreateCrtLoad("sites", crtLoad, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al cargar el certificado: %v", err)
		}

		// change bind
		_, https_bind, err := s.confClient.GetBind("internal_https", "frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("Error al obtener el binding: %v", err)
		}

		if len(https_bind.BindParams.DefaultCrtList) > 0 {
			https_bind.BindParams.DefaultCrtList = append(https_bind.BindParams.DefaultCrtList, fmt.Sprintf("@sites/%s_cert", getDashDomain(site.Domain)))
		} else {
			https_bind.BindParams.DefaultCrtList = []string{
				fmt.Sprintf("@sites/%s_cert", getDashDomain(site.Domain)),
			}
		}
		https_bind.Ssl = true

		err = s.confClient.EditBind("internal_https", "frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), https_bind, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al modificar el binding: %v", err)
		}

		_, aclList, err := s.confClient.GetACLs("frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("Error al obtener ACL: %v", err)
		}
		aclIndex := len(aclList)
		// add ack and rule backend
		acl_https := &models.ACL{
			ACLName:   fmt.Sprintf("host_%s", getDashDomain(site.Domain)), // Usar el campo ACLName
			Criterion: "hdr(host) -i -m end",                              // Modificar el campo Criterion para usar -m end
			Value:     site.Domain,                                        // Añadir un punto delante del dominio
			// Criterion: "hdr(host) -i",                                     // usar el campo Criterion
			// Value:     site.Domain,                                        // usar el campo Value
		}
		err = s.confClient.CreateACL(int64(aclIndex), "frontend", fmt.Sprintf("fe_%d_https", site.ListenPort), acl_https, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al crear ACL: %v", err)
		}

		_, switchingRules, err := s.confClient.GetBackendSwitchingRules(fmt.Sprintf("fe_%d_https", site.ListenPort), "")
		if err != nil {
			return fmt.Errorf("Error al obtener la regla de cambio de backend: %v", err)
		}
		switchingRuleIndex := len(switchingRules)
		httpsUseBackendRule := &models.BackendSwitchingRule{
			Name:     fmt.Sprintf("be_%s", getDashDomain(site.Domain)),
			Cond:     "if",
			CondTest: acl_https.ACLName,
		}
		err = s.confClient.CreateBackendSwitchingRule(int64(switchingRuleIndex), fmt.Sprintf("fe_%d_https", site.ListenPort), httpsUseBackendRule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al crear la regla de cambio de backend: %v", err)
		}

	}

	transaction, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("Error al confirmar la transacción: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil
}

func (s *HAProxyServiceImpl) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.stopHAProxy()

}

func (s *HAProxyServiceImpl) Reload() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.reloadHAProxy()
}

func (s *HAProxyServiceImpl) HotReloadRemoveConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Archivos a eliminar, excluyendo PidFile y SocketFile
	filesToRemove := []string{
		s.HAProxyConfigFile,
		s.SpoeConfigFile,
	}

	// Guardar las rutas absolutas de PidFile y SocketFile para comparar después
	pidFileAbs, _ := filepath.Abs(s.PidFile)
	socketFileAbs, _ := filepath.Abs(s.SocketFile)

	// Directorios a eliminar
	dirsToRemove := []string{
		s.TransactionDir,
		s.SpoeTransactionDir,
		s.CertDir,
	}

	// Tratamiento especial de filepath.Dir(s.HAProxyConfigFile)
	configDir := filepath.Dir(s.HAProxyConfigFile)
	if configDir != "" {
		// Listar primero todos los archivos del directorio
		entries, err := os.ReadDir(configDir)
		if err == nil {
			// Eliminar archivos uno a uno, omitiendo PidFile y SocketFile
			for _, entry := range entries {
				if entry.IsDir() {
					continue // Omitir subdirectorios, los directorios se procesan por separado
				}

				filePath := filepath.Join(configDir, entry.Name())
				filePathAbs, _ := filepath.Abs(filePath)

				// Comprobar si es el archivo a conservar
				if filePathAbs == pidFileAbs || filePathAbs == socketFileAbs {
					continue // Omitir PidFile y SocketFile
				}

				s.logger.Info().Msgf("Eliminando archivo: %s", filePath)
				if err := os.Remove(filePath); err != nil {
					s.logger.Error().Msgf("Error al eliminar el archivo %s: %v", filePath, err)
					// Seguir eliminando otros archivos, sin devolver error inmediatamente
				}
			}
		}
	}

	// Eliminar archivo
	for _, file := range filesToRemove {
		if file == "" {
			continue // Omitir rutas vacías
		}

		// Comprobar si el archivo existe
		if _, err := os.Stat(file); err == nil {
			// Confirmar que no es un archivo a conservar
			fileAbs, _ := filepath.Abs(file)
			if fileAbs == pidFileAbs || fileAbs == socketFileAbs {
				continue // Omitir PidFile y SocketFile
			}

			s.logger.Info().Msgf("Eliminando archivo: %s", file)
			if err := os.Remove(file); err != nil {
				s.logger.Error().Msgf("Error al eliminar el archivo %s: %v", file, err)
				// Seguir eliminando otros archivos, sin devolver error inmediatamente
			}
		}
	}

	// Como debemos conservar ciertos archivos de configDir, no se puede eliminar configDir directamente
	// Eliminar otros directorios
	for _, dir := range dirsToRemove {
		if dir == "" || dir == configDir {
			continue // Omitir rutas vacías y el directorio de configuración
		}

		// Comprobar si el directorio existe
		if _, err := os.Stat(dir); err == nil {
			s.logger.Info().Msgf("Eliminando directorio: %s", dir)
			if err := os.RemoveAll(dir); err != nil {
				s.logger.Error().Msgf("Error al eliminar el directorio %s: %v", dir, err)
				// Seguir eliminando otros directorios, sin devolver error inmediatamente
			}
		}
	}

	s.logger.Info().Msg("Limpieza de configuración completada")
	return nil
}

func (s *HAProxyServiceImpl) RemoveConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Archivos a eliminar
	filesToRemove := []string{
		s.HAProxyConfigFile,
		s.SpoeConfigFile,
		s.PidFile,
		s.SocketFile,
	}

	// Directorios a eliminar
	dirsToRemove := []string{
		filepath.Dir(s.HAProxyConfigFile),
		s.TransactionDir,
		s.SpoeTransactionDir,
		s.CertDir,
	}

	// Eliminar archivo
	for _, file := range filesToRemove {
		if file == "" {
			continue // Omitir rutas vacías
		}

		// Comprobar si el archivo existe
		if _, err := os.Stat(file); err == nil {
			s.logger.Info().Msgf("Eliminando archivo: %s", file)
			if err := os.Remove(file); err != nil {
				s.logger.Error().Msgf("Error al eliminar el archivo %s: %v", file, err)
				// Seguir eliminando otros archivos, sin devolver error inmediatamente
			}
		}
	}

	// Eliminar directorio
	for _, dir := range dirsToRemove {
		if dir == "" {
			continue // Omitir rutas vacías
		}

		// Comprobar si el directorio existe
		if _, err := os.Stat(dir); err == nil {
			s.logger.Info().Msgf("Eliminando directorio: %s", dir)
			if err := os.RemoveAll(dir); err != nil {
				s.logger.Error().Msgf("Error al eliminar el directorio %s: %v", dir, err)
				// Seguir eliminando otros directorios, sin devolver error inmediatamente
			}
		}
	}

	s.logger.Info().Msg("Limpieza de configuración completada")
	return nil
}

// InitHAProxyConfig inicializa la configuración de HAProxy
func (s *HAProxyServiceImpl) InitHAProxyConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Crear los directorios necesarios
	dirs := []string{
		filepath.Dir(s.HAProxyConfigFile),
		s.TransactionDir,
		s.SpoeDir,
		s.SpoeTransactionDir,
		s.CertDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("unable to create directory %s: %v", dir, err)
		}
	}

	if _, err := os.Stat(s.HAProxyConfigFile); err == nil {
		// El archivo existe, devolver error
		return fmt.Errorf("El archivo de configuración de haproxy ya existe: %s", s.HAProxyConfigFile)
	} else if !os.IsNotExist(err) {
		// Ocurrió un error distinto de "El archivo no existe"
		return fmt.Errorf("Error al comprobar el archivo de configuración de haproxy: %v", err)
	}

	username := os.Getenv("USER")
	if username == "" {
		username = "haproxy"
	}

	// Definición de la plantilla de configuración
	configTemplate := `# _version = 1
global
    log stdout format raw local0
{{if gt .Thread 0}}    nbthread {{.Thread}} # Número de hilos
{{end}} 
    # user {{.Username}}
    # group {{.Username}}
	# maxconn 4000 # número máximo de conexiones
defaults http
    mode http
    log global
    option httplog
    
    # Configuración del modo de cierre de conexiones del servidor
    option http-server-close    # cierre de conexión del lado del servidor, optimiza la reutilización
    
    # Configuración básica de timeouts
    timeout connect 5s          # tiempo de espera de conexión
    timeout client 30s          # tiempo de espera del cliente
    timeout server 30s          # tiempo de espera del servidor
    timeout tunnel 1h           # timeout del túnel WebSocket - ajuste clave
    
    # Timeouts relacionados con HTTP
    timeout http-request 10s    # timeout de procesamiento de petición HTTP
    timeout http-keep-alive 10s # timeout de keep-alive HTTP
    
    # Otras opciones de optimización
    option forwardfor           # reenviar la IP real del cliente
    option dontlognull          # no registrar conexiones vacías
    option redispatch           # redistribuir en caso de fallo del servidor
    retries 3                   # reintentos ante fallo de conexión
    
    # Configuración de balanceo de carga - optimizado para WebSocket
    balance leastconn           # algoritmo de menos conexiones, ideal para conexiones largas WebSocket
defaults tcp
    mode tcp
    log global
    option tcplog
    
    # Configuración básica de timeouts
    timeout connect 5s          # timeout de conexión
    timeout client 3h           # timeout de cliente extendido, para conexiones largas
    timeout server 3h           # timeout de servidor extendido
    
    # Configuración de keep-alive TCP
    option tcpka                # habilitar keep-alive TCP
    option clitcpka             # keep-alive del lado del cliente
    option srvtcpka             # keep-alive del lado del servidor
    
    # Otros ajustes
    option dontlognull          # no registrar conexiones vacías
    retries 3                   # reintentos ante fallo de conexión
    
    # Configuración de balanceo de carga
    balance leastconn           # algoritmo óptimo para conexiones largas
frontend stats from http
  mode http
  bind *:8404
  stats enable
  stats uri /stats
  stats show-modules
# The following part will be dynamically configured
`
	// Preparar los datos de la plantilla
	data := struct {
		Username string
		Thread   int
	}{
		Username: username,
		Thread:   s.thread,
	}

	// Analizar la plantilla
	tmpl, err := template.New("haproxy-config").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Ejecutar la plantilla y escribir el archivo
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if err := os.WriteFile(s.HAProxyConfigFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to create basic config file: %v", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) InitSpoeConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.ensureSpoeClient(); err != nil {
		return fmt.Errorf("Error al inicializar el cliente SPOE: %v", err)
	}

	spoeFileName := filepath.Base(s.SpoeConfigFile)

	if _, err := os.Stat(s.SpoeConfigFile); err == nil {
		// El archivo existe, devolver error
		return fmt.Errorf("El archivo de configuración SPOE ya existe: %s", s.SpoeConfigFile)
	} else if !os.IsNotExist(err) {
		// Ocurrió un error distinto de "El archivo no existe"
		return fmt.Errorf("Error al comprobar el archivo de configuración SPOE: %v", err)
	}

	emptyReader := bytes.NewReader([]byte{})
	readCloser := io.NopCloser(emptyReader)

	_, err := s.spoeClient.Create(spoeFileName, readCloser)
	if err != nil {
		return fmt.Errorf("Error al crear el archivo de configuración SPOE: %v", err)
	}

	singleSpoe, err := s.spoeClient.GetSingleSpoe(spoeFileName)
	if err != nil {
		return fmt.Errorf("Error al obtener la configuración SPOE: %v", err)
	}
	version, err := singleSpoe.Transaction.TransactionClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("Error al obtener la versión SPOE: %v", err)
	}
	transaction, err := singleSpoe.Transaction.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("Error al iniciar la transacción SPOE: %v", err)
	}
	scopeName := models.SpoeScope("[coraza]")
	err = singleSpoe.CreateScope(&scopeName, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el ámbito SPOE: %v", err)
	}

	agent := &models.SpoeAgent{
		Name: StringP("coraza-agent"),
		// Según isResponseCheck, decidir si incluir procesamiento de respuestas
		Messages: func() string {
			if s.isResponseCheck {
				return "coraza-req coraza-res"
			}
			return "coraza-req"
		}(),
		OptionVarPrefix:   "coraza",
		OptionSetOnError:  "error",
		HelloTimeout:      2000,   // 2s (milisegundos)
		IdleTimeout:       120000, // 2m (milisegundos)
		ProcessingTimeout: 500,    // 500ms
		UseBackend:        "coraza-spoa",
		Log:               models.LogTargets{&models.LogTarget{Global: true}},
	}

	err = singleSpoe.CreateAgent(string(scopeName), agent, transaction.ID, 0)
	if err != nil {
		singleSpoe.Transaction.DeleteTransaction(transaction.ID)
		return fmt.Errorf("Error al crear el agente SPOE: %v", err)
	}

	// Crear mensaje coraza-req
	reqEvent := &models.SpoeMessageEvent{
		Name: StringP("on-frontend-http-request"),
	}
	reqMsg := &models.SpoeMessage{
		Name:  StringP("coraza-req"),
		Event: reqEvent,
		Args:  "app=str(coraza) src-ip=src src-port=src_port dst-ip=dst dst-port=dst_port method=method path=path query=query version=req.ver headers=req.hdrs body=req.body",
	}

	// Crear message bajo la sección coraza
	err = singleSpoe.CreateMessage(string(scopeName), reqMsg, transaction.ID, 0)
	if err != nil {
		singleSpoe.Transaction.DeleteTransaction(transaction.ID)
		return fmt.Errorf("Error al crear el mensaje de petición SPOE: %v", err)
	}

	// Crear mensaje coraza-res
	if s.isResponseCheck {
		resEvent := &models.SpoeMessageEvent{
			Name: StringP("on-http-response"),
		}
		resMsg := &models.SpoeMessage{
			Name:  StringP("coraza-res"),
			Event: resEvent,
			Args:  "app=str(coraza) id=var(txn.coraza.id) version=res.ver status=status headers=res.hdrs body=res.body",
		}

		err = singleSpoe.CreateMessage(string(scopeName), resMsg, transaction.ID, 0)
		if err != nil {
			singleSpoe.Transaction.DeleteTransaction(transaction.ID)
			return fmt.Errorf("Error al crear el mensaje de respuesta SPOE: %v", err)
		}

	}

	_, err = singleSpoe.Transaction.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("Error al confirmar la transacción SPOE: %v", err)
	}

	singleSpoe.Transaction.DeleteTransaction(transaction.ID)
	return nil
}

func (s *HAProxyServiceImpl) CreateHAProxyCrtStore() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Comprobar si existe el archivo de configuración de HAProxy
	if _, err := os.Stat(s.HAProxyConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("El archivo de configuración de HAProxy no existe: %s", s.HAProxyConfigFile)
	}

	// Asegurar que el cliente de configuración se inicialice
	if err := s.ensureConfClient(); err != nil {
		return err
	}
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("Error al obtener la versión: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("Error al iniciar la transacción: %v", err)
	}

	crtStore := &models.CrtStore{
		Name:    "sites",
		CrtBase: s.CertDir, // Directorio de certificados
		KeyBase: s.CertDir, // Directorio de claves privadas
	}
	err = s.confClient.CreateCrtStore(crtStore, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el almacén de certificados: %v", err)
	}
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("Error al confirmar la transacción: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil
}

func (s *HAProxyServiceImpl) AddCorazaBackend() error {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Comprobar si existe el archivo de configuración de HAProxy
	if _, err := os.Stat(s.HAProxyConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("El archivo de configuración de HAProxy no existe: %s", s.HAProxyConfigFile)
	}

	// Asegurar que el cliente de configuración se inicialice
	if err := s.ensureConfClient(); err != nil {
		return err
	}
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("Error al obtener la versión: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("Error al iniciar la transacción: %v", err)
	}

	coraza_backend := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    "coraza-spoa",
			Mode:    "tcp",
			From:    "tcp",
			Enabled: true,
		},
	}
	err = s.confClient.CreateBackend(coraza_backend, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el backend: %v", err)
	}

	server_coraza := &models.Server{
		Name:    "coraza-agent",
		Address: s.SpoeAgentAddress,      // Obtener la dirección de la estructura, admite dominio o IP
		Port:    Int64P(s.SpoeAgentPort), // Obtener el puerto de la estructura
	}
	err = s.confClient.CreateServer("backend", coraza_backend.Name, server_coraza, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el servidor: %v", err)
	}

	transaction, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("Error al confirmar la transacción: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil
}

func (s *HAProxyServiceImpl) Reset() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	appConfig, err := config.GetAppConfig()
	if err != nil {
		return fmt.Errorf("Error al obtener la configuración de la aplicación: %v", err)
	}

	s.thread = appConfig.Haproxy.Thread
	s.isResponseCheck = appConfig.IsResponseCheck
	s.isDebug = appConfig.IsDebug
	s.isK8s = appConfig.IsK8s

	if err := s.resetClients(); err != nil {
		return fmt.Errorf("Error al reiniciar el cliente: %v", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) GetStats() (models.NativeStats, error) {
	stats, err := s.getHAProxyStats()
	if err != nil {
		return models.NativeStats{}, err
	}
	return stats, nil
}

// ========================== internal method ==========================
func (s *HAProxyServiceImpl) initConfClient() error {
	confClient, err := configuration.New(s.ctx,
		cfg_opt.ConfigurationFile(s.HAProxyConfigFile),
		cfg_opt.HAProxyBin(s.HaproxyBin),
		cfg_opt.Backups(s.BackupsNumber),
		cfg_opt.UsePersistentTransactions,
		cfg_opt.TransactionsDir(s.TransactionDir),
		cfg_opt.MasterWorker,
		cfg_opt.UseMd5Hash,
	)
	if err != nil {
		return fmt.Errorf("init conf client failed: %v", err)
	}
	s.confClient = confClient
	return nil
}

func (s *HAProxyServiceImpl) initSpoeClient() error {
	prms := spoe.Params{
		SpoeDir:        s.SpoeDir,
		TransactionDir: s.SpoeTransactionDir,
	}
	spoeClient, err := spoe.NewSpoe(prms)
	if err != nil {
		return fmt.Errorf("init spoe client failed: %v", err)
	}
	s.spoeClient = spoeClient
	return nil
}

func (s *HAProxyServiceImpl) initRuntimeClient() error {
	// Comprobar si existe el archivo de socket
	if _, err := os.Stat(s.SocketFile); os.IsNotExist(err) {
		return fmt.Errorf("El archivo de socket no existe: %s", s.SocketFile)
	}

	ms := runtime_options.MasterSocket(s.SocketFile)
	runtimeClient, err := runtime_api.New(s.ctx, ms)
	if err != nil {
		return fmt.Errorf("init runtime client failed: %v", err)
	}
	s.runtimeClient = runtimeClient
	return nil
}

// ensureConfClient garantiza que el cliente de configuración esté inicializado
func (s *HAProxyServiceImpl) ensureConfClient() error {
	if s.confClient == nil {
		if err := s.initConfClient(); err != nil {
			return fmt.Errorf("Error al inicializar el cliente de configuración: %v", err)
		}
	}
	return nil
}
func (s *HAProxyServiceImpl) ensureSpoeClient() error {
	if s.spoeClient == nil {
		if err := s.initSpoeClient(); err != nil {
			return fmt.Errorf("Error al inicializar el cliente SPOE: %v", err)
		}
	}
	return nil
}

func (s *HAProxyServiceImpl) ensureRuntimeClient() error {
	if s.runtimeClient == nil {
		if err := s.initRuntimeClient(); err != nil {
			return fmt.Errorf("Error al inicializar el cliente de runtime: %v", err)
		}
	}
	return nil
}

func (s *HAProxyServiceImpl) ensureClientNative() error {
	if s.clientNative == nil {
		if err := s.initClients(); err != nil {
			return fmt.Errorf("Error al inicializar el cliente completo: %v", err)
		}
	}
	return nil
}

func (s *HAProxyServiceImpl) initClients() error {
	// Inicializar el cliente de configuración
	if err := s.initConfClient(); err != nil {
		return fmt.Errorf("init conf client failed: %v", err)
	}

	// Inicializar el cliente SPOE
	if err := s.initSpoeClient(); err != nil {
		return fmt.Errorf("init spoe client failed: %v", err)
	}

	// Inicializar el cliente de runtime
	if err := s.initRuntimeClient(); err != nil {
		return fmt.Errorf("init runtime client failed: %v", err)
	}

	// Cliente combinado
	clientOpts := []options.Option{
		options.Configuration(s.confClient),
		options.Runtime(s.runtimeClient),
		options.Spoe(s.spoeClient),
	}

	clientNative, err := client_native.New(s.ctx, clientOpts...)
	if err != nil {
		return fmt.Errorf("init client failed: %v", err)
	}
	s.clientNative = clientNative

	return nil
}

func (s *HAProxyServiceImpl) resetClients() error {
	s.confClient = nil
	s.spoeClient = nil
	s.runtimeClient = nil
	s.clientNative = nil

	return nil
}

// stopHAProxy detiene el proceso HAProxy
func (s *HAProxyServiceImpl) stopHAProxy() error {
	// Si la instancia guarda el comando HAProxy, usarlo para terminar el proceso
	if s.haproxyCmd != nil && s.haproxyCmd.Process != nil {
		// Intentar terminar el proceso de forma elegante
		if err := s.haproxyCmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("Error al enviar la señal de interrupción: %v", err)
			// Terminación forzada
			if err := s.haproxyCmd.Process.Kill(); err != nil {
				log.Printf("Error al forzar la terminación del proceso: %v", err)
			}
		}

		// Esperar a que el proceso termine completamente
		s.haproxyCmd.Wait()
		s.haproxyCmd = nil
	} else {
		// Intentar leer el archivo PID
		pid, err := s.getHAProxyPid()
		if err == nil && pid > 0 {
			process, err := os.FindProcess(pid)
			if err == nil {
				// Intentar terminar el proceso de forma elegante
				if err := process.Signal(os.Interrupt); err != nil {
					log.Printf("Error al enviar la señal de interrupción: %v", err)
					// Terminación forzada
					if err := process.Kill(); err != nil {
						log.Printf("Error al forzar la terminación del proceso: %v", err)
					}
				}

				// Esperar a que el proceso salga
				_, err = process.Wait()
				if err != nil {
					log.Printf("Error esperando a que el proceso termine: %v", err)
				}
			}
		} else {
			// No se pudo obtener el PID, intentar pkill
			exec.Command("pkill", "-f", s.HaproxyBin).Run()
		}
	}

	// Restablecer el cliente
	s.resetClients()

	// Eliminar el archivo de socket
	if _, err := os.Stat(s.SocketFile); err == nil {
		if err := os.Remove(s.SocketFile); err != nil {
			s.logger.Error().Msgf("Error al eliminar el archivo de socket: %v", err)
		}
	}

	// Eliminar el archivo PID
	if _, err := os.Stat(s.PidFile); err == nil {
		if err := os.Remove(s.PidFile); err != nil {
			s.logger.Error().Msgf("Error al eliminar el archivo PID: %v", err)
		}
	}

	s.status.Store(int32(StatusStopped))
	return nil
}

func (s *HAProxyServiceImpl) reloadHAProxy() error {

	if err := s.ensureRuntimeClient(); err != nil {
		return fmt.Errorf("Error al inicializar el cliente de runtime: %v", err)
	}
	_, err := s.runtimeClient.Reload()
	if err != nil {
		return fmt.Errorf("Error al recargar HAProxy: %v", err)
	}
	return nil
}

func (s *HAProxyServiceImpl) isHAProxyRunning() (bool, error) {
	// Si la instancia guarda el comando HAProxy, comprobarlo
	if s.haproxyCmd != nil && s.haproxyCmd.Process != nil {
		// Intentar enviar la señal 0 para comprobar si el proceso existe
		if err := s.haproxyCmd.Process.Signal(syscall.Signal(0)); err != nil {
			// El proceso no existe
			s.haproxyCmd = nil
			return false, nil
		}
		return true, nil
	}

	// Intentar leer el archivo PID
	pid, err := s.getHAProxyPid()
	if err != nil || pid <= 0 {
		return false, nil
	}

	// Comprobar si el proceso existe
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	// En sistemas Unix, FindProcess siempre tiene éxito; hay que enviar la señal 0 para confirmar que el proceso existe
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}

	return true, nil
}

// getHAProxyPid obtiene el ID del proceso HAProxy desde el archivo PID
func (s *HAProxyServiceImpl) getHAProxyPid() (int, error) {
	if _, err := os.Stat(s.PidFile); os.IsNotExist(err) {
		return 0, fmt.Errorf("El archivo PID no existe")
	}

	pidBytes, err := os.ReadFile(s.PidFile)
	if err != nil {
		return 0, fmt.Errorf("Error al leer el archivo PID: %v", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return 0, fmt.Errorf("Error al analizar el PID: %v", err)
	}

	return pid, nil
}

func (s *HAProxyServiceImpl) addSiteCert(site model.Site) error {
	// Comprobar si HTTPS está habilitado y la información del certificado no está vacía
	if !site.EnableHTTPS || site.Certificate.PublicKey == "" || site.Certificate.PrivateKey == "" {
		return fmt.Errorf("site cert invalid or not enable https")
	}

	// Asegurar que el directorio de certificados exista
	if err := os.MkdirAll(s.CertDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Construir la ruta del archivo de certificado
	certPath := filepath.Join(s.CertDir, site.Domain+".crt")
	keyPath := filepath.Join(s.CertDir, site.Domain+".key")

	// Escribir el archivo de certificado de clave pública (modo sobrescritura)
	if err := os.WriteFile(certPath, []byte(site.Certificate.PublicKey), 0644); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Escribir el archivo de clave privada (modo sobrescritura)
	if err := os.WriteFile(keyPath, []byte(site.Certificate.PrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) removeSiteCert(site model.Site) error {
	// Construir la ruta del archivo de certificado
	certPath := filepath.Join(s.CertDir, site.Domain+".crt")
	keyPath := filepath.Join(s.CertDir, site.Domain+".key")

	// Eliminar archivo de certificado, ignorar si no existe
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove certificate file: %w", err)
	}

	// Eliminar el archivo de clave privada, ignorar si no existe
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove private key file: %w", err)
	}

	return nil
}

func (s *HAProxyServiceImpl) getFeCombined(port int) (string, error) {
	if err := s.ensureConfClient(); err != nil {
		return "", fmt.Errorf("Error al inicializar el cliente de configuración: %v", err)
	}

	_, frontend, err := s.confClient.GetFrontend(fmt.Sprintf("fe_%d_combined", port), "")
	if err != nil {
		return "", fmt.Errorf("Error al obtener el frontend: %v", err)
	}
	return frontend.Name, nil
}

func (s *HAProxyServiceImpl) createFeCombined(port int, isHttpsRedirect bool) error {
	// Asegurar que el cliente de configuración se inicialice
	if err := s.ensureConfClient(); err != nil {
		return fmt.Errorf("Error al inicializar el cliente de configuración: %v", err)
	}
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("Error al obtener la versión: %v", err)
	}
	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("Error al iniciar la transacción: %v", err)
	}

	// Crear fe_(port)_combined
	fe_combined := &models.Frontend{
		FrontendBase: models.FrontendBase{
			Name:           fmt.Sprintf("fe_%d_combined", port),
			Mode:           "tcp",
			DefaultBackend: fmt.Sprintf("be_%d_https", port), // Establecer el backend por defecto
			Enabled:        true,
			From:           "tcp",
		},
	}
	err = s.confClient.CreateFrontend(fe_combined, transaction.ID, 0)

	if err != nil {
		return fmt.Errorf("Error al crear el frontend: %v", err)
	}

	bind := &models.Bind{
		BindParams: models.BindParams{
			Name: fmt.Sprintf("combined_%d", port),
		},
		Address: "*",
		Port:    Int64P(int64(port)),
	}
	err = s.confClient.CreateBind("frontend", fe_combined.Name, bind, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el binding: %v", err)
	}

	// rule

	tcpInspectDelay := &models.TCPRequestRule{
		Type:    "inspect-delay",
		Timeout: Int64P(2),
	}
	err = s.confClient.CreateTCPRequestRule(0, "frontend", fe_combined.Name, tcpInspectDelay, transaction.ID, 0)

	if err != nil {
		return fmt.Errorf("Error al crear la regla de petición TCP: %v", err)
	}

	tcpAcceptHTTP := &models.TCPRequestRule{
		Action:   "accept",
		Type:     "content",
		Cond:     "if",
		CondTest: "HTTP",
	}

	err = s.confClient.CreateTCPRequestRule(1, "frontend", fe_combined.Name, tcpAcceptHTTP, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear la regla de petición TCP: %v", err)
	}

	tcpAcceptSSL := &models.TCPRequestRule{
		Action:   "accept",
		Type:     "content",
		Cond:     "if",
		CondTest: "{ req.ssl_hello_type 1 }",
	}
	err = s.confClient.CreateTCPRequestRule(2, "frontend", fe_combined.Name, tcpAcceptSSL, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear la regla de petición TCP: %v", err)
	}

	// backend
	useBackendRule := &models.BackendSwitchingRule{
		Name:     fmt.Sprintf("be_%d_http", port),
		Cond:     "if",
		CondTest: "HTTP",
	}
	err = s.confClient.CreateBackendSwitchingRule(0, fe_combined.Name, useBackendRule, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear la regla de cambio de backend: %v", err)
	}

	// abstract backend abns@haproxy-{port}-http
	be_http := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    fmt.Sprintf("be_%d_http", port),
			Mode:    "tcp",
			Enabled: true,
			From:    "tcp",
		},
	}
	err = s.confClient.CreateBackend(be_http, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el backend: %v", err)
	}

	serverHTTP := &models.Server{
		ServerParams: models.ServerParams{
			SendProxyV2: "enabled", // Habilitar protocolo proxy v2
		},
		Name:    "loopback-for-http",
		Address: fmt.Sprintf("abns@haproxy-%d-http", port),
		Port:    Int64P(1000),
	}

	err = s.confClient.CreateServer("backend", be_http.Name, serverHTTP, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el servidor: %v", err)
	}

	be_https := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    fmt.Sprintf("be_%d_https", port),
			Mode:    "tcp",
			Enabled: true,
			From:    "tcp",
		},
	}
	err = s.confClient.CreateBackend(be_https, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el backend: %v", err)
	}

	serverHTTPS := &models.Server{
		ServerParams: models.ServerParams{
			SendProxyV2: "enabled", // Habilitar protocolo proxy v2
		},
		Name:    "loopback-for-https",
		Address: fmt.Sprintf("abns@haproxy-%d-https", port),
		Port:    Int64P(1000),
	}

	err = s.confClient.CreateServer("backend", be_https.Name, serverHTTPS, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el servidor: %v", err)
	}

	// create fe_(port)_http  fe_(port)_https

	// fe_(port)_http
	fe_http := &models.Frontend{
		FrontendBase: models.FrontendBase{
			Name:           fmt.Sprintf("fe_%d_http", port),
			Mode:           "http",
			DefaultBackend: fmt.Sprintf("p%d_backend", port),
			Enabled:        true,
			From:           "http",
			// El formato de log usa barra invertida para escapar espacios y caracteres especiales
			LogFormat: "\"%ci:%cp\\ [%t]\\ %ft\\ %b/%s\\ %Th/%Ti/%TR/%Tq/%Tw/%Tc/%Tr/%Tt\\ %ST\\ %B\\ %CC\\ %CS\\ %tsc\\ %ac/%fc/%bc/%sc/%rc\\ %sq/%bq\\ %hr\\ %hs\\ %{+Q}r\\ %[var(txn.coraza.id)]\\ spoa-error:\\ %[var(txn.coraza.error)]\\ waf-hit:\\ %[var(txn.coraza.status)]\"",
			Forwardfor: &models.Forwardfor{
				Enabled: StringP("enabled"),
				Ifnone:  true,
			},
		},
	}
	err = s.confClient.CreateFrontend(fe_http, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el frontend: %v", err)
	}

	fe_http_bind := &models.Bind{
		BindParams: models.BindParams{
			Name:        "internal_http",
			AcceptProxy: true,
		},
		Port:    Int64P(1000),
		Address: fmt.Sprintf("abns@haproxy-%d-http", port),
	}

	err = s.confClient.CreateBind("frontend", fe_http.Name, fe_http_bind, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el binding: %v", err)
	}

	// Añadir filtro spoe
	fe_http_filter := &models.Filter{
		Type:       "spoe",           // Tipo de filtro
		SpoeEngine: "coraza",         // Nombre del motor SPOE
		SpoeConfig: s.SpoeConfigFile, // Usar la ruta estándar del archivo de configuración
	}
	err = s.confClient.CreateFilter(0, "frontend", fe_http.Name, fe_http_filter, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el filtro: %v", err)
	}

	// Añadir regla de petición HTTP
	var fe_http_request_rule []struct {
		index int64
		rule  *models.HTTPRequestRule
	}
	if isHttpsRedirect {
		fe_http_request_rule = []struct {
			index int64
			rule  *models.HTTPRequestRule
		}{
			{0, &models.HTTPRequestRule{
				Type:       "redirect",
				RedirCode:  Int64P(301),
				RedirType:  "scheme",
				RedirValue: "https",
			}},
			{1, &models.HTTPRequestRule{
				Type:       "redirect",
				RedirCode:  Int64P(302),
				RedirType:  "location", // Especificar el tipo de redirección
				RedirValue: "%[var(txn.coraza.data)]",
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str redirect }",
			}},
			{2, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(403),
				HdrName:    "waf-block", // Establecer el nombre del header
				HdrFormat:  "request",   // Establecer el valor del header
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str deny }",
			}},
			{3, &models.HTTPRequestRule{
				Type:     "silent-drop",
				Cond:     "if",
				CondTest: "{ var(txn.coraza.action) -m str drop }",
			}},
			{4, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(500),
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
			}},
		}
	} else {

		fe_http_request_rule = []struct {
			index int64
			rule  *models.HTTPRequestRule
		}{
			{0, &models.HTTPRequestRule{
				Type:       "redirect",
				RedirCode:  Int64P(302),
				RedirType:  "location", // Especificar el tipo de redirección
				RedirValue: "%[var(txn.coraza.data)]",
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str redirect }",
			}},
			{1, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(403),
				HdrName:    "waf-block", // Establecer el nombre del header
				HdrFormat:  "request",   // Establecer el valor del header
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.action) -m str deny }",
			}},
			{2, &models.HTTPRequestRule{
				Type:     "silent-drop",
				Cond:     "if",
				CondTest: "{ var(txn.coraza.action) -m str drop }",
			}},
			{3, &models.HTTPRequestRule{
				Type:       "deny",
				DenyStatus: Int64P(500),
				Cond:       "if",
				CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
			}},
		}

	}

	for i, item := range fe_http_request_rule {
		err = s.confClient.CreateHTTPRequestRule(item.index, "frontend", fe_http.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al añadir la regla de petición HTTP #%d: %v", i, err)
		}
	}

	// Añadir regla de respuesta HTTP - asegurar la estructura correcta de reglas de respuesta HTTP
	fe_http_response_rule := []struct {
		index int64
		rule  *models.HTTPResponseRule
	}{
		{0, &models.HTTPResponseRule{
			Type:       "redirect",
			RedirCode:  Int64P(302),
			RedirType:  "location", // Especificar el tipo de redirección
			RedirValue: "%[var(txn.coraza.data)]",
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str redirect }",
		}},
		{1, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(403),
			HdrName:    "waf-block", // Establecer el nombre del header
			HdrFormat:  "response",  // Establecer el valor del header
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str deny }",
		}},
		{2, &models.HTTPResponseRule{
			Type:     "silent-drop",
			Cond:     "if",
			CondTest: "{ var(txn.coraza.action) -m str drop }",
		}},
		{3, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(500),
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
		}},
	}

	for i, item := range fe_http_response_rule {
		err = s.confClient.CreateHTTPResponseRule(item.index, "frontend", fe_http.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al añadir la regla de respuesta HTTP #%d: %v", i, err)
		}
	}

	// fe_(port)_https
	fe_https := &models.Frontend{
		FrontendBase: models.FrontendBase{
			Name:           fmt.Sprintf("fe_%d_https", port),
			Mode:           "http",
			DefaultBackend: fmt.Sprintf("p%d_backend", port),
			Enabled:        true,
			From:           "http",
			// El formato de log usa barra invertida para escapar espacios y caracteres especiales
			LogFormat: "\"%ci:%cp\\ [%t]\\ %ft\\ %b/%s\\ %Th/%Ti/%TR/%Tq/%Tw/%Tc/%Tr/%Tt\\ %ST\\ %B\\ %CC\\ %CS\\ %tsc\\ %ac/%fc/%bc/%sc/%rc\\ %sq/%bq\\ %hr\\ %hs\\ %{+Q}r\\ %[var(txn.coraza.id)]\\ spoa-error:\\ %[var(txn.coraza.error)]\\ waf-hit:\\ %[var(txn.coraza.status)]\"",
			Forwardfor: &models.Forwardfor{
				Enabled: StringP("enabled"),
				Ifnone:  true,
			},
		},
	}
	err = s.confClient.CreateFrontend(fe_https, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el frontend: %v", err)
	}
	fe_https_bind := &models.Bind{
		BindParams: models.BindParams{
			Name:        "internal_https",
			AcceptProxy: true,
		},
		Port:    Int64P(1000),
		Address: fmt.Sprintf("abns@haproxy-%d-https", port),
	}
	err = s.confClient.CreateBind("frontend", fe_https.Name, fe_https_bind, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el binding: %v", err)
	}

	// Añadir filtro spoe
	fe_https_filter := &models.Filter{
		Type:       "spoe",           // Tipo de filtro
		SpoeEngine: "coraza",         // Nombre del motor SPOE
		SpoeConfig: s.SpoeConfigFile, // Usar la ruta estándar del archivo de configuración
	}
	err = s.confClient.CreateFilter(0, "frontend", fe_https.Name, fe_https_filter, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el filtro: %v", err)
	}

	// Añadir regla de petición HTTPs
	fe_https_request_rule := []struct {
		index int64
		rule  *models.HTTPRequestRule
	}{
		{0, &models.HTTPRequestRule{
			Type:       "redirect",
			RedirCode:  Int64P(302),
			RedirType:  "location", // Especificar el tipo de redirección
			RedirValue: "%[var(txn.coraza.data)]",
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str redirect }",
		}},
		{1, &models.HTTPRequestRule{
			Type:       "deny",
			DenyStatus: Int64P(403),
			HdrName:    "waf-block", // Establecer el nombre del header
			HdrFormat:  "request",   // Establecer el valor del header
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str deny }",
		}},
		{2, &models.HTTPRequestRule{
			Type:     "silent-drop",
			Cond:     "if",
			CondTest: "{ var(txn.coraza.action) -m str drop }",
		}},
		{3, &models.HTTPRequestRule{
			Type:       "deny",
			DenyStatus: Int64P(500),
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
		}},
	}

	for i, item := range fe_https_request_rule {
		err = s.confClient.CreateHTTPRequestRule(item.index, "frontend", fe_https.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al añadir la regla de petición HTTP #%d: %v", i, err)
		}
	}

	// Añadir regla de respuesta HTTPs - asegurar la estructura correcta de reglas de respuesta HTTP
	fe_https_response_rule := []struct {
		index int64
		rule  *models.HTTPResponseRule
	}{
		{0, &models.HTTPResponseRule{
			Type:       "redirect",
			RedirCode:  Int64P(302),
			RedirType:  "location", // Especificar el tipo de redirección
			RedirValue: "%[var(txn.coraza.data)]",
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str redirect }",
		}},
		{1, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(403),
			HdrName:    "waf-block", // Establecer el nombre del header
			HdrFormat:  "response",  // Establecer el valor del header
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.action) -m str deny }",
		}},
		{2, &models.HTTPResponseRule{
			Type:     "silent-drop",
			Cond:     "if",
			CondTest: "{ var(txn.coraza.action) -m str drop }",
		}},
		{3, &models.HTTPResponseRule{
			Type:       "deny",
			DenyStatus: Int64P(500),
			Cond:       "if",
			CondTest:   "{ var(txn.coraza.error) -m int gt 0 }",
		}},
	}

	for i, item := range fe_https_response_rule {
		err = s.confClient.CreateHTTPResponseRule(item.index, "frontend", fe_https.Name, item.rule, transaction.ID, 0)
		if err != nil {
			return fmt.Errorf("Error al añadir la regla de respuesta HTTP #%d: %v", i, err)
		}
	}

	// default backend
	be_default := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    fmt.Sprintf("p%d_backend", port),
			Mode:    "http",
			Enabled: true,
			From:    "http",
			// Añadir la opción forwarded
			Forwardfor: &models.Forwardfor{
				Enabled: StringP("enabled"),
				Ifnone:  true,
			},
		},
	}
	err = s.confClient.CreateBackend(be_default, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el backend: %v", err)
	}

	be_default_server := &models.Server{
		Name:    "loopback-for-default",
		Address: "httpbin.org",
		Port:    Int64P(80),
	}
	err = s.confClient.CreateServer("backend", be_default.Name, be_default_server, transaction.ID, 0)
	if err != nil {
		return fmt.Errorf("Error al crear el servidor backend: %v", err)
	}

	transaction, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("Error al confirmar la transacción: %v", err)
	}

	s.confClient.DeleteTransaction(transaction.ID)

	return nil

}

func (s *HAProxyServiceImpl) createBackendServer(name, address string, port int, transactionID string, backendName string, isSsl bool) error {
	server := &models.Server{
		Name:    name,
		Address: address,
		Port:    Int64P(int64(port)),
	}

	if isSsl {
		server.ServerParams = models.ServerParams{
			Ssl: "enabled",
			Sni: fmt.Sprintf("str(%s)", address),
			// SslCafile: "",
			Verify: "none", // No verificar el certificado
		}
	}

	return s.confClient.CreateServer("backend", backendName, server, transactionID, 0)

}

// get haproxy stats
func (s *HAProxyServiceImpl) getHAProxyStats() (models.NativeStats, error) {
	if s.runtimeClient == nil {
		return models.NativeStats{}, fmt.Errorf("runtime client not initialized")
	}
	stats := s.runtimeClient.GetStats()
	if stats.Error != "" {
		return models.NativeStats{}, fmt.Errorf("Error al obtener el estado de HAProxy: %s", stats.Error)
	}
	return stats, nil
}

// Int64P devuelve un puntero a int64
func Int64P(v int64) *int64 {
	return &v
}

// StringP devuelve un puntero a string
func StringP(v string) *string {
	return &v
}

// getSafeString obtiene de forma segura el valor de un puntero a string
func GetSafeString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getSafeInt64 obtiene de forma segura el valor de un puntero int64
func GetSafeInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func getDashDomain(domain string) string {
	// Reemplazar los puntos del dominio por guiones bajos
	dashDomain := strings.ReplaceAll(domain, ".", "_")
	return dashDomain
}

func isIPAddress(domain string) bool {
	// Comprobar la dirección IPv4
	ip := net.ParseIP(domain)
	return ip != nil
}
