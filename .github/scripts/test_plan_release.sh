#!/bin/bash
set -euo pipefail

echo "Running release planning regression tests..."

SCRIPT="$(pwd)/.github/scripts/plan_release.sh"

# Mock git-tag-inc
export PATH="$(pwd)/mock_bin:$PATH"
mkdir -p mock_bin
cat << 'MOCK' > mock_bin/git-tag-inc
#!/bin/bash
# Remove dashes from the arguments so it doesn't look like double dashes or invalid characters
if [[ -n "$3" ]]; then
  echo "v9.9.9.print.version.only.$2.$3"
else
  echo "v9.9.9.print.version.only.$2"
fi

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

run_test "Valid override bypasses git-tag-inc" "release-major" "v2.0.0" 0 "v2.0.0"
run_test "Valid override without v" "release-patch" "3.1.2" 0 "v3.1.2"
run_test "Malformed override" "release-minor" "not-a-version" 1 "Invalid tag format"
run_test "Shell metacharacters rejection" "release-major" "v1.0.0;rm -rf /" 1 "Invalid tag format"

run_test "Stable mode release-major" "release-major" "" 0 "v9.9.9.print.version.only.major"
run_test "Stable mode release-minor" "release-minor" "" 0 "v9.9.9.print.version.only.minor"
run_test "Stable mode release-patch" "release-patch" "" 0 "v9.9.9.print.version.only.patch"
run_test "Prerelease mode release-rc" "release-rc" "" 0 "v9.9.9.print.version.only.patch.rc"
run_test "Prerelease mode release-test" "release-test" "" 0 "v9.9.9.print.version.only.patch.test"
run_test "Prerelease mode release-alpha" "release-alpha" "" 0 "v9.9.9.print.version.only.patch.alpha"

run_test "Unsupported release mode" "invalid-mode" "" 1 "Unsupported release mode"

rm -rf mock_bin
echo "All release planning tests passed."
