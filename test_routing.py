import sys

def test_route(test_name, event_name, ref_type, ref, mode, expected_checks, expected_release):
    run_code_checks = False
    run_release = False

    if event_name == "push":
        run_code_checks = True
        if ref.startswith("refs/tags/v"):
            run_release = True
    elif event_name == "pull_request":
        run_code_checks = True
    elif event_name == "workflow_dispatch":
        run_code_checks = True
        if mode == "lint-fix":
            pass
        elif mode.startswith("release-"):
            run_release = True
        elif mode == "publish-tag":
            if ref_type != "tag" or not ref.startswith("refs/tags/v"):
                run_release = False
            else:
                run_release = True
        elif mode == "monthly-maintenance":
            pass
    elif event_name == "schedule":
        run_code_checks = True

    if str(run_code_checks).lower() != str(expected_checks).lower():
        print(f"FAIL: {test_name} expected run_code_checks={expected_checks}, got {run_code_checks}")
        sys.exit(1)
    if str(run_release).lower() != str(expected_release).lower():
        print(f"FAIL: {test_name} expected run_release={expected_release}, got {run_release}")
        sys.exit(1)

    print(f"PASS: {test_name}")

test_route("normal branch push", "push", "branch", "refs/heads/main", "", True, False)
test_route("external tag push", "push", "tag", "refs/tags/v1.0.0", "", True, True)
test_route("manual release", "workflow_dispatch", "branch", "refs/heads/main", "release-major", True, True)
test_route("valid publish-tag", "workflow_dispatch", "tag", "refs/tags/v1.0.0", "publish-tag", True, True)
test_route("invalid publish-tag", "workflow_dispatch", "branch", "refs/heads/main", "publish-tag", True, False)

print("All routing regression tests passed.")
