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
2. Usa los **path aliases reales del proyecto** (los detectados en G2 y declarados en el "Diseño Screenplay+POM" del spec aprobado de G3) — no asumas una lista fija de otro proyecto. Si el proyecto no tiene aliases o estructura Screenplay+POM previa, créala siguiendo exactamente el diseño aprobado en el spec, aplicando SOLID (ver spec).
3. Usa los **fixtures y tags** establecidos (ej. fixture `cotizacion-pedido`, tags `@FC-*`).
4. Implementa **exactamente** la arquitectura declarada en la sección "Diseño Screenplay+POM" del spec aprobado — reutilizando componentes existentes o creando los nuevos ahí diseñados. Un archivo de test con selectores/locators crudos (`page.getByText`, `page.getByRole`, etc.) fuera de Interactions/Targets es una violación de G5, no una opción válida, sin importar si el proyecto es nuevo o maduro.
5. **NO** agregues esperas fijas innecesarias; usa esperas explícitas de Playwright.
6. **NO** agregues dependencias ni toques config global sin autorización.
7. **NO** modifiques tests fuera del alcance del spec.
8. **NO** guardes secretos, tokens ni credenciales en el código.

## Validación previa a declarar terminado

- `npx tsc --noEmit` sin errores.
- Ejecuta la prueba modificada/creada y verifica que pase.
- Revisa que reutilices componentes existentes donde aplique.

## Guardrails

- Implementas solo el spec aprobado; si descubres un requisito nuevo, márcalo como pendiente de aprobación y NO lo implementes (G3/G5).

## Comandos de referencia

- Tipos: `npx tsc --noEmit`.
- Ejecución: runner de Playwright del proyecto (spec específico).