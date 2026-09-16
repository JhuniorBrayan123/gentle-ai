package qastage_test

// E2E integration tests for the full qa-orchestrator-v2 stage lifecycle
// (Phase 11 / PR13). These tests exercise the REAL primitives across three
// layers in one temp-repo ledger, not isolated unit doubles:
//   - internal/sddstatus: the vocabulary-bound runtime ledger (Begin/Finish/
//     ApproveStage/Reset) and its replay-derived RuntimeStatus.
//   - internal/qastage: the qa stage vocabulary and artifact admission rules.
//   - internal/cli: the qa-status/qa-validate verbs, exactly as an agent
//     would invoke them.
//
// Scope (per PR13 task, Phase 11):
//  1. Full chain explore -> spec -> apply -> verify, with an audited approval
//     required and recorded before every advance (matches the ACTUAL ledger
//     behavior implemented in PR3/PR4: every advance in a vocabulary-bound
//     chain requires approval of its predecessor, not only advance-into-apply
//     as an earlier design draft assumed -- see runtime_stage_approval_test.go).
//  2. A simulated interruption mid-chain (no Finish call) followed by a
//     brand-new RuntimeStore handle (as a resumed process would open) proving
//     resume recovers the in-flight attempt purely from ledger replay -- no
//     cached pointer, matching qa-status's own "recomputed, never cached"
//     contract.
//  3. A mid-chain requirement change (re-running an earlier stage with new
//     evidence) correctly invalidating a downstream approval bound to the
//     old evidence revision (design D4 / spec QA-ORCH-10 backtrack scenario).
//
// Rollback boundary: delete this file. It adds no production code.
import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v2/internal/qastage"
	"github.com/gentleman-programming/gentle-ai/v2/internal/sddstatus"
)

func initE2ERepository(t *testing.T, dir string) {
	t.Helper()
	runE2EGit(t, dir, "init", "-q", "--initial-branch=main")
	runE2EGit(t, dir, "config", "user.name", "Test")
	runE2EGit(t, dir, "config", "user.email", "test@example.com")
	runE2EGit(t, dir, "commit", "-q", "--allow-empty", "-m", "Initial commit")
}

func runE2EGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func e2eQAStatus(t *testing.T, change, cwd string) cli.QAStatus {
	t.Helper()
	var buf bytes.Buffer
	if err := cli.RunQAStatus([]string{"--change", change, "--cwd", cwd}, &buf); err != nil {
		t.Fatalf("qa-status --change %s --cwd %s: %v", change, cwd, err)
	}
	var status cli.QAStatus
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		t.Fatalf("bad qa-status JSON: %v\n%s", err, buf.String())
	}
	return status
}

func e2eValidateArtifact(t *testing.T, payload, change, stage, sourceRevision string) cli.QAValidateResult {
	t.Helper()
	args := []string{"--input", "-", "--change", change, "--stage", stage}
	if sourceRevision != "" {
		args = append(args, "--source-revision", sourceRevision)
	}
	var buf bytes.Buffer
	if err := cli.RunQAValidateFromReader(args, strings.NewReader(payload), &buf); err != nil {
		t.Fatalf("qa-validate: %v", err)
	}
	var result cli.QAValidateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("bad qa-validate JSON: %v\n%s", err, buf.String())
	}
	return result
}

func e2eEvidence(fill byte) string {
	return "sha256:" + strings.Repeat(string(fill), 64)
}

// TestE2EFullChainApprovalAndResume is the primary Phase 11 scenario: begin a
// stage chain, require and record an approval before each advance, advance
// through apply and verify, and confirm a mid-chain interruption resumes
// correctly from persisted ledger state.
func TestE2EFullChainApprovalAndResume(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	initE2ERepository(t, repo)
	change := "e2e-full-chain"
	ledgerChange := qastage.LedgerChangeName(change)
	vocab := qastage.VocabularyV1()

	openStore := func(t *testing.T) sddstatus.RuntimeStore {
		t.Helper()
		store, err := sddstatus.OpenRuntimeStore(ctx, repo, ledgerChange)
		if err != nil {
			t.Fatalf("OpenRuntimeStore: %v", err)
		}
		return store.WithStageVocabulary(vocab)
	}

	// --- Stage 1: explore ---
	store := openStore(t)
	status, err := store.Begin(ctx, sddstatus.BeginAttemptRequest{
		RequestID: "begin-explore", WorkUnit: "explore", EvidenceGoal: "discover requirements",
		MaxAttempts: 3, MaxChangedLines: 400,
	})
	if err != nil {
		t.Fatalf("Begin explore: %v", err)
	}

	exploreArtifact := `{"scope":{"status":"complete","next_stage":"spec"}}`
	if v := e2eValidateArtifact(t, exploreArtifact, change, "explore", ""); !v.Valid {
		t.Fatalf("explore artifact rejected: %s", v.Reason)
	}
	exploreEvidence := e2eEvidence('a')

	status, err = store.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "finish-explore", Outcome: sddstatus.AttemptPassed,
		EvidenceRevision: exploreEvidence, Diagnosis: "requirements discovered",
		CleanupEvidence: "none", ProcessEvidence: "none", HarnessDisposition: sddstatus.HarnessReused,
	})
	if err != nil {
		t.Fatalf("Finish explore: %v", err)
	}
	if got := e2eQAStatus(t, change, repo).NextAction; got != sddstatus.RuntimeActionComplete {
		t.Fatalf("qa-status after explore finish NextAction = %q, want %q", got, sddstatus.RuntimeActionComplete)
	}

	// Advancing into spec WITHOUT an approval of explore must be refused.
	if _, err := store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-spec-unapproved", WorkUnit: "spec",
		EvidenceGoal: "write spec", MaxAttempts: 3, MaxChangedLines: 400,
	}); err != sddstatus.ErrRuntimeStageApprovalRequired {
		t.Fatalf("expected ErrRuntimeStageApprovalRequired advancing unapproved explore, got %v", err)
	}

	// Approve explore, then advance into spec.
	status, err = store.ApproveStage(ctx, sddstatus.ApproveStageRequest{
		ExpectedRevision: status.Revision, RequestID: "approve-explore", Stage: "explore",
	})
	if err != nil {
		t.Fatalf("ApproveStage explore: %v", err)
	}
	status, err = store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-spec", WorkUnit: "spec",
		EvidenceGoal: "write spec", MaxAttempts: 3, MaxChangedLines: 400,
	})
	if err != nil {
		t.Fatalf("Begin spec (approved): %v", err)
	}

	// --- Simulated interruption mid-spec: no Finish call, "process" ends. ---
	// A brand-new store handle -- exactly what a resumed process opens --
	// must reconstruct the exact same in-flight state purely from replay.
	resumedStore := openStore(t)
	resumedStatus, err := resumedStore.Status()
	if err != nil {
		t.Fatalf("resumed Status(): %v", err)
	}
	if resumedStatus.ActiveAttempt == nil || resumedStatus.ActiveAttempt.WorkUnit != "spec" {
		t.Fatalf("resumed status did not reconstruct the active spec attempt: %#v", resumedStatus.ActiveAttempt)
	}
	if resumedStatus.NextAction != sddstatus.RuntimeActionFinish {
		t.Fatalf("resumed NextAction = %q, want %q", resumedStatus.NextAction, sddstatus.RuntimeActionFinish)
	}
	if got := e2eQAStatus(t, change, repo).NextAction; got != sddstatus.RuntimeActionFinish {
		t.Fatalf("resumed qa-status NextAction = %q, want %q", got, sddstatus.RuntimeActionFinish)
	}

	// Resume the interrupted work using the resumed handle/revision -- this
	// proves the resumed handle is fully usable, not merely read-only.
	specArtifact := `{"scope":{"status":"complete","next_stage":"apply"},"predecessor_sha256":"` + exploreEvidence + `"}`
	if v := e2eValidateArtifact(t, specArtifact, change, "spec", exploreEvidence); !v.Valid {
		t.Fatalf("spec artifact rejected: %s", v.Reason)
	}
	specEvidence := e2eEvidence('b')
	status, err = resumedStore.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision: resumedStatus.Revision, RequestID: "finish-spec", Outcome: sddstatus.AttemptPassed,
		EvidenceRevision: specEvidence, Diagnosis: "spec written",
		CleanupEvidence: "none", ProcessEvidence: "none", HarnessDisposition: sddstatus.HarnessReused,
	})
	if err != nil {
		t.Fatalf("Finish spec (resumed): %v", err)
	}

	// --- Stage 3: apply, gated by approval of spec ---
	store = openStore(t)
	status, err = store.ApproveStage(ctx, sddstatus.ApproveStageRequest{
		ExpectedRevision: status.Revision, RequestID: "approve-spec", Stage: "spec",
	})
	if err != nil {
		t.Fatalf("ApproveStage spec: %v", err)
	}
	status, err = store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-apply", WorkUnit: "apply",
		EvidenceGoal: "implement", MaxAttempts: 3, MaxChangedLines: 400,
	})
	if err != nil {
		t.Fatalf("Begin apply (approved): %v", err)
	}
	if status.Objective == nil || status.Objective.StagePosition != 2 {
		t.Fatalf("apply objective StagePosition = %#v, want 2 (0-indexed: explore=0,spec=1,apply=2)", status.Objective)
	}

	applyArtifact := `{"scope":{"status":"complete","next_stage":"verify"},"predecessor_sha256":"` + specEvidence + `"}`
	if v := e2eValidateArtifact(t, applyArtifact, change, "apply", specEvidence); !v.Valid {
		t.Fatalf("apply artifact rejected: %s", v.Reason)
	}
	applyEvidence := e2eEvidence('c')
	status, err = store.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "finish-apply", Outcome: sddstatus.AttemptPassed,
		EvidenceRevision: applyEvidence, Diagnosis: "implemented",
		CleanupEvidence: "none", ProcessEvidence: "none", HarnessDisposition: sddstatus.HarnessReused,
	})
	if err != nil {
		t.Fatalf("Finish apply: %v", err)
	}

	// --- Stage 4: verify, gated by approval of apply ---
	status, err = store.ApproveStage(ctx, sddstatus.ApproveStageRequest{
		ExpectedRevision: status.Revision, RequestID: "approve-apply", Stage: "apply",
	})
	if err != nil {
		t.Fatalf("ApproveStage apply: %v", err)
	}
	status, err = store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-verify", WorkUnit: "verify",
		EvidenceGoal: "verify", MaxAttempts: 3, MaxChangedLines: 400,
	})
	if err != nil {
		t.Fatalf("Begin verify (approved): %v", err)
	}
	// docs is the opt-in terminal branch (never a required predecessor), so
	// verify's own artifact still names it as the legal next_stage.
	verifyArtifact := `{"scope":{"status":"complete","next_stage":"docs"},"predecessor_sha256":"` + applyEvidence + `"}`
	if v := e2eValidateArtifact(t, verifyArtifact, change, "verify", applyEvidence); !v.Valid {
		t.Fatalf("verify artifact rejected: %s", v.Reason)
	}
	verifyEvidence := e2eEvidence('d')
	status, err = store.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "finish-verify", Outcome: sddstatus.AttemptPassed,
		EvidenceRevision: verifyEvidence, Diagnosis: "verified",
		CleanupEvidence: "none", ProcessEvidence: "none", HarnessDisposition: sddstatus.HarnessReused,
	})
	if err != nil {
		t.Fatalf("Finish verify: %v", err)
	}

	finalStatus := e2eQAStatus(t, change, repo)
	if finalStatus.NextAction != sddstatus.RuntimeActionComplete {
		t.Fatalf("final qa-status NextAction = %q, want %q (docs stays opt-in, chain completes without it)", finalStatus.NextAction, sddstatus.RuntimeActionComplete)
	}
	if !finalStatus.Complete {
		t.Fatalf("final qa-status Complete = false, want true")
	}
	if len(finalStatus.Approvals) != 3 {
		t.Fatalf("expected 3 recorded stage approvals (explore, spec, apply), got %d: %#v", len(finalStatus.Approvals), finalStatus.Approvals)
	}
}

// TestE2EBacktrackInvalidatesDownstreamApproval exercises design D4 / spec
// QA-ORCH-10's backtrack scenario: a maintainer resets an approved,
// already-advanced-past stage back open (a mid-chain requirement change),
// re-runs it with new evidence, and the stale approval bound to the OLD
// evidence revision no longer authorizes the advance it used to.
func TestE2EBacktrackInvalidatesDownstreamApproval(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	initE2ERepository(t, repo)
	change := "e2e-backtrack"
	ledgerChange := qastage.LedgerChangeName(change)
	vocab := qastage.VocabularyV1()

	store, err := sddstatus.OpenRuntimeStore(ctx, repo, ledgerChange)
	if err != nil {
		t.Fatalf("OpenRuntimeStore: %v", err)
	}
	store = store.WithStageVocabulary(vocab)

	status, err := store.Begin(ctx, sddstatus.BeginAttemptRequest{
		RequestID: "begin-explore", WorkUnit: "explore", EvidenceGoal: "discover", MaxAttempts: 3, MaxChangedLines: 400,
	})
	if err != nil {
		t.Fatalf("Begin explore: %v", err)
	}
	originalEvidence := e2eEvidence('1')
	status, err = store.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "finish-explore", Outcome: sddstatus.AttemptPassed,
		EvidenceRevision: originalEvidence, Diagnosis: "first pass",
		CleanupEvidence: "none", ProcessEvidence: "none", HarnessDisposition: sddstatus.HarnessReused,
	})
	if err != nil {
		t.Fatalf("Finish explore: %v", err)
	}
	status, err = store.ApproveStage(ctx, sddstatus.ApproveStageRequest{
		ExpectedRevision: status.Revision, RequestID: "approve-explore", Stage: "explore",
	})
	if err != nil {
		t.Fatalf("ApproveStage explore: %v", err)
	}
	status, err = store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-spec", WorkUnit: "spec", EvidenceGoal: "spec", MaxAttempts: 3, MaxChangedLines: 400,
	})
	if err != nil {
		t.Fatalf("Begin spec (approved): %v", err)
	}
	status, err = store.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "finish-spec", Outcome: sddstatus.AttemptPassed,
		EvidenceRevision: e2eEvidence('2'), Diagnosis: "spec drafted",
		CleanupEvidence: "none", ProcessEvidence: "none", HarnessDisposition: sddstatus.HarnessReused,
	})
	if err != nil {
		t.Fatalf("Finish spec: %v", err)
	}

	// Requirement change discovered mid-chain: backtrack to explore via the
	// audited reset primitive (QA-ORCH-10's documented backtrack mechanism).
	status, err = store.Reset(ctx, sddstatus.ResetObjectiveRequest{
		ExpectedRevision: status.Revision, RequestID: "reset-to-explore",
		Reason: "requirement changed mid-chain", Actor: "maintainer",
	})
	if err != nil {
		t.Fatalf("Reset to explore: %v", err)
	}
	status, err = store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-explore-2", WorkUnit: "explore",
		EvidenceGoal: "re-discover with new requirement", MaxAttempts: 3, MaxChangedLines: 400,
	})
	if err != nil {
		t.Fatalf("Begin explore after reset: %v", err)
	}
	newEvidence := e2eEvidence('9')
	if newEvidence == originalEvidence {
		t.Fatalf("test setup error: new evidence must differ from original")
	}
	status, err = store.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "finish-explore-2", Outcome: sddstatus.AttemptPassed,
		EvidenceRevision: newEvidence, Diagnosis: "re-discovered with new requirement",
		CleanupEvidence: "none", ProcessEvidence: "none", HarnessDisposition: sddstatus.HarnessReused,
	})
	if err != nil {
		t.Fatalf("Finish explore (re-run): %v", err)
	}

	// The OLD approval for "explore" was bound to originalEvidence; it no
	// longer matches the new EvidenceRevision, so advancing into spec again
	// must be refused even though "explore" was approved once before.
	if _, err := store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-spec-stale-approval", WorkUnit: "spec",
		EvidenceGoal: "spec again", MaxAttempts: 3, MaxChangedLines: 400,
	}); err != sddstatus.ErrRuntimeStageApprovalRequired {
		t.Fatalf("expected ErrRuntimeStageApprovalRequired for approval invalidated by requirement change, got %v", err)
	}

	afterReset := e2eQAStatus(t, change, repo)
	if len(afterReset.Approvals) != 1 || afterReset.Approvals[0].ApprovalRevision != originalEvidence {
		t.Fatalf("expected the projection to still carry only the STALE original approval (%s), got %#v", originalEvidence, afterReset.Approvals)
	}

	// A fresh approval against the new evidence unblocks the same advance.
	status, err = store.ApproveStage(ctx, sddstatus.ApproveStageRequest{
		ExpectedRevision: status.Revision, RequestID: "approve-explore-2", Stage: "explore",
	})
	if err != nil {
		t.Fatalf("ApproveStage explore (re-approve): %v", err)
	}
	if _, err := store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: status.Revision, RequestID: "begin-spec-fresh-approval", WorkUnit: "spec",
		EvidenceGoal: "spec again", MaxAttempts: 3, MaxChangedLines: 400,
	}); err != nil {
		t.Fatalf("Begin spec after fresh approval of the re-run stage: %v", err)
	}
}

// TestE2ECWDResolvesRepositoryRootLikeReview mirrors the design's threat
// matrix row "Git repository selection" for qa-status: a relative --cwd, an
// absolute --cwd, and a linked worktree must all observe the identical
// replayed chain, and a non-repository --cwd must refuse outright rather
// than silently opening a store.
func TestE2ECWDResolvesRepositoryRootLikeReview(t *testing.T) {
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	initE2ERepository(t, repo)
	change := "e2e-cwd-selector"
	ledgerChange := qastage.LedgerChangeName(change)

	ctx := context.Background()
	store, err := sddstatus.OpenRuntimeStore(ctx, repo, ledgerChange)
	if err != nil {
		t.Fatalf("OpenRuntimeStore: %v", err)
	}
	store = store.WithStageVocabulary(qastage.VocabularyV1())
	if _, err := store.Begin(ctx, sddstatus.BeginAttemptRequest{
		RequestID: "begin-explore", WorkUnit: "explore", EvidenceGoal: "discover", MaxAttempts: 3, MaxChangedLines: 400,
	}); err != nil {
		t.Fatalf("Begin explore: %v", err)
	}

	absolute := e2eQAStatus(t, change, repo)
	if absolute.NextAction != sddstatus.RuntimeActionFinish {
		t.Fatalf("absolute --cwd NextAction = %q, want %q", absolute.NextAction, sddstatus.RuntimeActionFinish)
	}

	workingDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(workingDir, repo)
	if err != nil {
		t.Skipf("relative path from %s to %s unavailable on this platform: %v", workingDir, repo, err)
	}
	relativeStatus := e2eQAStatus(t, change, relative)
	if relativeStatus.NextAction != absolute.NextAction || len(relativeStatus.Approvals) != len(absolute.Approvals) {
		t.Fatalf("relative --cwd status diverged from absolute: %#v vs %#v", relativeStatus, absolute)
	}

	sibling := filepath.Join(t.TempDir(), "sibling-worktree")
	runE2EGit(t, repo, "worktree", "add", "-q", "-b", "e2e-sibling", sibling)
	worktreeStatus := e2eQAStatus(t, change, sibling)
	if worktreeStatus.NextAction != absolute.NextAction || len(worktreeStatus.Approvals) != len(absolute.Approvals) {
		t.Fatalf("linked worktree --cwd status diverged from repository root: %#v vs %#v", worktreeStatus, absolute)
	}

	nonRepo := t.TempDir()
	if err := cli.RunQAStatus([]string{"--change", change, "--cwd", nonRepo}, &bytes.Buffer{}); err == nil {
		t.Fatal("qa-status accepted a --cwd outside any git repository")
	}
}
