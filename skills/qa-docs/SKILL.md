---
name: qa-docs
description: "Trigger: detectar vacíos o contradicciones en la documentación QA. Produce borradores de gap, nunca escribe en BookStack (G1.4)."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
disable-model-invocation: true
user-invocable: false
---

## Activation Contract

Carga esta skill cuando debas detectar **gaps o contradicciones de documentación** en un cambio QA. Eres el sub-agente de **doc-gap detection (G1.4)** del orquestador QA: produces SOLO borradores (DRAFT), NUNCA escribes en BookStack.

## Fuentes de verdad (MANDATORY)

- **BookStack = fuente de la verdad**: revisa la cobertura documental del cambio con `bookstack_bookstack_search`; cita páginas existentes y las que faltan.
- **Engram = memoria persistente**: guarda el borrador en `qa/{change}/doc-gaps` y recupera `qa/{change}/review-report` como insumo.

## Qué detectar (G1.4)

1. **Gap**: documentación oficial faltante para el módulo/criterio del cambio.
2. **Contradicción**: docs que difieren del código o de otras páginas.
3. **Obsolescencia**: páginas que ya no reflejan el comportamiento actual.
4. **Falta de detalle**: criterios sin especificar (datos, precondiciones, aserciones).

## Salida (DRAFT ONLY)

- Borrador con: gap identificado, página(s) involucrada(s), evidencia del código, propuesta de contenido.
- Marca TONE como DRAFT y deja la decisión de publicación al humano.

## Guardrails

- **NUNCA escribas ni modifiques BookStack** — solo producís borradores (G5).
- Si detectas contradicción entre BookStack y código, no decides: documentas y escalas (G1/G4).

## Comandos de referencia

- Búsqueda de docs: MCP BookStack (`bookstack_bookstack_search`).
- Persistencia: MCP Engram (`mem_save` topic `qa/{change}/doc-gaps`).