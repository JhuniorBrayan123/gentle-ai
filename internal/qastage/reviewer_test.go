package qastage

import (
	"context"
	"errors"
	"testing"
)

// QACodeReviewer decouples qa-verify from RDD's concrete invocation shape
// (design decision 1.1 / contrato pendiente #8). 3A.11 scope is the
// interface plus a stub adapter — real integration with
// `gentle-ai review inspect-candidate` is Fase 3B, out of Core QA.

func TestRDDAdapter_ImplementsQACodeReviewer(t *testing.T) {
	var _ QACodeReviewer = NewRDDAdapter()
}

func TestRDDAdapter_StubReturnsNotImplementedError(t *testing.T) {
	adapter := NewRDDAdapter()

	_, err := adapter.Review(context.Background(), "some-change")
	if !errors.Is(err, ErrRDDIntegrationNotImplemented) {
		t.Fatalf("expected ErrRDDIntegrationNotImplemented, got %v", err)
	}
}

// fakeReviewer demonstrates that QACodeReviewer is a genuinely substitutable
// seam: a future qa-verify caller (or its tests) can inject any
// implementation without depending on RDDAdapter or real RDD invocation.
type fakeReviewer struct {
	findings []ReviewFinding
}

func (f *fakeReviewer) Review(ctx context.Context, change string) ([]ReviewFinding, error) {
	return f.findings, nil
}

func TestQACodeReviewer_InterfaceIsSubstitutable(t *testing.T) {
	var reviewer QACodeReviewer = &fakeReviewer{findings: []ReviewFinding{
		{Severity: "WARNING", Description: "unused variable", Path: "foo.go"},
	}}

	findings, err := reviewer.Review(context.Background(), "some-change")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 || findings[0].Severity != "WARNING" {
		t.Fatalf("unexpected findings: %+v", findings)
	}
}
