---
name: qa-doc-despliegue
description: "Trigger: leer, crear o actualizar páginas BookStack de Especificación y Despliegue Micro (TDL). Renderiza el formato con plantilla fija; nunca elimina."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

## Activation Contract

Carga esta skill cuando el usuario pida leer, crear o actualizar una página de BookStack con el formato "Formato de Especificación y Despliegue Micro" (código TDL-*). Para citar otra documentación PRD usa `qa-doc-access`.

## Hard Rules

- NUNCA elimines páginas, capítulos ni libros: no llames herramientas de borrado. Si lo piden, responde que no está permitido y que debe hacerse a mano en BookStack.
- NUNCA escribas el HTML a mano: renderiza siempre con `scripts/render.py`; tú solo completas el JSON.
- NUNCA escribas, pidas ni extraigas credenciales: la sección 5.2 es texto fijo. `render.py` rechaza claves tipo password, contraseña o clave.
- Antes de crear muestra un resumen; antes de actualizar muestra el diff campo a campo. Espera confirmación explícita.
- Si la página leída es legacy de Word, avisa que al actualizar quedará normalizada al formato de la plantilla.
- Si falta un dato, pregunta; no lo inventes.

## Decision Gates

| Pedido | Operación |
|--------|-----------|
| lee, muestra, resume | Leer |
| crea, nueva especificación | Crear |
| actualiza, cambia, agrega paso | Actualizar |
| elimina, borra | Rechazar (primera regla) |

## Execution Steps

Invoca los scripts siempre con `python` (los archivos se instalan sin permiso de ejecución).

**Leer**
1. Resuelve la página con `bookstack_list_pages` / `bookstack_list_chapters` y cita su URL exacta.
2. `bookstack_get_page`; guarda el campo HTML en `page.html`.
3. `python <skill-dir>/scripts/parse.py page.html > data.json` y presenta los campos pedidos.

**Crear**
1. Reúne los datos según `references/esquema.md`. Opcional: si hay MCP de Redmine, prellena número, título, autor y URL de la incidencia.
2. `python <skill-dir>/scripts/render.py data.json > page.html`.
3. Muestra resumen (código, nombre, pasos) y libro/capítulo destino; espera confirmación.
4. `bookstack_create_page` con `name` = `codigo`, `book_id`/`chapter_id` indicados y `html` = contenido de `page.html`.

**Actualizar**
1. Ejecuta los pasos de Leer.
2. Aplica solo los cambios pedidos a `data.json`.
3. Muestra el diff (antes → después) y la advertencia legacy si aplica; espera confirmación.
4. Renderiza y llama `bookstack_update_page` con el `id` y el HTML nuevo.

## Output Contract

Devuelve: operación realizada, URL exacta de la página, campos leídos o cambiados y advertencias (legacy normalizado, datos pendientes). Nunca incluyas contraseñas.

## References

- `references/esquema.md` — esquema JSON entre `render.py` y `parse.py`.
- `assets/plantilla.html` — plantilla fija del formato.
