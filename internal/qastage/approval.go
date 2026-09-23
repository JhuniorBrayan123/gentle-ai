package qastage

import (
	"context"
	"errors"
)

// Approval records that a human (or authorized agent) reviewed a completed
// stage's output. Only "apply" requires one today — of its predecessor
// "spec" — matching v2's requires_approval intent, now actually enforced on
// every attempt to begin apply, not only when the ledger happens to already
// have prior history (baseline Escenario 2).
type Approval struct {
	Stage  string `json:"stage"`
	Actor  string `json:"actor"`
	Reason string `json:"reason"`
}

var (
	ErrApprovalRequired = errors.New("qastage: apply requires a recorded approval of spec before it can begin")
	ErrNothingToApprove = errors.New("qastage: stage has not completed yet, there is nothing to approve")
)

const stageRequiringApproval = "apply"
const approvedPredecessorStage = "spec"

func isApproved(state ledgerState, stage string) bool {
	for _, approval := range state.Approvals {
		if approval.Stage == stage {
			return true
		}
	}
	return false
}

func stageHasCompleted(state ledgerState, stage string) bool {
	for _, attempt := range state.Attempts {
		if attempt.Stage == stage && attempt.Outcome == OutcomePassed {
			return true
		}
	}
	return false
}

// Approve records an approval for stage, which must have already completed
// (Outcome == passed) at least once. It does not yet bind to a specific
// artifact_revision — that invalidation-on-change behavior is added in 3A.8.
func (m *QAStateMachine) Approve(ctx context.Context, change, stage, actor, reason string) (Approval, error) {
	state, head, err := m.readState(ctx, change)
	if err != nil {
		return Approval{}, err
	}
	if !stageHasCompleted(state, stage) {
		return Approval{}, ErrNothingToApprove
	}

	approval := Approval{Stage: stage, Actor: actor, Reason: reason}
	state.Approvals = append(state.Approvals, approval)

	if err := m.commit(ctx, change, head.Revision, state); err != nil {
		return Approval{}, err
	}
	return approval, nil
}
