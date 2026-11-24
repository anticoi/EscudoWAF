# FAQ y solución de problemas — EscudoWAF

## El panel no responde en http://localhost:2333

```bash
docker compose ps                 # ¿los contenedores están "Up"?
docker compose logs -f escudo     # logs del WAF
docker compose logs -f mongodb    # logs de la base de datos
```

Causas frecuentes: MongoDB aún arrancando (espera ~30s), puerto 2333 ocupado por otro proceso, o el build falló a mitad.

## HAProxy no arranca / el motor aparece "Detenido"

- Comprueba **Configuración → Configuración general → Estado del motor**
- Revisa que el puerto de escucha del sitio no choque con otro servicio
- Cambios en hilos del motor, rutas de HAProxy o directivas requieren **Reiniciar** el motor desde el panel

## Mi app no recibe tráfico / error 503

- Verifica la IP y puerto del **servidor backend** del sitio — debe ser alcanzable *desde dentro del contenedor* (en Docker, `localhost` no es tu máquina; usa `host.docker.internal` en Mac/Windows o la IP real)
- Comprueba que el sitio está **Activo** en la lista de sitios

## Bloquea peticiones legítimas (falsos positivos)

- Deja el sitio en **Modo observación**, revisa **Registros → Eventos de ataque** e identifica la regla que dispara
- Añade una **regla personalizada de lista blanca** (por IP o por ruta) en **Reglas → Reglas personalizadas**
- O añade la IP del cliente a un **grupo de IPs** permitido

## No puedo iniciar sesión / "Acceso no autorizado"

- Credenciales por defecto: `admin` / `admin123` (solo primer arranque)
- Si perdiste la contraseña, reinicia los datos: `docker compose down -v` ⚠️ borra toda la configuración y logs guardados

## ¿Qué puertos usa EscudoWAF?

| Puerto | Uso |
|--------|-----|
| 2333 | Panel web + API |
| 80 / 443 | Tráfico HTTP/HTTPS de los sitios protegidos |
| 8080 | HAProxy (según sitios configurados) |
| 8404 | Estadísticas de HAProxy |
| 9443 | HAProxy SSL |
| 27017 | MongoDB (no exponer a internet) |

## ¿Cómo actualizo a una nueva versión?

```bash
git pull
docker compose up -d --build
```

Los datos (sitios, reglas, logs) persisten en los volúmenes `mongodb_data` y `escudo_data`.

## ¿Cómo empiezo de cero?

```bash
docker compose down -v   # borra contenedores + volúmenes (datos incluidos)
docker compose up -d --build
```

## ¿Qué no cubre EscudoWAF?

- No sustituye TLS/HTTPS: configura certificados o termina TLS en otro proxy
- No es un CDN ni hace cache de contenido
- Las reglas OWASP CRS vienen en inglés (son del propio ruleset oficial)
