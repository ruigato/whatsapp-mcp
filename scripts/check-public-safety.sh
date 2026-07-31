#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

failed=0

while IFS= read -r tracked_path; do
  case "$tracked_path" in
    whatsapp-bridge/store/*|whatsapp-bridge/store.bak.*/*|\
    *.db|*.sqlite|*.sqlite3|*.log|*.pem|*.key|.env|.env.*|\
    claude_desktop_config.json|whatsapp-bridge/whatsapp-bridge)
      printf 'Blocked sensitive tracked path: %s\n' "$tracked_path" >&2
      failed=1
      ;;
  esac
done < <(git ls-files)

scan_tracked() {
  local label="$1"
  local pattern="$2"
  local matches
  if matches="$(
    git grep -n -I -E "$pattern" -- . \
      ':(exclude)scripts/check-public-safety.sh' \
      ':(exclude).githooks/pre-push' 2>/dev/null
  )"; then
    printf 'Blocked %s:\n%s\n' "$label" "$matches" >&2
    failed=1
  fi
}

scan_tracked \
  "personal email address" \
  '[[:alnum:]][[:alnum:]._%+-]*@[[:alnum:].-]+\.[[:alpha:]]{2,}'
scan_tracked \
  "numeric WhatsApp identifier" \
  '[0-9]{8,}@(s\.whatsapp\.net|lid|g\.us)'
scan_tracked \
  "local user path" \
  '(/Users/[^/[:space:]]+|C:\\Users\\[^\\[:space:]]+|/home/[^/[:space:]]+)'
scan_tracked \
  "private key or high-confidence credential" \
  '(BEGIN [A-Z ]*PRIVATE KEY|gh[opsu]_[A-Za-z0-9_]{20,}|sk-[A-Za-z0-9_-]{20,})'
scan_tracked \
  "Portuguese phone number" \
  '(\+351|00351)[[:space:].()-]*[0-9]([[:space:].()-]*[0-9]){8}'

range="${1:-}"
if [[ -n "$range" ]]; then
  non_private_emails="$(
    git log --format='%ae%n%ce' "$range" |
      sort -u |
      grep -Ev '^[0-9]+\+ruigato@users\.noreply\.github\.com$' || true
  )"
  if [[ -n "$non_private_emails" ]]; then
    printf 'Blocked non-noreply commit email in %s:\n%s\n' \
      "$range" "$non_private_emails" >&2
    failed=1
  fi
fi

if [[ "$failed" -ne 0 ]]; then
  exit 1
fi

printf 'Public-safety gate passed.\n'
