package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/cli"
)

func TestRunQAValidateHelp(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RunQAValidate([]string{"--help"}, &buf)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !strings.Contains(buf.String(), "qa-validate") {
		t.Errorf("expected help to mention qa-validate, got: %s", buf.String())
	}
}

func TestRunQAValidateMissingInput(t *testing.T) {
	err := cli.RunQAValidate([]string{"--change", "foo", "--stage", "explore"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--input") {
		t.Errorf("expected --input required error, got %v", err)
	}
}

func TestRunQAValidateMissingChange(t *testing.T) {
	err := cli.RunQAValidate([]string{"--input", "-", "--stage", "explore"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--change") {
		t.Errorf("expected --change required error, got %v", err)
	}
}

func TestRunQAValidateMissingStage(t *testing.T) {
	err := cli.RunQAValidate([]string{"--input", "-", "--change", "foo"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--stage") {
		t.Errorf("expected --stage required error, got %v", err)
	}
}

func TestRunQAValidateAdmitValidPayload(t *testing.T) {
	payload := `{"scope":{"status":"complete","next_stage":"spec"}}`
	var buf bytes.Buffer
	// Write to a temp file and pass via stdin simulation using the - flag
	err := cli.RunQAValidateFromReader(
		[]string{"--input", "-", "--change", "my-feature", "--stage", "explore"},
		strings.NewReader(payload),
		&buf,
	)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	var result cli.QAValidateResult
	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid=true, got valid=false, reason=%s", result.Reason)
	}
}

func TestRunQAValidateDenyInvalidPayload(t *testing.T) {
	// Missing scope entirely → bad JSON or invalid
	payload := `{"scope":{"status":"invalid-status"}}`
	var buf bytes.Buffer
	err := cli.RunQAValidateFromReader(
		[]string{"--input", "-", "--change", "my-feature", "--stage", "explore"},
		strings.NewReader(payload),
		&buf,
	)
	if err != nil {
		t.Fatalf("expected JSON output, got error: %v", err)
	}
	var result cli.QAValidateResult
	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if result.Valid {
		t.Error("expected valid=false for invalid status value")
	}
}

func TestRunQAStatusHelp(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RunQAStatus([]string{"--help"}, &buf)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !strings.Contains(buf.String(), "qa-status") {
		t.Errorf("expected help to mention qa-status, got: %s", buf.String())
	}
}

func TestRunQAStatusMissingChange(t *testing.T) {
	err := cli.RunQAStatus([]string{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--change") {
		t.Errorf("expected --change required error, got %v", err)
	}
}
