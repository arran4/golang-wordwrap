#!/bin/bash
set -euo pipefail

echo "Running release planning regression tests..."

SCRIPT="$(pwd)/.github/scripts/plan_release.sh"

# Mock git-tag-inc
export PATH="$(pwd)/mock_bin:$PATH"
mkdir -p mock_bin

CALL_LOG="$(pwd)/mock_bin/call.log"

function run_test() {
    local name="$1"
    local mode="$2"
    local override="$3"
    local expected_code="$4"
    local expected_output="$5"
    local expect_call_count="$6"
    local expected_args="$7"

    # Reset call tracking by recreating the mock script
    cat << MOCK_WRAPPER > mock_bin/git-tag-inc
#!/bin/bash
echo "\$@" >> "$CALL_LOG"
echo "v9.9.9"
MOCK_WRAPPER
    chmod +x mock_bin/git-tag-inc

    rm -f "$CALL_LOG"
    touch "$CALL_LOG"

    set +e
    output=$("$SCRIPT" "$mode" "$override" 2>&1)
    local code=$?
    set -e

    if [[ "$code" != "$expected_code" ]]; then
        echo "FAIL: $name - expected exit code $expected_code, got $code. Output: $output"
        exit 1
    fi

    # Output matching: if it expects exact output, verify it exactly
    if [[ -n "$expected_output" && "$output" != "$expected_output" ]]; then
        # For error messages, we can do substring match
        if [[ "$code" != "0" ]]; then
            if [[ "$output" != *"$expected_output"* ]]; then
                 echo "FAIL: $name - expected output to contain '$expected_output', got '$output'"
                 exit 1
            fi
        else
            echo "FAIL: $name - expected output exactly '$expected_output', got '$output'"
            exit 1
        fi
    fi

    local actual_call_count=$(wc -l < "$CALL_LOG")
    if [[ "$actual_call_count" != "$expect_call_count" ]]; then
        echo "FAIL: $name - expected git-tag-inc to be called $expect_call_count times, but it was called $actual_call_count times."
        exit 1
    fi

    if [[ "$expect_call_count" == "1" ]]; then
        local actual_args=$(cat "$CALL_LOG")
        if [[ "$actual_args" != "$expected_args" ]]; then
            echo "FAIL: $name - expected git-tag-inc arguments '$expected_args', got '$actual_args'"
            exit 1
        fi
    fi

    echo "PASS: $name"
}

run_test "Valid override bypasses git-tag-inc" "release-major" "v2.0.0" 0 "v2.0.0" 0 ""
run_test "Valid override without v" "release-patch" "3.1.2" 0 "v3.1.2" 0 ""
run_test "Malformed override" "release-minor" "not-a-version" 1 "Invalid tag format" 0 ""
run_test "Shell metacharacters rejection" "release-major" "v1.0.0;rm -rf /" 1 "Invalid tag format" 0 ""

run_test "Stable mode release-major" "release-major" "" 0 "v9.9.9" 1 "-print-version-only major"
run_test "Stable mode release-minor" "release-minor" "" 0 "v9.9.9" 1 "-print-version-only minor"
run_test "Stable mode release-patch" "release-patch" "" 0 "v9.9.9" 1 "-print-version-only patch"
run_test "Prerelease mode release-rc" "release-rc" "" 0 "v9.9.9" 1 "-print-version-only patch rc"
run_test "Prerelease mode release-test" "release-test" "" 0 "v9.9.9" 1 "-print-version-only patch test"
run_test "Prerelease mode release-alpha" "release-alpha" "" 0 "v9.9.9" 1 "-print-version-only patch alpha"

run_test "Unsupported release mode" "invalid-mode" "" 1 "Unsupported release mode" 0 ""

rm -rf mock_bin
echo "All release planning tests passed."
