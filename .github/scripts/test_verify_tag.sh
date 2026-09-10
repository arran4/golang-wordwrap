#!/bin/bash
set -euo pipefail

echo "Running tag verification regression tests..."

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

pushd $TEST_DIR > /dev/null

git init --quiet -b main
git config user.name "Test Bot"
git config user.email "test@example.com"

echo "init" > file.txt
git add file.txt
git commit --quiet -m "init"

# Set up a fake "origin" to test ls-remote
git clone --quiet --bare . origin.git
git remote add origin origin.git

# Create a lightweight tag
SHA_LIGHT=$(git rev-parse HEAD)
git tag v1.0.0
git push --quiet origin v1.0.0

# Create an annotated tag
echo "annotated" > file2.txt
git add file2.txt
git commit --quiet -m "annotated"
SHA_ANNOTATED=$(git rev-parse HEAD)
git tag -a v2.0.0 -m "Annotated tag"
git push --quiet origin v2.0.0

# Create a mismatch tag
echo "mismatch" > file3.txt
git add file3.txt
git commit --quiet -m "mismatch"
SHA_MISMATCH=$(git rev-parse HEAD)
git tag v3.0.0
git push --quiet origin v3.0.0

popd > /dev/null

# Make absolute path to script
SCRIPT="$(pwd)/.github/scripts/verify_tag.sh"

function run_test() {
    local name="$1"
    local tag="$2"
    local sha="$3"
    local expected_code="$4"

    pushd $TEST_DIR > /dev/null
    set +e
    "$SCRIPT" "$tag" "$sha" > /dev/null 2>&1
    local code=$?
    set -e
    popd > /dev/null

    if [[ "$code" != "$expected_code" ]]; then
        echo "FAIL: $name - expected exit code $expected_code, got $code"
        exit 1
    fi
    echo "PASS: $name"
}

run_test "Valid lightweight tag" "v1.0.0" "$SHA_LIGHT" 0
run_test "Valid annotated tag" "v2.0.0" "$SHA_ANNOTATED" 0
run_test "Mismatch tag" "v3.0.0" "$SHA_LIGHT" 1 # we pass the wrong expected SHA
run_test "Missing tag" "v4.0.0" "$SHA_LIGHT" 2

echo "All tag verification tests passed."
