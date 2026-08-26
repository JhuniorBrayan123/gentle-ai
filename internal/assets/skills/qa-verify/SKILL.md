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
- **Engram = memoria persistente**: guarda el reporte en `qa/{change}/verify-report` y evidencia asociada.

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

## Guardrails

- No modifiques código para "arreglar" hallazgos; reporta y deja la corrección a `qa-apply`.
- Si el resultado contradice BookStack, preséntalo como contradicción al humano (G1/G4).

## Comandos de referencia

- Tipos: `npx tsc --noEmit`.
- Ejecución: runner de Playwright del proyecto con reporter/trace habilitados.