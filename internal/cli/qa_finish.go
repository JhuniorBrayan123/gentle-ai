package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v2/internal/qastage"
	"github.com/gentleman-programming/gentle-ai/v2/internal/sddstatus"
)

type qaFinishFlagDefinition struct {
	name, usage string
	required    bool
}

var qaFinishFlagDefinitions = []qaFinishFlagDefinition{
	{name: "change", usage: "Change name whose active attempt is finishing", required: true},
	{name: "cwd", usage: "Repository root (defaults to current directory)"},
	{name: "request-id", usage: "Idempotency key for this finish call", required: true},
	{name: "outcome", usage: "passed, failed, or interrupted", required: true},
	{name: "evidence-revision", usage: "sha256:<64 lowercase hex>; never none", required: true},
	{name: "diagnosis", usage: "Trimmed single-line text (max 500 bytes)", required: true},
	{name: "harness-disposition", usage: "reused or invalidated", required: true},
	{name: "cleanup-evidence", usage: "Trimmed single-line text (max 500 bytes)", required: true},
	{name: "process-evidence", usage: "Trimmed single-line text (max 500 bytes)", required: true},
	{name: "expected-revision", usage: "Expected current ledger revision (auto-resolved from qa-status when omitted)"},
}

// QAFinishResult is the JSON-encoded result of qa-finish.
type QAFinishResult struct {
	Change       string `json:"change"`
	LedgerChange string `json:"ledger_change"`
	Revision     string `json:"revision"`
	NextAction   string `json:"next_action"`
	Complete     bool   `json:"complete"`
}

// RunQAFinish is the CLI entry point for `gentle-ai qa-finish`.
func RunQAFinish(args []string, stdout io.Writer) error {
	return runQAFinish(context.Background(), args, stdout)
}

func runQAFinish(ctx context.Context, args []string, stdout io.Writer) error {
	if hasQAFinishHelp(args) {
		return renderQAFinishHelp(stdout)
	}
	flags := flag.NewFlagSet("qa-finish", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	change := flags.String("change", "", "")
	cwd := flags.String("cwd", "", "")
	requestID := flags.String("request-id", "", "")
	outcome := flags.String("outcome", "", "")
	evidenceRevision := flags.String("evidence-revision", "", "")
	diagnosis := flags.String("diagnosis", "", "")
	harnessDisposition := flags.String("harness-disposition", "", "")
	cleanupEvidence := flags.String("cleanup-evidence", "", "")
	processEvidence := flags.String("process-evidence", "", "")
	expectedRevision := flags.String("expected-revision", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected qa-finish argument %q", flags.Arg(0))
	}
	if strings.TrimSpace(*change) == "" {
		return errors.New("qa-finish requires --change")
	}
	if strings.TrimSpace(*requestID) == "" {
		return errors.New("qa-finish requires --request-id")
	}
	if strings.TrimSpace(*outcome) == "" {
		return errors.New("qa-finish requires --outcome")
	}
	if strings.TrimSpace(*evidenceRevision) == "" {
		return errors.New("qa-finish requires --evidence-revision")
	}
	if strings.TrimSpace(*diagnosis) == "" {
		return errors.New("qa-finish requires --diagnosis")
	}
	if strings.TrimSpace(*harnessDisposition) == "" {
		return errors.New("qa-finish requires --harness-disposition")
	}
	if strings.TrimSpace(*cleanupEvidence) == "" {
		return errors.New("qa-finish requires --cleanup-evidence")
	}
	if strings.TrimSpace(*processEvidence) == "" {
		return errors.New("qa-finish requires --process-evidence")
	}

	repoRoot := *cwd
	if repoRoot == "" {
		var err error
		repoRoot, err = filepath.Abs(".")
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
	}

	ledgerChange := qastage.LedgerChangeName(*change)
	store, err := sddstatus.OpenRuntimeStore(ctx, repoRoot, ledgerChange)
	if err != nil {
		return fmt.Errorf("open qa runtime store: %w", err)
	}
	store = store.WithStageVocabulary(qastage.VocabularyV1())

	expected := strings.TrimSpace(*expectedRevision)
	if expected == "" {
		current, statusErr := store.Status()
		if statusErr != nil {
			return fmt.Errorf("resolve current qa runtime status: %w", statusErr)
		}
		expected = current.Revision
	}

	status, err := store.Finish(ctx, sddstatus.FinishAttemptRequest{
		ExpectedRevision:   expected,
		RequestID:          *requestID,
		Outcome:            sddstatus.AttemptOutcome(*outcome),
		EvidenceRevision:   *evidenceRevision,
		Diagnosis:          *diagnosis,
		HarnessDisposition: sddstatus.HarnessDisposition(*harnessDisposition),
		CleanupEvidence:    *cleanupEvidence,
		ProcessEvidence:    *processEvidence,
	})
	if err != nil {
		return fmt.Errorf("qa-finish: %w", err)
	}

	result := QAFinishResult{
		Change:       *change,
		LedgerChange: ledgerChange,
		Revision:     status.Revision,
		NextAction:   status.NextAction,
		Complete:     status.Complete,
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func hasQAFinishHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func renderQAFinishHelp(stdout io.Writer) error {
	_, _ = fmt.Fprintln(stdout, "Usage: gentle-ai qa-finish --change <name> [--cwd <repo>] --request-id <id> --outcome <passed|failed|interrupted> --evidence-revision <sha256:...> --diagnosis <text> --harness-disposition <reused|invalidated> --cleanup-evidence <text> --process-evidence <text>")
	_, _ = fmt.Fprintln(stdout, "\nFlags:")
	for _, def := range qaFinishFlagDefinitions {
		required := ""
		if def.required {
			required = " (required)"
		}
		_, _ = fmt.Fprintf(stdout, "  --%-22s %s%s\n", def.name, def.usage, required)
	}
	_, _ = fmt.Fprintln(stdout, "\nCompletes the active QA ledger attempt for --change.")
	return nil
}
