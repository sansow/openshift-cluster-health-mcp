# PR #52 Backporting Plan

## Summary

**PR**: https://github.com/tosin2013/openshift-cluster-health-mcp/pull/52
**Title**: fix: anomaly detection integration with coordination-engine
**Status**: Ready for merge (all CI checks passed)
**Blocking**: Requires 1 approval (branch protection on `main`)

## Changes Overview

### 1. Coordination Engine API v2 Migration
- **Breaking Change**: Replaces action/parameters model with structured resource + issue model
- **Old Request**: `{ action, parameters, priority, confidence }`
- **New Request**: `{ incident_id, namespace, resource{kind, name}, issue{type, severity, description} }`
- **Old Response**: `{ action_id, ... }`
- **New Response**: `{ workflow_id, deployment_method, estimated_duration }`
- **HTTP Status**: Now accepts both 200 OK and 202 Accepted

### 2. Anomaly Detection Configuration
- Default threshold: `0.7` → `0.3` (more sensitive)
- New env var: `ANOMALY_THRESHOLD` (0.0-1.0 range)
- JSON field fix: `patterns` → `anomalies`

### 3. Housekeeping
- Updated `.gitignore` for binaries

## Files Changed
- `internal/tools/analyze_anomalies.go` (12 additions, 2 deletions)
- `internal/tools/trigger_remediation.go` (69 additions, 101 deletions)
- `pkg/clients/coordination_engine.go` (18 additions, 23 deletions)
- `.gitignore` (2 additions)

## CI Status
✅ All checks passed:
- Build (14s)
- Helm (7s)
- Lint (1m9s)
- Security (17s)
- Test (2m34s)
- Trivy (2s)
- Container Build (2m19s)

## ADR Updates
I've updated **ADR-006 (Integration Architecture)** to document:
- Coordination Engine API v2 specification
- API evolution history (v1 → v2)
- Migration notes
- Anomaly detection threshold configuration
- Environment variable documentation

**Commit**: `69ae788` - "docs(adr): Update ADR-006 for Coordination Engine API v2"

---

## Backporting Instructions

### Prerequisites
1. Ensure you have approval to merge PR #52 to `main`
2. Verify Coordination Engine is updated to v2 API on all environments

### Step 1: Merge to Main
```bash
# Get approval for PR #52 first
gh pr review 52 --approve --body "LGTM - API v2 alignment, all checks passed"

# Once approved, merge
gh pr merge 52 --squash --delete-branch
```

### Step 2: Backport to release-4.20
```bash
# Create backport branch from release-4.20
git fetch origin
git checkout -b backport/pr-52-to-4.20 origin/release-4.20

# Cherry-pick the merge commit (use actual commit hash after merge)
git cherry-pick -x <merge-commit-hash>

# Cherry-pick ADR update
git cherry-pick -x 69ae788

# Push and create PR
git push origin backport/pr-52-to-4.20
gh pr create \
  --base release-4.20 \
  --head backport/pr-52-to-4.20 \
  --title "[4.20] fix: anomaly detection integration with coordination-engine" \
  --body "Backport of #52 to release-4.20

## Changes
- Coordination Engine API v2 migration
- Anomaly threshold: 0.7 → 0.3 (configurable via ANOMALY_THRESHOLD)
- JSON field alignment: patterns → anomalies

## Original PR
#52

## Testing
- All CI checks passed on main
- No conflicts during cherry-pick"
```

### Step 3: Backport to release-4.19
```bash
git checkout -b backport/pr-52-to-4.19 origin/release-4.19
git cherry-pick -x <merge-commit-hash>
git cherry-pick -x 69ae788
git push origin backport/pr-52-to-4.19
gh pr create \
  --base release-4.19 \
  --head backport/pr-52-to-4.19 \
  --title "[4.19] fix: anomaly detection integration with coordination-engine" \
  --body "Backport of #52 to release-4.19

## Changes
- Coordination Engine API v2 migration
- Anomaly threshold: 0.7 → 0.3 (configurable via ANOMALY_THRESHOLD)
- JSON field alignment: patterns → anomalies

## Original PR
#52"
```

### Step 4: Backport to release-4.18
```bash
git checkout -b backport/pr-52-to-4.18 origin/release-4.18
git cherry-pick -x <merge-commit-hash>
git cherry-pick -x 69ae788
git push origin backport/pr-52-to-4.18
gh pr create \
  --base release-4.18 \
  --head backport/pr-52-to-4.18 \
  --title "[4.18] fix: anomaly detection integration with coordination-engine" \
  --body "Backport of #52 to release-4.18

## Changes
- Coordination Engine API v2 migration
- Anomaly threshold: 0.7 → 0.3 (configurable via ANOMALY_THRESHOLD)
- JSON field alignment: patterns → anomalies

## Original PR
#52"
```

---

## Release Branch Protection Requirements

Per ADR-014, release branches have stricter requirements:
- **release-4.18**: 2 required approvals
- **release-4.19**: 2 required approvals
- **release-4.20**: 2 required approvals

Ensure you get appropriate approvals for each backport PR.

---

## Validation Checklist

After merging to each branch:

### Main Branch
- [ ] PR #52 merged
- [ ] ADR-006 updated (commit 69ae788)
- [ ] Container image built and pushed
- [ ] Helm chart version bumped (if needed)

### Release Branches (4.18, 4.19, 4.20)
- [ ] Backport PR created
- [ ] 2 approvals obtained
- [ ] CI checks passed
- [ ] No merge conflicts
- [ ] ADR-006 updated on release branch
- [ ] Container image tagged for release branch

---

## Deployment Verification

After backporting, verify on each environment:

```bash
# Check if Coordination Engine API is v2
curl -X POST http://coordination-engine:8080/api/v1/remediation/trigger \
  -H 'Content-Type: application/json' \
  -d '{
    "incident_id": "test",
    "namespace": "default",
    "resource": {"kind": "Deployment", "name": "test"},
    "issue": {"type": "pod_crash", "severity": "high", "description": "test"},
    "dry_run": true
  }'

# Expected response should include workflow_id, not action_id

# Test anomaly detection threshold
kubectl set env deployment/cluster-health-mcp ANOMALY_THRESHOLD=0.3 -n cluster-health-mcp-dev
kubectl rollout status deployment/cluster-health-mcp -n cluster-health-mcp-dev
```

---

## Rollback Plan

If issues arise after backporting:

```bash
# Revert on specific branch
git checkout <release-branch>
git revert <merge-commit-hash>
git push origin <release-branch>

# Or revert the backport PR
gh pr revert <backport-pr-number>
```

---

## Communication

After completing all backports:
1. Update the original PR #52 with backport links
2. Notify team in Slack/email:
   - API v2 migration completed
   - New ANOMALY_THRESHOLD environment variable available
   - Updated ADR-006 for reference

---

## Questions?

- **Why backport to all releases?** This aligns with the updated Coordination Engine API across all supported versions
- **Is this a breaking change?** Yes, for Coordination Engine integration. Ensure CE is updated first
- **Can we skip a release?** Not recommended - creates API version skew issues

---

## Next Steps

1. Get approval and merge PR #52 to `main`
2. Execute backport steps sequentially (4.20 → 4.19 → 4.18)
3. Verify deployments on each environment
4. Update documentation with new API examples
