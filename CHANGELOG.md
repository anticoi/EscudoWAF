# Changelog

Todos los cambios destacables de EscudoWAF se documentan aquí.

## [No publicado]

### Añadido
- Traducción completa al español: ~3.600 literales y ~4.200 comentarios traducidos en backend Go, frontend React, anotaciones Swagger y plantillas HAProxy
- Nuevo locale `es` en el frontend (`web/public/locales/es/translation.json`); español como idioma por defecto y fallback
- `inicio-rapido.sh`: despliegue en un comando con verificación de prerrequisitos
- `doc/GUIA-INICIO-RAPIDO.md`: tutorial paso a paso para usuarios no expertos
- `doc/FAQ.md`: solución de problemas comunes
- `docker-compose.prod.yml`: configuración de producción endurecida (secrets por env, healthchecks, MongoDB sin puerto expuesto)
- `.env.example`, `.dockerignore`, `CONTRIBUTING.md`

### Cambiado
- Rebrand completo RuiQi → EscudoWAF (módulos Go `github.com/anticoi/EscudoWAF`, contenedores, binarios, UI)
- `index.html`: `lang="es"`, metadatos en español
- Footer de login internacionalizado (antes texto fijo en inglés)

### Eliminado
- Artefactos de desarrollo del upstream en chino: `prompt.txt`, `context.txt`, plantillas de issues `_cn.yml`, workflows de CI no transferibles, `release.guide.md`, `doc/proposal/*_zh.md`

---

Para el historial de cambios del proyecto original, ver [RuiQi](https://github.com/HUAHUAI23/RuiQi/releases).
