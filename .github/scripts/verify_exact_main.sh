#!/bin/bash
set -euo pipefail

# Usage: ./verify_exact_main.sh <ref_name> <current_sha>
REF_NAME="$1"
CURRENT_SHA="$2"

if [[ "$REF_NAME" != "refs/heads/main" && "$REF_NAME" != "refs/heads/master" ]]; then
  echo "Error: Manual release preparation must run on refs/heads/main or refs/heads/master, got $REF_NAME" >&2
  exit 1
fi

# Extract the branch name (e.g. main or master)
BRANCH="${REF_NAME#refs/heads/}"

git fetch --quiet origin "$BRANCH" 2>/dev/null
MAIN_SHA=$(git rev-parse "origin/$BRANCH" 2>/dev/null || echo "")

if [[ -z "$MAIN_SHA" ]]; then
  echo "Error: Could not resolve origin/$BRANCH" >&2
  exit 1
fi

if [[ "$MAIN_SHA" != "$CURRENT_SHA" ]]; then
  echo "Error: Requested release against $CURRENT_SHA but origin/$BRANCH is at $MAIN_SHA" >&2
  exit 1
fi

echo "Verified exactly on origin/$BRANCH ($MAIN_SHA)"
exit 0
