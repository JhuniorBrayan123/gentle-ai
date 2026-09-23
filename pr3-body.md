<!-- ⚠️ READ BEFORE SUBMITTING
  Every PR must be linked to an issue that has the "status:approved" label.
  PRs without a linked approved issue will be automatically rejected by CI.
  See CONTRIBUTING.md for the full contribution workflow.
-->

## 🔗 Linked Issue

Closes #N 
<!-- Update this to point to the approved issue for this slice -->

---

## 🏷️ PR Type

What kind of change does this PR introduce?

- [ ] `type:bug` — Bug fix (non-breaking change that fixes an issue)
- [ ] `type:feature` — New feature (non-breaking change that adds functionality)
- [ ] `type:docs` — Documentation only
- [ ] `type:refactor` — Code refactoring (no functional changes)
- [ ] `type:chore` — Build, CI, or tooling changes
- [ ] `type:breaking-change` — Breaking change (fix or feature that changes existing behavior)

---

## 📝 Summary

This PR implements Phase 2 of the `qa-orchestrator-v2` stack, enforcing strict stage ordering in the runtime ledger. It ensures that objective advances in the runtime ledger follow the deterministic stage sequence defined by the `StageVocabulary`.

---

## 📂 Changes

| File / Area | What Changed |
|-------------|-------------|
| `internal/sddstatus/runtime_compact.go` | Minor update for ledger compatibility |
| `internal/sddstatus/runtime_ledger.go` | Adds `StageVocabulary`/`StagePosition` and enforces sequential stage validation (`ErrRuntimeStageOutOfOrder`, `ErrRuntimeStageUnknown`) |
| `internal/sddstatus/runtime_stage_order_test.go` | Tests for stage order enforcement and structural integrity |

Total: 3 files changed, 220 additions, 30 deletions.

---

## 🧪 Test Plan

**Unit Tests**
```bash
go test ./... -v
```
> **Known pre-existing failures (not blocking this PR):** `internal/update` and `internal/update/upgrade` clusters present on `main` are not introduced by this slice; verified identical via `git stash` baseline. `internal/sddstatus` tests pass cleanly and Phase 0's byte-identical regression gate continues to pass unaffected.

**Go Format**
```bash
go run ./internal/gofmtcheck
```

**E2E Tests** (Docker required)
```bash
cd e2e && ./docker-test.sh
```

**Benchmark Validation**

N/A - This change only affects runtime ledger validation.

- [x] Unit tests pass (`go test ./...`) *(with pre-existing failures noted above)*
- [x] Go format passes (`go run ./internal/gofmtcheck`)
- [ ] E2E tests pass (`cd e2e && ./docker-test.sh`)
- [ ] Manually tested locally

---

## ✅ Contributor Checklist

- [ ] PR is linked to an issue with `status:approved`
- [x] PR stays within 400 changed lines, or I have requested/obtained maintainer-applied `size:exception` with rationale documented
- [ ] I have added the appropriate `type:*` label to this PR
- [x] Unit tests pass (`go test ./...`)
- [x] Go format passes (`go run ./internal/gofmtcheck`)
- [ ] E2E tests pass (`cd e2e && ./docker-test.sh`)
- [x] Benchmark validation completed, or this change is not applicable to the benchmark (explain why in the Test Plan).
- [ ] I have updated documentation if necessary
- [x] My commits follow [Conventional Commits](https://www.conventionalcommits.org/) format
- [x] My commits do not include `Co-Authored-By` trailers

---

## Pending maintainer actions

The following are **maintainer-applied per `pr-check.yml` and CONTRIBUTING.md** in this repo — not within contributor scope:

- [ ] `type:feature` label applied to this PR
- [ ] Fork workflow approval (if required for action_required gate)

---

## 💬 Notes for Reviewers

- PR 3 of 13 for `qa-orchestrator-v2`. Stacked on top of PR `#2` (`feat/qa-orchestrator-v2-pr2-sddstatus-foundation`).
- For production Go changes in `internal/sddstatus`:
  - [ ] Identify any qualifying security, integrity, admission, repair, or governance guard and challenge its legitimate input population against real-world evidence.
  - [ ] Confirm its `guard:population` direction and claim are adjacent and accurate.
