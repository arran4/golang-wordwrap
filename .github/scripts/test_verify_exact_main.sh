#!/bin/bash
set -euo pipefail

echo "Running exact main verification regression tests..."

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

# Main SHA
SHA_MAIN=$(git rev-parse HEAD)

# Create a stale main situation: push a new commit to origin, but keep local at SHA_MAIN
echo "stale" > file2.txt
git add file2.txt
git commit --quiet -m "stale"
SHA_STALE=$(git rev-parse HEAD)
git push --quiet origin main
git reset --hard HEAD~1 --quiet

popd > /dev/null

SCRIPT="$(pwd)/.github/scripts/verify_exact_main.sh"

function run_test() {
    local name="$1"
    local ref="$2"
    local sha="$3"
    local expected_code="$4"

    pushd $TEST_DIR > /dev/null
    set +e
    "$SCRIPT" "$ref" "$sha" > /dev/null 2>&1
    local code=$?
    set -e
    popd > /dev/null

    if [[ "$code" != "$expected_code" ]]; then
        echo "FAIL: $name - expected exit code $expected_code, got $code"
        exit 1
    fi
    echo "PASS: $name"
}

# 1. Feature branch rejection
run_test "Feature branch rejection" "refs/heads/feature" "$SHA_MAIN" 1

# 2. Exact main success (using the advanced SHA that is currently on origin/main)
# We test with SHA_STALE which is the commit origin/main points to.
run_test "Exact main success" "refs/heads/main" "$SHA_STALE" 0

# 3. Stale-main SHA rejection (using SHA_MAIN which is behind origin/main)
run_test "Stale main rejection" "refs/heads/main" "$SHA_MAIN" 1

echo "All exact main verification tests passed."
