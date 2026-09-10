#!/bin/bash
set -euo pipefail

echo "Running release planning regression tests..."

SCRIPT="$(pwd)/.github/scripts/plan_release.sh"

# Mock git-tag-inc
export PATH="$(pwd)/mock_bin:$PATH"
mkdir -p mock_bin

# Tracker file to assert bypass
CALL_LOG="$(pwd)/mock_bin/call.log"

function run_test() {
    local name="$1"
    local mode="$2"
    local override="$3"
    local expected_code="$4"
    local expected_output="$5"
    local expect_bypass="$6"

    # Reset call tracking by injecting an extra wrapper in mock_bin
    cat << MOCK_WRAPPER > mock_bin/git-tag-inc
#!/bin/bash
echo "called" > "$CALL_LOG"
if [[ "\$1" != "-print-version-only" ]]; then
  echo "Error: git-tag-inc expected -print-version-only, got \$1" >&2
  exit 1
fi
ARG_STR="\$@"
FORMATTED_ARGS="\${ARG_STR// /.}"
FORMATTED_ARGS="\${FORMATTED_ARGS//-/.}"
echo "v9.9.9.\${FORMATTED_ARGS}"
MOCK_WRAPPER
    sed -i 's/exit/exit/g' mock_bin/git-tag-inc
    chmod +x mock_bin/git-tag-inc

    rm -f "$CALL_LOG"

    set +e
    output=$("$SCRIPT" "$mode" "$override" 2>&1)
    local code=$?
    set -e

    if [[ "$code" != "$expected_code" ]]; then
        echo "FAIL: $name - expected exit code $expected_code, got $code. Output: $output"
        exit 1
    fi

    if [[ -n "$expected_output" && "$output" != *"$expected_output"* ]]; then
        echo "FAIL: $name - expected output to contain '$expected_output', got '$output'"
        exit 1
    fi

    if [[ "$expect_bypass" == "true" ]]; then
        if [[ -f "$CALL_LOG" ]]; then
            echo "FAIL: $name - expected override to bypass git-tag-inc, but it was called."
            exit 1
        fi
    fi

    echo "PASS: $name"
}

run_test "Valid override bypasses git-tag-inc" "release-major" "v2.0.0" 0 "v2.0.0" "true"
run_test "Valid override without v" "release-patch" "3.1.2" 0 "v3.1.2" "true"
run_test "Malformed override" "release-minor" "not-a-version" 1 "Invalid tag format" "true"
run_test "Shell metacharacters rejection" "release-major" "v1.0.0;rm -rf /" 1 "Invalid tag format" "true"

run_test "Stable mode release-major" "release-major" "" 0 "v9.9.9..print.version.only.major" "false"
run_test "Stable mode release-minor" "release-minor" "" 0 "v9.9.9..print.version.only.minor" "false"
run_test "Stable mode release-patch" "release-patch" "" 0 "v9.9.9..print.version.only.patch" "false"
run_test "Prerelease mode release-rc" "release-rc" "" 0 "v9.9.9..print.version.only.patch.rc" "false"
run_test "Prerelease mode release-test" "release-test" "" 0 "v9.9.9..print.version.only.patch.test" "false"
run_test "Prerelease mode release-alpha" "release-alpha" "" 0 "v9.9.9..print.version.only.patch.alpha" "false"

run_test "Unsupported release mode" "invalid-mode" "" 1 "Unsupported release mode" "false"

rm -rf mock_bin
echo "All release planning tests passed."
