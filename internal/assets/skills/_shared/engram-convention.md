# Engram Convention (reference documentation)

Critical Engram calls (`mem_search`, `mem_save`, `mem_get_observation`) belong in the calling skill's own contract. This reference explains ODD task recovery; do not use a search preview as the full task document.

## ODD task naming and recovery

For a substantial authorized feature, the local task document is `odd/tasks/{feature-name}.md` and its project-scoped Engram mirror has topic key `odd/{feature-name}/tasks`. The mirror contains the full document and repository-relative locator, not a summary. Use a stable filename-safe feature name; never overwrite another feature. Both writes must be read back because they are not atomic.

To resume, obtain session context, search the current project and feature for the task topic, retrieve its full observation with `mem_get_observation`, then read the local task file. Reconcile before continuing. Preserve both versions of irreconcilable edits and ask only about the actual conflict. If Engram is unavailable, keep local progress and mark the mirror pending; resynchronize when available. Never treat a stale mirror as authority over observed code or tests.

Memory lifecycle rule (when Engram exposes lifecycle metadata/tooling):
- At session start or before architecture-sensitive work, call `mem_review` with action `list` for the current project when the tool is available.
- If `mem_review` is unavailable, do not fail the task. Continue with normal `mem_context`/`mem_search`, and still apply lifecycle metadata from any returned observations when present.
- `active` memories may be used normally.
- `needs_review` memories are stale context, not trusted facts.
- Surface `needs_review` context and verify it against current evidence before relying on it.
- Do NOT call `mem_review` with action `mark_reviewed` automatically. Only call `mark_reviewed` after explicit user confirmation or through a dedicated memory maintenance command.

## Writing and updating

Use the resolved project name for project-scoped memory. A `mem_save` with the same `topic_key`, project and scope is an upsert: it overwrites previous content, not an audit trail. Read the current observation before updating an evolving topic and preserve unrelated progress. Use `mem_update` when the exact observation ID is available. For the automated task-document mirror, set `capture_prompt: false` when supported; for normal human-directed memory saves, retain the normal capture behavior. If the tool does not support that field, omit it rather than failing. Never store secrets or user data in a task mirror.

After each verified task transition, update the local document and its full Engram mirror, then read both back. Record actual checks and work-unit evidence, including failures and pending mirrors; do not claim a successful mirror when its write failed.

## SDD Artifact Convention

This section is owned by the fork's SDD layer and applies to the SDD workflow only. It adds to the ODD task-recovery contract above; it never replaces it.

ALL SDD artifacts persisted to Engram follow this deterministic naming:

```
title:     sdd/{change-name}/{artifact-type}
topic_key: sdd/{change-name}/{artifact-type}
type:      architecture
project:   {detected or current project name}
scope:     project
capture_prompt: false
```

Set `capture_prompt: false` when the Engram tool schema supports it; if an older schema rejects or does not expose the field, omit it rather than failing.

### Artifact types

| Artifact Type | Produced By | Description |
|---------------|-------------|-------------|
| `explore` | sdd-explore | Exploration analysis |
| `proposal` | sdd-propose | Change proposal |
| `spec` | sdd-spec | Delta specifications (all domains concatenated) |
| `design` | sdd-design | Technical design |
| `tasks` | sdd-tasks | Task breakdown |
| `apply-progress` | sdd-apply | Implementation progress (one per batch) |
| `verify-report` | sdd-verify | Verification report |
| `archive-report` | sdd-archive | Archive closure with lineage |
| `state` | orchestrator | Optional recovery hint; actual artifacts remain authoritative |

### Research artifacts

Use `sdd/{change-name}/research` for optional source-backed notes when persistence is requested. Preserve historical research and preproposal observations. Their schema, revision or agreement with files is not proposal admission authority.

### Optional state hint

An existing `sdd/{change-name}/state` observation is an optional recovery hint, not a required YAML snapshot or a second authority. Recover using native status and its resolved artifact locators; retrieve full observations with `mem_get_observation`. Preserve historical snapshots, but verify their claims against actual artifacts. A state-only observation does not establish active work; the actual `archive-report` remains the closure marker.

### Recovery protocol

```
Step 1: mem_search(query: "sdd/{change-name}/{artifact-type}", project: "{project}") → truncated preview + ID
Step 2: mem_get_observation(id: {observation-id}) → complete content
```

When retrieving multiple artifacts, group all searches first, then all retrievals, so every dependency is fetched from its full observation rather than a truncated preview.

### Writing artifacts

Standard write:

```
mem_save(
  title: "sdd/{change-name}/{artifact-type}",
  topic_key: "sdd/{change-name}/{artifact-type}",
  type: "architecture",
  project: "{project}",
  capture_prompt: false,
  content: "{full markdown content}"
)
```

Update an existing artifact with `mem_update(id: {observation-id}, content: "{updated full content}")` when the exact observation ID is known; otherwise `mem_save` with the same `topic_key` upserts.

`capture_prompt: false` is REQUIRED for SDD artifacts when the Engram tool schema supports it. Engram v1.15.3 captures user prompts by default for human/proactive saves, but SDD artifacts are automated pipeline outputs. Do not infer this from `type` because both SDD artifacts and human architecture decisions use `architecture`. If an older schema rejects or does not expose `capture_prompt`, omit it rather than failing.

### Upsert behavior

Same `topic_key` + `project` + `scope` means UPDATE (overwrite), not INSERT. Previous content is lost; `revision_count` increments but old content is not kept. This is by design: Engram is working memory, not an audit trail. For iteration history or team collaboration, use the `openspec` or `hybrid` mode.

### Why this convention

- Deterministic titles let recovery match by exact key.
- `topic_key` enables upserts without duplicates.
- The `sdd/` prefix namespaces all SDD artifacts.
- Two-step recovery keeps search previews from being mistaken for full content: `mem_get_observation` is the only way to get the complete artifact.
- Lineage keeps the `archive-report` complete for traceability.
