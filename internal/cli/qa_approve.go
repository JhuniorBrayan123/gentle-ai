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

type qaApproveFlagDefinition struct {
	name, usage string
	required    bool
}

var qaApproveFlagDefinitions = []qaApproveFlagDefinition{
	{name: "change", usage: "Change name whose stage is being approved", required: true},
	{name: "cwd", usage: "Repository root (defaults to current directory)"},
	{name: "stage", usage: "Completed stage label to approve (e.g. spec)", required: true},
	{name: "evidence-revision", usage: "sha256:<64 lowercase hex> the approval must bind (the completed stage's own evidence revision)", required: true},
	{name: "actor", usage: "Human or agent identifier recording this approval", required: true},
	{name: "reason", usage: "Trimmed single-line text explaining the approval", required: true},
	{name: "request-id", usage: "Optional idempotency key (defaults to approve-<stage>)"},
	{name: "expected-revision", usage: "Expected current ledger revision (auto-resolved from qa-status when omitted)"},
}

// QAApproveResult is the JSON-encoded result of qa-approve.
type QAApproveResult struct {
	Change       string `json:"change"`
	LedgerChange string `json:"ledger_change"`
	Stage        string `json:"stage"`
	Actor        string `json:"actor"`
	Reason       string `json:"reason"`
	Revision     string `json:"revision"`
	NextAction   string `json:"next_action"`
}

// RunQAApprove is the CLI entry point for `gentle-ai qa-approve`.
func RunQAApprove(args []string, stdout io.Writer) error {
	return runQAApprove(context.Background(), args, stdout)
}

func runQAApprove(ctx context.Context, args []string, stdout io.Writer) error {
	if hasQAApproveHelp(args) {
		return renderQAApproveHelp(stdout)
	}
	flags := flag.NewFlagSet("qa-approve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	change := flags.String("change", "", "")
	cwd := flags.String("cwd", "", "")
	stage := flags.String("stage", "", "")
	evidenceRevision := flags.String("evidence-revision", "", "")
	actor := flags.String("actor", "", "")
	reason := flags.String("reason", "", "")
	requestID := flags.String("request-id", "", "")
	expectedRevision := flags.String("expected-revision", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected qa-approve argument %q", flags.Arg(0))
	}
	if strings.TrimSpace(*change) == "" {
		return errors.New("qa-approve requires --change")
	}
	if strings.TrimSpace(*stage) == "" {
		return errors.New("qa-approve requires --stage")
	}
	if strings.TrimSpace(*evidenceRevision) == "" {
		return errors.New("qa-approve requires --evidence-revision")
	}
	if strings.TrimSpace(*actor) == "" {
		return errors.New("qa-approve requires --actor")
	}
	if strings.TrimSpace(*reason) == "" {
		return errors.New("qa-approve requires --reason")
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

	current, err := store.Status()
	if err != nil {
		return fmt.Errorf("resolve current qa runtime status: %w", err)
	}

	// ApproveStageRequest (internal/sddstatus) records the approval bound to
	// the ledger's OWN current EvidenceRevision automatically; it carries no
	// EvidenceRevision/Actor/Reason fields of its own (see design #3281 D4 vs.
	// the shipped runtime_ledger.go). --evidence-revision is therefore
	// enforced HERE, as a fail-fast precondition, so a caller passing a stale
	// evidence revision (e.g. from an Engram artifact that predates a
	// re-run of the stage) is refused with a clear message instead of
	// silently approving the wrong evidence. --actor/--reason are accepted
	// and echoed in this command's own JSON result for the caller's audit
	// trail, but the underlying ledger record does not yet persist them —
	// this is a known, documented gap (see apply-progress), not a silent
	// drop.
	trimmedEvidence := strings.TrimSpace(*evidenceRevision)
	if current.EvidenceRevision != trimmedEvidence {
		return fmt.Errorf("qa-approve: --evidence-revision %q does not match the ledger's current evidence revision %q for stage %q; re-run qa-status to confirm the latest completed evidence before approving", trimmedEvidence, current.EvidenceRevision, *stage)
	}

	requestIDValue := strings.TrimSpace(*requestID)
	if requestIDValue == "" {
		requestIDValue = "approve-" + *stage
	}

	expected := strings.TrimSpace(*expectedRevision)
	if expected == "" {
		expected = current.Revision
	}

	status, err := store.ApproveStage(ctx, sddstatus.ApproveStageRequest{
		ExpectedRevision: expected,
		RequestID:        requestIDValue,
		Stage:            *stage,
	})
	if err != nil {
		return fmt.Errorf("qa-approve: %w", err)
	}

	result := QAApproveResult{
		Change:       *change,
		LedgerChange: ledgerChange,
		Stage:        *stage,
		Actor:        *actor,
		Reason:       *reason,
		Revision:     status.Revision,
		NextAction:   status.NextAction,
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func hasQAApproveHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func renderQAApproveHelp(stdout io.Writer) error {
	_, _ = fmt.Fprintln(stdout, "Usage: gentle-ai qa-approve --change <name> [--cwd <repo>] --stage <label> --evidence-revision <sha256:...> --actor <name> --reason <text>")
	_, _ = fmt.Fprintln(stdout, "\nFlags:")
	for _, def := range qaApproveFlagDefinitions {
		required := ""
		if def.required {
			required = " (required)"
		}
		_, _ = fmt.Fprintf(stdout, "  --%-22s %s%s\n", def.name, def.usage, required)
	}
	_, _ = fmt.Fprintln(stdout, "\nRecords an audited approval for --stage bound to --evidence-revision.")
	return nil
}
