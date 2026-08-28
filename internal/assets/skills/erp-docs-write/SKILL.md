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
  - Nunca decidas solo con la búsqueda si existe o no una página previa —
    sigue la sección "Confirmación de destino" de abajo antes de redactar.
- **Código Screenplay+POM del proyecto de automatización** —
  `src/screenplay/{tasks,interactions,questions,targets}/**`,
  `src/actors/**`, y también `src/task/**` (convención más antigua del
  mismo proyecto, con Tasks reales como `EmitirGuiaRemitente.task.ts`) —
  fuente de la verdad para cada paso, cada dato de pantalla, cada mensaje
  de confirmación y cada validación de resultado que aparezca en el
  borrador. Busca en AMBAS convenciones antes de asumir que un flujo "no
  tiene automatización" — puede estar en la carpeta antigua. NUNCA se lee
  el código fuente del backend/frontend de ERP2 — solo el proyecto de
  automatización.

## Confirmación de destino (MANDATORY — antes de redactar, no después)

**Regla dura: esta skill NUNCA actualiza una página existente.** Su único
destino posible es crear una página nueva. Publicar en el libro/capítulo
equivocado tampoco se corrige editando el contenido después — hay que
evitarlo antes de escribir una sola línea. Por eso, antes de producir el
borrador, la skill DEBE presentar al humano lo siguiente y esperar su
confirmación (no asumir ni decidir sola):

1. **Búsqueda de páginas existentes (solo para avisar, no para decidir
   actualizar)**: ejecuta `bookstack_bookstack_search` con el nombre del
   flujo/término y sinónimos de negocio razonables (ej. "boleta",
   "comprobante", "emisión de boleta"). Lista TODOS los resultados
   relevantes encontrados (título, ID, libro/capítulo, URL) — incluso
   coincidencias parciales o dudosas. Esto es solo para que el humano
   sepa que ya existe contenido relacionado y decida si igual quiere una
   página nueva, si prefiere que él mismo la actualice manualmente, o si
   cancela — la skill jamás propone ni ejecuta una actualización. Pregunta
   explícitamente:
   > "Encontré estas páginas relacionadas: [lista]. Esta skill solo crea
   > páginas nuevas, nunca actualiza una existente — ¿igual quieres que
   > cree una nueva (enlazándola a estas), o prefieres actualizar tú
   > mismo la que ya existe?"
   Si la búsqueda no encuentra nada, dilo explícitamente ("no encontré
   ninguna página existente para este flujo") en vez de asumir
   silenciosamente que no existe.
2. **Ubicación propuesta**: antes de redactar, propone en qué libro y
   capítulo iría cada página (`Concepto -` y `Guía -`), basándote en la
   taxonomía vigente (`SOP001`, `Estándar - Cómo documentar`) y en dónde
   viven las páginas `Guía -`/`Concepto -` ya existentes de ese mismo
   módulo. Pregunta explícitamente:
   > "Propongo publicar el Concepto en [libro/capítulo] y la Guía en
   > [libro/capítulo]. ¿Confirmas, o va en otro lugar?"
3. Solo después de que el humano responda ambos puntos, redacta el
   borrador — con el destino (siempre página nueva, en el libro/capítulo
   confirmado) ya fijado y guardado junto al borrador en Engram, para que
   `erp-docs-publish` no tenga que volver a adivinarlo.

## Regla dura: nada de pasos o datos inventados (mecánica obligatoria, no opcional)

**PROHIBIDO redactar el procedimiento de memoria/conocimiento general de
ERPs, aunque suene plausible.** El único método válido para escribir el
"Procedimiento paso a paso" es:

1. Localiza el/los archivo(s) de Task/Interaction que implementan el flujo
   (busca en `src/task/**`, `src/screenplay/tasks/**`,
   `src/screenplay/interactions/**` por nombre del flujo — ej. para
   "emitir guía de remisión remitente" es
   `EmitirGuiaRemitente.task.ts`/`NavegarAGuiaRemitente.task.ts`, no un
   nombre parecido ni un archivo distinto "que debería existir").
2. Lee el archivo completo y extrae, EN ORDEN, cada `test.step('...')`
   (o Interaction/Task equivalente) que ejecuta. Ese orden es el único
   orden válido para el procedimiento — no lo reordenes, no lo resumas
   saltándote pasos condicionales, y no fusiones dos `test.step` en uno
   salvo que sean literalmente la misma acción de UI.
3. Cada paso de la guía = un `test.step` real (o un grupo mínimo cuando
   varios `test.step` seguidos son la misma pantalla/acción, ej. "abrir
   menú X" + "seleccionar Y" si son navegación directa a un mismo lugar).
   Si el código tiene una rama condicional (ej. `if (data.modalidad ===
   'PUBLICA')`), la guía debe reflejar AMBAS ramas como alternativas
   explícitas (no elegir una y omitir la otra).
4. Si un paso del flujo real no tiene una traducción evidente a lenguaje
   de cliente (ej. un campo interno de test), no lo inventes ni lo
   ocultes sin más: pregunta al humano cómo describirlo o dilo
   explícitamente como "[pendiente de confirmar con el equipo]".
5. Antes de entregar el borrador, verifica tú mismo: ¿cada paso escrito
   corresponde a un `test.step` que puedes señalar por nombre de archivo
   y línea? Si no puedes responder eso para algún paso, ese paso está
   inventado — bórralo o reemplázalo por la verificación real.

Un guía que omite un paso real del código (ej. seleccionar la caja antes
de poder elegir producto, o completar punto de partida/llegada en una
guía de remisión) o que inventa un paso que el código no tiene (ej. un
campo "Observaciones" que no existe en el Task) es una violación de esta
regla, sin importar qué tan natural suene en la redacción.

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

**Lista cerrada de secciones — no agregues, quites ni renombres ninguna.**
El `Concepto` tiene exactamente 4 secciones de contenido (más la tabla de
metadatos) y la `Guía` exactamente 7 (más la tabla de metadatos), listadas
abajo. Nombres de sección PROHIBIDOS por reincidentes — si aparecen en tu
borrador, bórralos: "Requisitos previos", "Datos de salida esperados",
"Validaciones del sistema", "Nota importante" (como sección aparte —
cualquier aclaración de ese tipo va integrada en el paso que corresponde,
no en una sección nueva).

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
- **2. Procedimiento paso a paso** — pasos numerados, uno por acción real
  de UI, en el orden real de ejecución. Cuando un paso implica cambiar de
  pantalla o tiene sub-acciones, usa sub-numeración `N.1`, `N.2`, ... bajo
  ese paso (no lo aplanes ni lo mezcles con el paso principal). Cuando un
  paso pide capturar datos con varios campos (ej. datos del cliente,
  medio de pago), preséntalos como tabla `Campo | Descripción` en vez de
  prosa. NO agregues marcadores de "Captura sugerida" ni ningún otro
  indicador de screenshot — no es función de esta skill sugerir dónde va
  una imagen; el texto del paso debe bastar por sí solo.
- **3. Confirmación** — el mensaje de éxito real (si el código/BookStack
  lo confirma) y los campos garantizados que muestra el comprobante
  resultante (ver regla dura arriba).
- **4. Acciones disponibles después** — qué puede hacer el usuario con el
  resultado (visualizar, imprimir, descargar, enviar, etc.), solo las que
  el código respalda.
- **5. Consultar esto más adelante** — cómo volver a encontrar el
  resultado (pantalla de búsqueda real, filtros reales disponibles, cómo
  leer el estado desde la misma grilla sin pasos de más).
- **6. Problemas frecuentes** — tabla `Situación | Acción recomendada`,
  solo con casos que el código o BookStack respaldan (mensajes de error
  reales, validaciones reales) — si no hay evidencia de un problema
  frecuente real, no la inventes.
- **7. Flujo resumido** — el procedimiento reducido a una lista corta de
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
