#!/bin/bash
set -euo pipefail

# Usage: ./verify_tag.sh <tag> <expected_sha>
TAG="$1"
EXPECTED_SHA="$2"

git fetch --tags --force origin 2>/dev/null || true

# Check peeled annotated tag first
REMOTE_TAG_SHA=$(git ls-remote --tags origin "refs/tags/${TAG}^{}" | awk '{print $1}')
if [[ -z "$REMOTE_TAG_SHA" ]]; then
  REMOTE_TAG_SHA=$(git ls-remote --tags origin "refs/tags/${TAG}" | awk '{print $1}')
fi

if [[ -z "$REMOTE_TAG_SHA" ]]; then
  # Tag does not exist
  exit 2
fi

if [[ "$REMOTE_TAG_SHA" == "$EXPECTED_SHA" ]]; then
  echo "Tag $TAG exists and points to $EXPECTED_SHA. Safe to retry."
  exit 0
else
  echo "Error: Tag already exists and points to $REMOTE_TAG_SHA (expected $EXPECTED_SHA): $TAG" >&2
  exit 1
fi
