# Release Process

This document describes the release process for terraform-provider-wallix-bastion.

## Quick Start

For a standard patch release:

```bash
# 1. Prepare the release
make release-patch

# 2. Push any remaining changes
git push origin develop

# 3. Create GitHub release from the new tag
```

## Detailed Process

### 1. Prerequisites

- Clean git working directory
- On `develop` branch (recommended)
- Go 1.25+ installed
- golangci-lint installed
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

### 6. Rollback

If you need to undo a release:

```bash
# Delete local tag
git tag -d v0.14.8

# Delete remote tag (if already pushed)
git push origin --delete v0.14.8

# Reset any committed changes
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

#### "Not on develop branch"

```bash
git checkout develop
git pull origin develop
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

1. **On tag push**: Triggers release workflow
2. **Release workflow**: Builds artifacts, runs tests, creates GitHub release
3. **Documentation**: Updates provider registry documentation

Check `.github/workflows/` for workflow definitions.

### 11. Best Practices

- Always test releases in a staging environment first
- Keep CHANGELOG.md up to date throughout development
- Use feature branches for new functionality
- Tag releases promptly after merging to main
- Monitor release workflows for failures
- Communicate breaking changes clearly in release notes
