# Release Scripts

This directory contains scripts to help with the release process of the terraform-provider-wallix-bastion.

## prepare-release.sh

A comprehensive script that automates the release preparation process.

### Usage

```bash
# Prepare a patch release (default)
./scripts/prepare-release.sh

# Prepare a minor release
./scripts/prepare-release.sh --minor

# Prepare a major release
./scripts/prepare-release.sh --major

# Use a specific version
./scripts/prepare-release.sh --version v0.15.0

# Skip tests (for emergency releases)
./scripts/prepare-release.sh --skip-tests

# Dry run to see what would be done
./scripts/prepare-release.sh --dry-run

# Show help
./scripts/prepare-release.sh --help
```

### What the script does

1. **Checks prerequisites**: Ensures git working directory is clean and you're on the correct branch
2. **Updates dependencies**: Runs `go get -u all` and `go mod tidy`
3. **Runs quality checks**: Executes golangci-lint and unit tests
4. **Updates changelog**: Reminds to update the changelog (manual step)
5. **Creates git tag**: Creates an annotated tag with the new version
6. **Generates release notes**: Shows commits and changed files since last release

### Prerequisites

- Git repository with existing tags following semantic versioning (vX.Y.Z)
- Go 1.25+ installed (matches the `go` directive in `go.mod`)
- golangci-lint installed (optional, will be skipped if not available)
- Clean git working directory
- Recommended to be on `main`, with `develop` already merged in

### Example workflow

```bash
# 1. Merge develop into main and make sure main is up to date
git checkout main
git pull origin main
git merge origin/develop
git push origin main

# 2. Run the release script
./scripts/prepare-release.sh --minor

# 3. Review the generated changelog and commit any manual updates
git add CHANGELOG.md
git commit -m "docs: finalize changelog for v0.15.0"

# 4. Push the tag (if not done automatically)
git push origin v0.15.0

# 5. Create GitHub release from the tag
```

### Version numbering

The script follows semantic versioning:

- **Patch** (X.Y.Z+1): Bug fixes and small improvements
- **Minor** (X.Y+1.0): New features, backward compatible
- **Major** (X+1.0.0): Breaking changes

### Environment variables

You can set these environment variables to customize behavior:

- `DRY_RUN=true`: Run in dry-run mode
- `SKIP_TESTS=true`: Skip running tests and linters

### Troubleshooting

#### "Working directory is not clean"

- Commit or stash your changes before running the script

#### "golangci-lint not found"

- Install golangci-lint or use `--skip-tests` flag

#### "Not on main branch"

- Merge `develop` into `main` and switch to `main`, or continue with the current
  branch when prompted

#### "Tag already exists"

- Use `--version` to specify a different version or increment the existing tag manually
