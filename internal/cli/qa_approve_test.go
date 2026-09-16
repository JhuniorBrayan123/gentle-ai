package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/cli"
)

func TestRunQAApproveHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := cli.RunQAApprove([]string{"--help"}, &buf); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !strings.Contains(buf.String(), "qa-approve") {
		t.Errorf("expected help to mention qa-approve, got: %s", buf.String())
	}
}

func TestRunQAApproveMissingChange(t *testing.T) {
	err := cli.RunQAApprove([]string{
		"--stage", "spec", "--evidence-revision", evidenceRevisionFor("x"), "--actor", "a", "--reason", "r",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--change") {
		t.Errorf("expected --change required error, got %v", err)
	}
}

func TestRunQAApproveMissingStage(t *testing.T) {
	err := cli.RunQAApprove([]string{
		"--change", "foo", "--evidence-revision", evidenceRevisionFor("x"), "--actor", "a", "--reason", "r",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--stage") {
		t.Errorf("expected --stage required error, got %v", err)
	}
}

func TestRunQAApproveMissingEvidenceRevision(t *testing.T) {
	err := cli.RunQAApprove([]string{
		"--change", "foo", "--stage", "spec", "--actor", "a", "--reason", "r",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--evidence-revision") {
		t.Errorf("expected --evidence-revision required error, got %v", err)
	}
}

func TestRunQAApproveMissingActor(t *testing.T) {
	err := cli.RunQAApprove([]string{
		"--change", "foo", "--stage", "spec", "--evidence-revision", evidenceRevisionFor("x"), "--reason", "r",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--actor") {
		t.Errorf("expected --actor required error, got %v", err)
	}
}

func TestRunQAApproveMissingReason(t *testing.T) {
	err := cli.RunQAApprove([]string{
		"--change", "foo", "--stage", "spec", "--evidence-revision", evidenceRevisionFor("x"), "--actor", "a",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--reason") {
		t.Errorf("expected --reason required error, got %v", err)
	}
}

func TestRunQAApproveSuccessfulAfterSpecFinish(t *testing.T) {
	repo := initQABeginRepo(t)
	change := "qa-approve-success"

	mustQABegin(t, repo, change, "explore", "begin-explore", "discover")
	mustQAFinish(t, repo, change, "finish-explore", evidenceRevisionFor("explore"))
	mustQABegin(t, repo, change, "spec", "begin-spec", "write spec")
	specEvidence := evidenceRevisionFor("spec")
	mustQAFinish(t, repo, change, "finish-spec", specEvidence)

	result := mustQAApprove(t, repo, change, "spec", specEvidence)
	if result.Stage != "spec" || result.Actor != "test-actor" {
		t.Errorf("unexpected result: %#v", result)
	}

	// The recorded approval now unblocks apply -- this is the actual gated
	// transition the whole change exists to prove.
	beginResult := mustQABegin(t, repo, change, "apply", "begin-apply", "implement")
	if beginResult.Stage != "apply" {
		t.Errorf("expected apply to begin after approval, got %#v", beginResult)
	}
}

// TestRunQAApproveStaleEvidenceRevisionRefusesCleanly proves qa-approve fails
// fast with a clear message when the caller's --evidence-revision no longer
// matches the ledger's current evidence (e.g. the stage was re-run since the
// caller last read it), instead of silently approving against wrong evidence.
func TestRunQAApproveStaleEvidenceRevisionRefusesCleanly(t *testing.T) {
	repo := initQABeginRepo(t)
	change := "qa-approve-stale-evidence"

	mustQABegin(t, repo, change, "explore", "begin-explore", "discover")
	mustQAFinish(t, repo, change, "finish-explore", evidenceRevisionFor("explore"))
	mustQABegin(t, repo, change, "spec", "begin-spec", "write spec")
	mustQAFinish(t, repo, change, "finish-spec", evidenceRevisionFor("spec"))

	err := cli.RunQAApprove([]string{
		"--change", change, "--cwd", repo, "--stage", "spec",
		"--evidence-revision", evidenceRevisionFor("stale-and-wrong"),
		"--actor", "test-actor", "--reason", "test approval",
	}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "does not match the ledger's current evidence revision") {
		t.Fatalf("expected stale evidence-revision refusal, got %v", err)
	}
}
