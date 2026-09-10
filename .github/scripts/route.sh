#!/bin/bash
set -euo pipefail

run_code_checks=true
run_pr_meta_checks=false
run_cleanup=false
run_release=false
run_publisher=false
is_monthly=false
is_nightly=false

case "$EVENT_NAME" in
  push)
    if [[ "$REF_TYPE" == "tag" && "$REF" == refs/tags/v* ]]; then
      run_publisher=true
    fi
    ;;
  pull_request)
    if [[ "${EVENT_ACTION:-}" == "closed" ]]; then
      run_cleanup=true
    else
      run_pr_meta_checks=true
    fi
    ;;
  workflow_dispatch)
    case "$INPUT_MODE" in
      lint-fix)
        is_nightly=true
        ;;
      release-*)
        run_release=true
        ;;
      publish-tag)
        if [[ "$REF_TYPE" != "tag" || ! "$REF" =~ ^refs/tags/v.* ]]; then
          echo "publish-tag mode requires an eligible tag context (e.g. refs/tags/v*)" >&2
          exit 1
        fi
        run_code_checks=false
        run_publisher=true
        ;;
      monthly-maintenance)
        is_monthly=true
        ;;
    esac
    ;;
  schedule)
    if [[ "${EVENT_SCHEDULE:-}" == "17 3 1 * *" ]]; then
      is_monthly=true
    fi
    if [[ "${EVENT_SCHEDULE:-}" == "41 2 * * *" ]]; then
      is_nightly=true
    fi
    ;;
esac

echo "run_code_checks=$run_code_checks" >> "$GITHUB_OUTPUT"
echo "run_pr_meta_checks=$run_pr_meta_checks" >> "$GITHUB_OUTPUT"
echo "run_cleanup=$run_cleanup" >> "$GITHUB_OUTPUT"
echo "run_release=$run_release" >> "$GITHUB_OUTPUT"
echo "run_publisher=$run_publisher" >> "$GITHUB_OUTPUT"
echo "is_monthly=$is_monthly" >> "$GITHUB_OUTPUT"
echo "is_nightly=$is_nightly" >> "$GITHUB_OUTPUT"
