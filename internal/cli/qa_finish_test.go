package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v2/internal/sddstatus"
)

func TestRunQAFinishHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := cli.RunQAFinish([]string{"--help"}, &buf); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !strings.Contains(buf.String(), "qa-finish") {
		t.Errorf("expected help to mention qa-finish, got: %s", buf.String())
	}
}

func TestRunQAFinishMissingChange(t *testing.T) {
	err := cli.RunQAFinish([]string{
		"--request-id", "r1", "--outcome", "passed", "--evidence-revision", evidenceRevisionFor("x"),
		"--diagnosis", "d", "--harness-disposition", "reused", "--cleanup-evidence", "c", "--process-evidence", "p",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--change") {
		t.Errorf("expected --change required error, got %v", err)
	}
}

func TestRunQAFinishMissingOutcome(t *testing.T) {
	err := cli.RunQAFinish([]string{
		"--change", "foo", "--request-id", "r1", "--evidence-revision", evidenceRevisionFor("x"),
		"--diagnosis", "d", "--harness-disposition", "reused", "--cleanup-evidence", "c", "--process-evidence", "p",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--outcome") {
		t.Errorf("expected --outcome required error, got %v", err)
	}
}

func TestRunQAFinishMissingEvidenceRevision(t *testing.T) {
	err := cli.RunQAFinish([]string{
		"--change", "foo", "--request-id", "r1", "--outcome", "passed",
		"--diagnosis", "d", "--harness-disposition", "reused", "--cleanup-evidence", "c", "--process-evidence", "p",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--evidence-revision") {
		t.Errorf("expected --evidence-revision required error, got %v", err)
	}
}

func TestRunQAFinishMissingDiagnosis(t *testing.T) {
	err := cli.RunQAFinish([]string{
		"--change", "foo", "--request-id", "r1", "--outcome", "passed", "--evidence-revision", evidenceRevisionFor("x"),
		"--harness-disposition", "reused", "--cleanup-evidence", "c", "--process-evidence", "p",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--diagnosis") {
		t.Errorf("expected --diagnosis required error, got %v", err)
	}
}

func TestRunQAFinishSuccessfulAfterBegin(t *testing.T) {
	repo := initQABeginRepo(t)
	change := "qa-finish-success"
	mustQABegin(t, repo, change, "explore", "begin-explore", "discover")

	result := mustQAFinish(t, repo, change, "finish-explore", evidenceRevisionFor("explore"))
	if result.Change != change {
		t.Errorf("unexpected result: %#v", result)
	}
	if result.NextAction != sddstatus.RuntimeActionComplete {
		t.Errorf("NextAction = %q, want %q", result.NextAction, sddstatus.RuntimeActionComplete)
	}
	if !result.Complete {
		t.Errorf("expected Complete=true after finishing the only objective, got false")
	}
}

// TestRunQAFinishWithoutActiveAttemptRefusesCleanly proves qa-finish surfaces
// ErrRuntimeNoActiveAttempt as a clean wrapped error, not a crash, when there
// is nothing to finish.
func TestRunQAFinishWithoutActiveAttemptRefusesCleanly(t *testing.T) {
	repo := initQABeginRepo(t)
	change := "qa-finish-no-active-attempt"

	// A revision only exists once at least one attempt has been recorded, so
	// the "no active attempt" precondition is exercised by finishing once
	// (closing the only open attempt) and then finishing again with nothing
	// active, rather than against a genesis ledger with no revision at all.
	mustQABegin(t, repo, change, "explore", "begin-explore", "discover")
	mustQAFinish(t, repo, change, "finish-explore", evidenceRevisionFor("explore"))

	err := cli.RunQAFinish([]string{
		"--change", change, "--cwd", repo, "--request-id", "finish-nothing",
		"--outcome", "passed", "--evidence-revision", evidenceRevisionFor("x"),
		"--diagnosis", "d", "--harness-disposition", "reused",
		"--cleanup-evidence", "c", "--process-evidence", "p",
	}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected ErrRuntimeNoActiveAttempt, got nil")
	}
	if !errors.Is(err, sddstatus.ErrRuntimeNoActiveAttempt) {
		t.Fatalf("expected ErrRuntimeNoActiveAttempt, got %v", err)
	}
}
