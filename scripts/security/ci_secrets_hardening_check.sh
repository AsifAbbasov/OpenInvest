#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

fail=0

echo "[ci] verifying workflow token is globally read-only"
if ! grep -Eq '^permissions:[[:space:]]*$' .github/workflows/ci.yml ||
   ! grep -Eq '^[[:space:]]+contents:[[:space:]]+read[[:space:]]*$' .github/workflows/ci.yml; then
  echo "FAIL: CI workflow must declare permissions: contents: read" >&2
  fail=1
fi

echo "[ci] rejecting pull_request_target in executable workflows"
if git grep -n 'pull_request_target' -- '.github/workflows/*.yml' '.github/workflows/*.yaml' 2>/dev/null; then
  echo "FAIL: pull_request_target found in workflow" >&2
  fail=1
fi

echo "[supply-chain] requiring immutable SHA pins for external actions"
while IFS= read -r line; do
  ref="${line#*@}"
  ref="${ref%% *}"
  if [[ ! "$ref" =~ ^[0-9a-fA-F]{40}$ ]]; then
    echo "FAIL: unpinned GitHub Action: $line" >&2
    fail=1
  fi
done < <(grep -RhoE 'uses:[[:space:]]+[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+@[A-Za-z0-9_.-]+' .github/workflows || true)

echo "[secrets] scanning executable/config surfaces for common committed credential forms"
secret_pattern='-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----|AKIA[0-9A-Z]{16}|github_pat_[A-Za-z0-9_]+|ghp_[A-Za-z0-9]{20,}'
if git grep -nE "$secret_pattern" -- \
  ':!docs/**' \
  ':!**/*_test.go' \
  ':!.env.example' \
  ':!scripts/security/ci_secrets_hardening_check.sh'
then
  echo "FAIL: possible committed credential material found" >&2
  fail=1
fi

echo "[secrets] rejecting tracked local environment files"
tracked_env="$(git ls-files | grep -E '(^|/)\.env($|\.)' | grep -vE '(^|/)\.env\.example$' || true)"
if [[ -n "$tracked_env" ]]; then
  echo "FAIL: tracked environment files:" >&2
  echo "$tracked_env" >&2
  fail=1
fi

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo "CI_SECRETS_HARDENING=PASS"
