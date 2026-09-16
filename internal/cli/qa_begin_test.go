package cli_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v2/internal/sddstatus"
)

func initQABeginRepo(t *testing.T) string {
	t.Helper()
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runQABeginGit(t, repo, "init", "-q", "--initial-branch=main")
	runQABeginGit(t, repo, "config", "user.name", "Test")
	runQABeginGit(t, repo, "config", "user.email", "test@example.com")
	runQABeginGit(t, repo, "commit", "-q", "--allow-empty", "-m", "init")
	return repo
}

func runQABeginGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func TestRunQABeginHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := cli.RunQABegin([]string{"--help"}, &buf); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !strings.Contains(buf.String(), "qa-begin") {
		t.Errorf("expected help to mention qa-begin, got: %s", buf.String())
	}
}

func TestRunQABeginMissingChange(t *testing.T) {
	err := cli.RunQABegin([]string{"--stage", "explore", "--request-id", "r1", "--evidence-goal", "goal"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--change") {
		t.Errorf("expected --change required error, got %v", err)
	}
}

func TestRunQABeginMissingStage(t *testing.T) {
	err := cli.RunQABegin([]string{"--change", "foo", "--request-id", "r1", "--evidence-goal", "goal"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--stage") {
		t.Errorf("expected --stage required error, got %v", err)
	}
}

func TestRunQABeginMissingRequestID(t *testing.T) {
	err := cli.RunQABegin([]string{"--change", "foo", "--stage", "explore", "--evidence-goal", "goal"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--request-id") {
		t.Errorf("expected --request-id required error, got %v", err)
	}
}

func TestRunQABeginMissingEvidenceGoal(t *testing.T) {
	err := cli.RunQABegin([]string{"--change", "foo", "--stage", "explore", "--request-id", "r1"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--evidence-goal") {
		t.Errorf("expected --evidence-goal required error, got %v", err)
	}
}

func TestRunQABeginUnknownStageRefused(t *testing.T) {
	repo := initQABeginRepo(t)
	err := cli.RunQABegin([]string{
		"--change", "qa-begin-unknown-stage", "--stage", "not-a-real-stage", "--cwd", repo,
		"--request-id", "begin-1", "--evidence-goal", "goal",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unknown stage") {
		t.Fatalf("expected unknown-stage refusal, got %v", err)
	}
}

func TestRunQABeginSuccessfulFirstStage(t *testing.T) {
	repo := initQABeginRepo(t)
	var buf bytes.Buffer
	err := cli.RunQABegin([]string{
		"--change", "qa-begin-success", "--stage", "explore", "--cwd", repo,
		"--request-id", "begin-explore", "--evidence-goal", "discover requirements",
	}, &buf)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	var result cli.QABeginResult
	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if result.Stage != "explore" || result.Change != "qa-begin-success" {
		t.Errorf("unexpected result: %#v", result)
	}
	if result.NextAction != sddstatus.RuntimeActionFinish {
		t.Errorf("NextAction = %q, want %q", result.NextAction, sddstatus.RuntimeActionFinish)
	}
}

// TestRunQABeginApplyWithoutApprovalRefusesCleanly pins the hard blocker this
// whole CLI exists to prove: advancing into "apply" without a prior
// "qa-approve" for "spec" must surface ErrRuntimeStageApprovalRequired as a
// clean, wrapped error -- never a panic or stacktrace.
func TestRunQABeginApplyWithoutApprovalRefusesCleanly(t *testing.T) {
	repo := initQABeginRepo(t)
	change := "qa-begin-unapproved-apply"

	mustQABegin(t, repo, change, "explore", "begin-explore", "discover")
	mustQAFinish(t, repo, change, "finish-explore", evidenceRevisionFor("explore"))
	mustQABegin(t, repo, change, "spec", "begin-spec", "write spec")
	mustQAFinish(t, repo, change, "finish-spec", evidenceRevisionFor("spec"))

	err := cli.RunQABegin([]string{
		"--change", change, "--stage", "apply", "--cwd", repo,
		"--request-id", "begin-apply-unapproved", "--evidence-goal", "implement",
	}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected ErrRuntimeStageApprovalRequired, got nil")
	}
	if !errors.Is(err, sddstatus.ErrRuntimeStageApprovalRequired) {
		t.Fatalf("expected ErrRuntimeStageApprovalRequired, got %v", err)
	}
}

// --- shared helpers used by qa_begin_test.go / qa_finish_test.go / qa_approve_test.go ---

func evidenceRevisionFor(label string) string {
	// A stable, distinct, well-shaped sha256 per stage label so cross-stage
	// evidence-revision chaining in tests never accidentally collides. Each
	// byte of the label is folded into the 0-9a-f alphabet so the result is
	// always exactly 64 valid lowercase-hex characters.
	const hexAlphabet = "0123456789abcdef"
	digest := make([]byte, 64)
	for i := range digest {
		digest[i] = '0'
	}
	for i := 0; i < len(label); i++ {
		digest[i] = hexAlphabet[int(label[i])%len(hexAlphabet)]
	}
	return "sha256:" + string(digest)
}

func mustQABegin(t *testing.T, repo, change, stage, requestID, evidenceGoal string) cli.QABeginResult {
	t.Helper()
	var buf bytes.Buffer
	if err := cli.RunQABegin([]string{
		"--change", change, "--stage", stage, "--cwd", repo,
		"--request-id", requestID, "--evidence-goal", evidenceGoal,
	}, &buf); err != nil {
		t.Fatalf("qa-begin %s/%s: %v", change, stage, err)
	}
	var result cli.QABeginResult
	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		t.Fatalf("bad qa-begin JSON: %v", err)
	}
	return result
}

func mustQAFinish(t *testing.T, repo, change, requestID, evidenceRevision string) cli.QAFinishResult {
	t.Helper()
	var buf bytes.Buffer
	if err := cli.RunQAFinish([]string{
		"--change", change, "--cwd", repo, "--request-id", requestID,
		"--outcome", "passed", "--evidence-revision", evidenceRevision,
		"--diagnosis", "done", "--harness-disposition", "reused",
		"--cleanup-evidence", "none", "--process-evidence", "none",
	}, &buf); err != nil {
		t.Fatalf("qa-finish %s: %v", change, err)
	}
	var result cli.QAFinishResult
	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		t.Fatalf("bad qa-finish JSON: %v", err)
	}
	return result
}

func mustQAApprove(t *testing.T, repo, change, stage, evidenceRevision string) cli.QAApproveResult {
	t.Helper()
	var buf bytes.Buffer
	if err := cli.RunQAApprove([]string{
		"--change", change, "--cwd", repo, "--stage", stage,
		"--evidence-revision", evidenceRevision, "--actor", "test-actor", "--reason", "test approval",
	}, &buf); err != nil {
		t.Fatalf("qa-approve %s/%s: %v", change, stage, err)
	}
	var result cli.QAApproveResult
	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		t.Fatalf("bad qa-approve JSON: %v", err)
	}
	return result
}
