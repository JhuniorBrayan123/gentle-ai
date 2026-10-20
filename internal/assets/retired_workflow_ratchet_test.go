package assets

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// retiredWorkflowReferenceAllowlist lists embedded files that may still name
// SDD or OpenSpec, mapped to the reason. Only retirement inventory qualifies:
// content whose sole purpose is to recognize and clean up retired SDD assets.
// Prose offered to an agent never qualifies. Every entry must still match, so
// the list can only shrink.
var retiredWorkflowReferenceAllowlist = map[string]string{
	"antigravity/sdd-orchestrator.md":                  "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-apply.md":                       "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-archive.md":                     "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-design.md":                      "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-explore.md":                     "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-init.md":                        "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-onboard.md":                     "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-propose.md":                     "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-research.md":                    "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-spec.md":                        "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-tasks.md":                       "fork QA/SDD layer keeps this asset",
	"claude/agents/sdd-verify.md":                      "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-apply.md":              "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-archive.md":            "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-continue.md":           "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-explore.md":            "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-ff.md":                 "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-init.md":               "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-new.md":                "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-onboard.md":            "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-research.md":           "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-status.md":             "fork QA/SDD layer keeps this asset",
	"claude/commands/gentle-sdd-verify.md":             "fork QA/SDD layer keeps this asset",
	"claude/sdd-orchestrator-workflow.md":              "fork QA/SDD layer keeps this asset",
	"claude/sdd-orchestrator.md":                       "fork QA/SDD layer keeps this asset",
	"codex/sdd-orchestrator.md":                        "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-apply.md":                       "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-archive.md":                     "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-design.md":                      "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-explore.md":                     "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-init.md":                        "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-onboard.md":                     "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-propose.md":                     "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-research.md":                    "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-spec.md":                        "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-tasks.md":                       "fork QA/SDD layer keeps this asset",
	"cursor/agents/sdd-verify.md":                      "fork QA/SDD layer keeps this asset",
	"cursor/sdd-orchestrator.md":                       "fork QA/SDD layer keeps this asset",
	"gemini/sdd-orchestrator.md":                       "fork QA/SDD layer keeps this asset",
	"generic/sdd-orchestrator.md":                      "fork QA/SDD layer keeps this asset",
	"hermes/sdd-orchestrator.md":                       "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-apply.md":                         "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-apply.yaml":                       "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-archive.md":                       "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-archive.yaml":                     "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-design.md":                        "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-design.yaml":                      "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-explore.md":                       "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-explore.yaml":                     "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-init.md":                          "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-init.yaml":                        "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-onboard.md":                       "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-onboard.yaml":                     "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-propose.md":                       "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-propose.yaml":                     "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-research.md":                      "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-research.yaml":                    "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-spec.md":                          "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-spec.yaml":                        "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-tasks.md":                         "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-tasks.yaml":                       "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-verify.md":                        "fork QA/SDD layer keeps this asset",
	"kimi/agents/sdd-verify.yaml":                      "fork QA/SDD layer keeps this asset",
	"kimi/sdd-orchestrator.md":                         "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-apply.md":                         "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-archive.md":                       "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-design.md":                        "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-explore.md":                       "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-init.md":                          "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-onboard.md":                       "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-propose.md":                       "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-research.md":                      "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-spec.md":                          "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-tasks.md":                         "fork QA/SDD layer keeps this asset",
	"kiro/agents/sdd-verify.md":                        "fork QA/SDD layer keeps this asset",
	"kiro/sdd-orchestrator.md":                         "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-apply.md":                   "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-archive.md":                 "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-continue.md":                "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-explore.md":                 "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-ff.md":                      "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-init.md":                    "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-new.md":                     "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-onboard.md":                 "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-research.md":                "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-status.md":                  "fork QA/SDD layer keeps this asset",
	"opencode/commands/sdd-verify.md":                  "fork QA/SDD layer keeps this asset",
	"opencode/plugins-v2/sdd-task-result-artifacts.ts": "fork QA/SDD layer keeps this asset",
	"opencode/plugins/sdd-task-result-artifacts.ts":    "fork QA/SDD layer keeps this asset",
	"opencode/sdd-orchestrator.md":                     "fork QA/SDD layer keeps this asset",
	"opencode/sdd-overlay-multi.json":                  "fork QA/SDD layer keeps this asset",
	"opencode/sdd-overlay-single.json":                 "fork QA/SDD layer keeps this asset",
	"qwen/sdd-orchestrator.md":                         "fork QA/SDD layer keeps this asset",
	"skills/_shared/openspec-convention.md":            "fork QA/SDD layer keeps this asset",
	"skills/_shared/sdd-orchestrator-sections.md":      "fork QA/SDD layer keeps this asset",
	"skills/_shared/sdd-phase-common.md":               "fork QA/SDD layer keeps this asset",
	"skills/_shared/sdd-status-contract.md":            "fork QA/SDD layer keeps this asset",
	"skills/qa-doc-reference/SKILL.md":                 "fork QA/SDD layer keeps this asset",
	"skills/sdd-apply":                                 "fork QA/SDD layer keeps this asset",
	"skills/sdd-apply/SKILL.md":                        "fork QA/SDD layer keeps this asset",
	"skills/sdd-apply/strict-tdd.md":                   "fork QA/SDD layer keeps this asset",
	"skills/sdd-archive":                               "fork QA/SDD layer keeps this asset",
	"skills/sdd-archive/SKILL.md":                      "fork QA/SDD layer keeps this asset",
	"skills/sdd-design":                                "fork QA/SDD layer keeps this asset",
	"skills/sdd-design/SKILL.md":                       "fork QA/SDD layer keeps this asset",
	"skills/sdd-explore":                               "fork QA/SDD layer keeps this asset",
	"skills/sdd-explore/SKILL.md":                      "fork QA/SDD layer keeps this asset",
	"skills/sdd-init":                                  "fork QA/SDD layer keeps this asset",
	"skills/sdd-init/SKILL.md":                         "fork QA/SDD layer keeps this asset",
	"skills/sdd-init/references/init-details.md":       "fork QA/SDD layer keeps this asset",
	"skills/sdd-onboard":                               "fork QA/SDD layer keeps this asset",
	"skills/sdd-onboard/SKILL.md":                      "fork QA/SDD layer keeps this asset",
	"skills/sdd-propose":                               "fork QA/SDD layer keeps this asset",
	"skills/sdd-propose/SKILL.md":                      "fork QA/SDD layer keeps this asset",
	"skills/sdd-research":                              "fork QA/SDD layer keeps this asset",
	"skills/sdd-research/SKILL.md":                     "fork QA/SDD layer keeps this asset",
	"skills/sdd-spec":                                  "fork QA/SDD layer keeps this asset",
	"skills/sdd-spec/SKILL.md":                         "fork QA/SDD layer keeps this asset",
	"skills/sdd-tasks":                                 "fork QA/SDD layer keeps this asset",
	"skills/sdd-tasks/SKILL.md":                        "fork QA/SDD layer keeps this asset",
	"skills/sdd-verify":                                "fork QA/SDD layer keeps this asset",
	"skills/sdd-verify/SKILL.md":                       "fork QA/SDD layer keeps this asset",
	"skills/sdd-verify/references/report-format.md":    "fork QA/SDD layer keeps this asset",
	"skills/sdd-verify/strict-tdd-verify.md":           "fork QA/SDD layer keeps this asset",
	"windsurf/sdd-orchestrator.md":                     "fork QA/SDD layer keeps this asset",
	"windsurf/workflows/sdd-new.md":                    "fork QA/SDD layer keeps this asset",
}

// TestEmbeddedAssetsHaveNoRetiredWorkflowReferences is a ratchet over every
// embedded asset Gentle AI can install or render into a runtime. SDD and
// OpenSpec were retired in v4.0.0; their names must not reappear in shipped
// paths or content.
func TestEmbeddedAssetsHaveNoRetiredWorkflowReferences(t *testing.T) {
	retired := regexp.MustCompile(`(?i)sdd|openspec`)
	allowlisted := map[string]bool{}
	scanned := 0
	err := fs.WalkDir(FS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if retired.MatchString(entry.Name()) {
			if _, ok := retiredWorkflowReferenceAllowlist[path]; ok {
				allowlisted[path] = true
			} else {
				t.Errorf("%s: embedded path names a retired workflow", path)
			}
		}
		if entry.IsDir() {
			return nil
		}
		scanned++
		for index, line := range strings.Split(MustRead(path), "\n") {
			if !retired.MatchString(line) {
				continue
			}
			if _, ok := retiredWorkflowReferenceAllowlist[path]; ok {
				allowlisted[path] = true
				continue
			}
			t.Errorf("%s:%d: retired workflow reference: %s", path, index+1, strings.TrimSpace(line))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scanned == 0 {
		t.Fatal("no embedded assets scanned")
	}
	for path := range retiredWorkflowReferenceAllowlist {
		if !allowlisted[path] {
			t.Errorf("%s is allowlisted but no longer names a retired workflow; remove the entry", path)
		}
	}
}
