---
name: gitlab-release-tag
description: Create a semantic version tag (release) for a merged GitLab merge request, following the SmartClic release standard (semver X.Y.Z + changelog referencing the MR). Use when the user merges a feature branch into main/develop and wants the release tag, or asks to "tag a merge", "crear tag", "release", "versionar el merge".
allowed-tools: GitLab MCP (gitlab_list_tags, gitlab_list_merge_requests, gitlab_get_merge_request, gitlab_get_merge_request_diff, gitlab_create_tag)
license: Apache-2.0
metadata:
  author: JhuniorBrayan123
  version: "1.0"
---

Create a GitLab release tag for a merged merge request, following the team standard:

- Tag name = semantic version `X.Y.Z` (read from the most recent existing tag, not from package.json).
- Tag is created on the merge commit (target branch, e.g. `main` or `develop`).
- Tag message = changelog in Markdown that links to the MR and lists changes, exactly like the SmartClic standard.

## Release note format (tag message)

```markdown
## [2.133.1](https://gitlab.sreasons.com/<group>/<project>/-/merge_requests/<iid>) - DD/MM/YYYY

### Nuevas características
- <feat items, or "No Aplica">

### Incidencias solucionadas
- <fix items>
```

## Semver rules (confirmed with the team)

| Type of change | Commit prefix | Bump |
|---|---|---|
| Bug fix (backwards compatible) | `fix(...)` | PATCH: `2.133.1` → `2.133.2` |
| New feature (backwards compatible) | `feat(...)` | MINOR: `2.133.1` → `2.134.0` |
| Breaking change | `feat!`, `fix!`, `BREAKING CHANGE` | MAJOR: `2.133.1` → `3.0.0` |
| Docs / refactor / chore / test (no user-visible change) | `docs`, `refactor`, `chore`, `test`, `perf` | No bump by itself; contributes to the changelog only if grouped with a feat/fix, otherwise suggest "No Aplica" |

- Read the current version from the **most recent semver tag** (`gitlab_list_tags`), never from package.json.
- If no semver tags exist, ask the user for the starting version (e.g. `1.0.0`).
- MINOR/MAJOR bumps reset lower segments to 0 (e.g. `2.133.1` + feat → `2.134.0`, + breaking → `3.0.0`).

## Workflow

### 1. Identify the MR

- If the user names the MR (by URL or IID), use `gitlab_get_merge_request(project_path, mr_iid)`.
- Otherwise list recent merged MRs: `gitlab_list_merge_requests(project_path, state="merged")` and pick the most recent one merged into the target branch (`main` or `develop`).
- The skill triggers after a merge, so the MR `state` MUST be `merged`. If not merged yet, stop and tell the user to merge first.

### 2. Determine the next version

- Call `gitlab_list_tags(project_path)`.
- Extract the highest semver tag (`X.Y.Z`); parse `major.minor.patch` numerically (not string sort).
- Call `gitlab_get_merge_request_diff(project_path, mr_iid)` to inspect the changes.
- Classify the changes:
  - Any `feat!` / `fix!` / `BREAKING CHANGE` → **MAJOR** (sube `X`, resetea `Y` y `Z`).
  - Any `feat(...)` (no breaking) → **MINOR** (sube `Y`, resetea `Z`).
  - Only `fix(...)` → **PATCH** (sube `Z`).
  - Mixed fix+feat (no breaking) → **MINOR**.
- Show the user the proposed version and the reasoning before creating anything.

### 3. Build the changelog

- List commits/items from the MR diff grouped into:
  - **Nuevas características** → `feat` items.
  - **Incidencias solucionadas** → `fix` items.
- If a group has no items, write `- No Aplica`.
- Date format: `DD/MM/YYYY` (use the merge date from the MR).
- MR URL: `https://gitlab.sreasons.com/<project_path>/-/merge_requests/<iid>`.

### 4. Create the tag

- Call `gitlab_create_tag(project_path, tag_name, ref, message)`:
  - `tag_name`: the new `X.Y.Z`.
  - `ref`: the **target branch** of the MR (`main` or `develop`) — the tag lands on the merge commit.
  - `message`: the full changelog Markdown (multi-line → GitLab also creates a release note).
- Verify the result (`success: true`) and report the tag name + commit.

## Guardrails

- **Never** create a tag on a non-merged MR — confirm `state == "merged"` first.
- **Never** guess the version when tags are missing — ask the user.
- **Never** force-overwrite an existing tag; if `tag_name` already exists, bump again and inform the user.
- **Always** show the proposed version + changelog and get explicit confirmation from the user before calling `gitlab_create_tag` (it is a write action).
- If the MR diff is empty or ambiguous, ask the user how to classify the bump.
- Tags apply to GitLab only; syncing to the GitHub mirror (if any) is out of scope unless the user asks.
- Keep the exact release-note format shown above; do not invent new sections.

## Example

User: "mergeé feature-19061 a develop, sacá el tag"

1. `gitlab_get_merge_request("grupo/erp-x", 309)` → state `merged`, target `develop`.
2. `gitlab_list_tags("grupo/erp-x")` → last `2.133.1`.
3. Diff shows only `fix(selector)` and `fix(altura)` → **PATCH** → propose `2.133.2`.
4. Changelog:

```markdown
## [2.133.2](https://gitlab.sreasons.com/grupo/erp-x/-/merge_requests/309) - 05/08/2026

### Nuevas características
- No Aplica

### Incidencias solucionadas
- Corrección de selectores no adecuados por implementación de slider
- Corrección de altura de cards en lg para que al manejar 3 líneas de texto aplique bien
```

5. Confirm with user, then `gitlab_create_tag("grupo/erp-x", "2.133.2", "develop", <message>)`.