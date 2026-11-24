#!/usr/bin/env bash
# EscudoWAF - Despliegue en un comando
# Uso: ./inicio-rapido.sh
set -e

echo "🛡️  EscudoWAF - Inicio rápido"
echo "================================"

# 1. Comprobar prerrequisitos
if ! command -v docker &> /dev/null; then
    echo "❌ Docker no está instalado. Instálalo desde https://docs.docker.com/get-docker/"
    exit 1
fi
if ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose no está disponible."
    exit 1
fi
echo "✅ Docker encontrado: $(docker --version | head -1)"

# 2. Comprobar puertos libres
for port in 2333 27017; do
    if lsof -i :$port &> /dev/null 2>&1 || nc -z localhost $port 2>/dev/null; then
        echo "⚠️  El puerto $port ya está en uso. Libéralo o edita docker-compose.yaml."
    fi
done

# 3. Levantar los servicios
echo ""
echo "📦 Construyendo e iniciando los contenedores (la primera vez tarda varios minutos)..."
docker compose up -d --build

# 4. Esperar a que el panel responda
echo ""
echo "⏳ Esperando a que EscudoWAF esté listo..."
for i in $(seq 1 60); do
    if curl -sf -o /dev/null http://localhost:2333/ 2>/dev/null; then
        break
    fi
    if [ "$i" -eq 60 ]; then
        echo "⚠️  El panel aún no responde. Revisa los logs con: docker compose logs -f escudo"
        exit 1
    fi
    sleep 3
done

echo ""
echo "================================"
echo "✅ EscudoWAF está en marcha"
echo ""
echo "   🌐 Panel web:   http://localhost:2333"
echo "   📚 API docs:    http://localhost:2333/swagger/index.html"
echo "   👤 Usuario:     admin"
echo "   🔑 Contraseña:  admin123"
echo ""
echo "⚠️  ANTES DE USARLO EN PRODUCCIÓN:"
echo "   1. Cambia la contraseña de admin (te la pedirá en el primer inicio de sesión)"
echo "   2. Cambia MONGO_INITDB_ROOT_PASSWORD y DB_URI en docker-compose.yaml"
echo "   3. Cambia JWT_SECRET en docker-compose.yaml por un valor aleatorio"
echo "   4. Activa los sitios primero en 'Modo observación' y luego en 'Modo protección'"
echo ""
echo "📖 Guía completa: doc/GUIA-INICIO-RAPIDO.md"
echo "🛑 Para detener: docker compose down"
