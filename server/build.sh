#!/bin/bash

set -e

echo "🚀 Compilando EscudoWAF..."

# Comprobación del entorno
echo "🔍 Comprobando el entorno de compilación..."

# Comprobar la versión de Node.js
REQUIRED_NODE="23.10.0"
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version | sed 's/v//')
    echo "📦 Versión de Node.js: $NODE_VERSION (requerida: $REQUIRED_NODE)"
    if [ "$NODE_VERSION" != "$REQUIRED_NODE" ]; then
        echo "⚠️  Advertencia: versión de Node.js no coincide, se recomienda v$REQUIRED_NODE"
    fi
else
    echo "❌ Error: Node.js no encontrado, instale Node.js $REQUIRED_NODE primero"
    exit 1
fi

# Comprobar la versión de pnpm
REQUIRED_PNPM="10.11.0"
if command -v pnpm &> /dev/null; then
    PNPM_VERSION=$(pnpm --version)
    echo "📦 Versión de pnpm: $PNPM_VERSION (requerida: $REQUIRED_PNPM)"
    if [ "$PNPM_VERSION" != "$REQUIRED_PNPM" ]; then
        echo "⚠️  Advertencia: versión de pnpm no coincide, se recomienda $REQUIRED_PNPM"
    fi
else
    echo "❌ Error: pnpm no encontrado, instale pnpm $REQUIRED_PNPM primero"
    echo "💡 Comando de instalación: npm install -g pnpm@$REQUIRED_PNPM"
    exit 1
fi

# Comprobar la versión de Go
REQUIRED_GO="1.24.1"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "🔧 Versión de Go: $GO_VERSION (requerida: $REQUIRED_GO)"
    if [ "$GO_VERSION" != "$REQUIRED_GO" ]; then
        echo "⚠️  Advertencia: versión de Go no coincide, se recomienda $REQUIRED_GO"
    fi
else
    echo "❌ Error: Go no encontrado, instale Go $REQUIRED_GO primero"
    exit 1
fi

echo "✅ Comprobación del entorno completada"
echo ""

# 1. Compilar el frontend
echo "📦 Compilando recursos del frontend..."
cd ../web
pnpm install
pnpm build
cd ../server

# 2. Copiar recursos del frontend al directorio embebido
echo "📋 Copiando recursos del frontend..."
mkdir -p public/dist
cp -r ../web/dist/* public/dist/

# 3. Compilar el backend
echo "🔧 Compilando el servicio backend..."
go mod tidy
go build -o escudo-waf .

echo "✅ ¡Compilación completada!"
echo "📍 Ubicación del ejecutable: server/escudo-waf" 