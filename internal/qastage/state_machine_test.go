package qastage

import (
	"context"
	"errors"
	"testing"
)

func newTestMachine(t *testing.T) *QAStateMachine {
	t.Helper()
	return NewQAStateMachine(NewPersistentQAStateStore(t.TempDir()))
}

func TestBegin_FirstAttemptOfNewChangeRejectsNonExploreStage(t *testing.T) {
	machine := newTestMachine(t)
	ctx := context.Background()

	// This is the corrected behavior for the gap the baseline found in v2
	// (Escenario 3): a fresh change's first qa-begin used to accept ANY
	// stage label unchecked. v3 rejects it.
	_, err := machine.Begin(ctx, "change-a", "apply")
	if !errors.Is(err, ErrStageOutOfOrder) {
		t.Fatalf("expected ErrStageOutOfOrder for a first begin on a non-explore stage, got %v", err)
	}
}

func TestBegin_FirstAttemptOfNewChangeAcceptsExplore(t *testing.T) {
	machine := newTestMachine(t)
	ctx := context.Background()

	attempt, err := machine.Begin(ctx, "change-b", "explore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempt.Stage != "explore" || attempt.Outcome != OutcomeRunning {
		t.Fatalf("unexpected attempt: %+v", attempt)
	}
}

func TestBegin_RejectsWhileAnAttemptIsAlreadyActive(t *testing.T) {
	machine := newTestMachine(t)
	ctx := context.Background()

	if _, err := machine.Begin(ctx, "change-c", "explore"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := machine.Begin(ctx, "change-c", "explore")
	if !errors.Is(err, ErrAttemptAlreadyActive) {
		t.Fatalf("expected ErrAttemptAlreadyActive on a retry while running, got %v", err)
	}
}

func TestBeginFinish_AdvancingBeginAcceptsImmediateSuccessor(t *testing.T) {
	machine := newTestMachine(t)
	ctx := context.Background()

	if _, err := machine.Begin(ctx, "change-d", "explore"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := machine.Finish(ctx, "change-d"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	attempt, err := machine.Begin(ctx, "change-d", "spec")
	if err != nil {
		t.Fatalf("expected spec to be accepted as explore's immediate successor, got %v", err)
	}
	if attempt.Stage != "spec" {
		t.Fatalf("expected stage=spec, got %q", attempt.Stage)
	}
}

func TestBeginFinish_AdvancingBeginRejectsSkippedStage(t *testing.T) {
	machine := newTestMachine(t)
	ctx := context.Background()

	if _, err := machine.Begin(ctx, "change-e", "explore"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := machine.Finish(ctx, "change-e"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Skips spec entirely, jumping straight to apply — this is exactly the
	// real skip the baseline confirmed v2 rejects on an advancing begin
	// (Escenario 3, Prueba B). v3 preserves this rejection.
	_, err := machine.Begin(ctx, "change-e", "apply")
	if !errors.Is(err, ErrStageOutOfOrder) {
		t.Fatalf("expected ErrStageOutOfOrder for a skipped stage, got %v", err)
	}
}

func TestFinish_RejectsWithoutAnActiveAttempt(t *testing.T) {
	machine := newTestMachine(t)
	ctx := context.Background()

	_, err := machine.Finish(ctx, "change-f")
	if !errors.Is(err, ErrNoActiveAttempt) {
		t.Fatalf("expected ErrNoActiveAttempt, got %v", err)
	}
}

func TestBeginFinish_FullVocabularyInOrderSucceeds(t *testing.T) {
	machine := newTestMachine(t)
	ctx := context.Background()

	for _, stage := range QAStageVocabulary() {
		if _, err := machine.Begin(ctx, "change-g", stage); err != nil {
			t.Fatalf("begin %q: unexpected error: %v", stage, err)
		}
		if _, err := machine.Finish(ctx, "change-g"); err != nil {
			t.Fatalf("finish %q: unexpected error: %v", stage, err)
		}
	}

	// Nothing comes after "docs" — the vocabulary is exhausted.
	_, err := machine.Begin(ctx, "change-g", "docs")
	if !errors.Is(err, ErrStageOutOfOrder) {
		t.Fatalf("expected ErrStageOutOfOrder once the vocabulary is exhausted, got %v", err)
	}
}
