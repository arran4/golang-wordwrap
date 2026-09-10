import yaml
import sys

def test_job_reachability():
    with open('.github/workflows/ci.yml', 'r') as f:
        ci = yaml.safe_load(f)

    jobs = ci.get('jobs', {})

    # 1. Verify all jobs listed in needs arrays actually exist
    for job_name, job_data in jobs.items():
        needs = job_data.get('needs', [])
        if isinstance(needs, str):
            needs = [needs]
        for needed_job in needs:
            if needed_job not in jobs:
                print(f"FAIL: Job '{job_name}' depends on non-existent job '{needed_job}'")
                sys.exit(1)

    # 2. Goreleaser should NOT depend on prepare-release-tag
    goreleaser = jobs.get('goreleaser', {})
    needs = goreleaser.get('needs', [])
    if isinstance(needs, str):
        needs = [needs]

    if 'prepare-release-tag' in needs:
        print("FAIL: goreleaser still depends on prepare-release-tag")
        sys.exit(1)

    # 3. Release-validation should gracefully handle when run_code_checks is false and use always()
    release_validation = jobs.get('release-validation', {})
    rv_if = release_validation.get('if', '')
    if "always()" not in rv_if:
        print("FAIL: release-validation does not use always()")
        sys.exit(1)
    if "needs.route.outputs.run_code_checks != 'true'" not in rv_if:
        print("FAIL: release-validation does not gracefully handle skipped code checks")
        sys.exit(1)

    # 4. Goreleaser should check run_publisher, not run_release
    gr_if = goreleaser.get('if', '')
    if "needs.route.outputs.run_publisher == 'true'" not in gr_if:
        print("FAIL: goreleaser does not run on run_publisher condition")
        sys.exit(1)

    print("PASS: Job reachability tests")

if __name__ == '__main__':
    test_job_reachability()
