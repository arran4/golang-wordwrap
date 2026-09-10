#!/bin/bash
set -euo pipefail

# Dummy GITHUB_OUTPUT file
export GITHUB_OUTPUT=$(mktemp)

function test_route() {
  local test_name="$1"
  export EVENT_NAME="$2"
  export REF_TYPE="$3"
  export REF="$4"
  export INPUT_MODE="$5"
  local expected_exit_code="$6"
  local expected_checks="$7"
  local expected_release="$8"
  local expected_publisher="$9"

  # Clear dummy file
  > "$GITHUB_OUTPUT"

  echo "Running test: $test_name..."

  # Run the logic
  set +e
  bash .github/scripts/route.sh > /dev/null 2>&1
  local exit_code=$?
  set -e

  if [[ "$exit_code" != "$expected_exit_code" ]]; then
    echo "FAIL: $test_name expected exit code $expected_exit_code, got $exit_code"
    exit 1
  fi

  if [[ "$expected_exit_code" == "0" ]]; then
      local run_code_checks=$(grep "run_code_checks=" "$GITHUB_OUTPUT" | cut -d= -f2)
      local run_release=$(grep "run_release=" "$GITHUB_OUTPUT" | cut -d= -f2)
      local run_publisher=$(grep "run_publisher=" "$GITHUB_OUTPUT" | cut -d= -f2)

      if [[ "$run_code_checks" != "$expected_checks" ]]; then
        echo "FAIL: $test_name expected run_code_checks=$expected_checks, got $run_code_checks"
        exit 1
      fi
      if [[ "$run_release" != "$expected_release" ]]; then
        echo "FAIL: $test_name expected run_release=$expected_release, got $run_release"
        exit 1
      fi
      if [[ "$run_publisher" != "$expected_publisher" ]]; then
        echo "FAIL: $test_name expected run_publisher=$expected_publisher, got $run_publisher"
        exit 1
      fi
  fi

  echo "PASS: $test_name"
}

# 1. Normal branch push
test_route "normal branch push" "push" "branch" "refs/heads/main" "" 0 "true" "false" "false"

# 2. External v* tag push
test_route "external tag push" "push" "tag" "refs/tags/v1.0.0" "" 0 "true" "false" "true"

# 3. Manual release-*
test_route "manual release" "workflow_dispatch" "branch" "refs/heads/main" "release-major" 0 "true" "true" "false"

# 4. Valid internal publish-tag
test_route "valid publish-tag" "workflow_dispatch" "tag" "refs/tags/v1.0.0" "publish-tag" 0 "false" "false" "true"

# 5. Invalid branch-context publish-tag
test_route "invalid publish-tag" "workflow_dispatch" "branch" "refs/heads/main" "publish-tag" 1 "" "" ""

echo "All routing regression tests passed."
rm "$GITHUB_OUTPUT"
