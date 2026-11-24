# Contribuir a EscudoWAF

¡Gracias por tu interés! Este proyecto es un fork traducido al español de [RuiQi](https://github.com/HUAHUAI23/RuiQi).

## Antes de empezar

- **¿Bug o mejora de la traducción?** → abre un issue en este repo (EscudoWAF)
- **¿Bug del motor, HAProxy o Coraza?** → probablemente pertenece al upstream: [issues de RuiQi](https://github.com/HUAHUAI23/RuiQi/issues) o [corazawaf/coraza](https://github.com/corazawaf/coraza/issues)

## Entorno de desarrollo

- Go 1.24+, Node.js 22+, pnpm, Docker
- Ver [README](./README.md#desarrollo-local) para levantar el entorno

## Convenciones

- **Código y comentarios en español** — es la identidad de este fork
- Mensajes al usuario (UI, API) pasan por i18n (`web/public/locales/es|en|zh`) o, en backend, literales en español
- `gofmt -w .` antes de commitear Go; `pnpm lint` en el frontend
- Si cambias anotaciones Swagger: `swag init --dir ./server,./pkg` desde la raíz

## Pull requests

1. Fork + rama descriptiva (`fix/...`, `feat/...`)
2. Commits con mensaje claro en español
3. Describe el "por qué" del cambio y cómo probarlo
4. Si toca UI, incluye una captura
