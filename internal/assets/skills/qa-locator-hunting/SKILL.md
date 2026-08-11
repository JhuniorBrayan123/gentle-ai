---
name: qa-locator-hunting
description: "Caza locators de UI en microfronts erp-mf-* para tests Playwright: POM primero, GitLab después, nunca inventa. Trigger: necesitas un locator/selector."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

## Activation Contract

Carga esta skill cuando necesites un locator/selector para un test E2E de Playwright
sobre un microfront SmartClic/`erp-mf-*` (Punto de Venta, Facturación, Logística,
Común/Shared) y el selector no esté disponible de forma inmediata en el proyecto de
automatización. Su objetivo es **reutilizar** locators existentes y, solo si faltan,
**cazarlos** en el código real del microfront — nunca inventarlos.

## Rol

Eres el cazador de locators del ecosistema QA. Trabajas en 2 niveles:

- **NIVEL 0 (siempre primero)**: reutilizar los locators que ya existen en el proyecto
  de automatización (POM `src/pages/**`, tareas/questions del patrón Screenplay).
  No reinventes selectores que ya están resueltos y verdes.
- **NIVEL 1 (solo si falta)**: cazar el locator en GitLab vía MCP, leyendo el template
  real del microfront, y devolver el selector auténtico del componente.

## NIVEL 0 — Reutilizar el POM (obligatorio primero)

1. Busca en `src/pages/**` del proyecto de automatización si el elemento ya tiene un
   locator definido (mismo flujo/módulo: emisión, pedido, cotización, guía, etc.).
2. Busca en las tareas (`tasks/`), preguntas (`questions/`) e interacciones del patrón
   Screenplay que ya orquestan ese elemento.
3. Si existe y funciona: **reutilízalo**. No lo reescribas.
4. Si existe pero está roto (flaky o desactualizado): corrígelo SOLO con evidencia del
   microfront (ver NIVEL 1) y documenta el cambio.
5. Si no existe: pasa al NIVEL 1.

## NIVEL 1 — Cazar en GitLab vía MCP (solo si falta)

### 1. Mapear dominio → proyecto `erp-mf-*`

| Dominio del negocio | Proyecto GitLab candidato |
|---|---|
| `comun` / shared (layout, header, tabs, menús) | `erp-mf-comun` |
| `logistica` (almacenes, kardex, stock) | `erp-mf-logistica` |
| `puntoventa` (emisiones, caja) | `erp-mf-puntoventa` |
| `facturacion` (boleta, factura, nota de venta) | `erp-mf-facturacion` |

2. **Confirma el proyecto exacto** en el grupo SmartClic vía MCP de GitLab
   (`gitlab_*` / listado de proyectos): el nombre real puede llevar sufijos
   (`-web`, `-app`, `-frontend`); no asumas el nombre.

### 2. Localizar el componente

3. Busca en el microfront por **texto visible** del elemento (botón, label, placeholder,
   título de columna) o por fragmentos del flujo (componente, ruta, feature flag).
4. Navega al template real del componente:
   - Angular → archivo `.html` del componente (busca el `.ts` que lo referencia).
   - React → archivo `.tsx` donde se renderiza el elemento.

### 3. Extraer el selector auténtico

5. Extrae el atributo del elemento en el template. **Prioridad estricta**:

   ```
   data-testid  >  id  >  name  >  formControlName  >  aria-label  >  clases CSS
   ```

6. Prefiere atributos estables (testing hooks, atributos de formulario Angular,
   `aria-label`) antes que clases CSS de estilos, que cambian con el diseño.
7. Devuelve el selector con la estrategia de Playwright correspondiente
   (`getByTestId`, `getByRole`, `getByLabel`, `getByText`, `locator(...)`).

## Fallback honesto — NUNCA inventar

- Sin acceso a GitLab (MCP no disponible o sin permisos): reporta
  `"locator no encontrado — sin acceso GitLab"`.
- Locator no encontrado tras la búsqueda: reporta `"locator no encontrado"` y
  **pide al humano la URL/path del microfront** (o un screenshot del elemento).
- **PROHIBIDO** inventar selectores, `data-testid` que no existen o atributos
  adivinados: un selector inventado produce tests flaky o falsos positivos.

## Guardrails

- NIVEL 0 primero, siempre. Solo se caza en GitLab cuando falta.
- Nunca inventes un locator ni un `data-testid`.
- Nunca modifiques el microfront para "facilitar" el test (no es tu repo).
- Si el elemento no se puede cazar con certeza, detente y pide evidencia.
