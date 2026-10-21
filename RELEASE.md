# Release Process

This document describes the release process for terraform-provider-wallix-bastion.

Releases are tagged from `main`, once `develop` is merged into it and `CHANGELOG.md` is finalized.
Pushing the tag publishes the release: GoReleaser builds, signs and publishes the GitHub release and
its notes, and the Terraform Registry picks it up.

## Quick Start

```bash
# 1. On develop, through a pull request: finalize CHANGELOG.md, renaming "## Unreleased" to
#    "## X.Y.Z (Month DD, YYYY)", e.g. "## 0.16.0 (October 21, 2026)"

# 2. Merge develop into main
git checkout main
git pull origin main
git merge origin/develop

# 3. Check that everything is ready: shows the version CHANGELOG.md calls for and lists every
#    problem, without pushing anything
make prepare-release

# 4. Release: runs the checks, pushes main, then creates and pushes the tag (asking before each push)
make release-minor   # or release-patch, release-major, as CHANGELOG.md calls for

# 5. Follow the release workflow, then add a "## Unreleased" section to CHANGELOG.md on develop
gh run watch
```

Don't create the GitHub release by hand: the release workflow publishes it.

## Detailed Process

### 1. Prerequisites

- The current branch is `main`, clean, and contains `origin/main` and `origin/develop`. The script
  fetches them and refuses to go on otherwise.
- `CHANGELOG.md` is finalized for the version (see [CHANGELOG](#3-changelog)), through a pull
  request on `develop`.
- Dependencies are up to date on `develop`, through Dependabot or pull requests (`make update-deps`).
  The script doesn't update them, so that a release only ships what went through the CI of
  `develop`.
- Go as required by `go.mod` (`toolchain` directive, Go 1.26.9 today). With the default
  `GOTOOLCHAIN=auto`, Go downloads it.
- golangci-lint v2, and for the docs check, `tfplugindocs` and `npx` (markdownlint-cli2). The script
  warns and skips the lint or the docs check if they are missing, as the CI runs them on pull
  requests.

### 2. Versions

The script compares the bump to the top section of `CHANGELOG.md`:

| Top section has | Release |
| --- | --- |
| `BREAKING CHANGES:` | major from 1.0.0 on; minor before (0.x) |
| `FEATURES:` | minor |
| only `ENHANCEMENTS:`, `BUG FIXES:`, `NOTES:` | patch |

- Without `--major`, `--minor` or `--patch`, it uses the bump the CHANGELOG calls for
  (`make prepare-release`, `make release`).
- A smaller bump is refused, unless `--ignore-changelog-type` is given.
- A larger one is allowed, e.g. `--major` for 1.0.0.

```bash
make release-patch                                # v0.16.0 -> v0.16.1
make release-minor                                # v0.15.2 -> v0.16.0
make release-major                                # v0.16.0 -> v1.0.0
make release ARGS="--version v0.16.1"             # a given version
```

Flags must go through `ARGS`: `make release-minor --version vX` passes `--version` to make, which
prints its own version.

#### Release candidates

```bash
make release ARGS="--major --rc"                  # v0.16.0 -> v1.0.0-rc1, then v1.0.0-rc2, ...
make release-major                                # v1.0.0-rc2 -> v1.0.0
make release ARGS="--version v1.0.0-rc3"          # a given candidate
```

Release candidates don't count as releases when computing the next version. GoReleaser publishes
them as pre-releases (`prerelease: auto`).

### 3. CHANGELOG

For a release `vX.Y.Z`, the first level-2 heading of `CHANGELOG.md` must be `## X.Y.Z (Month DD, YYYY)`,
for example `## 0.16.0 (October 21, 2026)`. A date other than today only gives a warning.

For a release candidate, it may also be `## Unreleased` (with a warning) or
`## X.Y.Z-rcN (Month DD, YYYY)`.

### 4. What the Release Script Does

1. **Version:** fetches the tags, takes the latest `vX.Y.Z` release, applies the bump (or
   `--version`), and checks it against the CHANGELOG sections.
2. **Git checks:** fetches `main`, `develop` and the tags from `origin`, then checks that:
   - the current branch is `main`, and the working directory is clean;
   - `HEAD` contains `origin/main` and `origin/develop`;
   - the tag exists neither locally nor on `origin`.
3. **CHANGELOG:** checks that it is finalized for the version.
4. **Quality checks** on the commit to tag:
   - golangci-lint, `go vet`, unit tests, `make build`;
   - govulncheck, with the Go of `go.mod` and the same govulncheck version as the CI;
   - a docs check: `docs/` must match a fresh `tfplugindocs generate` and markdownlint-cli2;
   - the working directory must still be clean afterwards.
5. **Publish:** pushes `main` first, then creates and pushes the annotated tag. It asks before each
   push. Answering no stops without a tag, so a stale local tag can't break the next run.

`--dry-run` runs steps 1 to 3 and lists every problem, without the quality checks and without
pushing. It exits with an error if something must be fixed, so it can also run in CI.

### 5. Emergency Releases

```bash
./scripts/prepare-release.sh --patch --skip-tests
```

`--skip-tests` skips the quality checks of step 4. Nothing else runs them before the tag is pushed:
the release workflow only builds and publishes (see [CI/CD Integration](#9-cicd-integration)). So
this means accepting that risk knowingly, not deferring verification to CI.

### 6. Rollback

If you need to undo a release (replace `v0.16.0` with the actual tag):

```bash
# Delete the local and remote tag. This does NOT remove a GitHub release that GoReleaser already
# published from it: delete that separately from the GitHub releases page.
git tag -d v0.16.0
git push origin --delete v0.16.0
```

A version that reached the Terraform Registry should not be reused: release a new patch version
with the fix instead.

### 7. Maintenance Tasks

Regular maintenance, on `develop`, without creating a release:

```bash
make maintenance     # all maintenance tasks
make update-deps     # update dependencies (commit them through a pull request on develop)
make dev-check       # linting, tests and build
```

### 8. Troubleshooting

| Error | Fix |
| --- | --- |
| `Current branch is '...'` | `git checkout main` |
| `Working directory is not clean` | commit or stash the changes |
| `HEAD doesn't contain origin/main` | `git pull origin main` |
| `HEAD doesn't contain origin/develop` | `git merge origin/develop` (develop moved since the last merge) |
| `CHANGELOG.md isn't finalized` | finalize it through a pull request on develop, then merge develop into main |
| `... is a patch release, but CHANGELOG.md calls for a minor one` | use the bump the error gives |
| `Tag ... already exists` | the version is taken: release the next one |
| `docs/ is out of date` | run `tfplugindocs generate` and markdownlint on develop, and merge it |

### 9. CI/CD Integration

1. **On tag push**: any `v*` tag triggers `.github/workflows/release.yml`.
2. **Release workflow**: GoReleaser builds with the Go of `go.mod`, signs the artifacts, and
   publishes the GitHub release with notes generated from the commits. It does **not** re-run
   lint, tests or build: those only run locally through the script before the tag is pushed (see
   [Emergency Releases](#5-emergency-releases) for what `--skip-tests` bypasses).
3. **Documentation**: the Terraform Registry reads `docs/` from the tagged commit when it indexes the
   new release.
