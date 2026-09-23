package qastage

import (
	"context"
	"errors"
)

// ReviewFinding is one structured result from an automated code reviewer.
type ReviewFinding struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Path        string `json:"path,omitempty"`
}

// QACodeReviewer decouples qa-verify from any single reviewer's concrete
// invocation shape (design decision 1.1 / contrato pendiente #8: qa-verify
// should never invoke RDD's CLI directly). qa-verify's "Functional QA" work
// (Playwright, evidence, BookStack) stays its own; this interface is only
// for the code/diff-review portion that 3.7's RDD already does well.
type QACodeReviewer interface {
	Review(ctx context.Context, change string) ([]ReviewFinding, error)
}

// ErrRDDIntegrationNotImplemented marks RDDAdapter as a placeholder: real
// integration with `gentle-ai review inspect-candidate` is Fase 3B, out of
// Core QA (Fase 3A) scope.
var ErrRDDIntegrationNotImplemented = errors.New("qastage: RDDAdapter real integration is deferred to Fase 3B")

// RDDAdapter is the (stub) QACodeReviewer backed by Gentle-AI 3.7's RDD.
// It satisfies the interface today so callers can be wired against it ahead
// of the real integration, which will replace only this type's body.
type RDDAdapter struct{}

// NewRDDAdapter returns the stub RDD-backed reviewer.
func NewRDDAdapter() *RDDAdapter {
	return &RDDAdapter{}
}

// Review is not yet implemented — see ErrRDDIntegrationNotImplemented.
func (r *RDDAdapter) Review(ctx context.Context, change string) ([]ReviewFinding, error) {
	return nil, ErrRDDIntegrationNotImplemented
}
