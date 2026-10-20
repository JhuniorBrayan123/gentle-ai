package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v4/internal/model"
	"github.com/gentleman-programming/gentle-ai/v4/internal/planner"
	"github.com/gentleman-programming/gentle-ai/v4/internal/state"
	"github.com/gentleman-programming/gentle-ai/v4/internal/statecoord"
	"github.com/gentleman-programming/gentle-ai/v4/internal/system"
	"github.com/gentleman-programming/gentle-ai/v4/internal/verify"
)

// sddSharedConventionFiles are the fork's SDD-layer convention files. They
// belong to the SDD component: an install that does not select SDD must not
// leave them behind, and an install that selects SDD must keep them.
var sddSharedConventionFiles = []string{
	"openspec-convention.md",
	"sdd-phase-common.md",
	"sdd-status-contract.md",
}

// sddLayerSelection builds a selection for the SDD retention contract. The
// Skills component is always selected because it owns the ordinary shared
// references the SDD convention files sit beside.
func sddLayerSelection(agent model.AgentID, withSDD bool) model.Selection {
	components := []model.ComponentID{model.ComponentEngram, model.ComponentSkills}
	if withSDD {
		components = append(components, model.ComponentSDD)
	}
	return model.Selection{
		Agents:     []model.AgentID{agent},
		Components: components,
		Skills:     []model.SkillID{model.SkillGoTesting},
		Persona:    model.PersonaGentleman,
		Preset:     model.PresetFullGentleman,
		SDDMode:    model.SDDModeSingle,
	}
}

// installSDDLayer runs the TUI install pipeline for selection in home and
// returns the resolved plan.
func installSDDLayer(t *testing.T, home string, selection model.Selection) planner.ResolvedPlan {
	t.Helper()
	resolved, err := planner.NewResolver(planner.MVPGraph()).Resolve(selection)
	if err != nil {
		t.Fatalf("resolve plan: %v", err)
	}
	execution, _, _ := executeTUIInstall(home, selection, resolved, ResolveInstallProfile(system.DetectionResult{}), "", "", nil)
	if execution.Err != nil {
		t.Fatalf("install error = %v", execution.Err)
	}
	return resolved
}

// persistSDDLayerSelection records the selection the way the TUI does after a
// successful install, so a later sync restores the same component list.
func persistSDDLayerSelection(t *testing.T, home string, selection model.Selection) {
	t.Helper()
	agents := make([]string, 0, len(selection.Agents))
	for _, agent := range selection.Agents {
		agents = append(agents, string(agent))
	}
	installState := state.InstallState{
		InstalledAgents:     agents,
		SelectionConfigured: true,
		Persona:             string(selection.Persona),
	}
	installState.SetSelection(selection)
	if err := statecoord.WithLock(home, func() error { return state.WriteReconciled(home, installState) }); err != nil {
		t.Fatalf("persist install state: %v", err)
	}
}

func claudeSDDSharedPath(home, name string) string {
	return filepath.Join(home, ".claude", "skills", "_shared", name)
}

func claudeSDDSkillPath(home string) string {
	return filepath.Join(home, ".claude", "skills", "sdd-apply", "SKILL.md")
}

func claudeSDDCommandPath(home string) string {
	return filepath.Join(home, ".claude", "commands", "gentle-sdd-init.md")
}

// TestSDDSelectedInstallRetainsSDDLayer pins contract 1: with SDD selected the
// SDD component is installed, its shared convention files survive, and the
// post-apply verification completes without a rollback.
func TestSDDSelectedInstallRetainsSDDLayer(t *testing.T) {
	home := installTestHome(t)
	selection := sddLayerSelection(model.AgentClaudeCode, true)
	resolved := installSDDLayer(t, home, selection)

	report := runPostApplyVerification(postApplyVerificationInput{
		HomeDir:      home,
		WorkspaceDir: home,
		Scope:        ScopeGlobal,
		Selection:    selection,
		Resolved:     resolved,
	})
	if !report.Ready {
		t.Fatalf("post-apply verification not ready:\n%s", verify.RenderReport(report))
	}
	for _, name := range sddSharedConventionFiles {
		path := claudeSDDSharedPath(home, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("SDD shared convention file missing with SDD selected: %s (%v)", path, err)
		}
	}
	if _, err := os.Stat(claudeSDDSkillPath(home)); err != nil {
		t.Errorf("SDD skill missing with SDD selected: %v", err)
	}
	if _, err := os.Stat(claudeSDDCommandPath(home)); err != nil {
		t.Errorf("SDD command missing with SDD selected: %v", err)
	}
}

// TestSDDNotSelectedInstallOmitsSDDLayer pins contract 2: without SDD
// selected the upstream ODD-only behaviour holds, so no SDD asset is left
// behind.
func TestSDDNotSelectedInstallOmitsSDDLayer(t *testing.T) {
	home := installTestHome(t)
	selection := sddLayerSelection(model.AgentClaudeCode, false)
	installSDDLayer(t, home, selection)

	for _, name := range sddSharedConventionFiles {
		path := claudeSDDSharedPath(home, name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("SDD shared convention file present without SDD selected: %s (%v)", path, err)
		}
	}
	if _, err := os.Stat(claudeSDDSkillPath(home)); !os.IsNotExist(err) {
		t.Errorf("SDD skill present without SDD selected: %v", err)
	}
	if _, err := os.Stat(claudeSDDCommandPath(home)); !os.IsNotExist(err) {
		t.Errorf("SDD command present without SDD selected: %v", err)
	}
}

// TestSDDSelectedSyncKeepsSDDLayerAndIsIdempotent pins contracts 3 and 4: a
// sync that keeps SDD selected never removes the SDD assets, and running it
// twice converges on the same file set.
func TestSDDSelectedSyncKeepsSDDLayerAndIsIdempotent(t *testing.T) {
	home := installTestHome(t)
	selection := sddLayerSelection(model.AgentClaudeCode, true)
	installSDDLayer(t, home, selection)
	persistSDDLayerSelection(t, home, selection)

	if _, err := RunSyncWithSelection(home, selection); err != nil {
		t.Fatalf("first RunSyncWithSelection() error = %v", err)
	}
	for _, name := range sddSharedConventionFiles {
		path := claudeSDDSharedPath(home, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("sync removed the SDD shared convention file: %s (%v)", path, err)
		}
	}
	if _, err := os.Stat(claudeSDDSkillPath(home)); err != nil {
		t.Errorf("sync removed the SDD skill: %v", err)
	}
	if _, err := os.Stat(claudeSDDCommandPath(home)); err != nil {
		t.Errorf("sync removed the SDD command: %v", err)
	}

	before := snapshotManagedTree(t, home)
	if _, err := RunSyncWithSelection(home, selection); err != nil {
		t.Fatalf("second RunSyncWithSelection() error = %v", err)
	}
	assertSameManagedTree(t, "second sync", before, snapshotManagedTree(t, home))
}

// TestSDDNotSelectedSyncRetiresSDDLayer pins contract 2 on the sync path: a
// selection without SDD removes the SDD assets an earlier SDD install wrote.
func TestSDDNotSelectedSyncRetiresSDDLayer(t *testing.T) {
	home := installTestHome(t)
	selection := sddLayerSelection(model.AgentClaudeCode, true)
	installSDDLayer(t, home, selection)
	persistSDDLayerSelection(t, home, selection)

	withoutSDD := sddLayerSelection(model.AgentClaudeCode, false)
	if _, err := RunSyncWithSelection(home, withoutSDD); err != nil {
		t.Fatalf("RunSyncWithSelection() error = %v", err)
	}
	for _, name := range sddSharedConventionFiles {
		path := claudeSDDSharedPath(home, name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("SDD shared convention file survived a sync without SDD: %s (%v)", path, err)
		}
	}
	if _, err := os.Stat(claudeSDDSkillPath(home)); !os.IsNotExist(err) {
		t.Errorf("SDD skill survived a sync without SDD: %v", err)
	}
}

// TestUninstallRemovesOnlyManagedSDDAssets pins contract 5: uninstalling the
// SDD component removes the assets gentle-ai manages and keeps every
// user-authored neighbour.
func TestUninstallRemovesOnlyManagedSDDAssets(t *testing.T) {
	home := installTestHome(t)
	selection := sddLayerSelection(model.AgentClaudeCode, true)
	installSDDLayer(t, home, selection)

	userShared := filepath.Join(home, ".claude", "skills", "_shared", "my-notes.md")
	userCommand := filepath.Join(home, ".claude", "commands", "my-command.md")
	for _, path := range []string{userShared, userCommand} {
		mustWriteFile(t, path, []byte("user"))
	}

	if _, err := RunUninstallWithSelection(home, "", []model.AgentID{model.AgentClaudeCode}, []model.ComponentID{model.ComponentSDD}); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	for _, name := range sddSharedConventionFiles {
		path := claudeSDDSharedPath(home, name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("managed SDD shared convention file survived uninstall: %s (%v)", path, err)
		}
	}
	if _, err := os.Stat(claudeSDDSkillPath(home)); !os.IsNotExist(err) {
		t.Errorf("managed SDD skill survived uninstall: %v", err)
	}
	if _, err := os.Stat(claudeSDDCommandPath(home)); !os.IsNotExist(err) {
		t.Errorf("managed SDD command survived uninstall: %v", err)
	}
	for _, path := range []string{userShared, userCommand} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("user file removed by uninstall: %s (%v)", path, err)
		}
	}
}
