# Política de compuertas QA — G1 a G6

> **Provenance**: BookStack es la fuente OFICIAL. Este archivo es la ÚNICA
> renderización in-repo.
> Página: "06. Reglas del Agente Orquestador QA" · ID: `TODO-MAINTAINER-FILL-BOOKSTACK-PAGE-ID` ·
> URL: `TODO-MAINTAINER-FILL-BOOKSTACK-URL` · Recuperado: `2026-09-16`
> Prohibido copiar este texto en otro SKILL.md — refiéranlo por ruta.
> Ante conflicto con BookStack: gana BookStack; corrige este archivo, no el skill.

> **NOTA PARA EL MANTENEDOR**: el ID y la URL de BookStack de arriba son
> placeholders explícitos, no valores reales. Un mantenedor con acceso a
> BookStack debe reemplazarlos por la página real "06. Reglas del Agente
> Orquestador QA" antes de considerar esta provenance completa.

## Regla Cero (obligatoria antes de cualquier cosa)

Antes de proponer, escribir una línea de código o delegar una implementación,
**DEBES**:

1. Usar el MCP de **BookStack** (`bookstack_bookstack_search`) para buscar la
   documentación oficial que sustenta la solicitud (PRD del módulo, criterios
   de aceptación reales del negocio, páginas oficiales del Agente QA) y
   **citar** las páginas/capítulos/secciones usadas.
2. Si la solicitud lo amerita, usar el MCP de **GitLab** para verificar el
   estado real del código que se tocará (último MR, rama, tags).
3. NO inventar convenciones ni supuestos cuando exista documentación oficial
   aplicable.

En la respuesta, enlaza el documento consultado bajo el título "Reglas
aplicables - Documentación consultada": cada página consultada se renderiza
como **ficha de documentación PRD** (formato completo en `qa-doc-reference`).

### La fuente de las reglas

La lógica es operativa, no inventada. Vive en BookStack:

- **Página "06. Reglas del Agente Orquestador QA"** — reglas G1 a G6
  (texto completo en la sección siguiente).
- **Página "07. Plantillas del Agente QA"** — plantilla de respuesta
  (10 secciones) y checklist final.
- **Páginas PRD del módulo** según la solicitud (ej. para Punto de Venta,
  libro `ERP_PV_Punto Venta`).

## Reglas G1 a G6 (obligatorias, priman sobre todo)

Debes respetar y verificar cada una antes de autorizar cualquier
implementación:

- **G1 — Documentación como fuente oficial**: consultar BookStack antes de
  proponer; citar páginas; NO inventar convenciones; si falta/incompleta/
  contradice → detener e informar el vacío; si BookStack difiere del código
  actual, NO decidas tú, presenta la contradicción al humano.
- **G2 — Análisis previo**: buscar tests similares en el módulo; revisar
  fixtures/helpers/Tasks/Questions/Pages reutilizables; verificar
  convenciones de nombres y ubicación; revisar la config de Playwright;
  evaluar setup y prerrequisitos; medir impacto en otras pruebas.
- **G3 — Planificación obligatoria**: para cambios medianos/grandes NO
  implementes directo. Entrega un plan (objetivo, documentación consultada,
  pruebas similares, componentes reutilizables, archivos a crear/modificar,
  riesgos, validaciones, alcance/fuera-de-alcance). La implementación SOLO
  tras aprobación humana.
- **G4 — Manejo de incertidumbre**: distingue hechos-de-BookStack vs
  observados-en-código vs inferencias vs recomendaciones vs
  pendiente-de-confirmar. Si un criterio no está definido, pide aclaración.
  NO conviertas una suposición en regla de negocio.
- **G5 — Control de riesgos**: no toques config global sin autorización; no
  agregues dependencias sin justificar; no elimines código sin analizar
  referencias; no modifiques tests fuera del alcance; no guardes
  secretos/tokens/contraseñas; no ejecutes comandos destructivos; no
  sobreescribas en BookStack durante la primera fase.
- **G6 — Validación de la implementación**: al declarar finalizada una
  implementación, exige: `npx tsc --noEmit`; ejecutar la prueba modificada;
  revisar lint; verificar que no haya credenciales; verificar que no haya
  esperas fijas innecesarias; verificar reutilización de componentes;
  comparar el resultado contra la documentación consultada; entregar el
  comando de ejecución y el resultado.

## Checklist de evidencia G6 (obligatoria, TODA)

Al declarar finalizada una implementación o validación, exige y ejecuta:

1. `npx tsc --noEmit` — tipos compilan.
2. Ejecutar la prueba modificada/creada — pasa.
3. Revisar lint del proyecto — limpio.
4. Verificar que NO haya credenciales/secretos hardcodeados
   (`waitForTimeout`, `sleep`, etc. cuentan como esperas fijas, no como
   credenciales — verificar ambas por separado).
5. Verificar que NO haya esperas fijas innecesarias.
6. Verificar que el test siga **Screenplay+POM** (sin locators crudos en el
   archivo de test; actor/tasks/questions/targets usados según el diseño del
   spec) — reutilizando componentes existentes (Page Objects, Fixtures,
   Helpers, Tasks, Interactions, Questions) cuando aplica, o con la
   estructura nueva creada según SOLID cuando el proyecto no tenía patrón
   previo.
7. Comparar el resultado contra la documentación consultada (BookStack).
8. Entregar el comando de ejecución exacto y el resultado/evidencia
   obtenido (screenshots, traces, videos o salida de reporter de
   Playwright).

## Uso de este archivo

- Cada `SKILL.md` de la familia `qa-*` referencia este archivo por ruta
  (`skills/_shared/qa-gate-policy.md`) en vez de copiar el texto de arriba.
- Si necesitas citar una regla puntual en un skill, usa el identificador
  corto (`G1`, `G2`, ... `G6`) y enlaza aquí — nunca reproduzcas el párrafo
  completo.
