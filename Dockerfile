# Compilación multi-etapa
# Etapa 1: compilar el frontend Node.js
# Versión fijada para producción
FROM node:23.10.0-alpine AS frontend-builder
# Instalar pnpm
RUN npm install -g pnpm@10.11.0
# Establecer el directorio de trabajo
WORKDIR /app
# Copiar los archivos del proyecto frontend
COPY web/ ./
# Instalar dependencias y compilar el frontend
RUN pnpm install
RUN pnpm build

# Etapa 2: compilar el backend Go
FROM golang:1.24.1-alpine AS backend-builder
# Configurar variables de entorno de Go
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
# Establecer el directorio de trabajo
WORKDIR /build
# Copiar toda la estructura del proyecto
COPY coraza-spoa/ ./coraza-spoa/
COPY pkg/ ./pkg/
COPY server/ ./server/
COPY go.work ./
COPY geo-ip/ ./geo-ip/
# Copiar los artefactos del frontend al lugar correcto
COPY --from=frontend-builder /app/dist ./server/public/dist
# Compilar usando la función de workspace de Go
RUN go work use ./coraza-spoa ./pkg ./server
RUN cd server && go build -o ../escudo-waf main.go

# Etapa 3: imagen final - usa la imagen oficial de HAProxy 3.0.10
FROM haproxy:3.0.10

# Asegurar la configuración inicial como root
USER root

# Instalar herramientas de gestión de capabilities de Linux
RUN apt-get update && apt-get install -y libcap2-bin && \
    rm -rf /var/lib/apt/lists/*

# Crear usuario y grupo escudo
RUN groupadd --gid 1000 escudo && \
    useradd --uid 1000 --gid escudo --home-dir /home/escudo --create-home --shell /bin/bash escudo

# Añadir el usuario escudo al grupo haproxy para permisos de operación
RUN usermod -a -G haproxy escudo

# Crear el directorio de la aplicación y establecer permisos
WORKDIR /app
RUN chown escudo:escudo /app

# Copiar el binario Go desde el builder
COPY --from=backend-builder /build/escudo-waf .

# Copiar los archivos de documentación Swagger
COPY --from=backend-builder /build/server/docs/ ./docs/

# Establecer permisos de los archivos de la aplicación
RUN chown -R escudo:escudo /app && chmod +x /app/escudo-waf

# Crear el directorio escudo-waf en el home del usuario escudo y copiar la carpeta geo-ip
RUN mkdir -p /home/escudo/escudo-waf
COPY --from=backend-builder /build/geo-ip/ /home/escudo/escudo-waf/geo-ip/
RUN chown -R escudo:escudo /home/escudo/escudo-waf

# Paso clave: dar a HAProxy y a la aplicación la capacidad de enlazar puertos privilegiados
RUN setcap 'cap_net_bind_service=+ep' /usr/local/sbin/haproxy && \
    setcap 'cap_net_bind_service=+ep' /app/escudo-waf

# Verificar las capabilities (opcional, para depuración)
RUN getcap /usr/local/sbin/haproxy /app/escudo-waf

# Ahora se puede cambiar al usuario escudo de forma segura
USER escudo

# Configurar variables de entorno
ENV GIN_MODE=release

# Restablecer ENTRYPOINT (sobrescribe el docker-entrypoint.sh de la imagen base)
ENTRYPOINT []

# Exponer puertos: 2333 (aplicación)
EXPOSE 2333

# Ejecutar la aplicación
CMD ["/app/escudo-waf"]