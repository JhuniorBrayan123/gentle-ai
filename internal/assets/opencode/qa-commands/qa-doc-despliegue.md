---
description: Leer, crear o actualizar en BookStack una Especificacion y Despliegue Micro sin eliminar paginas
agent: gentle-orchestrator
---

Load `qa-doc-despliegue` first, then use it for the requested BookStack
operation: "$ARGUMENTS".

This is a standalone QA documentation tool. It may run without an active
`{change}`, never advances the QA ledger, always renders through its fixed
template, asks for explicit confirmation before creating or updating, and
never deletes BookStack content.
