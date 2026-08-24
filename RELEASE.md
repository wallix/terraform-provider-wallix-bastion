# Release Process

This document describes the release process for terraform-provider-wallix-bastion.

## Quick Start

For a standard patch release:

```bash
# 1. Merge develop into main and make sure main is up to date
git checkout main
git pull origin main
git merge origin/develop
git push origin main

# 2. Prepare the release: updates deps, runs lint/test/build, creates an
#    annotated tag, and prompts you to push the tag to origin
make release-patch

# 3. Push any commits the script made on main (changelog/dependency updates)
git push origin main

# 4. If you answered "no" to the push-tag prompt, push it now - the release
#    workflow only runs on a tag actually reaching the remote, on any branch
git push origin vX.Y.Z

# 5. Create GitHub release from the new tag
```

## Detailed Process

### 1. Prerequisites

- Clean git working directory
- On `main` branch, with `develop` already merged in — releases are tagged from
  `main` after it's brought up to date with `develop`, not from `develop` directly
- Go 1.25+ installed (matches the `go` directive in `go.mod`)
- golangci-lint installed (recommended, not enforced — `prepare-release.sh` skips
  linting with just a warning if it isn't found, so don't rely on it as a safety net)
- Git configured with proper user information

### 2. Release Types

#### Patch Release (X.Y.Z+1)

For bug fixes and small improvements:

```bash
make release-patch
# or
./scripts/prepare-release.sh --patch
```

#### Minor Release (X.Y+1.0)

For new features (backward compatible):

```bash
make release-minor
# or
./scripts/prepare-release.sh --minor
```

#### Major Release (X+1.0.0)

For breaking changes:

```bash
make release-major
# or
./scripts/prepare-release.sh --major
```

#### Custom Version

For specific version numbers:

```bash
./scripts/prepare-release.sh --version v1.0.0
```

### 3. What the Release Script Does

1. **Validates Environment**
   - Checks git working directory is clean
   - Ensures you're on correct branch
   - Verifies prerequisites

2. **Updates Dependencies**
   - Updates Go module dependencies
   - Runs `go mod tidy` and `go mod verify`

3. **Quality Checks**
   - Runs golangci-lint
   - Executes unit tests
   - Builds the provider to ensure compilation

4. **Creates Tag**
   - Determines next version number
   - Creates annotated git tag
   - Optionally pushes tag to remote

5. **Generates Release Notes**
   - Shows commits since last release
   - Lists changed files
   - Provides template for GitHub release

### 4. Manual Steps

After running the release script, you may need to:

1. **Update CHANGELOG.md**
   - Finalize release notes
   - Set correct release date
   - Commit changes if needed

2. **Create GitHub Release**
   - Go to GitHub releases page
   - Create release from the new tag
   - Use generated release notes as template
   - Attach any additional assets

3. **Verify Release**
   - Check that GitHub Actions workflows complete
   - Verify release artifacts are created
   - Test the released version

### 5. Emergency Releases

For urgent fixes that need to skip some checks:

```bash
./scripts/prepare-release.sh --patch --skip-tests
```

`--skip-tests` skips lint, unit tests, and the local build check — nothing else
runs them before the tag is pushed (the tag-triggered release workflow only
builds and publishes, see [CI/CD Integration](#10-cicd-integration)), so treat
this as accepting that risk knowingly, not as deferring verification to CI.

### 6. Rollback

If you need to undo a release (replace `v0.14.8` with the actual tag):

```bash
# Delete local tag
git tag -d v0.14.8

# Delete remote tag (if already pushed) - this does NOT remove a GitHub Release
# that GoReleaser already published from it; delete that separately from the
# GitHub releases page if the workflow already ran
git push origin --delete v0.14.8

# Only if the release script committed changes (e.g. dependency updates) that
# have NOT been pushed yet - this rewrites local history, so never run it on a
# commit already pushed to main; coordinate a revert instead in that case
git reset --hard HEAD~1
```

### 7. Maintenance Tasks

Regular maintenance without creating a release:

```bash
# Run all maintenance tasks
make maintenance

# Individual tasks
make update-deps    # Update dependencies only
make dev-check      # Run linting, tests, and build
./scripts/maintenance.sh test  # Run tests only
./scripts/maintenance.sh clean # Clean build artifacts
```

### 8. Version Numbering Guidelines

Follow semantic versioning:

- **MAJOR**: Breaking changes, incompatible API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible changes

Examples:

- `v0.14.7` → `v0.14.8` (patch: bug fix)
- `v0.14.7` → `v0.15.0` (minor: new feature)
- `v0.14.7` → `v1.0.0` (major: breaking change)

### 9. Troubleshooting

#### "Working directory is not clean"

```bash
git status
git add .
git commit -m "chore: prepare for release"
```

#### "Not on main branch"

```bash
git checkout main
git pull origin main
git merge origin/develop
git push origin main
```

#### "Tests failing"

```bash
make dev-check  # Run individual checks
./scripts/maintenance.sh test  # Run tests only
```

#### "Tag already exists"

```bash
git tag -d v0.14.8  # Delete local tag
./scripts/prepare-release.sh --version v0.14.9  # Use next version
```

#### "Dependencies update failed"

```bash
go mod tidy
go clean -modcache
go mod download
```

### 10. CI/CD Integration

The release process integrates with GitHub Actions:

1. **On tag push**: Any `v*` tag pushed to any branch triggers `.github/workflows/release.yml`
2. **Release workflow**: Runs `go mod tidy`, builds and signs artifacts with GoReleaser,
   and creates the GitHub release — it does **not** re-run lint/test/build, so those
   quality gates only ever run locally via `prepare-release.sh` before you tag (see
   [Emergency Releases](#5-emergency-releases) for what `--skip-tests` bypasses)
3. **Documentation**: Not a CI step — the Terraform Registry picks up `docs/` from the
   tagged commit on its own when it indexes the new release

Check `.github/workflows/` for workflow definitions.

### 11. Best Practices

- Always test releases in a staging environment first
- Keep CHANGELOG.md up to date throughout development
- Use feature branches for new functionality
- Merge `develop` into `main` before tagging — releases are cut from `main`, not
  `develop`, so a release that skips this step ships whatever `main` last had
- Confirm the tag actually reached the remote before walking away — since
  `release.yml` triggers on any pushed `v*` tag, an un-pushed local tag just does
  nothing silently, and a tag pushed from the wrong branch/commit still triggers
  a real release
- Monitor release workflows for failures
- Communicate breaking changes clearly in release notes
