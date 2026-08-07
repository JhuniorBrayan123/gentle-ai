---
name: gitlab-mr-flow
description: "Trigger: crear MR, mergear MR, merge a main, feature branch a main, crear merge request, aprobar MR, versionar merge. Orquesta un MR de GitLab de punta a punta: push, crear (con assigner+reviewer configurables), verificar aprobación obligatoria antes del merge, mergear por API y delegar a gitlab-release-tag."
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

Crear y mergear un Merge Request de GitLab de punta a punta, siguiendo el estándar del equipo QA (GU-CalidadTI).

## Configurables (por defecto)

- `MR_ASSIGNEE` (default `jgutierrez`): asignado del MR.
- `MR_REVIEWER` (default `jgutierrez`): revisor del MR.
- Si el usuario da otro username en el prompt, úsalo en lugar del default para ese MR.

## Gates (NO opcionales)

1. **Antes de crear**: confirmar con el usuario `source_branch`, `target_branch` (default `main`) y `title`.
2. **Antes de mergear**: el MR DEBE estar aprobado. Sin aprobación NO se mergea.
   - Leer aprobación con `gitlab_get_merge_request_approvals(project_path, mr_iid)`.
   - Si `approved == false` → NO mergear. Aprobar con `gitlab_approve_merge_request` (si el usuario lo aprueba) y re-verificar, o pedir al usuario que apruebe desde la UI y esperar confirmación.
   - Solo cuando `approved == true` se llama `gitlab_merge_merge_request`.
3. **Mergea por API**: el MCP ya expone el tool `gitlab_merge_merge_request` (PUT `/merge`). No asumas que no existe. Si el merge devuelve 404, revisa que el server.py del MCP use `api_put` (GitLab exige PUT en `/merge`, no POST).

## Flujo

```
1. git push origin <source>   (si no está ya en remoto)
2. CREAR: gitlab_create_merge_request(project_path, source, target, title,
     reviewer_usernames=[MR_REVIEWER], assignee_username=MR_ASSIGNEE)
   -> retorna iid + web_url.
3. Si el create no persistió assignee/reviewer (fallback), usar
   gitlab_update_merge_request(project_path, iid, reviewer_usernames, assignee_username).
4. GATE DE APROBACIÓN:
   gitlab_get_merge_requests_approvals(project_path, iid)
   - approved == true  -> ok, a mergear.
   - approved == false -> gitlab_approve_merge_request(project_path, iid)  [solo con OK
     del usuario] y re-verificar con gitlab_get_merge_requests_approvals.
5. MERGE: gitlab_merge_merge_request(project_path, iid).
6. Tag: delegar a la skill GITLAB-RELEASE-TAG (crea el tag semver + changelog).
```

## Guardrails

- **Nunca** mergear sin `<approved == true>`. Es la regla irrompible del equipo.
- **Nunca** crear el MR sin confirmar origen/destino/título con el usuario.
- El merge y el tag son ACCIONES DE ESCRITURA: ejecutarlas solo con confirmación
  explícita del usuario para ese MR puntual.
- `project_path` es el path_with_namespace completo, ej `GU-CalidadTI/erpperu2-automation`.
- Si el MR tiene conflictos (`has_conflicts`) o un pipeline fallido, detente y avisa antes de mergear.
- El username es configurable: usar el que dé el usuario, default `jgutierrez`.

## Output Contract

Al terminar reportar:
- `web_url` del MR creado.
- Estado de aprobación antes y después del gate.
- URL del MR o confirmación del merge (success + state).
- Tag creado (via gitlab-release-tag).