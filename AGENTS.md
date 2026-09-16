# Gentle AI — Agent Skills Index

When working on this project, load the relevant skill(s) BEFORE writing any code.

Naming convention: `gentle-ai-*` skills are repo-specific workflow skills. Unprefixed skills are portable writing or work-unit skills and intentionally keep their canonical names.

## How to Use

1. Check the trigger column to find skills that match your current task
2. Load the skill by reading the SKILL.md file at the listed path
3. Follow ALL patterns and rules from the loaded skill
4. Multiple skills can apply simultaneously

## Skills

| Skill | Trigger | Path |
|-------|---------|------|
| `gentle-ai-issue-creation` | When creating a GitHub issue, reporting a bug, or requesting a feature. | [`skills/issue-creation/SKILL.md`](skills/issue-creation/SKILL.md) |
| `gentle-ai-branch-pr` | When creating a pull request, opening a PR, or preparing changes for review. | [`skills/branch-pr/SKILL.md`](skills/branch-pr/SKILL.md) |
| `gentle-ai-chained-pr` | When a change is too large for one review, or when creating chained/stacked pull requests. | [`skills/chained-pr/SKILL.md`](skills/chained-pr/SKILL.md) |
| `cognitive-doc-design` | When writing docs that must reduce cognitive load for readers or reviewers. | [`skills/cognitive-doc-design/SKILL.md`](skills/cognitive-doc-design/SKILL.md) |
| `comment-writer` | When drafting human comments, PR feedback, issue replies, or async updates. | [`skills/comment-writer/SKILL.md`](skills/comment-writer/SKILL.md) |
| `work-unit-commits` | When splitting implementation work into deliverable commits or chained PRs. | [`skills/work-unit-commits/SKILL.md`](skills/work-unit-commits/SKILL.md) |
| `rdd-defect-workflow` | When RDD defects involve receipts, authority, recovery, delivery gates, or kill switches. | [`skills/rdd-defect-workflow/SKILL.md`](skills/rdd-defect-workflow/SKILL.md) |
| `rdd-advisory-transport` | When changing reviewer transport, adapters, lens prompts/schemas, or transport capability policy. | [`skills/rdd-advisory-transport/SKILL.md`](skills/rdd-advisory-transport/SKILL.md) |
| `issue-root-resolution` | When auditing backlog roots, proposing cluster fixes, or closing resolved/outdated issues. | [`skills/issue-root-resolution/SKILL.md`](skills/issue-root-resolution/SKILL.md) |
| `systemic-issue-triage` | When triaging issues, bugs, backlogs, root causes, dead ends, or blocked users. | [`skills/systemic-issue-triage/SKILL.md`](skills/systemic-issue-triage/SKILL.md) |
| `gentle-ai-bench` | When touching `bench/`, journeys, driven mode, the journey corpus, or bench axes. | [`skills/gentle-ai-bench/SKILL.md`](skills/gentle-ai-bench/SKILL.md) |
| `gitlab-mr-flow` | When creating, approving, or merging a GitLab merge request end-to-end (push, create with assigner/reviewer, approval gate, merge, tag). | [`skills/gitlab-mr-flow/SKILL.md`](skills/gitlab-mr-flow/SKILL.md) |
| `gitlab-release-tag` | When a merged GitLab MR needs its semantic version tag + changelog (SmartClic standard). | [`skills/gitlab-release-tag/SKILL.md`](skills/gitlab-release-tag/SKILL.md) |
| `qa-supervisor` | When supervising any QA automation request before writing code: validates rules G1–G6 using BookStack/GitLab MCPs and only then delegates implementation. | [`skills/qa-supervisor/SKILL.md`](skills/qa-supervisor/SKILL.md) |
| `qa-locator-hunting` | When hunting or reusing UI locators for SmartClic/erp-mf-* microfront tests: POM first, GitLab hunt via MCP second, never invents selectors. | [`skills/qa-locator-hunting/SKILL.md`](skills/qa-locator-hunting/SKILL.md) |
| `qa-doc-reference` | When citing BookStack PRD documentation during exploration or on-demand doc search: renders each page as the 13-field ficha (metadata table + Qué contiene) with exact URL and STOP on divergence. | [`skills/qa-doc-reference/SKILL.md`](skills/qa-doc-reference/SKILL.md) |
| `qa-explore` | Trigger: explorar un cambio QA antes de especificar. Analiza tests, fixtures y docs previas (G2) consultando BookStack y Engram. | [`skills/qa-explore/SKILL.md`](skills/qa-explore/SKILL.md) |
| `qa-spec` | Trigger: diseñar la prueba QA antes de implementar. Escenarios, datos y riesgos (G3) con la documentación BookStack como base. | [`skills/qa-spec/SKILL.md`](skills/qa-spec/SKILL.md) |
| `qa-apply` | Trigger: implementar un cambio QA aprobado. Escribe tests con Screenplay+POM (G5) siguiendo el spec y las convenciones del proyecto. | [`skills/qa-apply/SKILL.md`](skills/qa-apply/SKILL.md) |
| `qa-verify` | Trigger: validar la implementación QA contra el spec. Ejecuta pruebas y produce evidencia reproducible (G6) siguiendo la checklist completa. | [`skills/qa-verify/SKILL.md`](skills/qa-verify/SKILL.md) |
| `qa-docs` | Trigger: detectar vacíos o contradicciones en la documentación QA. Produce borradores de gap, nunca escribe en BookStack (G1.4). | [`skills/qa-docs/SKILL.md`](skills/qa-docs/SKILL.md) |
| `qa-review` | Trigger: revisión adversarial de un cambio QA. Evalúa contra BookStack, AGENTS.md y G1-G6 encontrando problemas sin modificar código (G1). | [`skills/qa-review/SKILL.md`](skills/qa-review/SKILL.md) |

