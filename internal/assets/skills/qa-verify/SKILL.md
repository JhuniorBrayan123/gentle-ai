---
name: qa-verify
description: "Trigger: validar la implementación QA contra el spec. Ejecuta pruebas y produce evidencia reproducible (G6) siguiendo la checklist completa."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
disable-model-invocation: true
user-invocable: false
---

## Activation Contract

Carga esta skill cuando debas validar una implementación QA contra su spec. Eres el sub-agente de **validación y evidencia (G6)** del orquestador QA. NO corriges el código: reportas hallazgos verificables.

## Fuentes de verdad (MANDATORY)

- **Spec** (`qa/{change}/spec`) = contrato de validación.
- **BookStack = fuente de la verdad**: el resultado debe compararse contra la documentación consultada; cita las páginas usadas (G1).
- **Engram = memoria persistente**: registra el progreso en `qa/{change}/verify-report`.
- **Ledger nativo = el gate real**: `verify` no requiere aprobación manual (el humano ya aprobó en spec), pero el ledger DEBE cerrarse para que la cadena avance a `complete`.

## Entrada (obligatorio, antes de validar)

1. Ejecuta `gentle-ai qa-begin --change {change} --stage verify --cwd <repo> --request-id <id-idempotente> --evidence-goal <objetivo de esta validación>`.
2. Si el comando se rechaza, **DETENTE** y devuelve el rechazo tal cual al orquestador/supervisor.
3. Solo si `qa-begin` responde con éxito continúa con la validación de abajo.

## Checklist G6 (obligatoria, TODA)

1. `npx tsc --noEmit` — tipos compilan.
2. Ejecutar la prueba modificada/creada — pasa.
3. Revisar lint del proyecto — limpio.
4. Verificar que NO haya credenciales/secretos en el diff.
5. Verificar que NO haya esperas fijas innecesarias.
6. Verificar que el test siga Screenplay+POM (sin locators crudos en el archivo de test, actor/tasks/questions usados según el diseño del spec) — reutilizando lo existente cuando aplica, o con la estructura nueva creada según SOLID cuando el proyecto no tenía patrón previo.
7. Comparar el resultado contra la documentación BookStack consultada.
8. Entregar el comando de ejecución y el resultado obtenido.

## Evidencia reproducible (G6)

- Captura **screenshots**, **traces** (en fallo), **videos** o salida de **reporter** de Playwright.
- Registra artefactos con rutas y comandos exactos de reproducción.

## Salida (obligatorio, cierre de la etapa)

1. **Cuerpo del reporte en Engram**: `mem_save` con `topic_key: "qa/{change}/verify-report"` — hallazgos, comandos ejecutados y resultados de la validación.
2. **Admisión anti-alucinación**: arma el envelope `gentle-ai.qa-stage-artifact/v1` (ver `internal/qastage/artifact.go`) y pásalo por `gentle-ai qa-validate --input - --change {change} --stage verify --source-revision <artifact_revision_del_apply>`.
3. **Cierre del ledger**: con el `artifact_revision` admitido, ejecuta `gentle-ai qa-finish --change {change} --cwd <repo> --request-id <id-idempotente> --outcome passed --evidence-revision <artifact_revision> --diagnosis <resumen> --harness-disposition <reused|invalidated> --cleanup-evidence <texto> --process-evidence <texto>`. Si hay hallazgos graves o la prueba falla, usa `--outcome failed`.
4. Solo tras un `qa-finish` exitoso reportas la etapa `verify` como cerrada al orquestador/supervisor.

## Guardrails

- No modifiques código para "arreglar" hallazgos; reporta y deja la corrección a `qa-apply`.
- Si el resultado contradice BookStack, preséntalo como contradicción al humano (G1/G4).

## Comandos de referencia

- Tipos: `npx tsc --noEmit`.
- Ejecución: runner de Playwright del proyecto con reporter/trace habilitados.
- Reglas G1-G6: lee `skills/_shared/qa-gate-policy.md` (fuente única in-repo).