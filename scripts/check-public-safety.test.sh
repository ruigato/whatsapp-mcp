#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
temporary="$(mktemp -d)"
trap 'rm -rf "$temporary"' EXIT

repo="$temporary/repo"
mkdir -p "$repo/scripts"
cp "$root/scripts/check-public-safety.sh" "$repo/scripts/"
cd "$repo"

git init -q -b main
git config user.name upstream-author
git config user.email "$(printf 'public%s%s' '@' 'example.org')"
printf 'upstream\n' >README.md
git add README.md
git commit -q -m upstream
upstream_commit="$(git rev-parse HEAD)"
git update-ref refs/remotes/upstream/main "$upstream_commit"

"$repo/scripts/check-public-safety.sh" "$upstream_commit" >/dev/null

git config user.name ruigato
noreply_domain="users.noreply.github.com"
git config user.email "11317160+ruigato@${noreply_domain}"
printf 'safe\n' >>README.md
git add README.md
git commit -q -m safe
"$repo/scripts/check-public-safety.sh" "$upstream_commit..HEAD" >/dev/null

safe_commit="$(git rev-parse HEAD)"
git config user.name local-author
git config user.email "$(printf 'private%s%s' '@' 'example.org')"
printf 'unsafe metadata\n' >>README.md
git add README.md
git commit -q -m unsafe
if "$repo/scripts/check-public-safety.sh" "$safe_commit..HEAD" \
  >/dev/null 2>&1; then
  printf 'Expected a local non-noreply commit to be blocked.\n' >&2
  exit 1
fi

printf 'Public-safety tests passed.\n'
