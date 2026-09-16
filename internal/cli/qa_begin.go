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

type qaBeginFlagDefinition struct {
	name, usage string
	required    bool
}

var qaBeginFlagDefinitions = []qaBeginFlagDefinition{
	{name: "change", usage: "Change name to begin/advance", required: true},
	{name: "stage", usage: "Stage label to begin (explore|spec|apply|verify|docs)", required: true},
	{name: "cwd", usage: "Repository root (defaults to current directory)"},
	{name: "request-id", usage: "Idempotency key for this begin call", required: true},
	{name: "evidence-goal", usage: "Single-line objective for this stage", required: true},
	{name: "expected-revision", usage: "Expected current ledger revision (auto-resolved from qa-status when omitted)"},
	{name: "max-attempts", usage: "Optional attempt limit (default 2)"},
	{name: "max-changed-lines", usage: "Optional changed-line limit (default 200)"},
}

// QABeginResult is the JSON-encoded result of qa-begin.
type QABeginResult struct {
	Change       string `json:"change"`
	LedgerChange string `json:"ledger_change"`
	Stage        string `json:"stage"`
	Revision     string `json:"revision"`
	NextAction   string `json:"next_action"`
	Complete     bool   `json:"complete"`
}

// RunQABegin is the CLI entry point for `gentle-ai qa-begin`.
func RunQABegin(args []string, stdout io.Writer) error {
	return runQABegin(context.Background(), args, stdout)
}

func runQABegin(ctx context.Context, args []string, stdout io.Writer) error {
	if hasQABeginHelp(args) {
		return renderQABeginHelp(stdout)
	}
	flags := flag.NewFlagSet("qa-begin", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	change := flags.String("change", "", "")
	stage := flags.String("stage", "", "")
	cwd := flags.String("cwd", "", "")
	requestID := flags.String("request-id", "", "")
	evidenceGoal := flags.String("evidence-goal", "", "")
	expectedRevision := flags.String("expected-revision", "", "")
	maxAttempts := flags.Int("max-attempts", 0, "")
	maxChangedLines := flags.Int("max-changed-lines", 0, "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected qa-begin argument %q", flags.Arg(0))
	}
	if strings.TrimSpace(*change) == "" {
		return errors.New("qa-begin requires --change")
	}
	if strings.TrimSpace(*stage) == "" {
		return errors.New("qa-begin requires --stage")
	}
	if strings.TrimSpace(*requestID) == "" {
		return errors.New("qa-begin requires --request-id")
	}
	if strings.TrimSpace(*evidenceGoal) == "" {
		return errors.New("qa-begin requires --evidence-goal")
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
	vocab := qastage.VocabularyV1()
	// The vocabulary's own Position lookup is only enforced by the ledger on
	// an ADVANCING begin (runtimeObjectiveAdvanceAdmissible); a fresh first
	// begin silently accepts an unknown label (StagePosition defaults to 0).
	// Refusing an unknown stage label here, before ever opening the store,
	// closes that gap for every begin, not only advancing ones.
	if _, err := vocab.Position(*stage); err != nil {
		return fmt.Errorf("qa-begin: unknown stage %q; want one of %s", *stage, qaKnownStageLabels(vocab))
	}

	store, err := sddstatus.OpenRuntimeStore(ctx, repoRoot, ledgerChange)
	if err != nil {
		return fmt.Errorf("open qa runtime store: %w", err)
	}
	store = store.WithStageVocabulary(vocab)

	expected := strings.TrimSpace(*expectedRevision)
	if expected == "" {
		// qa-begin auto-resolves the CAS token from current status so callers
		// (the qa-* skills) never have to plumb a manual revision through an
		// extra qa-status round trip; --expected-revision remains available
		// for a caller that already holds one from a prior response.
		current, statusErr := store.Status()
		if statusErr != nil {
			return fmt.Errorf("resolve current qa runtime status: %w", statusErr)
		}
		expected = current.Revision
	}

	status, err := store.Begin(ctx, sddstatus.BeginAttemptRequest{
		ExpectedRevision: expected,
		RequestID:        *requestID,
		WorkUnit:         *stage,
		EvidenceGoal:     *evidenceGoal,
		MaxAttempts:      *maxAttempts,
		MaxChangedLines:  *maxChangedLines,
	})
	if err != nil {
		return fmt.Errorf("qa-begin: %w", err)
	}

	result := QABeginResult{
		Change:       *change,
		LedgerChange: ledgerChange,
		Stage:        *stage,
		Revision:     status.Revision,
		NextAction:   status.NextAction,
		Complete:     status.Complete,
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func qaKnownStageLabels(vocab sddstatus.StageVocabulary) string {
	labels := make([]string, 0, len(vocab.Stages))
	for _, stage := range vocab.Stages {
		labels = append(labels, stage.Label)
	}
	return strings.Join(labels, ", ")
}

func hasQABeginHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func renderQABeginHelp(stdout io.Writer) error {
	_, _ = fmt.Fprintln(stdout, "Usage: gentle-ai qa-begin --change <name> --stage <label> [--cwd <repo>] --request-id <id> --evidence-goal <text> [--expected-revision <rev>]")
	_, _ = fmt.Fprintln(stdout, "\nFlags:")
	for _, def := range qaBeginFlagDefinitions {
		required := ""
		if def.required {
			required = " (required)"
		}
		_, _ = fmt.Fprintf(stdout, "  --%-22s %s%s\n", def.name, def.usage, required)
	}
	_, _ = fmt.Fprintln(stdout, "\nOpens or advances the QA ledger for --change into --stage. Refuses with a")
	_, _ = fmt.Fprintln(stdout, "clean error (not a crash) on out-of-order, unknown, or unapproved stages.")
	return nil
}
