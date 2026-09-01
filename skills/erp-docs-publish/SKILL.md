---
name: erp-docs-publish
description: "Trigger: publicar en BookStack la ficha de flujo ERP2 (una sola página) ya aprobada por el humano. Único punto que escribe en BookStack."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "2.0"
---

## Activation Contract

Carga esta skill SOLO cuando el humano haya revisado y aprobado explícitamente
un borrador producido por `erp-docs-write` (ej. "publica esto en BookStack",
"aprobado, publícalo"). Es un punto de entrada **independiente**, NO forma
parte del pipeline `qa-supervisor` → `qa-explore` → `qa-spec` → `qa-apply` →
`qa-review` → `qa-verify` → `qa-docs`.

Esta es la ÚNICA skill del par `erp-docs-write` / `erp-docs-publish`
autorizada a escribir en BookStack — puede crear o actualizar, pero
**nunca decide cuál de las dos por su cuenta**: solo ejecuta el destino
que el humano confirmó explícitamente en `erp-docs-write` (o que confirma
ahora mismo, si publicas sin pasar antes por esa skill). Por defecto, si
no hay una decisión de "actualizar" explícita, el destino es crear página
nueva.

## Precondición obligatoria

- Requiere un borrador APROBADO: de esta misma sesión, o recuperado desde el
  topic Engram `erp-docs/{flow}/draft`. El borrador trae UNA sola página
  (la ficha del flujo, título `[Módulo] - [Submódulo] - [Tema]`).
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
   este flujo justo antes de publicar:
   - Si el destino confirmado es **crear** y aparece una página que no
     estaba cuando se armó el borrador (alguien más la creó mientras
     tanto), o si el libro/capítulo confirmado ya no coincide con
     `Estándar - Convención de títulos` / `Estándar - Enlaces cruzados` /
     `SOP001`, DETENTE y presenta el cambio al humano — nunca conviertas
     un "crear" en "actualizar" por tu cuenta, solo confirma si igual se
     crea la nueva (enlazándola) o si el humano prefiere cancelar.
   - Si el destino confirmado es **actualizar página [ID]**, verifica que
     ese ID todavía exista y siga siendo la página correcta (título/
     libro/capítulo no cambiaron de forma que ya no calce) antes de
     tocarla; si algo no cuadra, DETENTE y confirma con el humano en vez
     de actualizar igual.
4. **Confirma la versión semver antes de escribir (mismo gate que
   `gitlab-release-tag` antes de crear un tag — nunca se bumpea sin
   mostrarla primero).**
   - Si el destino es **crear**: `Versión` va `1.0.0`, sin excepción.
   - Si el destino es **actualizar**: leé la `Versión` actual de la
     página real (`bookstack_bookstack_get_page`, no la del borrador
     viejo), tomá la clasificación PATCH/MINOR/MAJOR que trae el borrador
     de `erp-docs-write` (ver esa skill para el criterio), y mostrale al
     humano la versión actual → propuesta + el motivo. Solo publicá
     después de que la confirme — si el humano corrige la clasificación,
     usá la que él diga.
5. Ejecuta exactamente el destino confirmado:
   - **Crear**: usa `bookstack_create_page` en el libro/capítulo
     confirmado.
   - **Actualizar**: usa `bookstack_update_page` ÚNICAMENTE sobre el ID
     de página que el humano confirmó explícitamente (en `erp-docs-write`
     o en esta misma conversación) — nunca sobre una página que tú
     elegiste por similitud. Reemplaza el contenido completo por el del
     borrador aprobado; no mezcles contenido viejo no revisado con el
     nuevo.
6. Publica EXACTAMENTE el contenido aprobado — ninguna edición no aprobada
   se cuela en la publicación. Nunca agregues de vuelta información
   técnica/interna (dependencias, permisos) que el borrador haya
   descartado a propósito del cuerpo publicado. Aplica los tags
   (`audiencia`, `tipo-contenido`, `modulo`, `version-producto`) y fija
   `estado: vigente`.
7. Persiste un log vía MCP Engram bajo el topic `erp-docs/{flow}/publish-log`
   con: qué se publicó (la ficha), URL/ID de la página, y contexto de
   fecha/hora.

## Guardrails

- PROHIBIDO publicar un borrador que el humano no haya aprobado en esta
  conversación — la aprobación no se infiere ni se asume.
- PROHIBIDO publicar mientras existan contradicciones sin resolver o pasos
  sin confirmar señalados en el borrador.
- PROHIBIDO introducir contenido nuevo no presente en el borrador aprobado
  al momento de publicar.
- PROHIBIDO usar `bookstack_update_page` sobre cualquier página cuyo ID
  no haya sido confirmado explícitamente por el humano para este flujo —
  "se parece" o "probablemente es esta" no es confirmación.
- PROHIBIDO decidir "actualizar" cuando el destino confirmado era "crear"
  (o viceversa) solo porque el estado de BookStack cambió — vuelve a
  preguntar en vez de reinterpretar la decisión original.
- Cita siempre la(s) página(s) resultante(s) (creada o actualizada,
  según corresponda) al confirmar la publicación al humano.
- PROHIBIDO bumpear la `Versión` sin mostrarle antes al humano la versión
  actual, la propuesta, y el motivo (PATCH/MINOR/MAJOR) — igual que
  `gitlab-release-tag` nunca crea un tag sin confirmación explícita.
- PROHIBIDO adivinar la `Versión` actual de memoria o del borrador viejo
  — siempre leela de la página real justo antes de publicar.

## Comandos de referencia

- Lectura de borrador: MCP Engram (`mem_get_observation` / `mem_search`
  topic `erp-docs/{flow}/draft`).
- Publicación: MCP BookStack (`bookstack_create_page` para destino nuevo,
  `bookstack_update_page` solo para el ID de página confirmado por el
  humano).
- Registro de publicación: MCP Engram (`mem_save` topic
  `erp-docs/{flow}/publish-log`).
