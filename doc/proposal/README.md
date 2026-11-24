# Documentos de diseño

Documentos técnicos sobre la arquitectura interna de EscudoWAF (heredados del proyecto original RuiQi, en inglés):

- [`micro_engine_design.md`](./micro_engine_design.md) — diseño del MicroEngine: motor de reglas propio para filtrado por IP/URL/ruta con condiciones compuestas (AND/OR), listas blancas/negras y caché de regex.
- [`condition-build-en.md`](./condition-build-en.md) — especificación del constructor de condiciones: modelo de datos de las condiciones de reglas y su validación.

> Para usuarios finales, la documentación principal es la [Guía de inicio rápido](../GUIA-INICIO-RAPIDO.md).
