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

// QAStatus is the JSON-encoded result of qa-status.
type QAStatus struct {
	Change       string                           `json:"change"`
	LedgerChange string                           `json:"ledger_change"`
	Stages       []sddstatus.Stage                `json:"stages"`
	Approvals    []sddstatus.RuntimeStageApproval `json:"approvals,omitempty"`
	NextAction   string                           `json:"next_action"`
	Complete     bool                             `json:"complete"`
}

// RunQAStatus is the CLI entry point for `gentle-ai qa-status`.
func RunQAStatus(args []string, stdout io.Writer) error {
	if hasQAStatusHelp(args) {
		return renderQAStatusHelp(stdout)
	}
	flags := flag.NewFlagSet("qa-status", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	change := flags.String("change", "", "Change name to query")
	cwd := flags.String("cwd", "", "Repository root (defaults to current directory)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected qa-status argument %q", flags.Arg(0))
	}
	if strings.TrimSpace(*change) == "" {
		return errors.New("qa-status requires --change")
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

	ctx := context.Background()
	store, err := sddstatus.OpenRuntimeStore(ctx, repoRoot, ledgerChange)
	if err != nil {
		return fmt.Errorf("open qa runtime store: %w", err)
	}

	runtimeStatus, err := store.Status()
	if err != nil {
		return fmt.Errorf("read qa runtime status: %w", err)
	}

	status := QAStatus{
		Change:       *change,
		LedgerChange: ledgerChange,
		Stages:       vocab.Stages,
		Approvals:    runtimeStatus.StageApprovals,
		Complete:     runtimeStatus.Complete,
		// NextAction is recomputed from replay — never cached.
		NextAction: runtimeStatus.NextAction,
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(status)
}

func hasQAStatusHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func renderQAStatusHelp(stdout io.Writer) error {
	_, _ = fmt.Fprintln(stdout, "Usage: gentle-ai qa-status --change <name> [--cwd <repo>]")
	_, _ = fmt.Fprintln(stdout, "\nFlags:")
	_, _ = fmt.Fprintln(stdout, "  --change <name>   Change name to query (required)")
	_, _ = fmt.Fprintln(stdout, "  --cwd <repo>      Repository root (optional, defaults to current directory)")
	_, _ = fmt.Fprintln(stdout, "\nOutputs JSON with current QA stage approvals and next action.")
	return nil
}
