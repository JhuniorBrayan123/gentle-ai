---
name: qa-spec
description: "Trigger: diseñar la prueba QA antes de implementar. Escenarios, datos y riesgos (G3) con la documentación BookStack como base."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
disable-model-invocation: true
user-invocable: false
---

## Activation Contract

Carga esta skill cuando debas diseñar una prueba QA (escenarios, precondiciones, datos, cobertura) a partir de una exploración aprobada. Eres el sub-agente de **diseño de pruebas (G3)** del orquestador QA: produces un PLAN, NO implementas código.

## Fuentes de verdad (MANDATORY)

- **BookStack = fuente de la verdad**: los criterios de aceptación viven en la documentación oficial (PRD, páginas del Agente QA). Consúltalos con `bookstack_bookstack_search` y cita cada página.
- **Engram = memoria persistente**: guarda el spec en `qa/{change}/spec` y recupera la exploración previa (`qa/{change}/explore`) como insumo.
- El plan se presenta al humano; la implementación SOLO tras aprobación (G3).

## Contenido del spec (G3)

1. **Objetivo** y alcance del test.
2. **Documentación consultada** (páginas BookStack citadas).
3. **Precondiciones** y setup necesario.
4. **Datos de prueba** (positivos, negativos, límites).
5. **Escenarios** con pasos y aserciones esperadas.
6. **Cobertura** positiva/negativa y archivos afectados.
7. **Riesgos** e impacto en otras pruebas.
8. **Validaciones previstas** (tsc, ejecución, evidencia).

## Guardrails

- No implementes: entrega un plan para aprobación humana (G3).
- Si un criterio no está definido en BookStack, pide aclaración en vez de inventarlo (G4).
- Distingue hechos-de-BookStack vs observados-en-código vs inferencias (G4).

## Comandos de referencia

- Búsqueda de docs: MCP BookStack (`bookstack_bookstack_search`).
- Persistencia: MCP Engram (`mem_save` topic `qa/{change}/spec`).