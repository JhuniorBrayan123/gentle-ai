# qa-doc-despliegue — Skill para el formato "Especificación y Despliegue Micro"

**Objetivo.** Skill QA que lee, crea y actualiza en BookStack páginas con el formato idéntico de "Formato de Especificación y Despliegue Micro". Nunca elimina páginas.

| Campo | Valor |
|---|---|
| Rama | `feat/qa-doc-despliegue` |
| Entrega | `ask-on-risk`. Forecast: ~900 líneas, de las cuales ~450 son la plantilla HTML; se consulta la estrategia de PR antes del push. |
| Runners | `python -m unittest` (scripts) · `go test` (registro) |
| TDD | Strict TDD activado (configuración de la sesión) |

## Specs

**S1 — Formato idéntico.** "una skill o una tool que haga ese formato identico". El HTML generado reproduce el formato de la página de referencia TDL-21762:
- logo;
- título en Book Antiqua 20pt;
- "Versión 1.0" a la derecha;
- carátula en Times New Roman 16pt;
- encabezados de tabla `rgb(79,129,189)`;
- filas de grupo `rgb(141,179,226)`;
- secciones 1 a 7.

El formato vive en una plantilla fija dentro de `assets/`. El modelo nunca escribe el HTML a mano: solo completa datos. La plantilla de partida es `scratchpad/ejemplo.min.html`, una copia limpia de la página 3921 de BookStack, con contenido ficticio.

**S2 — Leer.** "que lea el formato de depliegue". Un script convierte el HTML de una página existente al JSON del esquema de abajo. Tiene que funcionar con:
- las páginas generadas por la plantilla;
- las páginas legacy importadas desde Word: estilos `mso-*`, comentarios `<!--[endif]-->`, spans anidados y numeración "1.&nbsp;&nbsp;".

Los encabezados se reconocen por su texto ("Objetivo del requerimiento", "Secuencia de despliegue", etc.), no por su posición exacta.

**S3 — Crear.** "cree". Desde el JSON, la skill:
1. renderiza el HTML;
2. crea la página con el MCP de BookStack (`bookstack_create_page`) en el libro y capítulo que indique el usuario;
3. pone como nombre de la página el `codigo`.

Antes de crear, muestra un resumen y pide confirmación.

**S4 — Actualizar.** "actualice". El flujo es: leer con `bookstack_get_page`, parsear, aplicar los cambios pedidos al JSON, renderizar y llamar a `bookstack_update_page`. Antes de actualizar, muestra un diff de los campos que cambian y pide confirmación. Si la página era legacy de Word, avisa que quedará normalizada al formato de la plantilla.

**S5 — No eliminar.** "excepto eliminar". La skill nunca llama a herramientas de borrado. Si se pide eliminar, responde que no está permitido y que se haga a mano en BookStack.

**S6 — Credenciales (supuesto de seguridad).** La sección 5.2 siempre usa texto fijo: "Usuario: [usuario de prueba]" y "Contraseña: se entrega por el canal seguro del equipo". El parser no extrae contraseñas, y el renderer no acepta un campo de contraseña.

**S7 — Registro.** La skill se registra en el fork para que `gentle-ai install`/`sync` la instale en `~/.claude/skills/qa-doc-despliegue/`. Va en dos copias idénticas: `internal/assets/skills/qa-doc-despliegue/` y `skills/qa-doc-despliegue/`. Está incluida en los presets por defecto, igual que `qa-evidence`.

**S8 — Comando público OpenCode.** Como herramienta standalone, se instala también como `/qa-doc-despliegue`. El comando carga la skill, no exige `{change}`, nunca avanza el ledger y conserva los gates de confirmación para crear/actualizar.

### Esquema JSON (contrato entre render y parse)

```json
{
  "codigo": "TDL-00000", "version": "1.0", "fecha": "DD/MM/AAAA",
  "creado_por": "", "nombre": "", "validado_por": "", "redmine_url": "",
  "objetivo": ["..."], "solucion": ["..."], "descripcion_usuario": ["..."],
  "peticion": {
    "aviso_comercial":        {"descripcion": "", "respuesta": "SI|NO"},
    "modulo_afectado":        {"descripcion": "", "respuesta": ""},
    "funcionalidad_afectada": {"descripcion": "", "respuesta": ""},
    "componente_afectado":    {"descripcion": "", "respuesta": ""}
  },
  "rf": ["..."], "rnf": ["..."],
  "requiere_qa": true,
  "documentacion": [{"descripcion": "", "link": ""}],
  "despliegue": [{"grupo": "BASE DE DATOS", "pasos": [
    {"componente": ["Scripts", "DBX"], "descripcion": ["texto o pasos"], "link": "", "responsable": "", "tiempo": ""}
  ]}],
  "reversion": [{"grupo": "...", "pasos": []}],
  "tiempo_total": "",
  "implementacion": [{"inicio_fin": "", "corte": "", "observaciones": ""}],
  "notificaciones": [{"cargo": "", "nombre": ""}],
  "responsables": [{"rol": "", "nombre": "", "fecha": ""}],
  "observaciones": ""
}
```

Cuando una lista llega vacía, se renderiza "No aplica." (en rf y rnf) o una fila vacía (en las tablas). Esto es un supuesto que copia el comportamiento del original.

## Tasks

| ID | Specs | Ruta | Estado | Commit |
|---|---|---|---|---|
| T1 | Contenido de la skill (S1–S6) | Writer A | Completo | — |
| T2 | Registro y comando público (S7–S8) | Writer B + parent | Completo | — |
| T3 | Chequeo de unión: instalación aislada y prueba de ida y vuelta render → parse | Inline (parent) | Completo | — |

**T1 (Writer A).** Incluye `SKILL.md`, `assets/plantilla.html`, `scripts/render.py`, `scripts/parse.py`, `scripts/test_scripts.py` y el espejo en `skills/`.

**T2 (Writer B).** Toca `types.go`, `catalog/skills.go`, `presets.go`, `skill_picker.go` y su test, `assets_test.go`, el golden, `e2e_test.sh` y `AGENTS.md`.

**Por qué se delegan (trigger de escritura, razón a).** Son dos unidades de comportamiento independientes, con superficies de edición disjuntas, y cada una lleva varios minutos. Se lanzan juntas.

## Log

- **L1 (2026-10-22, usuario):**
  - "Como hacemos una skill o una tool que haga ese formato identico ? es que se quiere agilizar la documentacion"
  - "Si armalo pero donde lo hariamos? porquie tengo un proyecto de gentle-ai un fork de QA podriamos hacer una skill que lea el formato de depliegue, actualice, cree y excepto eliminar"
- **L2:** El mapa del registro salió de un explorador de solo lectura:
  - archivos a tocar: los de T2;
  - ninguna skill trae scripts todavía, pero el instalador copia cualquier archivo con modo 0644, así que se invoca con `python scripts/...`;
  - frontmatter: `description` de una sola línea, de hasta 160 caracteres y con `Trigger:`.
- **L3:** Defaults elegidos sin preguntar, por ser convención del repo: SKILL.md en español (como los demás `qa-*`), espejo en `skills/` y presets por defecto.
- **L4 (2026-10-22):** Integración cerrada:
  - corregidos los nombres MCP al contrato real: `bookstack_list_pages`, `bookstack_list_chapters`, `bookstack_get_page`, `bookstack_create_page`, `bookstack_update_page`;
  - añadido `/qa-doc-despliegue` al inyector selectivo de comandos OpenCode;
  - 21 pruebas Python de parse/render en verde;
  - paquetes Go de assets, catálogo, skills y selector TUI en verde;
  - instalación real en perfil OpenCode aislado: 38 verificaciones, 0 fallos; skill y comando presentes.
