# Reconciliación deadcode de la capa SDD/QA retenida sobre v4

## Contexto

El merge `c5b9a55ec` conserva deliberadamente la capa SDD/QA del fork sobre
upstream v4, cuyo cambio `e219644b2` había retirado SDD. El baseline importado
por el merge era byte por byte el baseline de upstream v4 y, por tanto, no
incluía las superficies retenidas del fork.

En Windows, el ratchet además analizaba el grafo del `GOOS` host mientras el
baseline y CI representan Linux/amd64. Esto producía falsos cambios por build
tags incluso después de normalizar separadores de ruta.

## Decisión

- El ratchet analiza siempre el grafo canónico Linux/amd64, también cuando se
  ejecuta desde Windows o macOS.
- El baseline se regenera una sola vez después de fijar ese target.
- Las entradas añadidas corresponden a APIs y superficies SDD/QA/TUI
  trasplantadas deliberadamente desde v3, cubiertas por tests pero no
  alcanzables estáticamente desde `cmd/gentle-ai` en la composición v4.
- Tres entradas de upstream que ahora son alcanzables o desaparecieron se
  eliminan del baseline.
- Este ajuste no concede permiso para que el inventario crezca en cambios
  posteriores; el ratchet continúa fallando ante cualquier entrada nueva.

## Evidencia

- `bash scripts/test-deadcode-ratchet.sh`
- `bash scripts/deadcode-ratchet.sh --update`
- `bash scripts/deadcode-ratchet.sh`
- Comparación contra:
  - fork v3 pre-merge `4ed4793f2`
  - upstream v4 limpio `4acaf4241`
  - rama de integración `sync/upstream-2026-10`

