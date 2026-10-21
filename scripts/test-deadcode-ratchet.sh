#!/usr/bin/env bash
# Proves the deadcode ratchet compares like-for-like on every platform.
#
# Why this exists: the ratchet compares the tool output against
# .deadcode-baseline.txt with comm(1). The tool prints the file path with the
# host separator, so a Windows run produced "internal\cli\run.go" while the
# baseline holds "internal/cli/run.go". Every baselined entry then looked
# removed and every baselined entry looked new, and the guard reported the
# whole repository as new dead code instead of the functions that really are.
#
# The ratchet stays a ratchet: this test proves the comparison, not the
# baseline.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Loading the script defines the production normalization function. The main
# ratchet remains top-level, so extract only that function definition here.
definition="$(sed -n '/^normalize_deadcode_output()/,/^}/p' "${repo_root}/scripts/deadcode-ratchet.sh")"
if [[ -z "${definition}" ]]; then
  printf 'normalize_deadcode_output not found in scripts/deadcode-ratchet.sh\n' >&2
  exit 1
fi
eval "${definition}"

expected=$'internal/tui/screens/welcome.go\tRenderWelcome'

unix_actual="$(printf '%s\n' \
  'internal/tui/screens/welcome.go:42:1: unreachable func: RenderWelcome' \
  | normalize_deadcode_output)"

windows_actual="$(printf '%s\n' \
  'internal\tui\screens\welcome.go:42:1: unreachable func: RenderWelcome' \
  | normalize_deadcode_output)"

if [[ "${unix_actual}" != "${expected}" ]]; then
  printf 'unix normalization mismatch:\nexpected: %q\nactual:   %q\n' "${expected}" "${unix_actual}" >&2
  exit 1
fi

if [[ "${windows_actual}" != "${expected}" ]]; then
  printf 'windows normalization mismatch:\nexpected: %q\nactual:   %q\n' "${expected}" "${windows_actual}" >&2
  exit 1
fi

# The baseline is compared with comm(1), which needs one sorted input, so the
# normalization must also drop duplicates.
deduped="$(printf '%s\n%s\n' \
  'internal/tui/screens/welcome.go:42:1: unreachable func: RenderWelcome' \
  'internal\tui\screens\welcome.go:42:1: unreachable func: RenderWelcome' \
  | normalize_deadcode_output)"

if [[ "${deduped}" != "${expected}" ]]; then
  printf 'duplicate entries survived the normalization:\nexpected: %q\nactual:   %q\n' "${expected}" "${deduped}" >&2
  exit 1
fi

printf 'deadcode path normalization: pass\n'
