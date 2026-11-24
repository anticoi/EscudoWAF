package internal

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// Implementación original con bufio.Scanner (para validación comparativa)
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

// TestGetHeaderValue prueba la corrección de la función getHeaderValue
func TestGetHeaderValue(t *testing.T) {
	testHeaders := []byte(`Host: example.com
User-Agent: Mozilla/5.0
X-Real-IP: 192.168.1.100
X-Forwarded-For: 10.0.0.1, 192.168.1.100
Content-Type: application/json
Content-Length: 1234
Authorization: Bearer token123`)

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"host", "host", "example.com"},
		{"Host", "Host", "example.com"},
		{"HOST", "HOST", "example.com"},
		{"x-real-ip", "x-real-ip", "192.168.1.100"},
		{"X-Real-IP", "X-Real-IP", "192.168.1.100"},
		{"x-forwarded-for", "x-forwarded-for", "10.0.0.1, 192.168.1.100"},
		{"content-type", "content-type", "application/json"},
		{"not-exist", "not-exist", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Probar la nueva implementación
			result, err := getHeaderValue(testHeaders, tt.header)
			if err != nil {
				t.Errorf("getHeaderValue() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("getHeaderValue() = %v, want %v", result, tt.expected)
			}

			// Probar consistencia con la implementación anterior
			oldResult, err := getHeaderValueOld(testHeaders, tt.header)
			if err != nil {
				t.Errorf("getHeaderValueOld() error = %v", err)
				return
			}
			if oldResult != result {
				t.Errorf("Resultado inconsistente: nuevo=%v, anterior=%v", result, oldResult)
			}
		})
	}
}

// TestGetHeaderValueEdgeCases prueba casos límite
func TestGetHeaderValueEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		headers  []byte
		header   string
		expected string
	}{
		{
			name:     "Headers vacíos",
			headers:  []byte(""),
			header:   "test",
			expected: "",
		},
		{
			name:     "Solo contiene saltos de línea",
			headers:  []byte("\n\n\n"),
			header:   "test",
			expected: "",
		},
		{
			name:     "Línea sin dos puntos",
			headers:  []byte("invalid line\nHost: example.com"),
			header:   "host",
			expected: "example.com",
		},
		{
			name:     "El valor tiene espacios al inicio o al final",
			headers:  []byte("Host:   example.com   "),
			header:   "host",
			expected: "example.com",
		},
		{
			name:     "La clave tiene espacios al inicio o al final",
			headers:  []byte("   Host   : example.com"),
			header:   "host",
			expected: "example.com",
		},
		{
			name:     "Valor vacío",
			headers:  []byte("Empty-Header: "),
			header:   "empty-header",
			expected: "",
		},
		{
			name:     "Múltiples dos puntos",
			headers:  []byte("Time: 12:34:56"),
			header:   "time",
			expected: "12:34:56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getHeaderValue(tt.headers, tt.header)
			if err != nil {
				t.Errorf("getHeaderValue() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("getHeaderValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGetRealClientIP prueba la corrección de la función getRealClientIP
func TestGetRealClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  []byte
		expected string
	}{
		{
			name: "X-Forwarded-For con una IP",
			headers: []byte(`Host: example.com
X-Forwarded-For: 192.168.1.100`),
			expected: "192.168.1.100",
		},
		{
			name: "X-Forwarded-For con múltiples IPs",
			headers: []byte(`Host: example.com
X-Forwarded-For: 192.168.1.100, 10.0.0.1, 172.16.0.1`),
			expected: "192.168.1.100",
		},
		{
			name: "X-Real-IP",
			headers: []byte(`Host: example.com
X-Real-IP: 192.168.1.200`),
			expected: "192.168.1.200",
		},
		{
			name: "CF-Connecting-IP",
			headers: []byte(`Host: example.com
CF-Connecting-IP: 1.2.3.4`),
			expected: "1.2.3.4",
		},
		{
			name: "Prueba de prioridad: X-Forwarded-For tiene prioridad",
			headers: []byte(`Host: example.com
X-Real-IP: 192.168.1.200
X-Forwarded-For: 192.168.1.100`),
			expected: "192.168.1.100",
		},
		{
			name: "Cabecera estándar Forwarded",
			headers: []byte(`Host: example.com
Forwarded: for=192.168.1.100;proto=https;by=proxy`),
			expected: "192.168.1.100",
		},
		{
			name:     "Sin cabecera de IP de cliente",
			headers:  []byte(`Host: example.com\nUser-Agent: test`),
			expected: "", // Como SrcIp no está configurada, debería devolver vacío
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &applicationRequest{
				Headers: tt.headers,
			}
			result := getRealClientIP(req)
			if result != tt.expected {
				t.Errorf("getRealClientIP() = %q, want %q", result, tt.expected)
			}
		})
	}
}
