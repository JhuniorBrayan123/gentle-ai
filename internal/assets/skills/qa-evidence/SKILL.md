---
name: qa-evidence
description: "Trigger: screenshots/traces/videos/reporter + G6 checklist + commands. Provides evidence collection guidelines."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

## Evidence Collection Protocol (MANDATORY)

1. Always capture screenshots, traces, videos, or reporter output for QA tests.
2. Follow the G5 and G6 checklist below before finalizing any test implementation.
3. Provide the exact commands used to run tests and collect evidence.
4. Ensure evidence is reproducible and clearly linked to the task.

## Mandatory QA Rules (G5 & G6)

You MUST verify these rules before declaring any implementation or testing task complete:

- **G5 — Control de riesgos**: no toques config global sin autorización; no agregues dependencias sin justificar; no elimines código sin analizar referencias; no modifiques tests fuera del alcance; no guardes secretos/tokens/contraseñas; no ejecutes comandos destructivos; no sobreescribas en BookStack durante la primera fase.
- **G6 — Validación de la implementación (Checklist)**: al declarar finalizada una implementación, exige y ejecuta:
  - `npx tsc --noEmit` para verificar tipos.
  - Ejecutar la prueba modificada.
  - Revisar lint.
  - Verificar que NO haya credenciales hardcodeadas.
  - Verificar que NO haya esperas fijas innecesarias (`waitForTimeout`, `sleep`, etc).
  - Verificar que el test siga el patrón **Screenplay+POM** (sin locators crudos en el archivo de test; actor/tasks/questions/targets usados según el diseño del spec) — reutilizando componentes existentes (Page Objects, Fixtures, Helpers, Tasks, Interactions, Questions) cuando aplica, o con la estructura nueva creada según SOLID cuando el proyecto no tenía patrón previo.
  - Comparar el resultado contra la documentación consultada (BookStack).
  - Entregar el comando de ejecución exacto y el resultado/evidencia obtenido.
