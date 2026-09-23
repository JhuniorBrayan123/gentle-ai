---
name: qa-verify
description: "Trigger: validar la implementación QA contra el spec. Ejecuta pruebas funcionales y produce evidencia reproducible (G6); la revisión de código/diff se delega a RDD, no se reimplementa aquí."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "3.0"
disable-model-invocation: true
user-invocable: false
---

## Activation Contract

Carga esta skill cuando `qa-supervisor` te delegue validar una implementación QA contra su spec. Eres el sub-agente de **validación funcional y evidencia (G6)** del orquestador QA. NO corriges el código: reportas hallazgos verificables.

## Fuentes de verdad (MANDATORY)

- **Spec** (`qa/{change}/spec`) = contrato de validación.
- **BookStack = fuente de la verdad**: el resultado debe compararse contra la documentación consultada; cita las páginas usadas (G1).
- **Engram = memoria persistente**: registra el progreso en `qa/{change}/verify-report`.
- **Ledger nativo = el gate real**: `verify` no requiere aprobación manual (el humano ya aprobó en spec), pero el ledger DEBE cerrarse para que la cadena avance a `docs`.

## Entrada (obligatorio, antes de validar)

1. Ejecuta `gentle-ai qa-begin --change {change} --stage verify --cwd <repo> --request-id <id-idempotente>`.
2. Si el comando se rechaza, **DETENTE** y devuelve el rechazo tal cual al orquestador/supervisor.
3. Solo si `qa-begin` responde con éxito continúa con la validación de abajo.

## Diseño híbrido: Functional QA (esta skill) + revisión de código (delegada)

A diferencia de v2, esta skill **no reimplementa** verificaciones de código/diff que Gentle-AI 3.7 ya resuelve mejor (secretos, riesgo, diff, mantenibilidad). Esa porción se delega a `QACodeReviewer` (interfaz Go, `internal/qastage/reviewer.go`, 3A.11) — hoy es un stub (`RDDAdapter`) que devuelve "no implementado"; la integración real con `gentle-ai review inspect-candidate` es Fase 3C. Mientras esa integración no exista:

- Corre igual la checklist funcional completa (abajo) — eso es 100% responsabilidad de esta skill y no depende de RDD.
- Para las verificaciones de código/diff que normalmente delegarías a `QACodeReviewer` (puntos 3-4 del checklist), si RDD real todavía no está disponible, ejecútalas manualmente como hoy (lint, revisión de secretos) y **anota explícitamente** en el reporte que fueron manuales porque la integración de Fase 3C aún no existe — no bloquees el cierre de `verify` por eso.
- Cuando Fase 3C entregue la integración real, esta sección deja de aplicar y el checklist se actualiza para invocar `QACodeReviewer` directamente en vez de hacerlo a mano.

## Checklist funcional G6 (obligatoria, TODA)

1. `npx tsc --noEmit` — tipos compilan.
2. Ejecutar la prueba modificada/creada — pasa.
3. Revisar lint del proyecto — limpio (manual hasta que exista `QACodeReviewer` real).
4. Verificar que NO haya credenciales/secretos en el diff (manual hasta que exista `QACodeReviewer` real).
5. Verificar que NO haya esperas fijas innecesarias.
6. Verificar que el test siga Screenplay+POM (sin locators crudos en el archivo de test, actor/tasks/questions usados según el diseño del spec) — reutilizando lo existente cuando aplica, o con la estructura nueva creada según SOLID cuando el proyecto no tenía patrón previo.
7. Comparar el resultado contra la documentación BookStack consultada.
8. Entregar el comando de ejecución y el resultado obtenido.

## Evidencia reproducible (G6)

- Captura **screenshots**, **traces** (en fallo), **videos** o salida de **reporter** de Playwright.
- Registra artefactos con rutas y comandos exactos de reproducción.

## Salida (obligatorio, cierre de la etapa)

1. **Cuerpo del reporte en Engram**: `mem_save` con `topic_key: "qa/{change}/verify-report"` — hallazgos, comandos ejecutados, resultados de la validación funcional, y qué puntos del checklist fueron manuales por falta de integración RDD real (Fase 3C).
2. **Admisión anti-alucinación**: arma el envelope canónico `gentle-ai.qa-stage-artifact/v1` (`findings`/`pending_questions`/`scope`/`predecessor_sha256` — nunca el envelope legacy que el Core real rechaza) y pásalo por `gentle-ai qa-validate --input - --change {change} --stage verify --source-revision <artifact_revision del apply>`.
3. **Cierre del ledger**: con el `artifact_revision` que `qa-validate` devolvió, ejecuta `gentle-ai qa-finish --change {change} --cwd <repo> --request-id <id-idempotente> --outcome passed --evidence-revision <artifact_revision>`. Si hay hallazgos graves o la prueba falla, usa `--outcome failed`.
4. Solo tras un `qa-finish` exitoso reportas la etapa `verify` como cerrada al orquestador/supervisor.

## Guardrails

- No modifiques código para "arreglar" hallazgos; reporta y deja la corrección a `qa-apply`.
- Si el resultado contradice BookStack, preséntalo como contradicción al humano (G1/G4).
- No inventes ni asumas un `artifact_revision` — es siempre el que devuelve `qa-validate`.
- No acoples esta skill al CLI interno de RDD directamente — cuando la integración de Fase 3C exista, pasa por `QACodeReviewer`, nunca invoques `gentle-ai review inspect-candidate` a mano desde aquí.

## Comandos de referencia

- Tipos: `npx tsc --noEmit`.
- Ejecución: runner de Playwright del proyecto con reporter/trace habilitados.
- Reglas G1-G6: lee `skills/_shared/qa-gate-policy.md` (fuente única in-repo).
- Ledger: `gentle-ai qa-begin` / `gentle-ai qa-validate` / `gentle-ai qa-finish` (ver `docs/migration/qa-orchestrator-v3-design.md`).
