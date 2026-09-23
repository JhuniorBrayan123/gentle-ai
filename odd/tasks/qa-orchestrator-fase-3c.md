# Fase 3C — Integraciones reales (QA-Orchestrator v3)

## Objetivo

Completar las integraciones reales que Fase 3B dejó deliberadamente diferidas,
en el orden que pidió el usuario, sin reabrir Core ni skills salvo una
incompatibilidad concreta encontrada durante la integración. Sin merge a
`main` hasta que se decida explícitamente.

## Por qué / contexto

Fase 3B (cerrada en commit `5a730728`) portó y empaquetó las 9 skills QA con
un smoke test E2E real pero **simbólico**: sin Engram/BookStack MCP
conectados, sin Playwright MCP, sin RDDAdapter real, sin MR real. Fase 3C
ejercita cada una de esas integraciones contra MCPs/sistemas reales.

## Alcance y orden (instrucción explícita del usuario)

1. **3C.1** — Engram + BookStack MCP
2. **3C.2** — locator-hunting NIVEL 0/1 + Playwright MCP
3. **3C.3** — QACodeReviewer + RDDAdapter real
4. **3C.4** — G6 funcional completo / qa-verify E2E
5. **3C.5** — creación real de MR vía GitLab MCP
6. **3C.6** — smoke E2E final sobre un repositorio de automatización real

TDD/reviews acotadas por bloque (candidato = un work-unit commit, nunca el
branch acumulado). Detenerse solo cuando se necesite autorización real
(credenciales/MCP/acciones externas). No revisar candidatos históricos
gigantes (base-diff del branch completo) — declinar sin volver a preguntar,
tal como ya se estableció.

## Mapeo técnico inicial (workflow de scoping, 2026-09-23)

Confirmado con evidencia real de código/MCP (no supuesto):

| Sub-fase | ¿Lista ahora? | Bloqueo real |
|---|---|---|
| 3C.1 | Sí | Ninguno — BookStack MCP y Engram ya conectados en este entorno |
| 3C.2 | Parcial | NIVEL 0 (POM local) y NIVEL 2 (GitLab) listos; NIVEL 1 requiere Playwright MCP (no conectado) + autorización explícita de navegar a un entorno real |
| 3C.3 | No | (a) `internal/cli` ya importa `internal/qastage` → un `RDDAdapter` real no puede vivir en `qastage` sin ciclo de imports; (b) generar hallazgos reales requiere un binario de agente externo instalado y autenticado (claude/codex/opencode/pi) — no es Go puro. Necesita decisión de arquitectura antes de TDD. |
| 3C.4 | Parcial | Los ítems 1,2,5,6,7,8 del checklist G6 ya son responsabilidad 100% de `qa-verify`; los ítems 3-4 (lint/secretos) dependen de 3C.3; el cierre completo también depende de 3C.6 (repo real) |
| 3C.5 | Parcial | Adaptador/plumbing (interfaz + stub + tests contra un cliente GitLab falso) sí puede construirse ahora; la invocación real de `create_merge_request` exige autorización explícita del usuario en cada ocasión — nunca general |
| 3C.6 | No | Depende de que el usuario indique qué repositorio real de automatización usar — decisión de producto, no técnica |

Decisiones de producto/arquitectura pendientes (se preguntan cuando se llegue
a esa sub-fase, no todas de una vez):
- 3C.3: ¿`RDDAdapter` real vive en `internal/cli` (nuevo tipo, cableado en el
  punto de `qa-verify`) o hace `shell out` al binario `gentle-ai` compilado
  desde `internal/qastage`?
- 3C.3: ¿`QACodeReviewer.Review(ctx, change)` sigue siendo una llamada
  bloqueante única, o `qa-verify` debe correr él mismo el protocolo completo
  STATUS→START→capture-result→acknowledge-approved?
- 3C.5: ¿la autorización de MR real se vuelve un gate auditable en código
  (como `qa-approve`) o se queda como stop-and-ask conversacional (ya
  demostrado en 3B.8)?
- 3C.6: ¿qué repositorio real de automatización se usa?
- 3C.2: ¿se conecta Playwright MCP ahora, y se autoriza navegar a un
  entorno real de dev/staging?

## Tareas

### 3C.1 — Engram + BookStack MCP

- [x] Verificar en vivo (no solo prosa) que el nombre real de la tool
  BookStack es `bookstack_search`, no `bookstack_bookstack_search` — corregido
  en `skills/qa-explore`, `qa-spec`, `qa-apply`, `qa-supervisor`,
  `skills/_shared/qa-gate-policy.md` (+ mirrors en `internal/assets/skills`).
  Verificado con llamadas reales a `bookstack_search`/`bookstack_get_page`.
- [x] Rellenar los placeholders `TODO-MAINTAINER-FILL-BOOKSTACK-*` de
  `qa-gate-policy.md` con la página real verificada (ID `3239`,
  `https://bookstack.sreasons.com/books/gestion-interna/page/06-reglas-del-agente-orquestador-qa`).
- [x] Documentar en `qa-doc-reference` el paso de resolución Libro/Capítulo
  (`bookstack_get_book`/`bookstack_get_chapter`, ya que `bookstack_get_page`
  solo devuelve IDs numéricos) y aclarar que "Responsables"/"Ticket
  Redmine"/"Versión del documento" no existen como campos estructurados en la
  API real — se reportan "no disponible" salvo que la página los declare en
  su propio texto.
- [x] Corregir la nota de migración obsoleta en `qa-docs` que afirmaba que
  `qa-supervisor/qa-explore/qa-spec/qa-apply/qa-verify` "todavía no existen
  en esta rama" (falso desde el commit `5a730728`).
- [x] **Hallazgo de arquitectura no anticipado** (fuera del scoping original,
  planteado por el usuario): ¿qué hace `qa-explore` cuando BookStack no
  documenta una funcionalidad *existente* y el código tampoco basta para
  conocer el comportamiento real de UI? Resuelto 100% a nivel de skill, cero
  cambios de Core (verificado leyendo `internal/qastage/state_machine.go` y
  `artifact.go` directamente): fallback opcional de captura Codegen dirigida
  por el humano, cerrado con `qa-finish --outcome interrupted` sobre un
  artefacto `scope.status:"blocked"` (primitivos ya existentes en el Core
  desde 3A.6), retomado en un `qa-begin` nuevo que extrae hechos "OBSERVED
  (Codegen)" al cuerpo libre de Engram — nunca al `findings[]` validado por
  el Core (esa clasificación sigue siendo estrictamente
  `DOCUMENTED|MISSING|NOT_APPLICABLE`, verificado en código). Ver memoria
  Engram `architecture/codegen-human-capture-fallback-for-qa-explore...`
  para el análisis completo. Implementado de forma aditiva (sin renumerar ni
  tocar los pasos 0-7 existentes de `qa-explore`) en `qa-explore` (paso
  nuevo condicional), `qa-spec` y `qa-apply` (guardrails de una línea cada
  uno) + sus mirrors.
- [ ] (Opcional, no bloqueante) Ejercitar un flujo real completo tipo 3B.8
  pero con Engram+BookStack conectados de verdad, para confirmar que no
  queda ningún otro mismatch de prosa-vs-API — se puede diferir sin bloquear
  el cierre de 3C.1.

**Ruta/commits**: pendiente de commit inicial de 3C.1 (bookstack tool-name +
gate-policy + doc-reference + docs stale note + fallback Codegen).

**Verificación**: `go test ./internal/assets/...` y
`./internal/components/skills/...` verdes tras cada tanda de ediciones;
`diff -q` confirmando paridad byte-a-byte entre `skills/qa-*` e
`internal/assets/skills/qa-*` en cada archivo tocado.

### 3C.2 — locator-hunting NIVEL 0/1 + Playwright MCP

- [x] Re-verificar (sin cambio de código) que NIVEL 0 (POM local) y NIVEL 2
  (GitLab) siguen correctos; ejercitado NIVEL 2 de verdad vía
  `search_projects("erp-mf-seguridad")` → `SmartClic/erp-mf-seguridad`,
  coincide exacto con la fila del catálogo, cero drift. NIVEL 0 no se puede
  ejercitar dentro de `gentle-ai` (no es un proyecto de automatización
  consumidor, no tiene `src/pages/**`) — correctamente diferido a 3C.6.
- [x] Re-confirmar paridad de mirror entre `skills/qa-locator-hunting/**` e
  `internal/assets/skills/qa-locator-hunting/**` (`diff -q`, idénticos).
- [ ] Preguntar al usuario si conectar Playwright MCP y autorizar navegación
  a un entorno real de dev/staging ahora, o diferir NIVEL 1 (y por lo tanto
  el cierre completo de 3C.2) para más adelante. **PENDIENTE — punto de
  autorización real, se pregunta explícitamente.**
- [x] No tratar GitLab/NIVEL 2 como sustituto de NIVEL 1 — las Hard Rules de
  la skill prohíben saltar niveles (confirmado, sin cambios necesarios).

### 3C.3 — QACodeReviewer + RDDAdapter real

- [ ] Presentar al usuario la decisión de arquitectura (adapter en
  `internal/cli` vs. subprocess) antes de escribir cualquier test.
- [ ] Corregir el comentario obsoleto en `reviewer.go`/`reviewer_test.go`
  ("Fase 3B" → debería decir 3C.3; y la referencia a
  `gentle-ai review inspect-candidate` como entry point es incorrecta).
- [ ] TDD (RED primero) una vez resuelta la decisión de arquitectura.
- [ ] Cablear el `RDDAdapter` real como caller real de `qa-verify` (hoy no
  existe ningún caller productivo).

### 3C.4 — G6 funcional completo / qa-verify E2E

- [ ] Agregar el requisito de cobertura negativa (estilo E2/E3) a
  `qa-verify/SKILL.md` y `qa-gate-policy.md` (prosa, sin código).
- [ ] Recortar la narrativa G6 de `qa-gate-policy.md` que duplica lo que RDD
  ya cubre mejor (una vez 3C.3 esté resuelto).
- [ ] No declarar "G6 completo" cerrado hasta que 3C.1 (ya real), 3C.3 y
  3C.6 estén todos en su lugar.

### 3C.5 — Creación real de MR vía GitLab MCP

- [ ] Diseñar e implementar `QAMergeRequestCreator` (interfaz Go + stub),
  mapeando un `{change}` cerrado a `project_path/source_branch/
  target_branch/title/description`.
- [ ] Tests unitarios contra un cliente GitLab falso — ninguna llamada real
  en esta tarea.
- [ ] Preguntar al usuario: ¿la descripción del MR se auto-puebla desde las
  4 secciones de `qa-docs`? ¿la autorización se vuelve un gate de código
  auditable o se queda conversacional?
- [ ] Nunca invocar `create_merge_request`/`update_merge_request`/
  `approve_merge_request` reales sin confirmación explícita del usuario en
  ese momento puntual.

### 3C.6 — Smoke E2E final sobre un repositorio real

- [ ] Preguntar al usuario qué repositorio real de automatización usar.
- [ ] Confirmar acceso (checkout local, `project_path`/`target_branch` de
  GitLab, y si se autoriza un entorno real de dev/staging para NIVEL 1).
- [ ] Correr el ciclo completo real `qa-explore → qa-spec → qa-apply →
  qa-verify → qa-docs`, ejercitando 3C.1-3C.5 de verdad.
- [ ] Documentar la evidencia en
  `docs/migration/qa-orchestrator-v3-design.md`, siguiendo el patrón de
  3B.8.

## Ruteo de delegación

Exploración/mapeo inicial de 3C.1-3C.6: delegado vía Workflow (5 lentes en
paralelo + síntesis, `wf_144039f0-626`), evidencia real de código y MCP en
cada hallazgo. Escritura de las correcciones de 3C.1 (skills): directa
inline — ediciones mecánicas de varios archivos, más el diseño aditivo del
fallback Codegen (arquitectura ya resuelta por el orquestador antes de
escribir, no delegada).

## TDD

Modo: estricto (RED→GREEN→REFACTOR), fuente: `CLAUDE.md` del usuario
("Strict TDD Mode: enabled"). 3C.1 no tuvo código Go que cambiar (confirmado
por lectura directa de `internal/qastage`), así que no aplica ciclo
RED/GREEN aquí — solo verificación funcional (tests de paquete +
`diff -q` de paridad). 3C.3 sí tendrá Go real y usará TDD estricto cuando se
resuelva su decisión de arquitectura.

## Próximo paso

Commit del work-unit de 3C.1, luego continuar con 3C.2 (slice sin
Playwright) en la misma sesión o la siguiente.
