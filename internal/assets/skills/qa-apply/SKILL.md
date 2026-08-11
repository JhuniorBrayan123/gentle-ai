---
name: qa-apply
description: "Trigger: implementar un cambio QA aprobado. Escribe tests con Screenplay+POM (G5) siguiendo el spec y las convenciones del proyecto."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
disable-model-invocation: true
user-invocable: false
---

## Activation Contract

Carga esta skill cuando vayas a implementar un cambio QA **aprobado** (spec validado por el humano). Eres el sub-agente de **implementación (G5)** del orquestador QA. Implementas SOLO lo especificado, sin alcance adicional.

## Fuentes de verdad (MANDATORY)

- **Spec aprobado** (`qa/{change}/spec`) = contrato de lo que se implementa.
- **BookStack = fuente de la verdad**: ante cualquier duda de convención o criterio, vuelve a `bookstack_bookstack_search` y cita la página; NO decidas divergencias tú (G1).
- **Engram = memoria persistente**: registra el progreso en `qa/{change}/apply-progress`.

## Reglas de implementación (G5)

1. Sigue el patrón **Screenplay + POM**: actores, tareas, preguntas, habilidades y páginas del proyecto.
2. Usa los **path aliases** del proyecto: `@task/*`, `@question/*`, `@pages/*`, `@fixtures/*`, `@helpers/*`, `@utils/*`.
3. Usa los **fixtures y tags** establecidos (ej. fixture `cotizacion-pedido`, tags `@FC-*`).
4. **NO** agregues esperas fijas innecesarias; usa esperas explícitas de Playwright.
5. **NO** agregues dependencias ni toques config global sin autorización.
6. **NO** modifiques tests fuera del alcance del spec.
7. **NO** guardes secretos, tokens ni credenciales en el código.

## Validación previa a declarar terminado

- `npx tsc --noEmit` sin errores.
- Ejecuta la prueba modificada/creada y verifica que pase.
- Revisa que reutilices componentes existentes donde aplique.

## Guardrails

- Implementas solo el spec aprobado; si descubres un requisito nuevo, márcalo como pendiente de aprobación y NO lo implementes (G3/G5).

## Comandos de referencia

- Tipos: `npx tsc --noEmit`.
- Ejecución: runner de Playwright del proyecto (spec específico).