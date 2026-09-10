import yaml
import sys

def test_job_reachability():
    with open('.github/workflows/ci.yml', 'r') as f:
        ci = yaml.safe_load(f)

    jobs = ci.get('jobs', {})

    # 1. Goreleaser should NOT depend on prepare-release-tag
    goreleaser = jobs.get('goreleaser', {})
    needs = goreleaser.get('needs', [])
    if isinstance(needs, str):
        needs = [needs]

    if 'prepare-release-tag' in needs:
        print("FAIL: goreleaser still depends on prepare-release-tag")
        sys.exit(1)

    # 2. Release-validation should gracefully handle when run_code_checks is false
    release_validation = jobs.get('release-validation', {})
    rv_if = release_validation.get('if', '')
    if "needs.route.outputs.run_code_checks != 'true'" not in rv_if:
        print("FAIL: release-validation does not gracefully handle skipped code checks")
        sys.exit(1)

    print("PASS: Job reachability tests")

if __name__ == '__main__':
    test_job_reachability()
