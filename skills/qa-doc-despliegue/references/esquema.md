# Esquema JSON (contrato entre render.py y parse.py)

Obligatorios: `codigo`, `nombre`. Los demás campos son opcionales; los ausentes se renderizan vacíos.

```json
{
  "codigo": "TDL-00000", "version": "1.0", "fecha": "DD/MM/AAAA",
  "creado_por": "", "nombre": "", "validado_por": "", "redmine_url": "",
  "objetivo": ["..."], "solucion": ["..."], "descripcion_usuario": ["..."],
  "peticion": {
    "aviso_comercial":        {"descripcion": "", "respuesta": "SI|NO"},
    "modulo_afectado":        {"descripcion": "", "respuesta": ""},
    "funcionalidad_afectada": {"descripcion": "", "respuesta": ""},
    "componente_afectado":    {"descripcion": "", "respuesta": ""}
  },
  "rf": ["..."], "rnf": ["..."],
  "requiere_qa": true,
  "documentacion": [{"descripcion": "", "link": ""}],
  "despliegue": [{"grupo": "BASE DE DATOS", "pasos": [
    {"componente": ["Scripts", "DBX"], "descripcion": ["texto o pasos"], "link": "", "responsable": "", "tiempo": ""}
  ]}],
  "reversion": [{"grupo": "...", "pasos": []}],
  "tiempo_total": "",
  "implementacion": [{"inicio_fin": "", "corte": "", "observaciones": ""}],
  "notificaciones": [{"cargo": "", "nombre": ""}],
  "responsables": [{"rol": "", "nombre": "", "fecha": ""}],
  "observaciones": ""
}
```

## Campos

- `codigo`, `version`, `fecha`, `creado_por`, `nombre`, `validado_por`: carátula. `nombre` también es el título de la página.
- `redmine_url`: enlace de la incidencia (se muestra como vínculo si empieza por http/https).
- `objetivo`, `solucion`, `descripcion_usuario`: secciones 1 a 3; una viñeta por elemento.
- `peticion`: tabla PETICION/DESCRIPCION/RESPUESTA. En `aviso_comercial.respuesta` solo `SI` o `NO`. Los saltos de línea (`\n`) se conservan.
- `rf`, `rnf`: secciones 4.1 y 4.2. Lista vacía se renderiza "No aplica.".
- `requiere_qa`: `true` marca "SI (X) / NO ( )"; `false` marca "SI ( ) / NO (X)".
- `documentacion`: sección 5.1; `link` puede ser URL o texto.
- Credenciales (5.2): texto fijo, sin campo en el JSON. Cualquier clave tipo password/contraseña/clave hace fallar `render.py`.
- `despliegue`, `reversion`: secciones 6.1 y 6.2, agrupadas por `grupo`. En cada paso, `componente` es una línea por elemento; `descripcion` con más de un elemento se muestra como pasos numerados; `link` va como vínculo. Los pasos se numeran en orden.
- `tiempo_total`: fila "Tiempo Total" de la tabla de reversión.
- `implementacion`, `notificaciones`, `responsables`: secciones 7.1 a 7.3. Lista vacía = una fila vacía.
- `observaciones`: tabla final.

## Notas

- Los textos se normalizan (espacios colapsados); las listas ignoran elementos vacíos.
- `parse.py` lee páginas de plantilla y legacy de Word; ignora la tabla 5.2.
