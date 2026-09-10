#!/bin/bash
set -euo pipefail

echo "Running release planning regression tests..."

SCRIPT="$(pwd)/.github/scripts/plan_release.sh"

# Mock git-tag-inc
export PATH="$(pwd)/mock_bin:$PATH"
mkdir -p mock_bin
cat << 'MOCK' > mock_bin/git-tag-inc
#!/bin/bash
echo "v1.2.4"
MOCK
chmod +x mock_bin/git-tag-inc

function run_test() {
    local name="$1"
    local mode="$2"
    local override="$3"
    local expected_code="$4"
    local expected_output="$5"

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

    echo "PASS: $name"
}

run_test "Valid override with v" "release-major" "v2.0.0" 0 "v2.0.0"
run_test "Valid override without v" "release-patch" "3.1.2" 0 "v3.1.2"
run_test "Malformed override" "release-minor" "not-a-version" 1 "Invalid tag format"
run_test "Shell metacharacters rejection" "release-major" "v1.0.0;rm -rf /" 1 "Invalid tag format"
run_test "Stable mode fallback to git-tag-inc" "release-minor" "" 0 "v1.2.4"
run_test "Unsupported release mode" "invalid-mode" "" 1 "Unsupported release mode"

rm -rf mock_bin
echo "All release planning tests passed."
