package qastage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// AttemptOutcome is the result of a finished (or in-progress) QA stage
// attempt. Only OutcomeRunning is used by 3A.4; OutcomePassed/Failed/
// Interrupted semantics beyond "advance vs. don't advance" are completed in
// 3A.6.
type AttemptOutcome string

const (
	OutcomeRunning     AttemptOutcome = "running"
	OutcomePassed      AttemptOutcome = "passed"
	OutcomeFailed      AttemptOutcome = "failed"
	OutcomeInterrupted AttemptOutcome = "interrupted"
)

// Attempt is one begin/finish cycle for a single QA stage. ResetBy/ResetReason
// are populated only when the attempt was closed via Reset (an audited manual
// recovery), not via an ordinary Finish call.
type Attempt struct {
	Ordinal     int            `json:"ordinal"`
	Stage       string         `json:"stage"`
	Outcome     AttemptOutcome `json:"outcome"`
	ResetBy     string         `json:"reset_by,omitempty"`
	ResetReason string         `json:"reset_reason,omitempty"`
}

// ledgerState is the JSON persisted in a QAStateStore Record for a change.
type ledgerState struct {
	Attempts  []Attempt  `json:"attempts"`
	Approvals []Approval `json:"approvals,omitempty"`
}

// Errors returned by QAStateMachine. These replace v2's
// ErrRuntimeStageOutOfOrder/ErrRuntimeStageApprovalRequired with a single,
// v3-native vocabulary (design decision 1.1: the semantics are preserved,
// the exact v2 names are not).
var (
	ErrStageOutOfOrder        = errors.New("qastage: stage is not the vocabulary's first stage or the immediate successor of the last completed stage")
	ErrAttemptAlreadyActive   = errors.New("qastage: change already has an active (unfinished) attempt")
	ErrNoActiveAttempt        = errors.New("qastage: change has no active attempt to finish")
	ErrInvalidOutcome         = errors.New("qastage: outcome must be passed, failed, or interrupted")
	ErrNoActiveAttemptToReset = errors.New("qastage: change has no active attempt to reset")
)

// QAStateMachine implements the fixed 5-stage QA vocabulary
// (explore/spec/apply/verify/docs) on top of a QAStateStore. Unlike v2's
// configurable StageVocabulary, the sequence is fixed by design (1.1).
type QAStateMachine struct {
	store QAStateStore
}

// NewQAStateMachine returns a state machine persisting through store.
func NewQAStateMachine(store QAStateStore) *QAStateMachine {
	return &QAStateMachine{store: store}
}

func (m *QAStateMachine) readState(ctx context.Context, change string) (ledgerState, Record, error) {
	head, err := m.store.Head(ctx, change)
	if err != nil {
		return ledgerState{}, Record{}, err
	}
	if head.Revision == "" {
		return ledgerState{}, head, nil
	}
	var state ledgerState
	if err := json.Unmarshal(head.Data, &state); err != nil {
		return ledgerState{}, Record{}, fmt.Errorf("qastage: decode ledger state for %q: %w", change, err)
	}
	return state, head, nil
}

func lastCompletedStage(state ledgerState) string {
	for i := len(state.Attempts) - 1; i >= 0; i-- {
		if state.Attempts[i].Outcome == OutcomePassed {
			return state.Attempts[i].Stage
		}
	}
	return ""
}

// nextExpectedStage returns the only stage label a qa-begin may use right
// now: the vocabulary's first stage if none is completed yet, the immediate
// successor of the last completed stage otherwise, or "" once the
// vocabulary is exhausted (nothing may follow "docs").
func nextExpectedStage(state ledgerState) string {
	vocab := QAStageVocabulary()
	last := lastCompletedStage(state)
	if last == "" {
		return vocab[0]
	}
	for i, stage := range vocab {
		if stage == last {
			if i+1 < len(vocab) {
				return vocab[i+1]
			}
			return ""
		}
	}
	return ""
}

// Begin opens a new attempt for stage. It is rejected when: an attempt is
// already active for change, or stage is not the vocabulary's first stage
// (for a change with no completed attempts) or the immediate successor of
// the last completed stage. Both checks apply from the very first qa-begin —
// closing the v2 gap where a fresh change's first begin bypassed validation
// entirely (baseline Escenario 3).
func (m *QAStateMachine) Begin(ctx context.Context, change, stage string) (Attempt, error) {
	state, head, err := m.readState(ctx, change)
	if err != nil {
		return Attempt{}, err
	}

	if len(state.Attempts) > 0 && state.Attempts[len(state.Attempts)-1].Outcome == OutcomeRunning {
		return Attempt{}, ErrAttemptAlreadyActive
	}

	if stage != nextExpectedStage(state) {
		return Attempt{}, ErrStageOutOfOrder
	}

	if stage == stageRequiringApproval && !isApproved(state, approvedPredecessorStage) {
		return Attempt{}, ErrApprovalRequired
	}

	attempt := Attempt{Ordinal: len(state.Attempts) + 1, Stage: stage, Outcome: OutcomeRunning}
	state.Attempts = append(state.Attempts, attempt)

	if err := m.commit(ctx, change, head.Revision, state); err != nil {
		return Attempt{}, err
	}
	return attempt, nil
}

// Finish closes the active attempt with outcome. Only OutcomePassed advances
// the state machine (lastCompletedStage/nextExpectedStage only look at
// passed attempts) — failed and interrupted leave the same stage as the next
// expected one, so a retry opens a brand-new attempt without touching
// history (design decision 1.1).
func (m *QAStateMachine) Finish(ctx context.Context, change string, outcome AttemptOutcome) (Attempt, error) {
	if outcome != OutcomePassed && outcome != OutcomeFailed && outcome != OutcomeInterrupted {
		return Attempt{}, ErrInvalidOutcome
	}

	state, head, err := m.readState(ctx, change)
	if err != nil {
		return Attempt{}, err
	}

	if len(state.Attempts) == 0 || state.Attempts[len(state.Attempts)-1].Outcome != OutcomeRunning {
		return Attempt{}, ErrNoActiveAttempt
	}

	idx := len(state.Attempts) - 1
	state.Attempts[idx].Outcome = outcome

	if err := m.commit(ctx, change, head.Revision, state); err != nil {
		return Attempt{}, err
	}
	return state.Attempts[idx], nil
}

// Reset is an audited manual recovery for a stuck active attempt (e.g. an
// agent process died mid-stage without ever calling Finish). It closes the
// attempt as interrupted, records who did it and why, and — like every other
// outcome — never removes it from history.
func (m *QAStateMachine) Reset(ctx context.Context, change, actor, reason string) (Attempt, error) {
	state, head, err := m.readState(ctx, change)
	if err != nil {
		return Attempt{}, err
	}

	if len(state.Attempts) == 0 || state.Attempts[len(state.Attempts)-1].Outcome != OutcomeRunning {
		return Attempt{}, ErrNoActiveAttemptToReset
	}

	idx := len(state.Attempts) - 1
	state.Attempts[idx].Outcome = OutcomeInterrupted
	state.Attempts[idx].ResetBy = actor
	state.Attempts[idx].ResetReason = reason

	if err := m.commit(ctx, change, head.Revision, state); err != nil {
		return Attempt{}, err
	}
	return state.Attempts[idx], nil
}

// Attempts returns the full, ordered attempt history for change — nothing is
// ever deleted or rewritten by Finish or Reset.
func (m *QAStateMachine) Attempts(ctx context.Context, change string) ([]Attempt, error) {
	state, _, err := m.readState(ctx, change)
	if err != nil {
		return nil, err
	}
	return state.Attempts, nil
}

func (m *QAStateMachine) commit(ctx context.Context, change, expectedRevision string, state ledgerState) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("qastage: encode ledger state for %q: %w", change, err)
	}
	_, err = m.store.Append(ctx, change, expectedRevision, payload)
	return err
}
