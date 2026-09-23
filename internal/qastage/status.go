package qastage

import "context"

// Status is the read-only, always-recalculated view of a change's ledger —
// never cached (design decision 1.1 / baseline Escenario 8, preserved).
type Status struct {
	Change     string
	Revision   string
	NextAction string // "begin" | "finish" | "complete"
	Complete   bool
}

// Status recomputes change's current status directly from its history.
func (m *QAStateMachine) Status(ctx context.Context, change string) (Status, error) {
	state, head, err := m.readState(ctx, change)
	if err != nil {
		return Status{}, err
	}

	status := Status{Change: change, Revision: head.Revision}

	if len(state.Attempts) > 0 && state.Attempts[len(state.Attempts)-1].Outcome == OutcomeRunning {
		status.NextAction = "finish"
		return status, nil
	}

	if nextExpectedStage(state) == "" {
		status.NextAction = "complete"
		status.Complete = true
		return status, nil
	}

	status.NextAction = "begin"
	return status, nil
}
