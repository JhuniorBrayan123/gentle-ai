package qastage

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v2/internal/sddstatus"
)

// VocabularyID is the stable identifier for the QA Orchestrator V2 vocabulary.
const VocabularyID = "gentle-ai.qa-orchestrator/v2"

// VocabularyV1 returns the standard QA Orchestrator V2 stage vocabulary.
// It enforces the exact sequence: explore -> spec -> apply -> verify -> docs.
// Only advancing into "apply" requires a recorded approval (design item b:
// one Go-enforced gate, spec->apply); the other transitions are order-only.
func VocabularyV1() sddstatus.StageVocabulary {
	return sddstatus.StageVocabulary{
		ID: VocabularyID,
		Stages: []sddstatus.Stage{
			{Label: "explore"},
			{Label: "spec"},
			{Label: "apply", RequiresApproval: true},
			{Label: "verify"},
			{Label: "docs"},
		},
	}
}

// LedgerChangeName prefixes the user's target change name with "qa_" to
// ensure perfect isolation from the primary SDD runtime ledger. This
// prevents cross-ledger pollution since QA Orchestrator uses the exact same
// underlying filecoord persistence mechanism.
//
// Deviation from design D7 (discovered by the PR13 E2E integration test):
// the design's literal draft used a "qa--" (double-hyphen) prefix, but
// OpenRuntimeStore's own change-name validator (internal/sddstatus,
// reviewBindingChange) requires alphanumeric segments joined by exactly ONE
// hyphen or underscore -- it rejects consecutive separators outright. A
// change of "qa--{change}" therefore made OpenRuntimeStore refuse EVERY QA
// ledger with "invalid SDD change name", which no unit test caught because
// vocabulary_test.go validated the prefixed name against the more lenient
// sddstatus.ValidateRuntimeText (a free-text validator), not the actual
// change-name regex OpenRuntimeStore enforces. "qa_" (single underscore) is
// the smallest change that keeps a non-hyphen namespace marker while
// satisfying the real validator; it still cannot collide with an ordinary
// hyphen-only SDD change name.
func LedgerChangeName(change string) string {
	change = strings.TrimSpace(change)
	if change == "" {
		return ""
	}
	return fmt.Sprintf("qa_%s", change)
}
