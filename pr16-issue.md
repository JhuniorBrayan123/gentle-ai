### Pre-flight Checklist

- [x] I have searched the issue tracker for a bug report that matches the one I want to file, without success.
- [x] I understand that this issue must receive `status:approved` before any PR is opened.

### Bug Description

El PR 15 conectó las etapas explore, spec, apply y supervisor al ledger nativo, pero **omitió por completo la etapa verify**.
Actualmente `qa-verify` sigue guardando la evidencia usando un simple `mem_save` en Engram, sin llamar a `qa-begin`, `qa-validate` ni `qa-finish`.

### Steps to Reproduce

1. Correr el agente en un pipeline completo de QA para crear un test.
2. Permitir que `qa-apply` termine y el ledger asigne next_action `verify`.
3. Observar cómo `qa-verify` solo hace un `mem_save` y no cierra la etapa en el ledger.
4. El orquestador o `qa-status` se queda bloqueado esperando.

### Expected Behavior

Al finalizar `qa-apply`, el ledger avanza a `verify`. La habilidad de `qa-verify` DEBE abrir con `qa-begin`, validar el artefacto con `qa-validate` y cerrar exitosamente con `qa-finish` para que el ledger devuelva un status de completitud a `qa-status`.

### Actual Behavior

Al no llamar a los binarios del ledger, el estado interno se queda atascado esperando indefinidamente un `qa-begin` de la etapa `verify`, bloqueando el orquestador.

### Gentle AI Version

latest

### Operating System

Windows

### AI Agent / Client

Gemini CLI

### Affected Area

Agent
