---
name: erp-docs-publish
description: "Trigger: publicar en BookStack un borrador ERP2 (Concepto + Guía) ya aprobado por el humano. Único punto que escribe en BookStack."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

## Activation Contract

Carga esta skill SOLO cuando el humano haya revisado y aprobado explícitamente
un borrador producido por `erp-docs-write` (ej. "publica esto en BookStack",
"aprobado, publícalo"). Es un punto de entrada **independiente**, NO forma
parte del pipeline `qa-supervisor` → `qa-explore` → `qa-spec` → `qa-apply` →
`qa-review` → `qa-verify` → `qa-docs`.

Esta es la ÚNICA skill del par `erp-docs-write` / `erp-docs-publish`
autorizada a escribir en BookStack. Y solo hace UNA cosa: **crear páginas
nuevas**. Nunca actualiza/modifica una página ya existente, bajo ninguna
circunstancia.

## Precondición obligatoria

- Requiere un borrador APROBADO: de esta misma sesión, o recuperado desde el
  topic Engram `erp-docs/{flow}/draft`. El borrador siempre trae DOS páginas
  (`Concepto - [Término]` y `Guía [NNN]- [Acción]`) — nunca publiques solo
  una si el borrador definía ambas, salvo que el humano apruebe publicar
  parcialmente.
- NUNCA publica contenido que no haya sido mostrado y aprobado por el humano
  en esta conversación. Si no hay evidencia de aprobación explícita, DETENTE
  y pide confirmación antes de escribir nada en BookStack.

## Comportamiento

1. Recupera el borrador aprobado (sesión actual o Engram
   `erp-docs/{flow}/draft`).
2. Si el borrador señaló contradicciones sin resolver entre código y
   BookStack, o pasos marcados "[pendiente de confirmar con el equipo]" /
   "[pendiente de verificar]", RECHAZA publicar esa parte hasta que el
   humano la resuelva.
3. **Re-verifica el destino, no confíes ciegamente en lo que decidió
   `erp-docs-write`.** Vuelve a correr `bookstack_bookstack_search` para
   este flujo justo antes de publicar: si aparece una página que no
   estaba cuando se armó el borrador (alguien más la creó mientras tanto),
   o si el libro/capítulo confirmado en el borrador ya no coincide con
   `Estándar - Convención de títulos` / `Estándar - Enlaces cruzados` /
   `SOP001`, DETENTE y presenta el cambio al humano. Esto NUNCA se
   resuelve actualizando la página que apareció — solo confirma si igual
   se crea la nueva (enlazándola) o si el humano prefiere cancelar.
4. Usa `bookstack_create_page` para crear la página nueva en el destino
   confirmado (en el borrador, o re-confirmado en el paso 3) — para
   `Guía`, verifica de nuevo el último NNN en uso justo antes de crear,
   por si otra página se publicó entre el borrador y ahora. Nunca uses
   `bookstack_update_page` ni ninguna otra operación que modifique una
   página existente.
5. Publica EXACTAMENTE el contenido aprobado — ninguna edición no aprobada
   se cuela en la publicación. Aplica los tags (`audiencia`, `tipo-
   contenido`, `modulo`, `version-producto`) y fija `estado: vigente`.
6. Persiste un log vía MCP Engram bajo el topic `erp-docs/{flow}/publish-log`
   con: qué se publicó (Concepto y/o Guía), URL/ID de cada página, y
   contexto de fecha/hora.

## Guardrails

- PROHIBIDO publicar un borrador que el humano no haya aprobado en esta
  conversación — la aprobación no se infiere ni se asume.
- PROHIBIDO publicar mientras existan contradicciones sin resolver o pasos
  sin confirmar señalados en el borrador.
- PROHIBIDO introducir contenido nuevo no presente en el borrador aprobado
  al momento de publicar.
- PROHIBIDO publicar una `Guía` con un número `NNN` ya usado por otra
  página del mismo capítulo — vuelve a verificar el correlativo antes de
  crear.
- PROHIBIDO usar `bookstack_update_page` o modificar de cualquier forma
  una página ya existente — esta skill solo crea páginas nuevas, siempre.
- Cita siempre la(s) página(s) nueva(s) creada(s) al confirmar la
  publicación al humano.

## Comandos de referencia

- Lectura de borrador: MCP Engram (`mem_get_observation` / `mem_search`
  topic `erp-docs/{flow}/draft`).
- Publicación: MCP BookStack (`bookstack_create_page` únicamente).
- Registro de publicación: MCP Engram (`mem_save` topic
  `erp-docs/{flow}/publish-log`).
