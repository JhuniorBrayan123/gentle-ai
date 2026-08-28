---
name: erp-docs-write
description: "Trigger: redactar Concepto + Guía paso a paso de un flujo ERP2 para cliente final, con BookStack y código Screenplay+POM. Produce SOLO un borrador."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

## Activation Contract

Carga esta skill cuando el humano pida documentación para el cliente final de
ERP2 sobre un flujo de negocio (ej. "documenta la emisión de una boleta",
"redacta la guía de este flujo"). Es un punto de entrada **independiente**,
NO forma parte del pipeline `qa-supervisor` → `qa-explore` → `qa-spec` →
`qa-apply` → `qa-review` → `qa-verify` → `qa-docs` (ese pipeline automatiza
pruebas Playwright; esta skill redacta documentación para el cliente que usa
el ERP2).

**Audiencia = SIEMPRE cliente final.** No redactes en tono de soporte, QA o
implementación — nada de nombres de clases CSS, IDs de elementos, jerga
técnica interna, ni pasos que solo tendría sentido seguir alguien de soporte.
Si una validación solo importa para diagnóstico interno, no entra en el
borrador cliente; repórtala aparte como nota para el humano.

Produce SIEMPRE un borrador. Nunca escribe en BookStack — esa acción
pertenece exclusivamente a `erp-docs-publish`, y solo corre tras aprobación
humana explícita.

## Fuentes de verdad (MANDATORY)

- **BookStack** (`bookstack_bookstack_search` / `bookstack_bookstack_get_page`)
  — fuente de la verdad para conceptos de negocio, reglas, y el estándar
  documental ERP2 vigente. Antes de escribir CUALQUIER borrador, consulta:
  - El libro **"99. Plantillas ERP2 y Estándares"** (`Estándar - Cómo
    documentar`, `Estándar - Convención de títulos`, `Estándar - Enlaces
    cruzados`, y el `SOP001 Propuesta de estructura documental ERP2`) para
    la taxonomía de libros/tags/nombres vigente — no la asumas de memoria,
    puede cambiar.
  - Páginas ya publicadas del mismo tipo en el módulo (busca `Guía -` /
    `Concepto -` del módulo) para igualar tono y detectar el siguiente
    número de guía en uso (las guías reales usan el patrón `Guía NNN-
    [Título]`; ubica el NNN más alto ya usado en el capítulo destino y
    continúa la numeración, no reinicies en 001 salvo que no exista
    ninguna).
  - Si ya existe una página para este mismo flujo, tu salida es una
    ACTUALIZACIÓN de esa página (mismo ID), no una página nueva duplicada.
- **Código Screenplay+POM del proyecto de automatización** —
  `src/screenplay/{tasks,interactions,questions,targets}/**` y
  `src/actors/**` — fuente de la verdad para cada paso, cada dato de
  pantalla, cada mensaje de confirmación y cada validación de resultado que
  aparezca en el borrador. NUNCA se lee el código fuente del backend/
  frontend de ERP2 — solo el proyecto de automatización.

## Regla dura: nada de pasos o datos inventados

Todo paso "crítico" (navegación entre pantallas, mensaje de confirmación,
campo de datos, criterio de verificación de resultado) debe anclarse en un
Target/Interaction/Question/tipo real del código, citando el nombre de
archivo como evidencia interna (no se imprime en la página final, pero debe
poder mostrarse si el humano lo pide). Si no encuentras respaldo en código
ni en BookStack para un paso, NO lo inventes: márcalo como
"[pendiente de confirmar con el equipo]" y sigue.

Ejemplo real de esta regla aplicada (emisión de boleta, Punto de Venta):
la secuencia "Nueva venta → entra a la caja → recién ahí se puede
seleccionar el producto" viene de una guía ya publicada, no es una
suposición de redacción. El paso de confirmación de resultado usa
`EmisionResult` (`serie`, `correlativo`, `montoTotalVenta`) como los
únicos campos garantizados que el sistema muestra al emitir — no se
listan campos que el tipo no garantiza. El paso de consulta posterior usa
los filtros reales de `ComprobantesFiltrosComponent` (fecha, cliente,
número/correlativo, serie, moneda, monto, tipo) y la columna de estado
real (`EstadoDelComprobante.enGrilla`, pestañas `TODOS`/`VENTAS`/
`FACTURACION`/`GUIAS`/`COTIZACIONES`/`PEDIDOS`) — no una descripción vaga
de "verificar que se emitió bien".

## Naming y taxonomía (MANDATORY — usa la del estándar vigente, no la fijes de memoria)

Confirma contra `Estándar - Convención de títulos` y `SOP001` antes de
titular, pero como referencia base:

- `Concepto - [Término]` — ej. `Concepto - Boleta de venta electrónica`.
- `Guía [NNN]- [Título descriptivo]` — sigue la numeración real ya en uso
  en el capítulo destino (ver arriba).
- Tags (BookStack): `audiencia: cliente-final`, `tipo-contenido:
  concepto|guia`, `modulo: <módulo real>`, `estado: borrador` (pasa a
  `vigente` solo cuando `erp-docs-publish` confirma la publicación),
  `version-producto: <la que indique el estándar vigente>`.
- Ubicación: `Concepto -` va en el libro/capítulo de producto del módulo
  (ej. glosario de facturación electrónica); `Guía -` va en el libro/
  capítulo de operación diaria del módulo (ej. el mismo capítulo donde ya
  viven las guías de ese módulo).

## Salida requerida — DOS páginas separadas, nunca mezcladas

### 1. `Concepto - [Término]`

- Tabla de metadatos: Tipo de documento (Concepto) | Audiencia (Cliente
  final) | Módulo | Estado documental | Responsable | Última actualización.
- **Respuesta rápida** (3-5 líneas, lenguaje llano).
- **Definición** completa, sin jerga técnica. Si existe página BookStack
  para el término, básate en ella y cítala; si no existe, redacta desde el
  dominio del negocio y márcalo como "[pendiente de validar con negocio]"
  en la nota interna del borrador (no en el cuerpo visible).
- **¿Cuándo lo vas a ver?** — en qué pantallas/flujos del ERP2 aparece.
- **Documentación relacionada** — enlaces a la(s) Guía(s) y Conceptos
  relacionados (obligatorio por `Estándar - Enlaces cruzados`).

### 2. `Guía [NNN]- [Acción]`

- Tabla de metadatos: igual que Concepto, con Tipo de documento = Guía
  paso a paso.
- **1. Objetivo** — qué logra el usuario siguiendo esta guía, en 1-2
  líneas.
- **2. Requisitos previos** — lista de condiciones necesarias antes de
  empezar (sesión iniciada, caja abierta si aplica, datos disponibles,
  etc.), sacadas de las precondiciones reales que valida el código/la
  guía existente del módulo.
- **3. Procedimiento paso a paso** — pasos numerados, uno por acción real
  de UI, en el orden real de ejecución. Cuando un paso implica cambiar de
  pantalla o tiene sub-acciones, usa sub-numeración `N.1`, `N.2`, ... bajo
  ese paso (no lo aplanes ni lo mezcles con el paso principal). Cuando un
  paso pide capturar datos con varios campos (ej. datos del cliente,
  medio de pago), preséntalos como tabla `Campo | Descripción` en vez de
  prosa. Marca `Captura sugerida: [qué debería verse en la pantalla]`
  después de cada paso donde una captura de pantalla ayude — nunca
  generes una imagen real, solo el marcador de texto para que el
  responsable la complete.
- **4. Confirmación** — el mensaje de éxito real (si el código/BookStack
  lo confirma) y los campos garantizados que muestra el comprobante
  resultante (ver regla dura arriba).
- **5. Acciones disponibles después** — qué puede hacer el usuario con el
  resultado (visualizar, imprimir, descargar, enviar, etc.), solo las que
  el código respalda.
- **6. Consultar esto más adelante** — cómo volver a encontrar el
  resultado (pantalla de búsqueda real, filtros reales disponibles, cómo
  leer el estado desde la misma grilla sin pasos de más).
- **7. Problemas frecuentes** — tabla `Situación | Acción recomendada`,
  solo con casos que el código o BookStack respaldan (mensajes de error
  reales, validaciones reales) — si no hay evidencia de un problema
  frecuente real, no la inventes.
- **8. Flujo resumido** — el procedimiento reducido a una lista corta de
  etapas, en **texto plano** (nunca imagen ni SVG), formato:

  ```
  Etapa 1
        ↓
  Etapa 2
        ↓
  Etapa 3
  ```

  usando separadores de texto simple (`↓`, líneas `------`), consistente
  con la regla RAG de "texto plano, sin decoración" — nada de HTML ni
  colores.

## Guardrails de formato (RAG-friendly, obligatorio — ver `SOP001` §7)

- Texto plano y estructurado: un dato por línea, sin HTML decorativo, sin
  cajas de color, sin mezclar estilos dentro de un mismo paso.
- Una página = una intención (Concepto separado de Guía, nunca fusionados).
- Títulos explícitos con las palabras que el cliente realmente buscaría.
- Sinónimos del negocio cuando ayuden (comprobante/documento,
  aceptado/validado por SUNAT).
- La metadata (estado, audiencia, módulo) va en tags/tabla de metadatos,
  nunca como frase suelta en el cuerpo ("Estado: DRAFT" está PROHIBIDO
  como texto narrativo — usa el campo "Estado documental" de la tabla).

## Guardrails (honestidad epistémica)

- Distingue SIEMPRE hechos observados en código, hechos citados de
  BookStack, e inferencias — nunca los mezcles sin marcarlos en la nota
  interna del borrador.
- PROHIBIDO inventar afirmaciones legales/regulatorias (ej. criterios de
  aceptación SUNAT) — solo afirma lo que el código o BookStack confirman
  de verdad; de lo contrario márcalo como "[pendiente de verificar]".
- NUNCA escribe en BookStack — la salida es siempre un borrador; publicar
  es una skill separada y explícita (`erp-docs-publish`).
- Si el flujo no tiene código de automatización correspondiente (sin
  artefactos Screenplay), decláralo explícitamente y construye el
  borrador solo con BookStack, señalando el vacío en vez de inventar
  pasos.
- Si detectas contradicción entre BookStack y el código, no decidas por
  tu cuenta: documenta la contradicción y escala al humano.

## Persistencia

- Persiste el borrador (ambas páginas) vía MCP Engram bajo el topic
  `erp-docs/{flow}/draft`, incluyendo la evidencia interna (rutas de
  código, páginas BookStack citadas) que no va en el cuerpo publicado.

## Comandos de referencia

- Búsqueda/lectura de docs: MCP BookStack (`bookstack_bookstack_search`,
  `bookstack_bookstack_get_page`).
- Persistencia: MCP Engram (`mem_save` topic `erp-docs/{flow}/draft`).
