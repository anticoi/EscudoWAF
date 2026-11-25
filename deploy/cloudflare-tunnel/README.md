# Despliegue con Cloudflare Tunnel

Expone EscudoWAF a Internet **sin abrir puertos** ni tener IP fija:

```
Internet → Cloudflare Edge (TLS gratis) → cloudflared → HAProxy → Coraza → tu app
```

Requisito: el dominio debe usar los nameservers de Cloudflare (plan gratis).

## Paso 1 — Mover el DNS a Cloudflare

1. Crea cuenta en [cloudflare.com](https://cloudflare.com) → *Add site* → tu dominio → plan **Free**.
2. Cloudflare te asigna 2 nameservers (ej. `lara.ns.cloudflare.com`).
3. **Antes de cambiar NS**, recrea en Cloudflare todos los registros actuales (los exporta automáticamente, pero verifica).
4. Cambia los NS en tu registrador/proveedor DNS. Propagación: minutos a horas.

> ⚠️ Registros de correo/ftp deben quedar en **DNS only** (nube gris), NO proxied.

## Paso 2 — Crear el túnel

1. Cloudflare Zero Trust → **Networks → Tunnels → Add tunnel** → Cloudflared.
2. Copia el **token** del túnel.
3. En *Public Hostnames* añade:
   - Subdomain: `gps` (o el que quieras) — Domain: `tudominio.cl`
   - Service: `http` → `escudo:8080` (nombre del servicio en el compose)

## Paso 3 — Levantar

```bash
cd deploy/cloudflare-tunnel
cp .env.example .env   # rellena MONGO_PASSWORD, JWT_SECRET, TUNNEL_TOKEN
docker compose up -d
```

## Paso 4 — Crear el sitio en el panel

En `http://localhost:2333` (admin/admin123 → cambia la contraseña):

- **Sitios → Nuevo sitio**
  - Dominio: `gps.tudominio.cl`
  - Puerto de escucha: `8080`
  - Backend: host `host.docker.internal`, puerto `8082` (o la IP/host:puerto de tu app)
  - WAF: activado, modo `protection` (empieza con `observation` si quieres ver primero qué bloquearía)
- **Motor → Recargar** para aplicar el sitio

## Paso 5 — Probar

```bash
curl -I https://gps.tudominio.cl
```

El TLS lo pone Cloudflare; el tráfico llega al WAF por HTTP interno y Coraza inspecciona todo antes del backend.

## Notas

- Múltiples hostnames: repite el paso 2 con `panel.tudominio.cl` → `http://escudo:2333` si quieres el panel público (activa Cloudflare Access delante — gratis — para que no quede abierto).
- Puerto 8080: solo accesible dentro de la red Docker; nada queda expuesto en tu router.
- Si tu app corre en el mismo Docker, usa el nombre del contenedor como backend host en vez de `host.docker.internal`.
