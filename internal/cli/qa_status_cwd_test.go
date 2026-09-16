package cli_test

// Threat-matrix --cwd tests for qa-status (Phase 11 / PR13). The design's
// threat matrix names "Git repository selection" as Applicable for the new
// qa-* CLI verbs and requires one test per selector: relative --cwd,
// absolute --cwd, non-repo --cwd, and a linked worktree (which must observe
// the same chain). qa_cli_test.go (PR7) covered flag validation only; these
// selector-shape cases were not yet covered by any test, so they are added
// here rather than duplicated.
import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/cli"
)

func initQAStatusCWDRepo(t *testing.T) string {
	t.Helper()
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runQAStatusCWDGit(t, repo, "init", "-q", "--initial-branch=main")
	runQAStatusCWDGit(t, repo, "config", "user.name", "Test")
	runQAStatusCWDGit(t, repo, "config", "user.email", "test@example.com")
	runQAStatusCWDGit(t, repo, "commit", "-q", "--allow-empty", "-m", "init")
	return repo
}

func runQAStatusCWDGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func qaStatusOutput(t *testing.T, args []string) (cli.QAStatus, string) {
	t.Helper()
	var buf bytes.Buffer
	if err := cli.RunQAStatus(args, &buf); err != nil {
		t.Fatalf("qa-status %v: %v", args, err)
	}
	var status cli.QAStatus
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		t.Fatalf("bad qa-status JSON for %v: %v\n%s", args, err, buf.String())
	}
	return status, buf.String()
}

// TestQAStatusCWDAbsoluteRelativeAndWorktreeAgree pins the threat-matrix row:
// a relative --cwd, an absolute --cwd, and a linked worktree all resolve to
// the identical repository Git common-dir and therefore observe the same
// ledger chain.
func TestQAStatusCWDAbsoluteRelativeAndWorktreeAgree(t *testing.T) {
	repo := initQAStatusCWDRepo(t)
	change := "cwd-selector-change"

	absolute, absoluteRaw := qaStatusOutput(t, []string{"--change", change, "--cwd", repo})

	workingDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	if relative, err := filepath.Rel(workingDir, repo); err == nil {
		relativeStatus, relativeRaw := qaStatusOutput(t, []string{"--change", change, "--cwd", relative})
		if relativeRaw != absoluteRaw {
			t.Fatalf("relative --cwd qa-status diverged from absolute:\nabsolute: %s\nrelative: %s", absoluteRaw, relativeRaw)
		}
		if relativeStatus.LedgerChange != absolute.LedgerChange {
			t.Fatalf("relative --cwd LedgerChange = %q, want %q", relativeStatus.LedgerChange, absolute.LedgerChange)
		}
	} else {
		t.Logf("relative path from %s to %s unavailable on this platform: %v", workingDir, repo, err)
	}

	sibling := filepath.Join(t.TempDir(), "sibling-worktree")
	runQAStatusCWDGit(t, repo, "worktree", "add", "-q", "-b", "qa-status-cwd-sibling", sibling)
	worktreeStatus, worktreeRaw := qaStatusOutput(t, []string{"--change", change, "--cwd", sibling})
	if worktreeRaw != absoluteRaw {
		t.Fatalf("linked worktree --cwd qa-status diverged from repository root:\nroot:      %s\nworktree:  %s", absoluteRaw, worktreeRaw)
	}
	if worktreeStatus.NextAction != absolute.NextAction {
		t.Fatalf("linked worktree NextAction = %q, want %q", worktreeStatus.NextAction, absolute.NextAction)
	}
}

// TestQAStatusCWDNonRepositoryRefused pins the threat-matrix requirement that
// a non-repo --cwd refuses rather than creating a store.
func TestQAStatusCWDNonRepositoryRefused(t *testing.T) {
	nonRepo := t.TempDir()
	var buf bytes.Buffer
	if err := cli.RunQAStatus([]string{"--change", "any-change", "--cwd", nonRepo}, &buf); err == nil {
		t.Fatalf("qa-status accepted a --cwd outside any git repository: %s", buf.String())
	}
}

// TestQAStatusCWDDefaultsToCurrentDirectory pins that omitting --cwd resolves
// against the process's current working directory, exactly like --cwd <cwd>.
func TestQAStatusCWDDefaultsToCurrentDirectory(t *testing.T) {
	repo := initQAStatusCWDRepo(t)
	change := "cwd-default-change"

	explicit, explicitRaw := qaStatusOutput(t, []string{"--change", change, "--cwd", repo})

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatal(err)
		}
	}()

	defaulted, defaultedRaw := qaStatusOutput(t, []string{"--change", change})
	if defaultedRaw != explicitRaw {
		t.Fatalf("default --cwd qa-status diverged from explicit --cwd:\nexplicit: %s\ndefault:  %s", explicitRaw, defaultedRaw)
	}
	if defaulted.LedgerChange != explicit.LedgerChange {
		t.Fatalf("default --cwd LedgerChange = %q, want %q", defaulted.LedgerChange, explicit.LedgerChange)
	}
}
