---
name: erp-docs-write
description: "Trigger: redactar Concepto, Guía paso a paso o ficha Funcional de un flujo ERP2, con BookStack y código Screenplay+POM. Produce SOLO un borrador."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

## Activation Contract

Carga esta skill cuando el humano pida documentación de un flujo o
funcionalidad de ERP2 — para cliente final (ej. "documenta la emisión de
una boleta", "redacta la guía de este flujo") o para uso interno (ej.
"arma la ficha funcional de Guías de Remisión"). Es un punto de entrada
**independiente**,
NO forma parte del pipeline `qa-supervisor` → `qa-explore` → `qa-spec` →
`qa-apply` → `qa-review` → `qa-verify` → `qa-docs` (ese pipeline automatiza
pruebas Playwright; esta skill redacta documentación para el cliente que usa
el ERP2).

**La audiencia depende del tipo de documento pedido:**
- `Concepto` y `Guía` son SIEMPRE para cliente final. No redactes en tono
  de soporte, QA o implementación — nada de nombres de clases CSS, IDs de
  elementos, jerga técnica interna, ni pasos que solo tendría sentido
  seguir alguien de soporte. Si una validación solo importa para
  diagnóstico interno, no entra en el borrador cliente; repórtala aparte
  como nota para el humano.
- `Funcional` es para audiencia interna (analistas, QA, soporte, equipo de
  desarrollo) — sí puede usar vocabulario técnico y de negocio (reglas,
  validaciones, dependencias, tablas de BD si se conocen). Aun así, nunca
  incluyas selectores/IDs de test ni detalles de implementación del
  proyecto de automatización en sí (eso pertenece al código, no a la
  ficha funcional).

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
  - **Excepción para `Funcional`: BookStack NO es la fuente de verdad del
    contenido.** Para Concepto y Guía, BookStack manda en definiciones de
    negocio. Para Funcional, el código manda — BookStack puede estar
    desactualizado respecto al comportamiento real del sistema. Usa
    BookStack solo para citar contexto/historia y para la taxonomía de
    naming, nunca para completar una sección de la ficha funcional
    (Flujo principal, Reglas de negocio, Validaciones, Casos especiales,
    Entradas y salidas, Errores frecuentes) cuando el código dice algo
    distinto o cuando BookStack simplemente no tiene el dato — en ese
    caso, confía en el código o márcalo "[pendiente de confirmar con el
    equipo]", nunca copies BookStack sin más por default.
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

**Regla dura: la skill NUNCA decide sola si crea o actualiza.** Puede
hacer ambas cosas — pero solo la que el humano confirme explícitamente
para esta página en concreto. Por defecto, ante la duda, propone crear
página nueva (no actualizar) — actualizar es la excepción explícita, no
la opción por defecto. Publicar en el libro/capítulo equivocado tampoco
se corrige editando el contenido después — hay que evitarlo antes de
escribir una sola línea. Por eso, antes de producir el borrador, la skill
DEBE presentar al humano lo siguiente y esperar su confirmación:

1. **Búsqueda de páginas existentes**: ejecuta `bookstack_bookstack_search`
   con el nombre del flujo/término y sinónimos de negocio razonables (ej.
   "boleta", "comprobante", "emisión de boleta"). Lista TODOS los
   resultados relevantes encontrados (título, ID, libro/capítulo, URL) —
   incluso coincidencias parciales o dudosas — y pregunta explícitamente:
   > "Encontré estas páginas relacionadas: [lista]. ¿Alguna es la que
   > debo actualizar (dime el ID exacto), o es contenido nuevo y creo una
   > página aparte?"
   Si el humano confirma una página puntual (por ID) para actualizar,
   usa esa página como destino — nunca actualices "la que más se parece"
   sin que el humano haya señalado el ID exacto. Si el humano no elige
   ninguna o no hay resultados, el destino es crear página nueva. Si la
   búsqueda no encuentra nada, dilo explícitamente ("no encontré ninguna
   página existente para este flujo") en vez de asumir silenciosamente
   que no existe.
2. **Ubicación propuesta** (solo aplica si el destino es página nueva):
   antes de redactar, propone en qué libro y capítulo iría cada página
   (`Concepto -`, `Guía -` y/o `Funcional`), basándote en la taxonomía
   vigente (`SOP001`, `Estándar - Cómo documentar`) y en dónde viven las
   páginas `Guía -`/`Concepto -`/`Funcional` ya existentes de ese mismo
   módulo. **Por defecto, `Funcional` va en el mismo libro donde ya están
   las demás páginas nuevas que ha creado esta skill (Concepto/Guía) —
   no propongas un libro nuevo salvo que el humano lo pida.** Pregunta
   explícitamente:
   > "Propongo publicar el Concepto en [libro/capítulo], la Guía en
   > [libro/capítulo] y el Funcional en [libro/capítulo]. ¿Confirmas, o
   > va en otro lugar?"
3. Solo después de que el humano responda, redacta el borrador — con el
   destino ya fijado y guardado junto al borrador en Engram (`crear` +
   libro/capítulo confirmado, o `actualizar` + ID de página confirmado
   por el humano), para que `erp-docs-publish` no tenga que volver a
   adivinarlo ni a preguntar de nuevo.

## Regla dura: nada de pasos o datos inventados (mecánica obligatoria, no opcional)

**PROHIBIDO redactar el procedimiento de memoria/conocimiento general de
ERPs, aunque suene plausible.** Esta regla aplica igual al "Procedimiento
paso a paso" de la Guía y al "Flujo principal" de la ficha Funcional — en
ambos casos el único método válido es:

0. **Antes de buscar el Task específico del documento, busca el/los Task(s)
   de entrada compartidos del módulo.** En Punto de Venta, TODO comprobante
   (Boleta, Factura, Cotización, Guía de Remisión, Nota de Venta, Nota de
   Crédito/Débito, etc.) entra por el mismo camino:
   `PuntoVentaSetupPage.navegarANuevaVenta()` ("Ventas y compras" → "Nueva
   venta") y luego la apertura/continuación de caja
   (`CajaPage.abrirCajaCompleta()`: "Aperturar caja" → "Apertura" → "Sí,
   aperturar"; o `continuarVendiendo()`/`ClickContinuarVendiendo`:
   "Continuar vendiendo" si ya estaba abierta). Esos pasos van SIEMPRE
   primero en el procedimiento, para cualquier tipo de comprobante — no son
   exclusivos de la guía de Boleta. **Formato fijo de estos pasos de
   entrada, en el Procedimiento paso a paso y en el Flujo resumido —
   siempre estas 3 etapas, en este orden y con este texto exacto, nunca
   menos ni reordenadas:**
   ```
   Ventas y Compras
         ↓
   Nueva Venta
         ↓
   Aperturar / Continuar caja
   ```
   "Aperturar / Continuar caja" es UNA sola etapa (aunque dentro del
   Procedimiento detallado sí se expliquen las dos variantes — caja
   cerrada vs. caja abierta — como sub-pasos de ese mismo paso). Nunca la
   partas en dos etapas separadas del Flujo resumido — "Entrar a Caja" +
   "Continuar vendiendo" como etapas independientes está PROHIBIDO. Nunca
   redactes un paso que salte directo de "entrar al módulo" a
   "seleccionar el tipo de comprobante" sin pasar por la caja, ni que
   omita "Ventas y Compras"/"Nueva Venta" como etapas del Flujo resumido,
   salvo que confirmes en código que ese flujo en particular de verdad no
   pasa por ahí.
1. Localiza el/los archivo(s) de Task/Interaction que implementan la parte
   específica del documento (busca en `src/task/**`,
   `src/screenplay/tasks/**`, `src/screenplay/interactions/**` por nombre
   del flujo — ej. para "emitir guía de remisión remitente" es
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

**Esta misma mecánica de anclaje aplica al resto de secciones del
Funcional que describen comportamiento del sistema** — no solo al Flujo
principal:
- **Reglas de negocio**: cada fila sale de una condición/validación real
  en el código (ej. funciones como `motivoAceptaDni`,
  `motivoRequiereComprador` en `EmitirGuiaRemitente.task.ts`), citando el
  archivo como evidencia interna. Si BookStack describe una regla que el
  código no confirma (o la contradice), prevalece el código — anota la
  discrepancia para el humano, no la ocultes ni la promedies.
- **Validaciones**: cada fila sale de un mensaje de error/validación real
  encontrado en Targets/Tasks (ej. `mensajeSinItems`,
  `mensajeNoEncontrado` en `CotizacionTargets`), con su texto exacto.
- **Casos especiales**: solo ramas condicionales reales del código
  (`if`/`switch` sobre motivo, modalidad, tipo de documento, etc.).
- **Entradas y salidas**: los tipos de datos reales que el Task recibe y
  devuelve (ej. `DatosEmisionSimple`, `EmisionResult`), no una lista
  genérica de "campos que suelen pedirse".
- **Errores frecuentes (Fact)**: solo errores que el código o BookStack
  respaldan con evidencia concreta — nunca "problemas típicos de un ERP".
- **Dependencias** (Módulo/Configuración/Servicio/Tabla BD) y **Planes
  relacionados** (permisos): el proyecto de automatización normalmente NO
  expone tablas de BD ni permisos de backend — si no puedes confirmarlo
  desde el código de automatización ni desde BookStack, dilo
  explícitamente como "[pendiente de confirmar con el equipo]" en vez de
  inventar una tabla o un permiso que suene razonable.

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
- `[Módulo] - [Submódulo] - [Tema]` — título de la ficha `Funcional`, tal
  cual la plantilla oficial (sin prefijo "Funcional -"; el tipo de página
  queda marcado por la tag `tipo-contenido: funcional`, no por el título).
- Tags (BookStack): `audiencia: cliente-final` para Concepto/Guía,
  `audiencia: interno` para Funcional; `tipo-contenido:
  concepto|guia|funcional`, `modulo: <módulo real>`, `estado: borrador`
  (pasa a `vigente` solo cuando `erp-docs-publish` confirma la
  publicación), `version-producto: <la que indique el estándar vigente>`.
- Ubicación: `Concepto -` va en el libro/capítulo de producto del módulo
  (ej. glosario de facturación electrónica); `Guía -` va en el libro/
  capítulo de operación diaria del módulo (ej. el mismo capítulo donde ya
  viven las guías de ese módulo); `Funcional` va, por defecto, en el
  mismo libro donde ya viven las páginas `Concepto -`/`Guía -` nuevas de
  este proyecto (ver "Confirmación de destino" arriba) — no un libro
  nuevo, salvo pedido explícito del humano.

## Salida requerida — hasta TRES páginas, nunca mezcladas

**Lista cerrada de secciones — no agregues, quites ni renombres ninguna.**
El `Concepto` tiene exactamente 4 secciones de contenido (más la tabla de
metadatos), la `Guía` exactamente estas 7 (más la tabla de metadatos) — la
última SIEMPRE es la 7, "Flujo resumido" — y el `Funcional` exactamente
las 12 de su plantilla oficial (ver más abajo, sección 3):

```
1. Objetivo
2. Procedimiento paso a paso
3. Confirmación
4. Acciones disponibles después
5. Consultar esto más adelante
6. Problemas frecuentes
7. Flujo resumido
```

Antes de entregar el borrador, cuenta las secciones numeradas de la Guía:
si la última no es exactamente "7. Flujo resumido" (ej. si llegaste a
"8. Flujo resumido"), coló una sección de más — revisa y bórrala, no
renumeres para que cuadre. Nombres de sección PROHIBIDOS por reincidentes
— si aparecen en tu borrador, bórralos: "Requisitos previos", "Datos de
salida esperados", "Validaciones del sistema", "Nota importante" (como
sección aparte — cualquier aclaración de ese tipo va integrada en el paso
que corresponde, no en una sección nueva).

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
  etapas, en **texto plano** (nunca imagen ni SVG). **Cada etapa y cada
  flecha van en su propia línea, en bloque vertical — NUNCA todo en una
  sola línea/párrafo separado por flechas inline.**

  **OBLIGATORIO: envuelve el bloque completo en una cerca de código
  Markdown (tres backticks, ` ``` `), no lo dejes como texto plano
  suelto.** En Markdown, líneas separadas por un solo salto de línea SIN
  una cerca de código se colapsan en un mismo párrafo al renderizar a
  HTML — es la causa exacta por la que un "Flujo resumido" bien escrito
  en la fuente terminó mostrándose todo en una sola línea en BookStack.
  La cerca de código (` ``` `) es la única forma de garantizar que los
  saltos de línea se respeten visualmente. Formato exacto a copiar,
  backticks incluidos:

  ````
  ```
  Etapa 1
        ↓
  Etapa 2
        ↓
  Etapa 3
  ```
  ````

  Mal ejemplo (PROHIBIDO): `Etapa 1 ↓ Etapa 2 ↓ Etapa 3` en un mismo
  renglón, o el bloque de etapas SIN la cerca de código — ambos se
  renderizan como una sola línea corrida en BookStack aunque la fuente
  tenga saltos de línea. Antes de entregar el borrador, verifica que el
  bloque de Flujo resumido esté envuelto en ` ``` `. Cada etapa del
  resumen debe corresponder 1 a 1 con un paso (o grupo de sub-pasos) real
  de la sección 2, en el mismo orden — nunca inventes ni reordenes
  etapas aquí que no aparecían en el procedimiento detallado.

### 3. `[Módulo] - [Submódulo] - [Tema]` (Funcional)

Documenta una funcionalidad, flujo o submódulo completo para audiencia
interna (analistas, QA, soporte, desarrollo). Se activa cuando el humano
pide explícitamente una "ficha funcional", "plantilla funcional" o
documentación técnica/de negocio de un módulo — no reemplaza a Concepto
ni a Guía, es un tercer tipo de página independiente.

**Encabezado de metadatos (formato exacto de la plantilla oficial, NO una
tabla — campos "Campo: valor" en texto plano, uno por línea):**

```
# [Módulo] - [Submódulo] - [Tema]
Estado:
Responsable:
Área:
Módulo:
Submódulo:
Prioridad:
Versión:
Última actualización:
Fuente:
```

`Fuente` cita el/los archivo(s) de código de donde salió el contenido
(ej. `EmitirGuiaRemitente.task.ts`) — es la evidencia de que esto no se
inventó. Campos que no puedas completar con evidencia real (Responsable,
Prioridad, Versión) quedan en blanco para que el humano los llene — no
los inventes.

**Las 12 secciones, en este orden exacto, ni una más ni una menos:**

1. **Objetivo** — para qué existe la funcionalidad y qué necesidad de
   negocio resuelve. Puede citar BookStack para el "por qué" de negocio,
   pero el "qué hace" debe poder verificarse en el código.
2. **Alcance** — Incluye / No incluye, basado en lo que el código
   realmente cubre (ramas implementadas) vs. lo que NO tiene Task/
   Interaction que lo respalde.
3. **Usuarios involucrados** — roles reales mencionados en el código o
   en BookStack (ej. Cajero, Vendedor); si no hay evidencia, márcalo
   pendiente en vez de suponer roles genéricos.
4. **Flujo principal** — igual mecánica que el "Procedimiento paso a
   paso" de la Guía (ver Regla dura arriba): pasos numerados = secuencia
   real de `test.step`/Task/Interaction, incluyendo la entrada compartida
   de Punto de Venta cuando aplique.
5. **Reglas de negocio** (tabla `Código | Regla | Descripción |
   Prioridad`) — cada fila anclada a una condición real del código (ver
   Regla dura arriba). Si BookStack describe una regla que el código
   contradice, el código gana — anota la discrepancia, no la escondas.
6. **Validaciones** (tabla `Código | Validación | Mensaje esperado |
   Tipo`) — mensajes de error/validación reales encontrados en el
   código, con su texto exacto. `Tipo` = Frontend si es una validación de
   UI/cliente (la que ve el proyecto de automatización); marca "[por
   confirmar]" si no puedes saber si también valida en Backend.
7. **Casos especiales** — ramas condicionales reales del código, no
   escenarios hipotéticos.
8. **Dependencias** (tabla `Tipo | Dependencia | Descripción`) — Módulo y
   Servicio pueden salir del código (imports, endpoints llamados);
   Configuración y Tabla BD normalmente NO son visibles desde el proyecto
   de automatización — márcalas "[pendiente de confirmar con el equipo]"
   en vez de adivinar.
9. **Planes relacionados** (tabla `Permiso | Descripción`) — solo si hay
   evidencia real (código o BookStack); si no, la tabla queda con la fila
   marcada "[pendiente de confirmar con el equipo]", nunca vacía sin
   explicación ni inventada.
10. **Entradas y salidas** — Entradas/Salidas reales según los tipos de
    datos que el Task recibe y devuelve (ej. `DatosEmisionSimple`,
    `EmisionResult`), no una lista genérica de campos típicos de un ERP.
11. **Errores frecuentes (Fact)** (tabla `Error | Causa | Solución`) —
    solo errores con evidencia real en código o BookStack.
12. **Documentación relacionada (Opcional)** — enlaces a Conceptos,
    Guías u otras fichas Funcionales relacionadas, si existen.

**Formato**: igual que Concepto/Guía — texto plano y estructurado, sin
HTML decorativo. Las tablas van en Markdown estándar (no code fences,
esas se reservan para el Flujo resumido de la Guía).

## Guardrails de formato (RAG-friendly, obligatorio — ver `SOP001` §7)

- Texto plano y estructurado: un dato por línea, sin HTML decorativo, sin
  cajas de color, sin mezclar estilos dentro de un mismo paso.
- Una página = una intención (Concepto, Guía y Funcional siempre
  separados, nunca fusionados en una sola página).
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
- Si detectas contradicción entre BookStack y el código en Concepto o
  Guía, no decidas por tu cuenta: documenta la contradicción y escala al
  humano. **Excepción para Funcional**: ahí el código es la fuente de
  verdad (BookStack puede estar desactualizado) — redacta según el
  código y AÚN ASÍ anota la discrepancia con BookStack para el humano,
  no la ocultes ni la resuelvas en silencio.

## Persistencia

- Persiste el/los borrador(es) que hayas producido (Concepto, Guía y/o
  Funcional, según lo pedido) vía MCP Engram bajo el topic
  `erp-docs/{flow}/draft`, incluyendo la evidencia interna (rutas de
  código, páginas BookStack citadas) que no va en el cuerpo publicado.

## Comandos de referencia

- Búsqueda/lectura de docs: MCP BookStack (`bookstack_bookstack_search`,
  `bookstack_bookstack_get_page`).
- Persistencia: MCP Engram (`mem_save` topic `erp-docs/{flow}/draft`).
