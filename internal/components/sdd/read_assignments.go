package sdd

import (
	"os"

	"github.com/gentleman-programming/gentle-ai/v2/internal/components/filemerge"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/opencode"
)

// configurableAgentSet is the set of valid agent names that may appear in
// opencode.json. It includes SDD, QA, Judgment Day, review, and coordinator agents.
var configurableAgentSet = buildConfigurableAgentSet()

func buildConfigurableAgentSet() map[string]bool {
	phases := opencode.ConfigurableAgentPhases()
	set := make(map[string]bool, len(phases)+2)
	for _, p := range phases {
		set[p] = true
	}
	set["qa-orchestrator"] = true
	// Backward-compatible read aliases for configs that have not been synced yet.
	set["gentle-orchestrator"] = true
	set["sdd-orchestrator"] = true
	return set
}

// ReadCurrentProfiles reads the named SDD profiles from opencode.json at
// settingsPath. It is a thin wrapper around DetectProfiles provided so that
// sync code can import a single symbol from this file.
func ReadCurrentProfiles(settingsPath string) ([]model.Profile, error) {
	return DetectProfiles(settingsPath)
}

// ReadCurrentModelAssignments reads the agent definitions from opencode.json
// at settingsPath and extracts the "model" field for each configurable agent.
//
// Only agents whose names match a configurable agent phase (SDD phases, QA agents,
// JD agents via opencode.ConfigurableAgentPhases()) or an orchestrator key
// (qa-orchestrator and the gentle-/sdd-orchestrator legacy aliases) are included.
// Agents without a "model" field, or with a malformed model value, are silently
// skipped.
//
// Returns an empty map (no error) when the file does not exist, contains no
// "agent" key, or has no matching phase agents with a valid model field.
func ReadCurrentModelAssignments(settingsPath string) (map[string]model.ModelAssignment, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]model.ModelAssignment{}, nil
		}
		return nil, err
	}

	root, err := filemerge.UnmarshalJSONObject(data)
	if err != nil {
		// Unparseable JSON — return empty map, no error.
		return map[string]model.ModelAssignment{}, nil
	}

	agentRaw, ok := root["agent"]
	if !ok {
		return map[string]model.ModelAssignment{}, nil
	}
	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		return map[string]model.ModelAssignment{}, nil
	}

	result := make(map[string]model.ModelAssignment)
	for name, defRaw := range agentMap {
		if !configurableAgentSet[name] {
			continue
		}
		defMap, ok := defRaw.(map[string]any)
		if !ok {
			continue
		}
		modelStr, ok := defMap["model"].(string)
		if !ok || modelStr == "" {
			continue
		}
		providerID, modelID, ok := model.SplitModelSpec(modelStr)
		if !ok {
			continue
		}
		effort, _ := defMap["variant"].(string)
		// Keep the raw agent key for every orchestrator alias; the canonical
		// qa-orchestrator key below is resolved deterministically after the loop.
		result[name] = model.ModelAssignment{
			ProviderID: providerID,
			ModelID:    modelID,
			Effort:     effort,
		}
	}

	// Normalize legacy orchestrator aliases to the canonical key with explicit
	// priority: qa-orchestrator > gentle-orchestrator > sdd-orchestrator.
	if _, hasCanonical := result["qa-orchestrator"]; !hasCanonical {
		if legacy, hasGentle := result["gentle-orchestrator"]; hasGentle {
			result["qa-orchestrator"] = legacy
		} else if legacy, hasSDD := result["sdd-orchestrator"]; hasSDD {
			result["qa-orchestrator"] = legacy
		}
	}
	delete(result, "gentle-orchestrator")
	delete(result, "sdd-orchestrator")

	return result, nil
}
