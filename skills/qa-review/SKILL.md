---
name: qa-review
description: "Trigger: revisar adversarialmente un cambio QA. Delega al motor de review nativo de gentle-ai; no reimplementa la revisión."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "2.0"
disable-model-invocation: true
user-invocable: false
---

## Activation Contract

Carga esta skill SOLO para enrutar una revisión QA al motor nativo. No es un stage del flujo qa-*.

## Hard Rules

- NUNCA ejecutes una revisión adversarial propia: el motor nativo es la única autoridad de review en este repo.
- NUNCA emitas un veredicto PASS/allow por tu cuenta ni inventes un recibo.
- Las reglas G1-G6 no se revisan aquí; su checklist de evidencia vive en `qa-verify`.

## Execution Steps

1. Ejecuta `gentle-ai review status --cwd <repo> --contract gentle-ai.review-integration/v2 --agent {{GENTLE_AI_RUNTIME_AGENT_ID}} --next-transition`.
2. Enruta ÚNICAMENTE desde el `next_transition` devuelto (`execute` / `collect` / `stop`). Nunca desde la prosa del status.
3. En `stop`, entrega el `reason_code` y su continuación documentada; detente.

## Output Contract

Devuelve el `next_transition` textual y el resultado de la operación ejecutada. Nada más.

## References

- `skills/_shared/qa-gate-policy.md` — reglas G1-G6 (fuente única in-repo).
