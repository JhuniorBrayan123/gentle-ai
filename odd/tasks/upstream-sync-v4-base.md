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

## Progress
- 2026-10-20: diagnóstico completo (mediciones de arriba). T0 en curso.

## Next step
T1: ejecutar `git merge upstream/main` y resolver Lote A + Lote C.

## Verification evidence
(pendiente)
