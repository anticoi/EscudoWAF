package benchmarks

import (
	"bufio"
	"bytes"
	"net/netip"
	"strings"
	"testing"
)

// Estas son funciones privadas importadas del paquete internal, accesibles vía interfaz
// O copiamos la implementación para el benchmark

// Implementación original con bufio.Scanner (para comparación de rendimiento)
func getHeaderValueOld(headers []byte, targetHeader string) (string, error) {
	s := bufio.NewScanner(bytes.NewReader(headers))
	for s.Scan() {
		line := bytes.TrimSpace(s.Bytes())
		if len(line) == 0 {
			continue
		}

		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) != 2 {
			continue
		}

		key, value := bytes.TrimSpace(kv[0]), bytes.TrimSpace(kv[1])
		if strings.EqualFold(string(key), targetHeader) {
			return string(value), nil
		}
	}
	return "", nil
}

// Implementación optimizada (copiada para benchmark)
func getHeaderValueOptimized(headers []byte, targetHeader string) (string, error) {
	// Preprocesar el header objetivo a minúsculas, evita conversiones repetidas
	targetLower := strings.ToLower(targetHeader)
	targetBytes := []byte(targetLower)
	targetLen := len(targetBytes)

	i := 0
	for i < len(headers) {
		// Encontrar el inicio de la línea
		lineStart := i

		// Encontrar el fin de la línea (\n o \r\n)
		for i < len(headers) && headers[i] != '\n' && headers[i] != '\r' {
			i++
		}
		lineEnd := i

		// Omitir saltos de línea
		if i < len(headers) && headers[i] == '\r' {
			i++
		}
		if i < len(headers) && headers[i] == '\n' {
			i++
		}

		// Omitir líneas vacías
		if lineEnd == lineStart {
			continue
		}

		// Buscar los dos puntos en la línea actual
		colonPos := -1
		for j := lineStart; j < lineEnd; j++ {
			if headers[j] == ':' {
				colonPos = j
				break
			}
		}

		if colonPos == -1 {
			continue // Sin dos puntos, omitir esta línea
		}

		// Extraer la clave (eliminar espacios al inicio y al final)
		keyStart := lineStart
		keyEnd := colonPos

		// Eliminar espacios antes de la clave
		for keyStart < keyEnd && (headers[keyStart] == ' ' || headers[keyStart] == '\t') {
			keyStart++
		}

		// Eliminar espacios después de la clave
		for keyEnd > keyStart && (headers[keyEnd-1] == ' ' || headers[keyEnd-1] == '\t') {
			keyEnd--
		}

		keyLen := keyEnd - keyStart

		// Comprobación rápida de longitud
		if keyLen != targetLen {
			continue
		}

		// Comparación manual sin distinción de mayúsculas
		match := true
		for j := 0; j < keyLen; j++ {
			c := headers[keyStart+j]
			// Convertir a minúsculas para comparar
			if c >= 'A' && c <= 'Z' {
				c = c + ('a' - 'A')
			}
			if c != targetBytes[j] {
				match = false
				break
			}
		}

		if !match {
			continue
		}

		// Encontrar el header que coincide y extraer el valor
		valueStart := colonPos + 1
		valueEnd := lineEnd

		// Eliminar espacios antes del valor
		for valueStart < valueEnd && (headers[valueStart] == ' ' || headers[valueStart] == '\t') {
			valueStart++
		}

		// Eliminar espacios después del valor
		for valueEnd > valueStart && (headers[valueEnd-1] == ' ' || headers[valueEnd-1] == '\t') {
			valueEnd--
		}

		// Devolver el valor
		if valueEnd > valueStart {
			return string(headers[valueStart:valueEnd]), nil
		}
		return "", nil
	}

	return "", nil
}

// ApplicationRequest estructura simulada para pruebas
type ApplicationRequest struct {
	Headers []byte
	SrcIp   netip.Addr
}

// Implementación original (para comparación de rendimiento)
func getRealClientIPOriginal(req *ApplicationRequest) string {
	if req == nil {
		return ""
	}

	// Probar distintas cabeceras por prioridad
	headers := []string{
		"x-forwarded-for",  // El más común, formato en cadena
		"x-real-ip",        // Común en Nginx
		"true-client-ip",   // Akamai
		"cf-connecting-ip", // Cloudflare
		"fastly-client-ip", // Fastly
		"x-client-ip",      // General
		"x-original-forwarded-for",
		"forwarded", // Cabecera estándar
		"x-cluster-client-ip",
	}

	// Intentar obtener la IP de las distintas cabeceras
	for _, header := range headers {
		if value, err := getHeaderValueOptimized(req.Headers, header); err == nil && value != "" {
			// Para X-Forwarded-For y formatos en cadena similares, extraer la primera IP
			if header == "x-forwarded-for" || header == "x-original-forwarded-for" {
				ips := strings.Split(value, ",")
				if len(ips) > 0 {
					ip := strings.TrimSpace(ips[0])
					if ip != "" {
						return ip
					}
				}
			} else if header == "forwarded" { // La cabecera Forwarded requiere tratamiento especial
				// Analizar la cabecera Forwarded, formato p. ej.: for=client;proto=https;by=proxy
				parts := strings.Split(value, ";")
				for _, part := range parts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 && strings.TrimSpace(kv[0]) == "for" {
						// Eliminar posibles comillas y corchetes de IPv6
						ip := strings.TrimSpace(kv[1])
						ip = strings.Trim(ip, "\"")

						// Manejar formatos especiales de dirección IPv6
						if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
							ip = ip[1 : len(ip)-1]
						}

						if ip != "" {
							return ip
						}
					}
				}
			} else { // Otras cabeceras devuelven el valor directamente
				ip := strings.TrimSpace(value)
				if ip != "" {
					return ip
				}
			}
		}
	}

	// Si ninguna cabecera está presente, devolver la IP de origen
	if req.SrcIp.IsValid() {
		return req.SrcIp.String()
	}

	return ""
}

// Implementación optimizada de getRealClientIP (simulada)
func getRealClientIPOptimized(req *ApplicationRequest) string {
	if req == nil {
		return ""
	}

	// Ruta rápida: buscar directamente la cabecera X-Forwarded-For más usada
	if value, err := getXForwardedForIP(req.Headers); err == nil && value != "" {
		return value
	}

	// Volver a la lógica original
	return getRealClientIPOriginal(req)
}

// Función de análisis dedicada para X-Forwarded-For
func getXForwardedForIP(headers []byte) (string, error) {
	target := []byte("x-forwarded-for")
	targetLen := len(target)

	i := 0
	for i < len(headers) {
		lineStart := i

		// Encontrar el fin de línea
		for i < len(headers) && headers[i] != '\n' && headers[i] != '\r' {
			i++
		}
		lineEnd := i

		// Omitir saltos de línea
		if i < len(headers) && headers[i] == '\r' {
			i++
		}
		if i < len(headers) && headers[i] == '\n' {
			i++
		}

		if lineEnd == lineStart {
			continue
		}

		// Buscar los dos puntos
		colonPos := -1
		for j := lineStart; j < lineEnd; j++ {
			if headers[j] == ':' {
				colonPos = j
				break
			}
		}

		if colonPos == -1 {
			continue
		}

		// Comprobar la clave
		keyStart := lineStart
		keyEnd := colonPos

		for keyStart < keyEnd && (headers[keyStart] == ' ' || headers[keyStart] == '\t') {
			keyStart++
		}
		for keyEnd > keyStart && (headers[keyEnd-1] == ' ' || headers[keyEnd-1] == '\t') {
			keyEnd--
		}

		keyLen := keyEnd - keyStart
		if keyLen != targetLen {
			continue
		}

		// Comparación sin distinción de mayúsculas
		match := true
		for j := 0; j < keyLen; j++ {
			c := headers[keyStart+j]
			if c >= 'A' && c <= 'Z' {
				c = c + ('a' - 'A')
			}
			if c != target[j] {
				match = false
				break
			}
		}

		if !match {
			continue
		}

		// Extraer el valor
		valueStart := colonPos + 1
		valueEnd := lineEnd

		for valueStart < valueEnd && (headers[valueStart] == ' ' || headers[valueStart] == '\t') {
			valueStart++
		}
		for valueEnd > valueStart && (headers[valueEnd-1] == ' ' || headers[valueEnd-1] == '\t') {
			valueEnd--
		}

		if valueEnd > valueStart {
			value := string(headers[valueStart:valueEnd])
			// Extraer la primera IP
			if commaPos := strings.IndexByte(value, ','); commaPos != -1 {
				firstIP := strings.TrimSpace(value[:commaPos])
				if firstIP != "" {
					return firstIP, nil
				}
			}
			return strings.TrimSpace(value), nil
		}
		return "", nil
	}

	return "", nil
}

// ==================== Pruebas de rendimiento ====================

func BenchmarkGetHeaderValue(b *testing.B) {
	testHeaders := []byte(`Host: example.com
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
X-Real-IP: 192.168.1.100
X-Forwarded-For: 10.0.0.1, 192.168.1.100, 172.16.0.1
X-Forwarded-Proto: https
X-Cluster-Client-IP: 10.0.0.1
Content-Type: application/json; charset=utf-8
Content-Length: 1234
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
Cache-Control: no-cache
Accept: application/json, text/plain, */*
Accept-Language: en-US,en;q=0.9,zh-CN;q=0.8
Accept-Encoding: gzip, deflate, br`)

	b.Run("New_Implementation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOptimized(testHeaders, "x-forwarded-for")
		}
	})

	b.Run("Old_Implementation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOld(testHeaders, "x-forwarded-for")
		}
	})

	// Simular escenario real: buscar múltiples headers (como en getRealClientIP)
	headers := []string{
		"x-forwarded-for", "x-real-ip", "true-client-ip", "cf-connecting-ip",
		"fastly-client-ip", "x-client-ip", "x-original-forwarded-for",
		"forwarded", "x-cluster-client-ip",
	}

	b.Run("New_Multiple_Headers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, header := range headers {
				_, _ = getHeaderValueOptimized(testHeaders, header)
			}
		}
	})

	b.Run("Old_Multiple_Headers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, header := range headers {
				_, _ = getHeaderValueOld(testHeaders, header)
			}
		}
	})
}

func BenchmarkGetHeaderValueWorstCase(b *testing.B) {
	// Crear una lista de headers muy larga, con el header objetivo al final
	var headerBuilder strings.Builder
	for i := 0; i < 20; i++ {
		headerBuilder.WriteString("Header-")
		headerBuilder.WriteString(strings.Repeat("X", 10))
		headerBuilder.WriteString(": value")
		headerBuilder.WriteString(strings.Repeat("Y", 50))
		headerBuilder.WriteByte('\n')
	}
	headerBuilder.WriteString("Target-Header: found-value\n")

	testHeaders := []byte(headerBuilder.String())

	b.Run("New_Worst_Case", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOptimized(testHeaders, "target-header")
		}
	})

	b.Run("Old_Worst_Case", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOld(testHeaders, "target-header")
		}
	})
}

func BenchmarkGetRealClientIPOptimization(b *testing.B) {
	// Probar headers de distintos escenarios
	testCases := []struct {
		name    string
		headers []byte
	}{
		{
			name: "XForwardedFor_First",
			headers: []byte(`X-Forwarded-For: 192.168.1.100, 10.0.0.1
Host: example.com
User-Agent: Mozilla/5.0`),
		},
		{
			name: "XForwardedFor_Middle",
			headers: []byte(`Host: example.com
X-Forwarded-For: 192.168.1.100, 10.0.0.1
User-Agent: Mozilla/5.0`),
		},
		{
			name: "XRealIP_Only",
			headers: []byte(`Host: example.com
X-Real-IP: 192.168.1.100
User-Agent: Mozilla/5.0`),
		},
		{
			name: "CloudFlare_Only",
			headers: []byte(`Host: example.com
CF-Connecting-IP: 192.168.1.100
User-Agent: Mozilla/5.0`),
		},
		{
			name: "No_Client_IP_Headers",
			headers: []byte(`Host: example.com
User-Agent: Mozilla/5.0
Content-Type: application/json`),
		},
		{
			name: "Many_Headers_XFF_Last",
			headers: []byte(`Host: example.com
User-Agent: Mozilla/5.0
Accept: application/json
Content-Type: application/json
Cache-Control: no-cache
Authorization: Bearer token
X-Forwarded-For: 192.168.1.100, 10.0.0.1`),
		},
	}

	// Crear SrcIp simulada
	srcIp, _ := netip.ParseAddr("203.0.113.1")

	for _, tc := range testCases {
		req := &ApplicationRequest{
			Headers: tc.headers,
			SrcIp:   srcIp,
		}

		b.Run(tc.name+"_Original", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = getRealClientIPOriginal(req)
			}
		})

		b.Run(tc.name+"_Optimized", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = getRealClientIPOptimized(req)
			}
		})
	}
}

// Prueba específica del rendimiento de la ruta rápida de X-Forwarded-For
func BenchmarkXForwardedForFastPath(b *testing.B) {
	testHeaders := []byte(`X-Forwarded-For: 192.168.1.100, 10.0.0.1, 172.16.0.1
Host: example.com
User-Agent: Mozilla/5.0`)

	b.Run("FastPath_XFF", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getXForwardedForIP(testHeaders)
		}
	})

	b.Run("Generic_XFF", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOptimized(testHeaders, "x-forwarded-for")
		}
	})
}

// Prueba de asignación de memoria
func BenchmarkHeaderParsingMemory(b *testing.B) {
	testHeaders := []byte(`Host: example.com
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
X-Real-IP: 192.168.1.100
X-Forwarded-For: 10.0.0.1, 192.168.1.100, 172.16.0.1
Content-Type: application/json; charset=utf-8`)

	b.Run("Old_Implementation_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOld(testHeaders, "x-forwarded-for")
		}
	})

	b.Run("New_Implementation_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = getHeaderValueOptimized(testHeaders, "x-forwarded-for")
		}
	})
}
