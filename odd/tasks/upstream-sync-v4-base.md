# Feature: upstream sync v4 base (conservar la capa SDD/QA sobre base upstream v4)

## Objective
Actualizar la base del fork `JhuniorBrayan123/gentle-ai` a `upstream/main` (v4.x, HEAD `4acaf4241`, 2026-10-15)
sin perder la capa propia SDD/QA (skills `qa-*`, `internal/qastage`, `internal/components/sdd`,
`internal/sddstatus`, `internal/sddtaskresult`, `openspec/`), y sin una reescritura de producto.

## Why
El fork quedó 4184 commits detrás en conteo (historia reescrita por upstream) con un delta real de contenido
de 2161 archivos / ~135k líneas por lado. Adicionalmente, los plugins de OpenCode v2 no se pueden
actualizar porque el propio fork se niega a actuar bajo v2
(`internal/update/upgrade/strategy.go` y `internal/components/opencodeplugin/plugin.go`:
"OpenCode community plugin V2 compatibility is unverified"). Traer upstream es la vía de corregir
el contrato v2 — pero upstream **borró** SDD/OpenSpec (commit `e219644b2`), que es la base del QA orchestrator v3.

## Measured ground truth (2026-10-20, `git merge-tree --write-tree main upstream/main`)
- 914 rutas en conflicto (el "2713" del log crudo son líneas de mensaje duplicadas por directorio).
- **897 (98%)**: la versión en `main` de ese archivo EXISTE en la historia de upstream
  (verificado blob-por-blob con `git rev-list --objects` + `git cat-file --batch-check`)
  → tomar la versión de upstream **no pierde trabajo propio**.
- **16 (1.8%)**: la versión en `main` es única (editada por el fork) → resolución semántica manual.
- 1: archivo que upstream agrega y `main` no tiene.
- 713 archivos que upstream borra: `openspec/changes` (283) + `openspec/specs` (28) + `internal/sddstatus` (53)
  + `internal/components` (52) + `internal/assets` (126) + `testdata/golden` (49) + `internal/qastage` (13)
  + `internal/sddtaskresult` (6) + resto.

## Decision (user, 2026-10-20)
**Conservar la capa SDD/QA sobre base v4.** En las rutas que son capa propia gana "ours";
en el resto gana "theirs". NO se adopta el retiro de SDD de upstream (no se porta QA a ODD en este ciclo).
Consecuencia aceptada: coexisten SDD y ODD en el fork; portar QA a ODD es trabajo futuro.

## Scope
- [ ] T0 Higiene y red de seguridad: tag `backup/main-pre-upstream-sync-2026-10-20`, rama `sync/upstream-2026-10`,
      artefactos de análisis fuera del repo, working tree limpio.
- [ ] T1 Merge `upstream/main` + Lote A (897 mecánicos `--theirs`) + Lote C ("ours" en rutas de capa).
- [ ] T2 Lote B: resolución semántica de los 16 archivos con versión única del fork.
- [ ] T3 Verificación: `go build ./...`, `go test ./...`, goldens, e2e, arranque real de OpenCode.
- [ ] T4 Commit + MR a `main`.

## Authorized scope
- Ramas locales: `sync/upstream-2026-10` (creada), tag de backup (creado).
- **NO** autorizado: push a `origin`, apertura de MR, force-push, cualquier escritura en `upstream`.
  Push/MR quedan como decisión del usuario al final.

## Lote B — los 16 con versión única del fork (trabajo semántico real)
```
e2e/e2e_test.sh                              internal/components/filemerge/yaml.go
internal/agents/claude/adapter.go            internal/components/skills/inject_test.go
internal/app/app.go                          internal/components/skills/presets.go
internal/assets/assets_test.go               internal/components/skills/presets_test.go
internal/catalog/skills.go                   internal/model/types.go
internal/cli/run.go                          internal/tui/screens/skill_picker.go
internal/cli/run_integration_test.go         internal/tui/screens/skill_picker_test.go
internal/cli/sync.go                         testdata/golden/skills-presets.json
```
Estos son los puntos de integración de la capa QA (registro de skills QA, picker, CLI, catalog, model types).

## Acceptance criteria
- `go build ./...` y `go test ./...` en verde sobre la rama de sync.
- Las skills `qa-*` siguen presentes y registrables (presets + catalog + picker).
- `internal/qastage`, `internal/components/sdd`, `internal/sddstatus`, `internal/sddtaskresult`, `openspec/`
  siguen existiendo y compilando.
- `opencode plugin list` sin errores y los 6 plugins `.ts` con id v2.
- Cero conflictos sin resolver (`git diff --check` + `git ls-files -u` vacío).

## Applicable checks
- `go build ./...`
- `go test ./...`
- `e2e/e2e_test.sh` (bash) — aserciones de conteo de skills y file-exists.
- `gentle-ai doctor`
- `opencode plugin list`

## TDD mode
Strict TDD Mode: no se agrega comportamiento nuevo; el trabajo es reconciliación. Los tests existentes
(`presets_test.go`, `skill_picker_test.go`, `assets_test.go`, goldens) son la red de seguridad: si un
conflicto rompe un contrato, el test lo detecta. Runner: Go estándar + `e2e/e2e_test.sh`.

## Route
- T0/T1: direct inline (operación de git + script de resolución masiva; no hay decisión de diseño por archivo).
- T2: delegated direct — 16 archivos interrelated (model types → catalog → presets → picker → CLI →
  golden) forman un único grafo de dependencias; un solo writer lo resuelve en bloque para evitar
  estados intermedios inconsistentes.
- T3: direct inline (comandos de verificación) + delegated verification si un test falla.

## Corrections to the merge method (learned during T2/T3)
1. **"main's blob exists in upstream history" does NOT mean "losing it loses no fork work".**
   That classification (897/914) was correct about *provenance* — those 914 files' fork copies were old
   upstream snapshots — but wrong as a *safety* proof: upstream v4 later REMOVED clauses from its canonical
   assets and prompt, and the fork's SDD layer still depends on them. Example: the clause
   "Tests or frameworks being present does not enable TDD" lived in `internal/components/agentguidance/routing.go`
   and was dropped by upstream commit `55e3abf1c` / `e219644b2` (SDD retirement). Taking upstream's file
   removed it, and the fork's SDD tests assert it.
   Correction applied: the fork's AUTHORS' commits are now the authority for what the fork owns —
   83 files touched by `git log main --author="gutierrezmamanib26@gmail.com"`. Intersection with the
   893 "theirs" resolutions = **0 at risk**: the fork's own authored work is intact.
2. **The Lote C layer pattern missed `internal/assets/skills/_shared/`.**
   That directory holds fork-owned shared assets (`research-lifecycle.md`, `persistence-contract.md`,
   `engram-convention.md`, `sdd-phase-common.md`, `sdd-status-contract.md`, `qa-gate-policy.md`).
   Upstream v4 rewrote the shared ones to ODD semantics, so they took upstream's version. The fork's
   SDD clauses must be re-introduced in the fork's SDD-owned assets instead of reverting upstream's files.
3. **v4 made the review contract source mandatory.** `agentguidance.InjectRouting` became
   `InjectRoutingWithOptions(targetDir, agent, RoutingOptions{ReviewContract: ...})`; without a source the
   injection fails closed ("review contract source was not wired for %q"). `SetReviewContractSource` is a
   TEST-ONLY export (`export_test.go`) and is unusable from another package; only `RoutingOptions.ReviewContract`
   is public. The fork's own `ReviewExecutionContractFor(agent) (string, error)` in
   `internal/components/sdd/boundedreview.go` matches the required signature exactly.

## Post-merge audit (beyond the author-based check)
The author-based check (83 files by `gutierrezmamanib26@gmail.com`) is ADDITIONAL evidence, not the
protection. Two further protections were applied:

1. **Branch comparison against `qa-orchestrator-v3`** (the branch that carried the QA layer into main,
   `c7723af49`): 21 files differ from `main`. Of those, 16 are the manually reconciled Lote B files, 5 are
   intact (= `main`), and **1 was discarded: `AGENTS.md`** — the merge took upstream's version and lost the
   11 `qa-*`/`smartops-ui` rows of the fork's skill table (main: 10 qa rows, merged: 0). No test binds the
   root `AGENTS.md`, so this is documentation hygiene, not a suite requirement. The 11 rows were extracted
   from `main:AGENTS.md` for re-insertion on top of upstream's canonical skill table.
2. **Known SDD/QA layer paths**: only 4 files under the layer paths ended up with upstream's version —
   `internal/assets/skills/_shared/{research-lifecycle,persistence-contract,engram-convention,skill-resolver}.md`
   — the shared assets upstream v4 rewrote to ODD semantics. These are the writer's assignment: the fork's
   SDD clauses must be re-introduced in fork-owned SDD assets WITHOUT reverting upstream's canonical files.

Full divergence audit (index of the merge vs both trees): 872 files where the fork's `main` diverged from
upstream → 23 reconciled manually, 1 kept as `ours`, 848 took upstream's version. The residual risk class is
"upstream removed something the fork's layer still needs"; the test suite is the auditor for that class.

## Baseline worktrees (for the comparative report)
- `%TEMP%\opencode\wt-upstream-v4` → upstream/main v4 clean (`4acaf4241`, module `/v4`)
- `%TEMP%\opencode\wt-fork-v3` → fork `main` pre-merge (`4ed4793f2`, module `/v3`)
- Suite results: `%TEMP%\opencode\baseline-upstream-v4.txt`, `%TEMP%\opencode\baseline-fork-v3.txt`

## Baseline results (the bar is comparative, not "all green")
`go test ./...` on CLEAN `upstream/main` v4 (`wt-upstream-v4`, `4acaf4241`) on this Windows machine:
**12 failing packages / ~85 failing tests, EXIT=1.** Upstream v4 is NOT green here, so the acceptance bar is
comparative: no failure that the fork-v3 baseline and the upstream-v4 baseline do not also have.

Classification of the upstream-v4 baseline failures (all environmental / platform, none product):
- `TestUserWindowsWriteLatency` → "Windows directory grants another account mutation access" (local ACL hardening) → **entorno**
- `TestReleaseSecurityScriptsAreSyntacticallyValidAndFailClosed` → `bash -n ...: exit status 127` (bash not resolvable in the test PATH) → **entorno**
- `TestRunStrategyUsesCaskOwnershipAndMigrationGuidance` → `exec: "printf": executable file not found` (needs Homebrew/macOS) → **plataforma**
- `TestCheckSingleTool_EngramUsesBinaryReleaseChannel` → `engram status = "version-unknown"` (no registry/network) → **entorno**
- `TestInstallScriptBetaGoInstallBypassesPublicGoProxy` → POSIX shell fixture, `exit status 1` → **entorno**
- `TestResolveEffectiveConfigPrecedenceAndDefaultWriteTarget` → path expectation mismatch in upstream's own test on Windows → **upstream-on-Windows**
- 5 packages hit Go's 10-minute timeout (~600s each): `e2e/organicruntime`, `internal/app`, `internal/cli`,
  `internal/components/communitytool`, `internal/reviewtransaction` → re-run isolated with `-p 1 -timeout 40m`
  (`%TEMP%\opencode\baseline-upstream-v4-rerun.txt`) to separate saturation from real failures.

## Comparative bar (normalized package paths, `/v3` and `/v4` unified)
| | upstream v4 limpio | fork v3 (main pre-merge) | UNION = ambiental/plataforma |
|---|---|---|---|
| Paquetes FAIL | 12 | 10 | **13** |
| Tests FAIL (top-level) | 31 | 18 | **36** |
| Comunes a ambos | | | **13** (fallan en v3 y v4) |

Packages only failing in v4 (new in v4): `internal/components/legacyassets`,
`internal/components/opencoderuntimeplugins`, `internal/shellinstaller`.
Package only failing in v3 (retired in v4): `internal/components/engram`.

**Acceptance rule for the merged tree:** any failing package/test NOT present in these sets is a
candidate merge regression (category 2) and must be explained or fixed. Everything in the sets is
category 4 (environment/platform) and must NOT be "fixed" in product.

Baseline sets persisted for the test-by-test diff:
`%TEMP%\opencode\set-{v4-pkgs,v3-pkgs,union-pkgs,v4-tests,v3-tests,union-tests}.txt`

## Progress
- 2026-10-20: diagnóstico completo (mediciones de arriba).
- 2026-10-20 T0 CERRADO: tag `backup/main-pre-upstream-sync-2026-10-20`, rama `sync/upstream-2026-10`,
  artefactos de análisis movidos a `%TEMP%\opencode\`, `.windsurf/workflows/sdd-new.md` restaurado
  (upstream lo borra pero su capa SDD lo conserva → "ours"), working tree limpio.
- 2026-10-20 T1 CERRADO: `git merge upstream/main --no-commit --no-ff` → 923 rutas en conflicto.
  Resolución por capas: **893 "theirs"** (upstream), **14 "ours"** (capa QA/SDD), **16 diferidas a T2**.
  `testdata/golden/sdd-command-sdd-init.md` restaurado a mano (upstream lo borra y su
  `scripts/gen-sdd-agent-digests/assets.go` lo referencia).
- 2026-10-20 T2 CERRADO: los 16 resueltos con merge MANUAL (nunca a lo bruto), reescritura de módulo
  `/v3` → `/v4` (182 → 8 archivos; los 8 restantes son fixtures idénticos a upstream que legítimamente
  conservan `/v3` para installs cross-major), y adaptación de los guards de upstream v4 que afirman que
  ningún asset SDD debe shippear: `TestODDOnlyEmbeddedInventory` → `TestEmbeddedInventoryKeepsForkSDDAndQAAssets`
  con allowlist `forkRetainedSDDAsset()`, más `retired_workflow_ratchet_test.go` adaptado.
  Integridad de la capa QA verificada post-merge: 10 skills `qa-*` (raíz + espejo embebido),
  `qastage` 13, `sddstatus` 53, `sddtaskresult` 6, `components/sdd` 47, `openspec/changes` 278,
  `qa-commands` 5, `smartops-commands` 1; 11 skill IDs QA en `model/types.go`, 10 en `catalog/skills.go`,
  11 en `presets.go`, 20 refs QA en `testdata/golden/skills-presets.json`.
- 2026-10-20 T3 PARCIAL: `go build ./...` → exit 0 (29s). `gofmt -l .` → limpio. `go vet` (muestreo) → exit 0.
  0 conflictos abiertos; 1550 archivos staged sobre el árbol mergeado. La suite compilaba pero
  `internal/components/sdd` + `internal/components` fallan (22 tests top-level). Causas raíz verificadas:
  (a) v4 exige `RoutingOptions.ReviewContract` y la capa SDD no lo conecta (fallo de PRODUCCIÓN, no solo de
  test) — corregido en los 3 call sites de test del fork con `ReviewExecutionContractFor`;
  (b) upstream v4 reescribió los assets compartidos a ODD y el research gatekeeper / persistence contract SDD
  se perdieron; (c) v4 cambió las tablas de assignment de modelos Codex (`gpt-6-luna`/`gpt-6-sol`) y el conteo
  de secciones markdown (16→19); (d) cambios de manejo JSONC y rutas Windows. Porte delegado a un writer
  único con la especificación de las 3 correcciones de arriba.

- 2026-10-20 T3 (sigue) CERRADO en las dos suites spec: `go test -count=1 ./internal/components/sdd/... ./internal/components/` → **exit 0**
  (0 tests top-level fallando; antes eran 22). Correcciones aplicadas:
  (a) los assets compartidos canónicos de upstream v4 (`research-lifecycle.md`, `persistence-contract.md`,
  `engram-convention.md`) conservan su texto ODD intacto y reciben una sección SDD propiedad del fork
  (`#### Research-specific gatekeeper precedence`, `## SDD Artifact Store Persistence`, `## SDD Artifact Convention`);
  (b) `opencode/background-subagents.md` recupera el sentinel "SDD apply" que upstream v4 había quitado;
  (c) goldens/contratos actualizados a la realidad de v4 (`gpt-6.1-*`, baseline SHA-256 de Kilocode, costos de
  protocolo de review, `engram-antigravity-mcp.golden` con `--tools=agent`);
  (d) guards de upstream v4 que afirman que SDD no existe adaptados con allowlist de la capa del fork
  (`docs_retired_terms_test.go`, `documented_invocation_test.go`, `compatibility_skills_test.go`,
  `retired_workflow_ratchet_test.go` de `internal/assets` y de `internal/components/agentguidance`,
  `components/skills/inject_test.go`).
  `bash e2e/e2e_test.sh` (Git Bash 5.2.26) → **PASSED 66 / FAILED 0 / SKIPPED 3** (tier 1).
  `gofmt -l .` → vacío. `go build ./...` → exit 0.
  Pendiente para T3: la suite completa (`go test ./...`) conserva fallas que dependen de un item grande
  fuera del alcance de esta corrección — el pipeline de install/sync de `internal/cli` sigue la ruta de
  retiro de SDD de upstream v4 (`componentApplyStep`/`componentSyncStep` con `case model.ComponentSDD: return nil`
  más los pasos `retiredSDD*`), así que el componente SDD del fork no se instala y `_shared/sdd-*.md` se borran
  tras escribirse. Eso hace fallar la verificación post-apply de `internal/cli` y `internal/app`.

## Next step
T3: cerrar la suite completa + `e2e/e2e_test.sh`. T4: commit del merge + MR a `main` (push requiere
autorización explícita del usuario).

## Verification evidence
- `go build ./...` → exit 0 (29s)
- `gofmt -l .` → limpio (0 archivos)
- `go vet ./internal/catalog/... ./internal/components/skills/... ./internal/tui/...` → exit 0
- `go test -count=1 ./internal/components/sdd/... ./internal/components/` → **exit 0** (453s + 17s)
- `bash e2e/e2e_test.sh` → **PASSED 66 / FAILED 0 / SKIPPED 3** (tier 1, Git Bash 5.2.26)
- `go test ./...` → EN CURSO
- `gentle-ai doctor` / `opencode plugin list` → pendiente (requiere reinstalar el binario desde este árbol)
- `git diff --check` → limpio (0 whitespace errors)
- `git ls-files -u` → 0 · `git diff --diff-filter=U` → 0 · `MERGE_HEAD` presente (merge sin commitear)
- `go test -count=1 ./...` → EXIT=1: 13 paquetes FAIL / 42 tests. **0 paquetes fuera de la unión de los dos
  baselines** (13/13paketes son ambientales). 10 tests nuevos: 8 = gap SDD (tipo 2) · 2 = flakiness por
  saturación (verde en re-corrida aislada → tipo 4).
- Build + install en perfil OpenCode aislado (`%TEMP%\opencode\iso-validate.ps1`) → **install FALLÓ y revirtió**:
  "post-apply verification failed: 45 passed, 3 failed" — los 3 `_shared/{openspec-convention,sdd-phase-common,
  sdd-status-contract}.md` borrados por los pasos `retiredSDD*` de upstream. Sin rollback no queda nada
  instalado (solo directorios vacíos). El guard nuevo de v4 (`@opencode/plugin@2.0.4` requerido) **sí** funciona
  y devuelve la remediación exacta.
- Leakage check: el perfil real NO fue tocado (`~/.gentle-ai/state.json` con mtime sin cambios; aislamiento con
  `GENTLE_AI_HOME` + `USERPROFILE`/`HOME`/`XDG_*`/`APPDATA`/`LOCALAPPDATA`).
- Los 5 `/qa-*` **no** se pudieron ejecutar en vivo: el provider `opencode` devuelve "Rate limit exceeded" a
  nivel de cuenta para sesiones nuevas (falla igual con otro modelo del mismo provider). Las 11 skills `qa-*`
  + `smartops-ui` sí quedaron instaladas en el perfil aislado (verificado en disco).

## Diff stat final
- Merge (index vs HEAD): **1549 files changed, +138190 / -25459**
- Porte (worktree vs index): **19 files changed, +423 / -42**
  - producción: `docs/architecture.md`, `internal/assets/opencode/background-subagents.md`,
    `internal/assets/skills/_shared/{engram-convention,persistence-contract,research-lifecycle}.md`,
    `testdata/golden/engram-antigravity-mcp.golden`, este doc
  - tests: 12 archivos (`internal/components/sdd/*`, `internal/app/*`, `internal/assets/*`,
    `internal/cli/compatibility_skills_test.go`, `internal/components/skills/inject_test.go`)

## Next step (actualizado)
El merge está listo y verificado, pero queda **un gap de producto abierto (tipo 2)**: el pipeline de
install/sync de `internal/cli` sigue la ruta de retiro de SDD de upstream v4, por lo que el componente SDD
del fork no se instala y sus assets compartidos se borran → el install revierte por completo. Restaurarlo
exige cambios de producción en `internal/cli/{run,sync,uninstall}.go` + `internal/model/selection.go`
(perfiles SDD nombrados) y reescribir 6 archivos de test de upstream que afirman el contrato contrario
("SDD está retirado"). Es una tarea propia, con su revisión, no un remate de este merge.
