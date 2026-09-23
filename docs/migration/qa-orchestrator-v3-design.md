# Diseño: QA-Orchestrator v3 sobre Gentle-AI 3.7.0

**Fecha:** 2026-09-23
**Fase:** 2 CERRADA (ver sección 1.1) → **Fase 3A en curso** (implementación del core, sección 7).
**Insumo:** [inventario v2→v3](qa-orchestrator-v2-to-v3.md) + [reporte de validación](qa-orchestrator-v2-to-v3-validation-report.md) + [checklist baseline ejecutado](qa-orchestrator-v2-baseline-checklist.md) + decisiones de arquitectura aprobadas por el usuario (secciones 1 y 1.1).

## 1. Decisiones de arquitectura aprobadas

Estas 7 decisiones resuelven las preguntas abiertas del inventario y son la base de este diseño. Quedan registradas como aprobadas, no como propuesta:

1. **Ledger QA como capa de dominio independiente.** No se reintroduce `StageVocabulary` dentro del core de `sddstatus`. Se reutilizan primitivas actuales de v3.7 (`OpenRuntimeStore`, Engram, patrón `next_action`) cuando son compatibles.
2. **Gate humano pre-apply se conserva**, mediante un contrato QA específico de aprobación de contenido/etapa — reutiliza primitivas del consentimiento actual (actor, reason, request-id, auditoría) sin mezclar semánticamente con permisos de escritura (`edit_authority`/`consent_contract` quedan intactos para su propósito original).
3. **BookStack y GitLab siguen como fuentes externas de verdad** (documental/funcional y código/MR respectivamente). Engram = memoria del agente. ODD = progreso/contexto de trabajo. Ninguno sustituye a otro.
4. **Se conserva la semántica** de "aprobación requerida" y "etapa fuera de orden", sin obligación de portar literalmente `ErrRuntimeStageApprovalRequired`/`ErrRuntimeStageOutOfOrder`. RDD no decide el orden ni los gates del workflow QA.
5. **`qa-verify` se rediseña como híbrido**: capa QA conserva Playwright/evidencia/BookStack/reporter; delega en RDD/lenses lo que 3.7 ya resuelve mejor (risk/reliability/resilience/readability/diff).
6. **Screenplay+POM/Playwright/locator-hunting siguen como familia de skills QA separada**, componiendo ODD/SDD por fuera, sin modificar el core genérico de SDD.
7. **Tests v2 = contratos de comportamiento/regresión.** Los tests v3 se reescriben contra la arquitectura nueva; no se portan literalmente los que dependen de tipos eliminados (`StageVocabulary`, `RuntimeStageApproval`).

## 1.1. Cierre de Fase 2 — decisiones finales (2026-09-23)

Estas resoluciones cierran los 10 contratos pendientes de la sección 5, incorporando la evidencia real del [checklist baseline](qa-orchestrator-v2-baseline-checklist.md) (`legacy-v2-custom`, ejecutado end-to-end). **Fase 2 queda cerrada.** No se reabren decisiones de arquitectura salvo que la implementación (Fase 3) revele una incompatibilidad concreta con Gentle-AI 3.7 — en ese caso, se documenta el hallazgo y se decide puntualmente, no se reabre todo el diseño.

**Flujo y stages**
- Flujo oficial: `explore → spec → approve_spec → apply → verify → docs → complete`. `docs` deja de ser una rama opcional sin dueño: tendrá un executor explícito, la skill `qa-docs`. `qa-supervisor` sigue sin decidir el orden — solo enruta según `next_action`.
- Orden de etapas se valida desde el **primer** `qa-begin` (corrige el hallazgo del Escenario 3): la única etapa válida como primer intento de un `{change}` nuevo es `explore`.
- Gate de aprobación aplica también cuando `apply` es el primer intento (corrige el hallazgo del Escenario 2) — no depende de que exista una etapa previa completada.

**Outcomes y reintentos**
- `passed` avanza la máquina de estados. `failed` e `interrupted` NO avanzan — dejan la etapa en estado reintentable.
- Reintentar crea un **nuevo intento**, nunca borra el historial de intentos anteriores (`attempts[]` se preserva, como ya se observó en `sdd-attempt status`).
- `reset` es una recuperación manual y auditada (quién, cuándo, por qué) — tampoco borra historial.

**Idempotencia y concurrencia**
- Mismo `request-id` + mismo payload → mismo resultado (replay seguro). Mismo `request-id` + payload distinto → conflicto explícito (nunca se sobreescribe silenciosamente).
- Toda mutación conserva `--expected-revision` (control de concurrencia optimista, ya presente en v2 y confirmado funcional).

**Aprobación de contenido**
- `QAStageApproval` queda ligada a un `artifact_revision` real (no a un hash autodeclarado). Cualquier nueva revisión del artefacto de `spec` invalida automáticamente la aprobación anterior — cierra el hallazgo del Escenario 6 (evidencia autodeclarada) para el caso específico de aprobación.

**Validación de artefactos**
- `qa-validate` permanece como verbo CLI, pero implementado sobre un `QAArtifactValidator` con un **schema canónico versionado** (una sola fuente de verdad — corrige la incompatibilidad schema-real vs. envelope-descrito-por-las-skills del Escenario 1). `qa-validate` pasa a devolver el `artifact_revision` real de lo que admitió, para que `qa-finish`/`qa-approve` lo consuman directamente.

**Autoridad y persistencia**
- Engram nunca es autoridad de estado QA — solo contexto/handoffs (ya fijado en la decisión #6 aprobada anteriormente; se reafirma aquí).
- Persistencia: `QAStateMachine → QAStateStore (interfaz) → OpenRuntimeStoreAdapter`.
- Revisión de código: `qa-verify → QACodeReviewer (interfaz) → RDDAdapter`.

**Skills**
- Las skills QA permanecen independientes y estáticas (archivos `.md`, como hoy) — no se incorporan al generador Go de plantillas SDD.

**Disposición de flags/campos legacy** (cierra el contrato pendiente #10):

| Flag/campo legacy | Disposición |
|---|---|
| `--expected-revision` | **CONSERVAR** tal cual |
| `--max-attempts` | **REEMPLAZAR** por `QAPolicy.max_attempts` (configuración de política, no flag por invocación) |
| `--max-changed-lines` | **RETIRAR** del workflow QA — RDD/policy ya se ocupa del riesgo del diff |
| `--evidence-goal` | **RETIRAR** como flag — el contrato de evidencia se deriva del schema canónico de cada stage |
| `--evidence-revision` | **CONSERVAR**, pero ahora ligado a un `artifact_revision` real (no solo validado por formato) |
| harness disposition (`--harness-disposition`) | **MOVER** a metadata estructurada del artefacto de `verify`, deja de ser un flag CLI aparte |

**Validaciones diferidas**: los escenarios no ejecutables del baseline (5a completo, 5b NIVEL 1, 7, y el checklist G6 funcional completo) permanecen como `DIFERIR VALIDACIÓN` — no bloquean el core de Fase 3A; se validan cuando haya repo de automatización real, MCP de Playwright, y autorización explícita para MR.

## 2. Arquitectura propuesta

```
                        QA-Orchestrator v3
                               │
                ┌──────────────┴──────────────┐
                │                              │
          Dominio QA                     Gentle-AI 3.7 (reutilizado)
                │                              │
    ┌───────────┼───────────┐        ┌─────────┼─────────┐
    │           │           │        │         │         │
QA State    QA Stage    Playwright  ODD       SDD       RDD
Machine     Approval    /Screenplay (tasks/   (agent   (lenses:
(nuevo)     (nuevo)     /POM/       progreso) templates risk/reliability/
    │           │       locators              generados) resilience/
    │           │       (conservado                       readability)
    │           │        de v2)
    │           │
OpenRuntimeStore  Consent primitives
(reutilizado)     (actor/reason/request-id,
                   reutilizado el patrón,
                   NO el tipo consent_contract)
    │
  Engram (memoria persistente — compartida con ODD/SDD)
    │
BookStack MCP (verdad funcional)   GitLab MCP (verdad de código/MR)
```

Principio rector: **la capa QA conserva lo que la hace única** (máquina de estados de 5 etapas, gate de aprobación de contenido, Playwright/Screenplay/POM, BookStack/GitLab como fuentes de negocio); **todo lo genérico que 3.7 ya resuelve mejor se delega**, sin que el core de Gentle-AI tenga que conocer conceptos QA.

## 3. Matriz de compatibilidad final

| Componente v2 | Qué reutiliza directamente de Gentle-AI 3.7 | Qué queda en la capa QA (código nuevo) | Qué código legacy deja de ser necesario |
|---|---|---|---|
| **`qa-begin`** | `sddstatus.OpenRuntimeStore(ctx, repo, change)` como backend de persistencia; patrón de namespacing por prefijo de `change` | Nuevo `QAStateMachine` con las 5 etapas fijas (`explore/spec/apply/verify/docs`) codificadas en la capa QA (no como vocabulario configurable genérico); reimplementación del verbo `qa-begin` sobre ese state machine | `internal/qastage/vocabulary.go` (`VocabularyV1`, `StageVocabulary`, `Stage{Label, RequiresApproval}` genéricos) — el vocabulario deja de ser configurable, se fija a las 5 etapas QA conocidas |
| **`qa-finish`** | mecanismo `mutate`/replay de `runtime_ledger.go` para persistir el cierre | Verbo `qa-finish` reescrito sobre `QAStateMachine`; validación de evidencia (hash) se mantiene como lógica de dominio QA | Dependencia directa de tipos `RuntimeStageApproval`/`ErrRuntimeStageVocabularyMismatch` eliminados |
| **`qa-approve`** | Patrón actor/reason/request-id/auditoría de `consent_contract.go` (solo el patrón, no el tipo `SDDIntegrationConsentSchema`) | Nuevo contrato `QAStageApproval` (aprobación de contenido de etapa, distinto de `WriteConsent`) | `sddstatus.RuntimeStageApproval` (tipo eliminado en 3.7, no se reintroduce) |
| **`qa-status`** | Patrón `next_action` idiomático de 3.7 (visto en `sdd_status.go`, `review_correction_context.go`, `bench/classify.go`) — se sigue ese lenguaje de diseño en vez de inventar uno nuevo | Verbo `qa-status` reescrito sobre `QAStateMachine`, devolviendo `next_action` en el mismo estilo que usa el resto de 3.7 | Lectura directa de `StageVocabulary` para calcular next_action |
| **QA ledger** | `OpenRuntimeStore` como backend físico de persistencia (namespace `qa_<change>`) | `QAStateMachine`: las 5 etapas fijas + reglas de transición (`apply` bloqueado sin `approve`, etapas en orden) expresadas como resultado estructurado: `{"allowed": false, "reason": "approval_required", "next_action": "approve_spec"}` en vez de errores tipados nuevos con nombres calcados de v2 | `internal/qastage` completo como paquete (vocabulario genérico); `StageVocabulary`/`Stage` no se reintroducen en `sddstatus` |
| **`qa-supervisor`** | Routing de fases ya idiomático en 3.7 (`internal/components/sdd/orchestrator.go` como referencia de patrón, no como código a heredar directamente) | Skill QA que lee `next_action` de `QAStateMachine` (nunca decide el orden por sí sola), aplica reglas G1-G6, consulta BookStack/GitLab MCP, cierra creando MR | Ninguno — la lógica de negocio (G1-G6, BookStack, GitLab) no tiene equivalente en 3.7 y se conserva completa en la capa QA |
| **`qa-explore`** | Estructura de "solo investigar, guardar en Engram con topic_key convencional, result contract" tomada como patrón de `sdd-explore.md` (sin heredar el generador Go de SDD) | Skill QA independiente con extensiones de dominio: BookStack, Screenplay+POM/fixtures, `qa-locator-hunting` | Ninguno de dominio; se conserva la skill como familia separada (decisión #6) |
| **`qa-spec`** | Estructura de `sdd-spec.md` como referencia de patrón (specs describen QUÉ, no CÓMO) | Skill QA que produce plan de prueba + invoca `QAStageApproval` antes de permitir `qa-apply` | Ninguno de dominio |
| **`qa-apply`** | Estructura de progreso/resumen de `sdd-apply.md` como referencia (TDD RED→GREEN→REFACTOR, marcar tareas) | Skill QA que valida `QAStageApproval` antes de empezar, extiende Screenplay+POM real del proyecto | Dependencia de `ErrRuntimeStageApprovalRequired`/`ErrRuntimeStageOutOfOrder` tipados; se reemplaza por el resultado estructurado de `QAStateMachine` |
| **`qa-verify`** | Sistema de lens de RDD (`gentle-ai review inspect-candidate --lens risk\|reliability\|resilience\|readability`) para las verificaciones de código/diff que 3.7 ya resuelve mejor | Capa "Functional QA" propia: Playwright, comportamiento, criterios, BookStack, screenshots, traces, reporter — se mantiene íntegra | Checklist G6 narrativo que hoy duplica lo que RDD ya hace mejor para secretos/diff/riesgo — se retira esa porción, se delega a RDD |

## 4. Riesgos de diseño y puntos a validar en Fase 3 (implementación)

- **`QAStateMachine` es código nuevo, no una migración 1:1.** Su diseño interno (cómo persiste sobre `OpenRuntimeStore`, cómo modela las 5 etapas fijas) debe implementarse y probarse contra los tests de regresión heredados de v2 (decisión #7), no simplemente "hacerlo compilar".
- **`QAStageApproval` vs. `edit_authority`**: hay que confirmar en implementación que ambos contratos pueden coexistir sin fricción cuando una misma tarea QA necesita a la vez permiso de escritura (SDD) y aprobación de etapa (QA) — el diseño los mantiene separados pero deben interoperar en la práctica.
- **Integración `qa-verify` ↔ RDD lens**: falta definir en Fase 3 el mecanismo exacto de invocación (¿`qa-verify` llama a `gentle-ai review inspect-candidate` como subproceso, o se orquesta desde el nivel de skill/agente?). Es la pieza de mayor incertidumbre técnica de esta matriz.
- ~~**Autoría de skills QA**: hay que decidir en Fase 3 si las skills QA siguen siendo archivos `.md` estáticos o se generan igual que SDD.~~ **RESUELTO en 1.1**: las skills QA permanecen estáticas e independientes, no se incorporan al generador Go de SDD.
- **Reescritura de tests**: antes de implementar cada componente, extraer de los tests v2 (`internal/qastage/*_test.go`, `internal/cli/qa_*_test.go`) las reglas de comportamiento (ej. "apply sin aprobación falla") como casos de prueba independientes del tipo concreto, para escribir los tests v3 contra `QAStateMachine`/`QAStageApproval` sin acoplarse a los tipos eliminados.

## 5. Contratos pendientes antes de Fase 3 (RESUELTOS — ver sección 1.1)

> **Estado: CERRADO el 2026-09-23.** Las 10 preguntas de esta sección fueron resueltas explícitamente en la sección 1.1 ("Cierre de Fase 2 — decisiones finales"), con evidencia real del checklist baseline. Esta sección se conserva como registro histórico de qué se preguntó y por qué — no como trabajo pendiente.

La matriz de la sección 3 fija responsabilidades, pero dejó 10 contratos sin cerrar en su momento (los mismos ya resueltos en la sección 1.1). Se conservan aquí solo como registro de qué se preguntó, no como trabajo pendiente.

1. **`qa-validate` no tiene lugar en la arquitectura.** En v2 era usado por `qa-explore`, `qa-spec`, `qa-apply` y `qa-verify` para validar artefactos antes de cerrar etapa; la matriz de la sección 3 no lo modela. Falta decidir: crear un `QAArtifactValidator` (schemas/contrato por etapa) y definir si sigue existiendo como verbo CLI `qa-validate` o queda como API interna de la capa QA.
2. **La etapa `docs` no tiene responsable.** `QAStateMachine` define 5 etapas (`explore/spec/apply/verify/docs`) pero solo hay skill para 4. Falta decidir si vuelve `qa-docs` explícito, si `qa-supervisor` absorbe la documentación, o si `docs` deja de ser una etapa del state machine.
3. **El contrato completo de `QAStateMachine` no está especificado.** Falta fijar, antes de programarlo: estados, transiciones válidas, transición fallida, outcomes (`passed`/`failed`/`interrupted`, heredados de v2), idempotencia por `request-id`, reintentos, reset/reanudación, y comportamiento ante una operación repetida.
4. **Concurrencia sin resolver.** v2 usaba `expected-revision` (control optimista); la matriz actual no lo menciona. Cada mutación debe declarar una revisión esperada y fallar si el ledger cambió desde entonces — necesario para que `qa-supervisor` y otro agente no avancen el mismo `{change}` a la vez.
5. **`QAStageApproval` debe invalidarse si cambia el contenido aprobado.** No basta con `actor/reason/request-id`; la aprobación debe ligarse a una revisión concreta del spec (`spec revision X → aprobación Y → apply autorizado`). Si el spec cambia después de aprobado, la aprobación anterior debe quedar inválida — de lo contrario se puede aprobar A, modificarlo a B, y ejecutar B con la aprobación de A.
6. **Regla de autoridad Engram vs. ledger.** Falta declarar explícitamente que Engram nunca es fuente autoritativa del estado QA — solo guarda contexto, resúmenes y handoffs. La autoridad única es `QAStateMachine`/su store. Sin esta regla, es posible llegar a un estado donde Engram dice "apply" mientras el ledger dice "approve_spec".
7. **`OpenRuntimeStore` debe aislarse detrás de una interfaz propia**, en vez de que la capa QA lo consuma directamente: `QAStateMachine → QAStateStore (interfaz) → OpenRuntimeStoreAdapter`. Consumirlo directo (como está hoy en la matriz de la sección 3) vuelve a acoplar la capa QA a un detalle interno de `sddstatus` que ya cambió una vez entre v2 y v3.
8. **`qa-verify` no debería invocar el CLI de RDD directamente** (ya señalado como riesgo abierto en la sección 4). Falta diseñar una interfaz `QACodeReviewer`/`RDDAdapter` — inicialmente puede envolver el CLI, después puede pasar a API Go — para que la skill solo pida revisión y reciba findings estructurados, sin acoplarse a la forma de invocación actual de RDD.
9. **Compatibilidad de cara al usuario QA (v2 → v3) sin definir.** Falta una tabla de contrato de experiencia: ¿sigue existiendo `qa-supervisor`? ¿los mismos comandos? ¿qué flags desaparecen? ¿qué valores puede devolver `next_action`? ¿cuándo pide aprobación humana? ¿qué cambia visiblemente para el QA que ya se capacitó con v2? Internamente todo puede cambiar; externamente, preservar la experiencia conocida reduce el costo de adopción.
10. **Flags/campos legacy sin disposición explícita.** `qa-begin` tenía `--max-attempts`, `--max-changed-lines`, `--evidence-goal`, `--expected-revision`; `qa-finish` tenía hash de evidencia y disposición del harness (el diseño actual solo conserva el hash). Cada uno debe marcarse `CONSERVAR`, `REEMPLAZAR` o `RETIRAR` explícitamente antes de Fase 3 — no deben desaparecer por omisión.

Adicionalmente, se identifica un componente que hoy está implícito y conviene hacer explícito en el diseño final: una **QA Policy** compartida (reglas G1-G6) consumida por `qa-supervisor`, `qa-explore`, `qa-spec`, `qa-apply` y `qa-verify` por igual, en vez de vivir duplicada/copiada en cada skill (riesgo de drift). Junto con `QAArtifactValidator` (punto 1), el flujo completo antes de programar quedaría:

```
qa-supervisor
     ↓
 explore → validate artifact
     ↓
 spec → validate artifact
     ↓
 HUMAN APPROVAL (QAStageApproval)
     ↓
 apply → validate artifact
     ↓
 verify (Functional QA + RDD vía QACodeReviewer)
     ↓
 docs
     ↓
 final validation → MR / complete
```

## 6. Próximo paso recomendado antes de implementar (COMPLETADO)

~~Correr los 8 escenarios de prueba manual en OpenCode sobre `legacy-v2-custom`~~ — **ejecutado**. Resultados completos en el [checklist baseline](qa-orchestrator-v2-baseline-checklist.md), incluyendo la tabla `Comportamiento v2 observado vs. contrato deseado para v3` que alimentó directamente las decisiones de la sección 1.1.

Las secciones 1–6 de este documento describen diseño (Fase 2, cerrada). La implementación real empieza en la sección 7 (Fase 3A).

## 7. Plan de Fase 3A — Core QA (planificación, sin implementar)

Implementación incremental del núcleo del ledger (state machine, aprobación, validación de artefactos) sobre `qa-orchestrator-v3` (base `v3.7.0`). Cada PR es un commit de trabajo pequeño, TDD estricto (RED → GREEN → REFACTOR), con su propia review/RDD antes de encadenar el siguiente. El orden respeta dependencias: no se puede escribir `QAStateMachine` sobre un `QAStateStore` que no existe.

| PR | Contenido | RED (test primero) | GREEN mínimo | Depende de |
|---|---|---|---|---|
| **3A.1** | Schema canónico versionado + `QAArtifactValidator` | Test: payload con el schema real (`findings`/`pending_questions`/`scope`/`predecessor_sha256`) es válido; payload con campos legacy inventados (`schema`/`facts`/`observations`) es rechazado; salida incluye `artifact_revision` (hash real del payload canónico) | Reimplementar `ValidateStageArtifactAdmission` detrás de una interfaz `QAArtifactValidator`, agregar el campo `artifact_revision` a la respuesta | — (base) |
| **3A.2** | `qa-validate` CLI compatible sobre 3A.1 | Test: `qa-validate --input - --change x --stage explore` con payload válido devuelve `artifact_revision` en el JSON de salida (nuevo campo, sin romper `valid`/`stage`/`change`/`reason` existentes) | Wire del comando CLI al nuevo `QAArtifactValidator` | 3A.1 |
| **3A.3** | `QAStateStore` (interfaz) + `PersistentQAStateStore` | Test: la implementación cumple la interfaz `QAStateStore` con CAS por `expected-revision`, historial append-only de intentos, sin depender de símbolos no exportados de `sddstatus` | Implementación propia y autocontenida de la capa QA (ver nota de incompatibilidad abajo) | — (paralelo a 3A.1) |
| **3A.4** | `QAStateMachine`: orden desde el primer `qa-begin` | Test: primer `qa-begin` con etapa ≠ `explore` es RECHAZADO (corrige el hallazgo del Escenario 3); primer `qa-begin --stage explore` es aceptado; avance fuera de orden sigue rechazado como hoy | Corregir la condición de "solo se valida en advancing begin" para incluir el primer begin | 3A.3 |
| **3A.5** | `QAStateMachine`: gate de aprobación desde el primer intento | Test: primer `qa-begin --stage apply` de un `{change}` nuevo es RECHAZADO sin aprobación (corrige el hallazgo del Escenario 2); sigue permitido tras `qa-approve` | Extender el chequeo de aprobación para no depender de una etapa previa completada | 3A.4 |
| **3A.6** | `QAStateMachine`: outcomes y reintentos | Test: `outcome=passed` avanza; `outcome=failed`/`interrupted` dejan la etapa reintentable sin avanzar; un reintento crea un nuevo intento en `attempts[]` sin borrar el anterior; `reset` queda auditado (actor/razón/timestamp) y tampoco borra historial | Especificar explícitamente las 3 transiciones de outcome + el flujo de `reset` | 3A.4 |
| **3A.7** | Idempotencia y concurrencia | Test: mismo `request-id` + mismo payload → mismo resultado exacto (replay); mismo `request-id` + payload distinto → conflicto explícito; toda mutación exige `--expected-revision` coincidente | Verificar/extender el manejo existente de CAS + idempotencia en el nuevo `QAStateMachine` | 3A.4, 3A.6 |
| **3A.8** | `QAStageApproval` ligada a `artifact_revision` | Test: aprobar `spec` en `artifact_revision` X, luego generar una nueva revisión Y del spec → la aprobación de X queda inválida para desbloquear `apply`; aprobar Y sí desbloquea | Enlazar `qa-approve`/el chequeo de `apply` al `artifact_revision` exacto de 3A.1/3A.2 | 3A.2, 3A.5 |
| **3A.9** | CLI: disposición de flags legacy | Test por flag: `--max-attempts` ya no existe como flag de `qa-begin`, se lee de `QAPolicy`; `--max-changed-lines` y `--evidence-goal` eliminados; `--evidence-revision` ahora se valida contra un `artifact_revision` real (no solo formato) — CLI rechaza un hash bien formado pero no admitido por `qa-validate` | Reescribir flags de `qa-begin`/`qa-finish`/`qa-approve` según la tabla de disposición de la sección 1.1 | 3A.8 |
| **3A.10** | Skill `qa-docs` (nueva) + actualización de `qa-supervisor` | Sin test de código (son archivos `.md`); validación: ejecutar el flujo completo `explore→spec→approve_spec→apply→verify→docs→complete` sobre `qa-orchestrator-v3` y confirmar que `qa-supervisor` enruta a `qa-docs` sin decidir contenido | Escribir `skills/qa-docs/SKILL.md` (executor explícito) y actualizar `qa-supervisor/SKILL.md` para incluir `docs` en el flujo oficial y corregir las referencias al schema/flags ya retirados | 3A.9 |
| **3A.11** | `QACodeReviewer`/`RDDAdapter` (interfaz, sin integración real todavía) | Test: la interfaz se puede stubear/mockear; `qa-verify` la consume en vez de invocar CLI de RDD directamente | Definir la interfaz y un stub; la integración real con `gentle-ai review inspect-candidate` queda para una Fase 3B separada (no bloquea el core) | 3A.10 |

**Incompatibilidad verificada durante la implementación (3A.3, 2026-09-23):** `sddstatus.RuntimeStore` en Gentle-AI 3.7.0 ya no expone `Begin()`/`Finish()` públicos — el propio código lo documenta (`internal/cli/sdd_attempt.go:15-16`: *"Runtime attempt admission, settlement and budgets are retired"*). Solo quedan `Status()` y `Grant()` (edit-authority), sin relación con progreso de etapas QA. Por lo tanto `OpenRuntimeStoreAdapter` (como se planteó originalmente) es inviable: no hay mecanismo público que envolver, y copiar símbolos no exportados de `sddstatus`/`runtime_ledger.go` violaría el encapsulamiento del paquete y volvería a acoplar la capa QA a un detalle interno que ya cambió una vez.

**Adaptación mínima adoptada** (no reabre Fase 2 — la arquitectura `QAStateMachine → QAStateStore → implementación` se mantiene sin cambios): `OpenRuntimeStoreAdapter` se reemplaza por `PersistentQAStateStore`, una implementación propia, mínima y autocontenida de la capa QA, sin dependencias del dominio SDD. Reutiliza únicamente primitivas genéricas y estables de filesystem/atomic-write si Gentle-AI 3.7 las expone como utilidad pública neutral (a confirmar en la implementación); si no existen, la persistencia vive enteramente en `internal/qastage`. Este cambio es puntual y de implementación, no de arquitectura: solo se sustituye qué hay detrás de la interfaz `QAStateStore`.

**Fuera de alcance de Fase 3A** (validaciones diferidas, no bloquean el core): NIVEL 0/1 de `qa-locator-hunting` (5a, 5b-nivel1) y creación real de MR en GitLab (Escenario 7) — se validan cuando haya repo de automatización real, MCP de Playwright, y autorización explícita.

Cada PR de esta lista es candidato a review/RDD individual antes de encadenar el siguiente, siguiendo la política de entrega ya vigente del repo (work-unit commits, tamaño acotado). No se implementa nada de esta tabla todavía — queda pendiente de tu aprobación para empezar 3A.1.
