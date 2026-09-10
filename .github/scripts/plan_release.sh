#!/bin/bash
set -euo pipefail

MODE="$1"
OVERRIDE="$2"

if [[ -n "$OVERRIDE" ]]; then
  # Accept "1.2.3" or "v1.2.3" override input.
  OVERRIDE="${OVERRIDE#v}"
  next_tag="v$OVERRIDE"
else
  case "$MODE" in
    release-major) level="major"; suffix="" ;;
    release-minor) level="minor"; suffix="" ;;
    release-patch) level="patch"; suffix="" ;;
    release-test)  level="patch"; suffix="test" ;;
    release-rc)    level="patch"; suffix="rc" ;;
    release-alpha) level="patch"; suffix="alpha" ;;
    *) echo "Unsupported release mode: $MODE" >&2; exit 1 ;;
  esac

  level="${level#-}"
  args=(-print-version-only "$level")
  [[ -n "$suffix" ]] && args+=("$suffix")

  # Allow tests to mock git-tag-inc, otherwise use real one
  if command -v git-tag-inc >/dev/null 2>&1; then
    next_tag=$(git-tag-inc "${args[@]}")
  else
    # For testing ONLY, when git-tag-inc is not available
    echo "git-tag-inc not found" >&2
    exit 1
  fi
fi

if [[ "$next_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.]+)?$ ]]; then
  echo "$next_tag"
else
  echo "Invalid tag format: $next_tag" >&2
  exit 1
fi
