# Guía de inicio rápido — EscudoWAF en 5 minutos

Esta guía te lleva de cero a un WAF protegiendo tu primera aplicación, sin necesidad de ser experto en HAProxy ni en Coraza.

## Paso 0 — ¿Qué vas a montar?

```
Internet → HAProxy (EscudoWAF inspecciona aquí) → tu aplicación
```

EscudoWAF se pone delante de tu app, inspecciona cada petición con las reglas OWASP y bloquea ataques (SQLi, XSS, etc.).

## Paso 1 — Levantar el sistema (1 comando)

```bash
./inicio-rapido.sh
```

O manualmente:

```bash
docker compose up -d --build
```

Cuando termine, abre **http://localhost:2333**.

## Paso 2 — Primer inicio de sesión

- Usuario: `admin` · Contraseña: `admin123`
- El sistema te obligará a **cambiar la contraseña** en el primer acceso — hazlo.

## Paso 3 — Proteger tu primer sitio

1. Ve a **Configuración → Gestión de sitios → Añadir sitio**
2. Rellena:
   - **Nombre**: el que quieras (p. ej. "mi-app")
   - **Dominio**: el dominio público de tu app (o `localhost` para pruebas)
   - **Puerto de escucha**: `80` (o `443` si usarás HTTPS)
   - **Servidores backend**: la IP/puerto real de tu app (p. ej. `192.168.1.50:3000`)
   - **Modo WAF**: empieza con **Modo observación** 👈 importante
3. Guarda. El sistema regenera la config de HAProxy solo.

**¿Por qué observación primero?** Así el WAF registra los ataques detectados pero *no bloquea* aún — evitas cortes por falsos positivos mientras ajustas.

## Paso 4 — HTTPS (opcional, para producción)

1. **Configuración → Gestión de certificados → Añadir certificado**
2. Sube tu certificado PEM y clave privada (o pégalos)
3. En el sitio, activa HTTPS y selecciona el certificado

## Paso 5 — Pasar a modo protección

Cuando el modo observación muestre que todo va bien:

1. Edita el sitio → **Modo protección**
2. Desde ese momento, las peticiones maliciosas se bloquean de verdad

## Paso 6 — Ver qué está pasando

- **Panel de seguridad**: mapa 3D de ataques, IPs atacantes, QPS en tiempo real
- **Registros → Eventos de ataque**: ataques agrupados por IP/dominio
- **Registros → Registros de protección**: detalle de cada petición bloqueada
- **Monitor**: estadísticas de tráfico (peticiones, bloqueos, errores)

## Checklist de seguridad antes de producción

- [ ] Contraseña de admin cambiada
- [ ] `MONGO_INITDB_ROOT_PASSWORD` + `DB_URI` cambiados en `docker-compose.yaml`
- [ ] `JWT_SECRET` cambiado por un valor aleatorio largo
- [ ] Sitios en modo protección (no observación)
- [ ] Puerto 2333 del panel **no expuesto a internet** (accede solo por VPN/localhost o detrás de otra auth)
- [ ] HTTPS activado en sitios públicos

## ¿Problemas? → [FAQ y solución de errores](./FAQ.md)
