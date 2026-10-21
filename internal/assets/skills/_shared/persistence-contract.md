# Persistence Contract for Organic Work

## Small and substantial work

Small, understood work needs no durable task artifact. For substantial authorized implementation, create `odd/tasks/{feature-name}.md` before the first source write and mirror the full current document under project-scoped Engram topic `odd/{feature-name}/tasks`. The document records objective, scope, constraints, actionable task IDs, acceptance checks, progress and next step. Do not create a planning artifact for read-only work. The local document and memory mirror are complementary recovery copies, not competing approval authorities.

Read back both writes. If Engram is unavailable, preserve the local document and explicitly mark the mirror pending without blocking independent safe work. If a local write is unsafe, preserve existing state and report the limitation. On resume, retrieve the full observation (search results are truncated), read the document and reconcile them before modifying either. Preserve conflicting versions until the human resolves a genuine conflict. Preserve historical user data; do not erase or replace it to make recovery appear clean.

## Delegated work

The parent selects relevant context and gives the worker its authorized edit surfaces, task document locator, the applicable test-first policy and runner from `## Implementation Routing`, and exact verification commands. The worker reads the task document before editing, stays inside those surfaces, and reports actual checks. Significant verified project discoveries can be saved with `mem_save` under the resolved project. The parent reconciles the task document and full mirror after each observed transition. A task checkbox is not review authority.

Skill paths are supplied in `## Skills to load before work`; workers read those exact `SKILL.md` files before task-specific work. See `skill-resolver.md` for matching and reporting.

## Safe writes and response order

Engram upserts replace existing topic content. Read, merge and write the complete task document rather than replacing earlier completed work with one batch. Set `capture_prompt: false` for an automated task mirror when the tool supports it; normal human-directed saves use the memory provider's default behavior. Never claim a write succeeded without readback. After persisting, finish with a complete text handoff; a memory tool result is not a reply. A worker does not close the parent's session.

Optional research findings stay in the conversation unless the user or authorized task contract requests persistence; disclose failed writes honestly. Native RDD review remains separate from task persistence, and delivery follows ordinary repository policy.

## SDD Artifact Store Persistence

This section is owned by the fork's SDD layer and applies to the SDD workflow only. It adds to the organic-work contract above; it never replaces it.

The orchestrator declares `artifact_store.mode` with one of: `engram | openspec | hybrid | none`.

### Mode roles

- **`engram`**: working memory between sessions. Upserts overwrite, so no iteration history is kept.
- **`openspec`**: source of truth. Files in the repository, git history, team-shareable, full audit trail.
- **`hybrid`**: files for the team plus Engram for recovery. Higher token cost.
- **`none`**: ephemeral. Lost when the conversation ends.

### Behavior per mode

| Mode | Read from | Write to | Project files |
|------|-----------|----------|---------------|
| `engram` | Engram | Engram | Never |
| `openspec` | Filesystem | Filesystem | Yes |
| `hybrid` | Resolved artifact locators | Both | Yes |
| `none` | Orchestrator prompt context | Nowhere | Never |

### Hybrid mode

Use the declared `artifactStore` and resolved `artifactPaths` from the orchestrator, as defined in `sdd-phase-common.md`. Read each resolved locator, not an Engram-first or filesystem fallback. Never silently substitute another store's copy.

Attempt both declared writes, following the Engram and OpenSpec conventions, and read back each successful write. Hybrid writes are not atomic: preserve successful writes, report any failed or unavailable mirror as partial, and name the outstanding locator. Do not claim synchronization or roll back valid progress. Preserve both versions of a genuine content conflict; ask only when it affects the next decision.

### Recovery

Recover from native status and the actual artifacts at their resolved locators, including full `mem_get_observation` content for topic keys. Existing `state.yaml` and `sdd/{change-name}/state` snapshots are optional recovery hints, not required per-phase writes or a second state authority. Missing or stale snapshots do not replace actual task progress or archive closure. Preserve existing snapshots, including `dependsOn` metadata, and historical artifacts.

READ-MERGE-WRITE cumulative tasks and apply-progress: retrieve the full current artifact, preserve prior completed and unrelated work, merge this batch, then persist and read back the full result.

### Common rules

- `none` → do not create or modify any project files; return results inline only.
- `engram` → do not write any project files; persist to Engram and return observation IDs.
- `openspec` → write files only to paths defined in `openspec-convention.md`.
- `hybrid` → persist to both Engram and the filesystem; follow both conventions.
- If the declared store or a required locator is unresolved, report it; do not select another store.

### Sub-agent context rules

Sub-agents launch with a fresh context and no access to the orchestrator's instructions or memory protocol. A phase with declared dependencies reads its artifacts directly from the backend; a phase without dependencies is read by nobody. The output-only SDD research collector returns evidence for the orchestrator to validate and persist, and never persists on its own.

When a sub-agent persists artifacts, the persistence call happens before the final text response: the absolute last output of a sub-agent is text, never a tool call. Sub-agents do not close the parent's session.

### Orchestrator prompt instructions for sub-agents

SDD (artifact-producing phases):

```
Artifact store: {declared artifactStore}
Artifact locators: {resolved artifactPaths and output locators}
Read required dependencies at those locators; topic keys require project-scoped
mem_search followed by full mem_get_observation, never a search preview.
Persist the full artifact through the declared store as in sdd-phase-common.md:
files for openspec, canonical topic keys for engram, both for hybrid, inline for none.
Read back writes and return their actual locators and outcomes. Report partial
persistence honestly; never force an Engram write for openspec or none.
Preserve prior progress with READ-MERGE-WRITE rather than replacing a batch history.
```

For SDD artifacts, `capture_prompt: false` is explicit and mandatory when the Engram tool schema supports it. Engram v1.15.3 defaults `capture_prompt` to true for normal human/proactive saves, but automated pipeline artifacts must not capture the user's prompt. Do not infer this from `type` because SDD artifacts and real human architecture decisions both use `architecture`. If an older schema rejects or does not expose `capture_prompt`, omit it rather than failing.

### Skill registry

The orchestrator pre-resolves skill paths from the skill registry and injects them as `## Skills to load before work` in the launch prompt. Sub-agents read those exact `SKILL.md` files before task-specific work. To generate or update the registry, run the `skill-registry` skill, or `sdd-init`.

### Detail level

The orchestrator may pass `detail_level`: `concise | standard | deep`. This controls output verbosity but does NOT affect what gets persisted — always persist the full artifact.
