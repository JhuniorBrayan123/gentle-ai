package qastage_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/qastage"
	"github.com/gentleman-programming/gentle-ai/v2/internal/sddstatus"
)

func TestVocabularyV1(t *testing.T) {
	vocab := qastage.VocabularyV1()
	if vocab.ID != qastage.VocabularyID {
		t.Errorf("expected ID %q, got %q", qastage.VocabularyID, vocab.ID)
	}
	if err := vocab.Validate(); err != nil {
		t.Fatalf("VocabularyV1 validation failed: %v", err)
	}
	expectedStages := []string{"explore", "spec", "apply", "verify", "docs"}
	if len(vocab.Stages) != len(expectedStages) {
		t.Fatalf("expected %d stages, got %d", len(expectedStages), len(vocab.Stages))
	}
	for i, expected := range expectedStages {
		if vocab.Stages[i].Label != expected {
			t.Errorf("stage %d: expected %q, got %q", i, expected, vocab.Stages[i].Label)
		}
	}
}

func TestLedgerChangeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"whitespace", "   ", ""},
		{"valid", "foo-bar", "qa_foo-bar"},
		{"with spaces", "foo bar", "qa_foo bar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := qastage.LedgerChangeName(tc.input)
			if result != tc.expected {
				t.Errorf("LedgerChangeName(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
			if result != "" {
				if err := sddstatus.ValidateRuntimeText(result, 128); err != nil {
					t.Errorf("expected LedgerChangeName output %q to pass ValidateRuntimeText: %v", result, err)
				}
			}
		})
	}
}

// TestLedgerChangeNamePassesRealRuntimeStoreValidation pins a regression
// found by the PR13 E2E test: an earlier "qa--" (double-hyphen) prefix
// passed the lenient ValidateRuntimeText check above but was refused outright
// by OpenRuntimeStore's actual change-name validator, which rejects
// consecutive separators. This test opens a REAL RuntimeStore against the
// prefixed name for an ordinary hyphenated change, the exact codepath every
// qa-status/qa-validate invocation exercises.
func TestLedgerChangeNamePassesRealRuntimeStoreValidation(t *testing.T) {
	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init", "-q", "--initial-branch=main")
	run("config", "user.name", "Test")
	run("config", "user.email", "test@example.com")
	run("commit", "-q", "--allow-empty", "-m", "init")

	ledgerChange := qastage.LedgerChangeName("my-feature-name")
	if _, err := sddstatus.OpenRuntimeStore(context.Background(), repo, ledgerChange); err != nil {
		t.Fatalf("OpenRuntimeStore(%q) rejected a LedgerChangeName-produced identity: %v", ledgerChange, err)
	}
}
