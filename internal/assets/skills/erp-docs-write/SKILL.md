---
name: erp-docs-write
description: "Trigger: redactar la ficha de flujo de un módulo ERP2 para cliente final, ancladla al código Screenplay+POM y a BookStack. Produce SOLO un borrador."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "2.0"
---

## Activation Contract

Carga esta skill cuando el humano pida documentación de un flujo o
funcionalidad de ERP2 para el cliente final (ej. "documenta la emisión de
una boleta", "arma la ficha de este flujo"). Es un punto de entrada
**independiente**, NO forma parte del pipeline `qa-supervisor` →
`qa-explore` → `qa-spec` → `qa-apply` → `qa-review` → `qa-verify` →
`qa-docs` (ese pipeline automatiza pruebas Playwright; esta skill redacta
documentación para el cliente que usa el ERP2).

**Un solo documento por flujo — no fragmentes en varias páginas.** Cada
flujo (ej. "emisión de boleta", "emisión de guía de remisión") produce
UNA sola ficha con las secciones de abajo, no un Concepto y una Guía por
separado.

**Audiencia = SIEMPRE cliente final.** No redactes en tono de soporte, QA
o implementación — nada de nombres de clases CSS, IDs de elementos,
códigos internos de triage (ej. "RN-001"), tipos de validación
Frontend/Backend, ni pasos que solo tendría sentido seguir alguien de
soporte. Si un dato solo importa para diagnóstico interno o de
arquitectura (dependencias de módulo/configuración/servicio/tabla BD,
permisos de acceso/RBAC internos), NO entra en la ficha del cliente —
repórtalo aparte como nota interna para el humano, nunca en el cuerpo
publicado.

**Excepción — el plan de suscripción SÍ es información de cliente.** No
confundas "permisos de backend" (interno, se excluye) con "qué plan de
suscripción incluye esta funcionalidad" (PYME / Emprendedor / Negocio /
Empresa / Corporativo) — eso el cliente lo necesita para saber si puede
usar el flujo que está leyendo. Va en la sección "9. Planes relacionados"
(ver "Salida requerida" abajo) — nunca en el encabezado de metadatos.

Produce SIEMPRE un borrador. Nunca escribe en BookStack — esa acción
pertenece exclusivamente a `erp-docs-publish`, y solo corre tras
aprobación humana explícita.

## Fuentes de verdad (MANDATORY)

- **Código Screenplay+POM del proyecto de automatización** —
  `src/screenplay/{tasks,interactions,questions,targets}/**`,
  `src/actors/**`, y también `src/task/**` (convención más antigua del
  mismo proyecto, con Tasks reales como `EmitirGuiaRemitente.task.ts`) —
  fuente de la verdad para cada paso, cada regla de negocio, cada
  validación, cada caso especial y cada dato de entrada/salida que
  aparezca en la ficha. Busca en AMBAS convenciones antes de asumir que un
  flujo "no tiene automatización" — puede estar en la carpeta antigua.
  NUNCA se lee el código fuente del backend/frontend de ERP2 — solo el
  proyecto de automatización.
- **BookStack** (`bookstack_bookstack_search` / `bookstack_bookstack_get_page`)
  — úsalo para: (a) el estándar documental vigente (naming, tags,
  ubicación — libro **"99. Plantillas ERP2 y Estándares"**: `Estándar -
  Cómo documentar`, `Estándar - Convención de títulos`, `Estándar -
  Enlaces cruzados`, `SOP001`), y (b) contexto de negocio complementario
  (el "por qué" detrás de una regla). **BookStack NO es la fuente de
  verdad del comportamiento del sistema — puede estar desactualizado.**
  Si BookStack describe un paso, una regla o un mensaje que el código no
  confirma (o contradice), el código gana: redacta según el código y
  anota la discrepancia como nota interna para el humano — nunca la
  ocultes ni la resuelvas copiando BookStack por default.
  - Nunca decidas solo con la búsqueda si existe o no una página previa —
    sigue la sección "Confirmación de destino" de abajo antes de redactar.
- **Matriz de planes de suscripción — EXCEPCIÓN a "el código manda".** El
  proyecto de automatización NO tiene forma de saber qué plan de
  suscripción (PYME/Emprendedor/Negocio/Empresa/Corporativo) incluye una
  funcionalidad — eso es una decisión comercial, no algo que un test E2E
  valide. Para la sección "9. Planes relacionados", usa este orden:
  1. **Snapshot fijo (fallback rápido, no la fuente que manda)** — matriz
     vigente al 2026-08-31, cada plan incluye todo lo del anterior más lo
     que agrega:
     - **PYME** — S/39/mes: Punto de venta (Facturación Electrónica hasta
       100 comprobantes), notas de venta y cotizaciones ilimitadas,
       gestión de clientes, reportes de ventas y contables, productos/
       servicios/combos, 2 usuarios.
     - **Emprendedor** — S/49/mes: + Facturación Electrónica ilimitada,
       gestión de cajas, ingresos y egresos, control de stock, inventarios
       y kardex, 4 usuarios / 1 caja / 1 almacén.
     - **Negocio** — S/69/mes: + proveedores/conductores/vendedores,
       reportes avanzados, compras/pedidos/listas, tienda virtual, cuentas
       por cobrar y pagar, 10 usuarios / 2 cajas / 2 almacenes / 1
       sucursal extra.
     - **Empresa** — S/99/mes: + tienda virtual avanzada, asistente IA por
       WhatsApp, caja chica financiera, usuarios/cajas/almacenes
       ilimitados, 2 sucursales extras, asesor exclusivo.
     - **Corporativo** — S/149/mes: + emisión masiva por Excel,
       integración API completa, 3 sucursales extras, backup mensual de
       comprobantes.
  2. **Fuente que manda** — la página BookStack "Concepto - Planes de
     suscripción ERP2 SmartClic" (id 3399, libro `erp-cfg-configuracion`,
     capítulo "CGF_Suscripcion, Planes y Pagos"). Si existe diferencia
     entre esa página y el snapshot de arriba, la página gana — el
     snapshot es solo para no depender de una búsqueda en cada ficha.
  3. Si ni el snapshot ni la página cubren la funcionalidad puntual (plan
     nuevo, feature no listada), pregúntale al humano directamente cuál es
     el plan mínimo — NUNCA lo adivines ni lo completes con el plan que
     "suena razonable".
  El snapshot de arriba puede quedar desactualizado con el tiempo — si el
  humano confirma que cambió algún precio/feature, actualízalo tanto en
  esta skill como en la página BookStack 3399, nunca solo en uno de los
  dos lugares.

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
   antes de redactar, propone en qué libro y capítulo iría la ficha,
   basándote en la taxonomía vigente (`SOP001`, `Estándar - Cómo
   documentar`) y en dónde viven las fichas ya existentes de ese mismo
   módulo. **Por defecto, va en el mismo libro donde ya están las demás
   páginas nuevas que ha creado esta skill — no propongas un libro nuevo
   salvo que el humano lo pida.** Pregunta explícitamente:
   > "Propongo publicar la ficha en [libro/capítulo]. ¿Confirmas, o va en
   > otro lugar?"
3. Solo después de que el humano responda, redacta el borrador — con el
   destino ya fijado y guardado junto al borrador en Engram (`crear` +
   libro/capítulo confirmado, o `actualizar` + ID de página confirmado
   por el humano), para que `erp-docs-publish` no tenga que volver a
   adivinarlo ni a preguntar de nuevo.

## Regla dura: nada de pasos, reglas o datos inventados (mecánica obligatoria, no opcional)

**PROHIBIDO redactar cualquier sección de comportamiento del sistema de
memoria/conocimiento general de ERPs, aunque suene plausible.** Esto
aplica a "Flujo principal", "Reglas de negocio", "Validaciones", "Casos
especiales", "Entradas y salidas" y "Errores frecuentes" por igual — el
único método válido es:

0. **Antes de buscar el Task específico del flujo, busca el/los Task(s)
   de entrada compartidos del módulo.** En Punto de Venta, TODO
   comprobante (Boleta, Factura, Cotización, Guía de Remisión, Nota de
   Venta, Nota de Crédito/Débito, etc.) entra por el mismo camino:
   `PuntoVentaSetupPage.navegarANuevaVenta()` ("Ventas y compras" → "Nueva
   venta") y luego la apertura/continuación de caja
   (`CajaPage.abrirCajaCompleta()`: "Aperturar caja" → "Apertura" → "Sí,
   aperturar"; o `continuarVendiendo()`/`ClickContinuarVendiendo`:
   "Continuar vendiendo" si ya estaba abierta). Esos pasos van SIEMPRE
   primero en el Flujo principal, para cualquier tipo de comprobante — no
   son exclusivos de la boleta. **Formato fijo de estos pasos de entrada,
   en el Flujo principal y en el Flujo resumido — siempre estas 3 etapas,
   en este orden y con este texto exacto, nunca menos ni reordenadas:**
   ```
   Ventas y Compras
         ↓
   Nueva Venta
         ↓
   Aperturar / Continuar caja
   ```
   "Aperturar / Continuar caja" es UNA sola etapa (aunque dentro del
   Flujo principal detallado sí se expliquen las dos variantes — caja
   cerrada vs. caja abierta — como sub-pasos de ese mismo paso). Nunca la
   partas en dos etapas separadas del Flujo resumido — "Entrar a Caja" +
   "Continuar vendiendo" como etapas independientes está PROHIBIDO. Nunca
   redactes un paso que salte directo de "entrar al módulo" a
   "seleccionar el tipo de comprobante" sin pasar por la caja, ni que
   omita "Ventas y Compras"/"Nueva Venta" como etapas del Flujo resumido,
   salvo que confirmes en código que ese flujo en particular de verdad no
   pasa por ahí.
1. Localiza el/los archivo(s) de Task/Interaction que implementan la parte
   específica del flujo (busca en `src/task/**`, `src/screenplay/tasks/**`,
   `src/screenplay/interactions/**` por nombre del flujo — ej. para
   "emitir guía de remisión remitente" es
   `EmitirGuiaRemitente.task.ts`/`NavegarAGuiaRemitente.task.ts`, no un
   nombre parecido ni un archivo distinto "que debería existir").
2. Lee el archivo completo y extrae, EN ORDEN, cada `test.step('...')`
   (o Interaction/Task equivalente) que ejecuta. Ese orden es el único
   orden válido para el Flujo principal — no lo reordenes, no lo resumas
   saltándote pasos condicionales, y no fusiones dos `test.step` en uno
   salvo que sean literalmente la misma acción de UI.
3. Cada paso del Flujo principal = un `test.step` real (o un grupo mínimo
   cuando varios `test.step` seguidos son la misma pantalla/acción, ej.
   "abrir menú X" + "seleccionar Y" si son navegación directa a un mismo
   lugar). Si el código tiene una rama condicional (ej. `if
   (data.modalidad === 'PUBLICA')`), la ficha debe reflejar AMBAS ramas
   como alternativas explícitas (no elegir una y omitir la otra).
4. Si un paso del flujo real no tiene una traducción evidente a lenguaje
   de cliente (ej. un campo interno de test), no lo inventes ni lo
   ocultes sin más: pregunta al humano cómo describirlo o dilo
   explícitamente como "[pendiente de confirmar con el equipo]".
5. Antes de entregar el borrador, verifica tú mismo: ¿cada paso escrito
   corresponde a un `test.step` que puedes señalar por nombre de archivo
   y línea? Si no puedes responder eso para algún paso, ese paso está
   inventado — bórralo o reemplázalo por la verificación real.

**La misma mecánica de anclaje aplica al resto de secciones que
describen comportamiento del sistema:**
- **Reglas de negocio**: cada fila sale de una condición/validación real
  en el código (ej. funciones como `motivoAceptaDni`,
  `motivoRequiereComprador` en `EmitirGuiaRemitente.task.ts`), citando el
  archivo como evidencia interna (no se imprime en la ficha, pero debe
  poder mostrarse si el humano lo pide). Si BookStack describe una regla
  que el código no confirma o contradice, prevalece el código — anota la
  discrepancia para el humano, no la escondas.
- **Validaciones**: cada fila sale de un mensaje de error/validación real
  encontrado en Targets/Tasks (ej. `mensajeSinItems`,
  `mensajeNoEncontrado` en `CotizacionTargets`), con su texto exacto.
- **Casos especiales**: solo ramas condicionales reales del código
  (`if`/`switch` sobre motivo, modalidad, tipo de documento, etc.), nunca
  escenarios hipotéticos.
- **Entradas y salidas**: los datos reales que el Task recibe y devuelve
  (ej. `DatosEmisionSimple`, `EmisionResult`), traducidos a lenguaje de
  cliente ("necesitas indicar..." / "vas a obtener..."), no una lista
  genérica de "campos que suelen pedirse".
- **Errores frecuentes**: solo errores con evidencia real en código o
  BookStack — nunca "problemas típicos de un ERP".

Una ficha que omite un paso real del código (ej. seleccionar la caja
antes de poder elegir producto, o completar punto de partida/llegada en
una guía de remisión) o que inventa un paso/regla/validación que el
código no tiene (ej. un campo "Observaciones" que no existe en el Task,
o un "monto mínimo" que no está en ninguna parte) es una violación de
esta regla, sin importar qué tan natural o plausible suene en la
redacción.

Ejemplo real de esta regla aplicada (emisión de boleta, Punto de Venta):
la secuencia "Nueva venta → entra a la caja → recién ahí se puede
seleccionar el producto" viene de una ficha ya publicada, no es una
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

- Título: `[Módulo] - [Submódulo] - [Tema]`, donde:
  - **Módulo** = el ítem de menú real de más alto nivel que el cliente
    hace clic en el ERP2 (ej. "Ventas y Compras", Logística, Facturación,
    Clientes).
  - **Submódulo** = el área funcional dentro de ese módulo a la que
    pertenece el flujo (ej. "Punto de Venta", dentro de "Ventas y
    Compras").
  - **Tema** = el caso puntual que documenta esta ficha (ej. "Agregar un
    combo con validación de stock").
  **NUNCA asumas esta jerarquía de memoria ni por el nombre de carpeta
  del proyecto de automatización — verificala contra la navegación real
  del código** (ej. `page.getByText('Ventas y compras').click()` seguido
  de `page.getByText('Nueva venta').click()` confirma que "Ventas y
  Compras" es el módulo real que el cliente clickea, y "Punto de Venta"
  es el área funcional dentro de él — aunque el proyecto de automatización
  use `PuntoVenta` como nombre de carpeta interno, eso es una convención
  de código, no necesariamente el menú que ve el cliente). Ejemplo
  completo grounded en código: "Ventas y Compras - Punto de Venta -
  Agregar un combo con validación de stock".
- Tags (BookStack): `audiencia: cliente-final`, `tipo-contenido: guia`,
  `modulo: <módulo real>`, `estado: borrador` (pasa a `vigente` solo
  cuando `erp-docs-publish` confirma la publicación), `version-producto:
  <la que indique el estándar vigente>`.
- Ubicación: por defecto, el mismo libro donde ya viven las páginas
  nuevas que ha creado esta skill en este proyecto (ver "Confirmación de
  destino" arriba) — no un libro nuevo, salvo pedido explícito.

## Salida requerida — UNA sola página, lista cerrada de secciones

**No agregues, quites ni renombres ninguna sección de esta lista.** Antes
de entregar el borrador, cuenta las secciones numeradas: si la última no
es exactamente "12. Flujo resumido", coló una sección de más — revisa y
bórrala, no renumeres para que cuadre. Nombres de sección PROHIBIDOS por
reincidentes — si aparecen, bórralos: "Requisitos previos", "Datos de
salida esperados", "Validaciones del sistema", "Nota importante" (como
sección aparte) y "Dependencias" (100% técnica — dependencias de módulo/
configuración/servicio/tabla BD no tienen versión válida para cliente
final; si el humano necesita esa información, va en la nota interna de
Engram, nunca en el cuerpo publicado). **"9. Planes relacionados" NO está
prohibida** — es una sección numerada del cuerpo (no un campo del
encabezado), y sí es información de cliente (ver arriba). No la confundas
con "Dependencias" ni con permisos de backend/RBAC — es exclusivamente
sobre qué plan de suscripción incluye la funcionalidad.

**Reincidencia adicional detectada — nombres inventados que reemplazan la
lista cerrada en vez de sumarse a ella: "Procedimiento paso a paso" (en
vez de "Flujo principal"), "Confirmación" (sección aparte), "Acciones
disponibles después", "Consultar esto más adelante".** Estos NO son
secciones válidas, aunque el contenido sea real y esté bien anclado al
código — el contenido no desaparece, se reubica:
- Lo que iba en "Confirmación" (mensaje de éxito, campos garantizados) es
  el último paso de **"4. Flujo principal"**, o pertenece a **"8. Entradas
  y salidas"** si describe los campos que el sistema devuelve.
- Lo que iba en "Acciones disponibles después" (enviar por WhatsApp,
  descargar PDF, imprimir, etc.) son pasos adicionales de **"4. Flujo
  principal"** si el código las expone como parte del mismo flujo — no
  una sección nueva.
- Lo que iba en "Consultar esto más adelante" (cómo buscar un comprobante
  ya emitido) es un caso de **"7. Casos especiales"** o, si el código lo
  trata como un flujo de consulta separado con su propio Task, va en
  **"2. Alcance"** aclarando que la consulta posterior está fuera o dentro
  del alcance de esta ficha puntual — nunca una sección numerada nueva.

**Checklist final obligatorio (ejecútalo literalmente antes de entregar,
no lo saltees):** lista los títulos de sección que escribiste, en orden,
y compáralos palabra por palabra contra esta lista exacta —
"1. Objetivo", "2. Alcance", "3. Usuarios involucrados", "4. Flujo
principal", "5. Reglas de negocio", "6. Validaciones", "7. Casos
especiales", "8. Entradas y salidas", "9. Planes relacionados",
"10. Errores frecuentes", "11. Documentación relacionada", "12. Flujo
resumido" — y también que el encabezado de metadatos (Estado/
Responsable/Área/Módulo/Submódulo/Tema/Prioridad/Versión/Última
actualización) esté presente completo arriba del todo. Si un solo título no calza exacto,
si falta uno, si sobra uno, o si la numeración salta (ej. de "1." a "3."),
el borrador NO está listo — corrígelo antes de mostrarlo al humano, no lo
entregues "a ver qué te parece" con la lista incompleta.

**Encabezado de metadatos — tabla `Campo | Valor`, NUNCA una cerca de
código.** El título (`# [Módulo] - [Submódulo] - [Tema]`) va como
encabezado Markdown normal. Los campos van en una tabla: una tabla no
sufre el bug de colapso de líneas (cada fila empieza con `|`, el
renderizador la respeta sin necesitar blanco ni cerca) y además se ve
como un cuadro de datos, no como un bloque de código — una cerca de
código aquí es un error de formato, no la solución (esa sí aplica al
"Flujo resumido", que es texto plano sin estructura de tabla). Ni
"Audiencia" ni "Plan requerido" ni "Fuente" van en esta tabla —
"Audiencia" es siempre "Cliente final" (constante, no hace falta
repetirla), "Plan requerido" se documenta en la sección "9. Planes
relacionados" del cuerpo, y "Fuente" (evidencia de código) va solo en la
nota interna de Engram, nunca en el cuerpo publicado.

```
| Campo | Valor |
|---|---|
| Estado | |
| Responsable | |
| Área | |
| Módulo | |
| Submódulo | |
| Tema | |
| Prioridad | |
| Versión | |
| Última actualización | |
```

`Módulo`, `Submódulo` y `Tema` en la tabla repiten exactamente los mismos
tres valores del título (`[Módulo] - [Submódulo] - [Tema]`) — no
redactes una versión distinta en la tabla.

`Responsable` — NO lo dejes en blanco por decisión propia sin preguntar.
Por defecto, es la persona que te pide y aprueba la publicación de esta
ficha en la conversación actual — si te dijo su nombre (como cuando el
humano dice "el responsable soy yo, [nombre]"), usa ese nombre
directamente. Si no lo sabés y no te lo dijeron, preguntá explícitamente
quién es el responsable antes de dejarlo vacío — no asumas que queda
"pendiente" sin más.

`Área` — si no tenés evidencia real ni el humano te lo confirmó, queda en
blanco para que el humano lo llene después — no la inventes.

`Prioridad` — clasificación cerrada de 3 niveles según qué tan central es
el flujo para la operación diaria del negocio; no la dejes en blanco sin
clasificar primero:
- **Alta** — flujo de operación diaria (ventas, facturación, cobros).
- **Media** — flujo frecuente pero no diario (configuración, reportes).
- **Baja** — caso especial o de borde (variante puntual dentro de un
  flujo más grande, ej. un combo sin stock dentro de Punto de Venta).
Si la clasificación no es evidente, preguntale al humano en vez de
adivinar — es una decisión de negocio, igual que "Planes relacionados".

`Versión` — semver `X.Y.Z`, mismo patrón que la skill `gitlab-release-tag`
(léela si necesitás el detalle completo de la mecánica), adaptado a
contenido de documentación en vez de código:
- **Ficha nueva** (destino = crear): siempre arranca en `1.0.0` — nunca
  otro valor de partida, ni preguntes por uno.
- **Ficha existente** (destino = actualizar): leé la `Versión` actual de
  la página que estás actualizando (nunca la inventes ni la calcules de
  memoria — sacala de la tabla de metadatos de la página real) y
  clasificá el cambio que estás por publicar:
  - **PATCH** (`X.Y.Z` → `X.Y.Z+1`) — correcciones que no cambian el
    comportamiento documentado: errores de tipeo, formato, enlaces
    rotos, redacción.
  - **MINOR** (`X.Y.Z` → `X.(Y+1).0`) — contenido nuevo que no invalida
    lo ya documentado: un caso especial nuevo, una fila nueva en una
    tabla, un paso con más detalle, un enlace relacionado nuevo.
  - **MAJOR** (`X.Y.Z` → `(X+1).0.0`) — un cambio de fondo: el
    procedimiento cambió de orden o de comportamiento, el alcance se
    redujo o amplió de forma significativa, o una regla/validación se
    invirtió o se eliminó.
  Mostrale al humano la versión propuesta y el motivo de la clasificación
  ANTES de que `erp-docs-publish` escriba nada — igual que
  `gitlab-release-tag` nunca crea el tag sin confirmación explícita, esta
  skill nunca bumpea la versión sin que el humano la vea primero.

**Las 12 secciones, en este orden exacto:**

1. **Objetivo** — para qué existe el flujo y qué necesidad de negocio
   resuelve, en 1-2 líneas.
2. **Alcance** — Incluye / No incluye, basado en lo que el código
   realmente cubre (ramas implementadas) vs. lo que NO tiene Task/
   Interaction que lo respalde.
3. **Usuarios involucrados** — roles reales mencionados en el código o en
   BookStack (ej. Cajero, Vendedor); si no hay evidencia, márcalo
   pendiente en vez de suponer roles genéricos.
4. **Flujo principal** — pasos numerados, uno por acción real de UI, en
   el orden real de ejecución (ver Regla dura arriba). Cuando un paso
   implica cambiar de pantalla o tiene sub-acciones, usa sub-numeración
   `N.1`, `N.2`, ... bajo ese paso. Cuando un paso pide capturar datos
   con varios campos, preséntalos como tabla `Campo | Descripción` en vez
   de prosa. NO agregues marcadores de "Captura sugerida" ni ningún otro
   indicador de screenshot.

   **Reincidencia del bug de colapso de Markdown, ahora dentro de listas
   numeradas.** El mismo problema que obliga a envolver el "Flujo
   resumido" en una cerca de código (líneas separadas por un solo salto
   se juntan en un mismo párrafo al renderizar) también ocurre **dentro
   de un ítem de lista numerada** cuando el título del paso y su
   instrucción van en líneas seguidas sin blanco entre medio — quedan
   pegados en un solo renglón ("1. Ingresar a Punto de Venta Dirígete al
   menú..."), perdiendo la separación visual. Por eso, cuando un paso
   tenga título + descripción en prosa, van SIEMPRE como dos párrafos
   distintos dentro del mismo ítem — título en negrita, línea en blanco,
   después la descripción (y, si hay alternativas, cada una en su propio
   viñeta, también separada por blanco del párrafo anterior):

   ```
   1. **Ingresar a Punto de Venta**

      Dirígete al menú lateral y selecciona "Ventas y Compras". Haz clic
      en "Nueva Venta".

   2. **Aperturar o continuar con la caja**

      El sistema muestra la pantalla de cajas.

      - Si la caja está cerrada, haz clic en "Aperturar caja", luego en
        "Apertura" y confirma con "Sí, aperturar".
      - Si la caja ya está abierta, haz clic en "Continuar vendiendo".
   ```

   Mal ejemplo (PROHIBIDO): título e instrucción en líneas seguidas sin
   blanco entre medio, aunque visualmente en el Markdown fuente parezcan
   separadas — sin la línea en blanco, BookStack las junta en un renglón.
5. **Reglas de negocio** (tabla `Regla | Descripción`, en lenguaje de
   cliente — sin códigos de triage tipo "RN-001" ni columna de
   prioridad interna) — ver Regla dura arriba.
6. **Validaciones** (tabla `Validación | Mensaje esperado`, sin código ni
   columna Frontend/Backend — eso es interno) — ver Regla dura arriba.
7. **Casos especiales** — ramas condicionales reales del código, en
   lenguaje de cliente.
8. **Entradas y salidas** — Entradas/Salidas reales, traducidas a
   lenguaje de cliente (ver Regla dura arriba).
9. **Planes relacionados** (tabla `Permiso | Descripción`) — qué plan de
   suscripción mínimo incluye esta funcionalidad (ver la fuente de verdad
   específica arriba: matriz de planes o la página BookStack, NUNCA el
   código de automatización). No es una tabla de permisos de backend/RBAC
   — "Permiso" acá significa "qué permite ese plan", no un rol interno.
   Si no hay evidencia de qué plan mínimo aplica, dejá la fila vacía y
   pregúntale al humano en vez de adivinar.
10. **Errores frecuentes** (tabla `Situación | Acción recomendada`) — solo
    con evidencia real; si no hay evidencia de un problema frecuente real,
    no la inventes.
11. **Documentación relacionada** (Opcional) — enlaces a otras fichas
    relacionadas, si existen.
12. **Flujo resumido** — el Flujo principal reducido a una lista corta de
    etapas, en **texto plano** (nunca imagen ni SVG). **Cada etapa y cada
    flecha van en su propia línea, en bloque vertical, y el bloque
    completo va envuelto en una cerca de código Markdown (tres backticks,
    ` ``` `) — NUNCA texto plano suelto sin cerca.** En Markdown, líneas
    separadas por un solo salto de línea SIN cerca de código se colapsan
    en un mismo párrafo al renderizar a HTML; la cerca es la única forma
    de que los saltos se respeten visualmente. Formato exacto a copiar,
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
    renglón, o el bloque SIN la cerca de código. Cada etapa debe
    corresponder 1 a 1 con un paso (o grupo de sub-pasos) real de la
    sección 4, en el mismo orden — nunca inventes ni reordenes etapas
    aquí que no aparecían en el Flujo principal.

## Guardrails de formato (RAG-friendly, obligatorio — ver `SOP001` §7)

- Texto plano y estructurado: un dato por línea, sin HTML decorativo, sin
  cajas de color, sin mezclar estilos dentro de un mismo paso.
- Una ficha = un flujo. No documentes dos flujos distintos en la misma
  página.
- Títulos explícitos con las palabras que el cliente realmente buscaría.
- Sinónimos del negocio cuando ayuden (comprobante/documento,
  aceptado/validado por SUNAT).
- La metadata (estado, audiencia, módulo) va en el encabezado/tags, nunca
  como frase suelta en el cuerpo ("Estado: DRAFT" está PROHIBIDO como
  texto narrativo — usa el campo "Estado" del encabezado).
- **Regla general contra el colapso de Markdown (ya detectada dos veces —
  encabezado de metadatos y Flujo principal):** cualquier bloque donde
  varias líneas deban verse SEPARADAS visualmente (una por dato, una por
  etapa, título separado de su descripción) pero estén unidas solo por un
  salto de línea simple, sin blanco ni cerca de código, se renderiza como
  un único párrafo corrido en BookStack. Antes de entregar el borrador,
  revisa cada bloque de este tipo y asegurate de envolverlo en una cerca
  de código o separar cada parte con una línea en blanco — no asumas que
  "se ve bien en el Markdown fuente" significa que se va a ver bien
  renderizado.

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
- Si detectas contradicción entre BookStack y el código, el código gana
  para todo lo que describe comportamiento del sistema — igual anota la
  discrepancia para el humano en la nota interna, nunca la resuelvas en
  silencio.
- Información puramente técnica/interna (dependencias de
  módulo/configuración/servicio/tabla BD, permisos de backend) que el
  humano pida conservar va SOLO en la nota interna de Engram, nunca en el
  cuerpo publicado de la ficha.

## Persistencia

- Persiste el borrador vía MCP Engram bajo el topic `erp-docs/{flow}/draft`,
  incluyendo la evidencia interna (rutas de código, páginas BookStack
  citadas, y cualquier dato técnico/interno descartado del cuerpo
  publicado) que no va en el cuerpo publicado.

## Comandos de referencia

- Búsqueda/lectura de docs: MCP BookStack (`bookstack_bookstack_search`,
  `bookstack_bookstack_get_page`).
- Persistencia: MCP Engram (`mem_save` topic `erp-docs/{flow}/draft`).
