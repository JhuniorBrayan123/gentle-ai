package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v4/internal/backup"
	"github.com/gentleman-programming/gentle-ai/v4/internal/model"
	"github.com/gentleman-programming/gentle-ai/v4/internal/state"
	"github.com/gentleman-programming/gentle-ai/v4/internal/system"
)

// claudeCommandSkillCollisions lists command names shadowed by skill directories.
// Claude Code resolves same-named skills ahead of commands.
func claudeCommandSkillCollisions(t *testing.T, home string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".claude", "commands"))
	if err != nil {
		t.Fatalf("ReadDir(~/.claude/commands): %v", err)
	}
	var collisions []string
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".md")
		if info, err := os.Stat(filepath.Join(home, ".claude", "skills", name)); err == nil && info.IsDir() {
			collisions = append(collisions, name)
		}
	}
	return collisions
}

func TestRunInstallRejectsRetiredSDDCommands(t *testing.T) {
	home := installTestHome(t)

	if _, err := RunInstall([]string{"--agents", "claude-code", "--components", "sdd"}, system.DetectionResult{}); err == nil {
		t.Fatal("install accepted a retired SDD component")
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "commands", "gentle-sdd-init.md")); !os.IsNotExist(err) {
		t.Fatalf("retired SDD command installed: %v", err)
	}
}

// TestRunSyncInstallsSDDCommandsWhenSDDIsSelected pins the fork's conditional
// SDD contract on the Claude Code command namespace. A selection that carries
// SDD installs the SDD commands and retires the legacy unprefixed name a
// pre-#2644 install managed, so the runtime never keeps both. A user's own
// command survives, and no command file is shadowed by a same-named skill
// directory.
func TestRunSyncInstallsSDDCommandsWhenSDDIsSelected(t *testing.T) {
	home := installTestHome(t)
	restoreBackupHome := backup.UserHomeDirFn
	backup.UserHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { backup.UserHomeDirFn = restoreBackupHome })
	if err := state.Write(home, state.InstallState{
		InstalledAgents:     []string{string(model.AgentClaudeCode)},
		SelectionConfigured: true,
		Components:          []model.ComponentID{model.ComponentSDD},
		Persona:             "neutral",
	}); err != nil {
		t.Fatalf("state.Write() error = %v", err)
	}
	legacy := filepath.Join(home, ".claude", "commands", "sdd-init.md")
	mustWriteFile(t, legacy, []byte("# user-edited former SDD command\n"))
	custom := filepath.Join(home, ".claude", "commands", "my-command.md")
	mustWriteFile(t, custom, []byte("keep"))

	result, err := RunSync([]string{"--agents", "claude-code"})
	if err != nil {
		t.Fatalf("RunSync() error = %v", err)
	}

	// The selected SDD layer installs its own command and retires the legacy
	// unprefixed name it replaced.
	installed := filepath.Join(home, ".claude", "commands", "gentle-sdd-init.md")
	if _, err := os.Stat(installed); err != nil {
		t.Errorf("sync did not install the SDD command: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("sync kept the legacy unprefixed SDD command: %v", err)
	}
	if !containsPath(result.ChangedFiles, legacy) {
		t.Errorf("ChangedFiles omits the retired legacy command %q", legacy)
	}
	if _, err := os.Stat(custom); err != nil {
		t.Errorf("sync removed a user-owned command: %v", err)
	}
	if collisions := claudeCommandSkillCollisions(t, home); len(collisions) > 0 {
		t.Fatalf("~/.claude/commands files shadowed by same-named ~/.claude/skills directories: %v", collisions)
	}
}
