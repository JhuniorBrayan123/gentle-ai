package assets

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// qaGateSentinels are literal G1-G6 rule statements verbatim from the
// canonical single-source policy file. If any of these strings appears in
// more than one embedded file, or in a file other than the canonical policy
// file itself, a SKILL.md has re-copied the policy instead of referencing it.
var qaGateSentinels = []string{
	"G1 — Documentación como fuente oficial",
	"G2 — Análisis previo",
	"G3 — Planificación obligatoria",
	"G4 — Manejo de incertidumbre",
	"G5 — Control de riesgos",
	"G6 — Validación de la implementación",
}

// qaGatePolicyCanonicalPath is the single in-repo rendering of the G1-G6
// policy. Every qa-* skill must reference it by this exact literal path
// instead of copying its text.
const qaGatePolicyCanonicalPath = "skills/_shared/qa-gate-policy.md"

// qaSkillNames enumerates every qa-* skill directory under skills/ that must
// carry a reference to the canonical G1-G6 policy file.
var qaSkillNames = []string{
	"qa-supervisor",
	"qa-explore",
	"qa-spec",
	"qa-apply",
	"qa-verify",
	"qa-review",
	"qa-docs",
	"qa-doc-access",
	"qa-doc-reference",
	"qa-evidence",
	"qa-locator-hunting",
}

// TestQAGatePolicyHasExactlyOneInRepoRendering walks every embedded asset and
// asserts that each G1-G6 sentinel appears in exactly one file: the canonical
// policy file. This fails the build the moment any SKILL.md re-copies the
// G1-G6 rule text instead of referencing the single source of truth.
func TestQAGatePolicyHasExactlyOneInRepoRendering(t *testing.T) {
	for _, sentinel := range qaGateSentinels {
		t.Run(sentinel, func(t *testing.T) {
			var matches []string
			err := fs.WalkDir(FS, ".", func(assetPath string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() {
					return nil
				}
				content, readErr := Read(assetPath)
				if readErr != nil {
					return readErr
				}
				if strings.Contains(content, sentinel) {
					matches = append(matches, assetPath)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walk embedded assets: %v", err)
			}

			if len(matches) != 1 {
				t.Fatalf("sentinel %q must appear in exactly one embedded file, found in %v", sentinel, matches)
			}
			if matches[0] != qaGatePolicyCanonicalPath {
				t.Fatalf("sentinel %q found in %q, want canonical policy file %q", sentinel, matches[0], qaGatePolicyCanonicalPath)
			}
		})
	}
}

// TestQASkillsReferenceGatePolicy asserts every qa-*/SKILL.md contains the
// literal path to the canonical G1-G6 policy file, so no qa-* skill can drift
// back to an inline (and therefore duplicable) rendering of the rules.
func TestQASkillsReferenceGatePolicy(t *testing.T) {
	for _, skill := range qaSkillNames {
		t.Run(skill, func(t *testing.T) {
			content := MustRead("skills/" + skill + "/SKILL.md")
			if !strings.Contains(content, qaGatePolicyCanonicalPath) {
				t.Fatalf("skills/%s/SKILL.md does not reference canonical policy file %q", skill, qaGatePolicyCanonicalPath)
			}
		})
	}
}

// TestQAGatePolicyPublicMirrorMatchesEmbeddedAsset guards the shared policy
// file itself: the top-level skills/_shared/qa-gate-policy.md must stay
// byte-identical to the embedded asset, following the same
// os.ReadFile-vs-Read parity convention as
// TestPublicBundledSkillsMatchEmbeddedAssets for named skills.
func TestQAGatePolicyPublicMirrorMatchesEmbeddedAsset(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", qaGatePolicyCanonicalPath))
	if err != nil {
		t.Fatalf("ReadFile(public source) error = %v", err)
	}

	embedded, err := Read(qaGatePolicyCanonicalPath)
	if err != nil {
		t.Fatalf("Read(embedded asset) error = %v", err)
	}
	if !bytes.Equal(source, []byte(embedded)) {
		t.Fatal("public qa-gate-policy.md and embedded distribution asset differ")
	}

	if strings.TrimSpace(embedded) == "" {
		t.Fatal("embedded qa-gate-policy.md is empty")
	}
	for _, sentinel := range qaGateSentinels {
		if !strings.Contains(embedded, sentinel) {
			t.Fatalf("canonical qa-gate-policy.md missing sentinel %q", sentinel)
		}
	}
}
