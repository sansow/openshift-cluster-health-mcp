# GitHub Actions Quick Start Guide

This guide will help you set up GitHub Actions to run tests against your OpenShift cluster in **3 easy steps**.

## What Was Done

✅ **Fixed Compilation Errors**
- Added `ListInferenceServices()` method to KServeClient
- Enhanced KServe client with Kubernetes CRD support
- All packages now compile successfully

✅ **Updated CI Workflow**
- Modified `.github/workflows/ci.yml` to support OpenShift authentication
- Added kubeconfig setup from GitHub Secrets
- Added connectivity verification step

✅ **Created Setup Scripts**
- `scripts/setup-github-actions-sa.sh` - Automated service account creation
- Full documentation in `docs/github-actions-setup.md`

## Quick Start (3 Steps)

### Step 1: Create Service Account (2 minutes)

Run the automated setup script:

```bash
./scripts/setup-github-actions-sa.sh
```

This will:
- Create a read-only service account
- Generate a 1-year token
- Output the credentials you need

**Example Output:**
```
OPENSHIFT_SERVER:
https://api.cluster-t2wns.t2wns.sandbox1039.opentlc.com:6443

OPENSHIFT_TOKEN:
eyJhbGciOiJSUzI1NiIsImtpZCI6...
```

### Step 2: Add Secrets to GitHub (1 minute)

1. Go to: https://github.com/tosin2013/openshift-cluster-health-mcp/settings/secrets/actions

2. Click **"New repository secret"**

3. Add first secret:
   - Name: `OPENSHIFT_SERVER`
   - Value: `https://api.cluster-t2wns.t2wns.sandbox1039.opentlc.com:6443`
   - Click **"Add secret"**

4. Add second secret:
   - Name: `OPENSHIFT_TOKEN`
   - Value: (paste the token from Step 1)
   - Click **"Add secret"**

### Step 3: Commit and Push Changes (1 minute)

Commit the workflow changes and push to trigger CI:

```bash
git add .github/workflows/ci.yml
git commit -m "feat: add OpenShift credentials support to GitHub Actions"
git push origin main
```

## Verify It's Working

1. Go to: https://github.com/tosin2013/openshift-cluster-health-mcp/actions

2. Click on the latest workflow run

3. Check these steps succeed:
   - ✅ Set up OpenShift kubeconfig
   - ✅ Verify OpenShift connectivity
   - ✅ Run tests

## Alternative: Use Your Current Credentials

If you want to skip the service account setup and use your current login token:

### Get Your Current Credentials

```bash
# Get cluster URL
oc whoami --show-server

# Get your token
oc whoami -t
```

### Add to GitHub Secrets

Follow Step 2 above, but use:
- `OPENSHIFT_SERVER`: Output from `oc whoami --show-server`
- `OPENSHIFT_TOKEN`: Output from `oc whoami -t`

⚠️ **Warning:** User tokens typically expire after 24 hours. For long-term use, create a service account instead.

## Troubleshooting

### "kubectl: command not found" in CI

**Fixed automatically** - The workflow now installs kubectl before running tests.

### "Unauthorized" error in CI

**Cause:** Token expired or invalid

**Fix:**
```bash
# Get a fresh token
oc whoami -t

# Update OPENSHIFT_TOKEN secret in GitHub
```

### Tests pass locally but fail in CI

**Common causes:**
1. **Network latency** - Add timeout adjustments
2. **Permissions** - Verify service account has required permissions:
   ```bash
   oc auth can-i get pods --as=system:serviceaccount:default:github-actions-sa
   ```
3. **Namespace access** - Check if service account can access test namespaces

### No secrets configured

If you see this warning in CI:
```
Skipping 'Set up OpenShift kubeconfig': secrets.OPENSHIFT_SERVER is not set
```

**Fix:** You haven't added the secrets yet. Go to Step 2.

## Security Notes

### ✅ What's Secure

- Service account has **read-only** access
- Token is stored encrypted in GitHub Secrets
- No write/delete permissions
- Limited to cluster monitoring only

### ⚠️ What to Watch

- **Token expiration** - Rotate every year
- **Audit logs** - Monitor service account usage
- **Secret leaks** - Never print secrets in logs

### 🔒 Best Practices

1. **Use service account** instead of user tokens
2. **Rotate secrets** every 90 days minimum
3. **Monitor access** in OpenShift audit logs
4. **Limit scope** to only required namespaces if possible

## What Gets Tested in CI

When GitHub Actions runs, it will:

1. ✅ Connect to your OpenShift cluster
2. ✅ Run all unit tests
3. ✅ Run integration tests (cluster health, pod listing, etc.)
4. ✅ Run race detection tests
5. ✅ Generate code coverage reports
6. ✅ Run linting checks
7. ✅ Build binaries
8. ✅ Run security scans
9. ✅ Lint Helm charts

## Next Steps

After GitHub Actions is working:

1. **Set up branch protection**
   - Require CI to pass before merging
   - Go to Settings → Branches → Add rule

2. **Configure notifications**
   - Get alerts on test failures
   - Settings → Notifications

3. **Monitor token expiration**
   - Set calendar reminder for 11 months from now
   - Or set up automated rotation

## Need Help?

- 📖 Full documentation: [docs/github-actions-setup.md](docs/github-actions-setup.md)
- 🔧 Service account script: [scripts/setup-github-actions-sa.sh](scripts/setup-github-actions-sa.sh)
- 🛡️ Security model: [docs/adrs/007-rbac-based-security-model.md](docs/adrs/007-rbac-based-security-model.md)

## Summary

✅ You now have:
- Updated CI workflow that supports OpenShift authentication
- Automated script to create service accounts
- Comprehensive documentation
- Security best practices

🎯 Next action:
1. Run `./scripts/setup-github-actions-sa.sh`
2. Add secrets to GitHub
3. Push changes
4. Watch CI pass! 🎉
