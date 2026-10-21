#!/bin/bash

# Terraform Provider Wallix-Bastion release script.
#
# Releases are tagged from main, once develop is merged into it and CHANGELOG.md is finalized
# (through a pull request on develop). The script:
# 1. Fetches main, develop and the tags, and checks that the current branch is main, that it is
#    clean, contains origin/main and origin/develop, and that the tag doesn't exist yet.
# 2. Determines the version from the requested bump (or the one CHANGELOG.md calls for), and
#    checks that CHANGELOG.md is finalized for it.
# 3. Runs the linters, unit tests, build, govulncheck and the docs check on the commit to tag.
# 4. Pushes main, then creates and pushes the annotated tag. The tag triggers the release
#    workflow: GoReleaser builds, signs and publishes the GitHub release and its notes, and the
#    Terraform Registry picks it up.
#
# Dependencies are not updated here: Dependabot and pull requests do it on develop, so that a
# release only ships what went through the CI of develop.

set -euo pipefail

REMOTE="${RELEASE_REMOTE:-origin}"
CHANGELOG_FILE="CHANGELOG.md"
GOVULNCHECK_VERSION="v1.9.0" # same as the CI (linters.yml, security.yml)

FINAL_RE='^v([0-9]+)\.([0-9]+)\.([0-9]+)$'
RC_RE='^v([0-9]+)\.([0-9]+)\.([0-9]+)-rc([0-9]+)$'

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

DRY_RUN=false
PROBLEMS=0

# fail stops the release. In dry-run mode, it reports the problem and goes on, so that a single
# run lists every problem; the dry run then exits with an error.
fail() {
    log_error "$1"
    if [[ "$DRY_RUN" == true ]]; then
        PROBLEMS=$((PROBLEMS + 1))
    else
        exit 1
    fi
}

usage() {
    cat <<'EOF'
Usage: scripts/prepare-release.sh [OPTIONS]

Options:
  --major                  Increment the major version
  --minor                  Increment the minor version
  --patch                  Increment the patch version
                           Without any of them, the bump CHANGELOG.md calls for is used.
  --rc                     Release a candidate of that version: vX.Y.Z-rc1, then -rc2, ...
  --version V              Release version V (vX.Y.Z or vX.Y.Z-rcN)
  --ignore-changelog-type  Allow a smaller bump than the CHANGELOG.md sections call for
  --skip-tests             Skip the linters, tests, build, govulncheck and docs checks
  --dry-run                Run the checks and show what would be done, without pushing
  --help, -h               Show this help message

Examples:
  scripts/prepare-release.sh --dry-run          # checks, and the version CHANGELOG.md calls for
  scripts/prepare-release.sh --minor            # v0.15.2 -> v0.16.0
  scripts/prepare-release.sh --major --rc       # v0.16.0 -> v1.0.0-rc1, then v1.0.0-rc2
  scripts/prepare-release.sh --major            # v1.0.0-rc2 -> v1.0.0 (v0.16.0 -> v1.0.0)
EOF
}

# latest_final_tag prints the highest vX.Y.Z tag, or v0.0.0 when there is none.
latest_final_tag() {
    local tag
    tag=$(git tag --list 'v*' | grep -E "$FINAL_RE" | sort -V | tail -1 || true)
    echo "${tag:-v0.0.0}"
}

# bump_version prints the version after a major, minor or patch bump of a vX.Y.Z version.
bump_version() {
    local version=$1 type=$2 major minor patch
    [[ "$version" =~ $FINAL_RE ]] || return 1
    major=${BASH_REMATCH[1]} minor=${BASH_REMATCH[2]} patch=${BASH_REMATCH[3]}
    case $type in
        major) echo "v$((major + 1)).0.0" ;;
        minor) echo "v${major}.$((minor + 1)).0" ;;
        patch) echo "v${major}.${minor}.$((patch + 1))" ;;
        *) return 1 ;;
    esac
}

# next_rc prints the next release candidate of a vX.Y.Z version: -rc1, or the one after the
# highest existing -rcN.
next_rc() {
    local base=$1 last
    last=$(git tag --list "${base}-rc*" | grep -E "$RC_RE" | sed 's/.*-rc//' | sort -n | tail -1 || true)
    echo "${base}-rc$(( ${last:-0} + 1 ))"
}

# bump_type prints the bump from a vX.Y.Z version to another: major, minor, patch, or none.
bump_type() {
    local from=$1 to=$2 f_major f_minor t_major t_minor
    [[ "$from" =~ $FINAL_RE ]] && f_major=${BASH_REMATCH[1]} f_minor=${BASH_REMATCH[2]}
    [[ "$to" =~ $FINAL_RE ]] && t_major=${BASH_REMATCH[1]} t_minor=${BASH_REMATCH[2]}
    if ((t_major != f_major)); then
        echo major
    elif ((t_minor != f_minor)); then
        echo minor
    elif [[ "$from" != "$to" ]]; then
        echo patch
    else
        echo none
    fi
}

type_rank() {
    case $1 in major) echo 3 ;; minor) echo 2 ;; patch) echo 1 ;; *) echo 0 ;; esac
}

# changelog_top_heading prints the first "## " heading of CHANGELOG.md: the section to release.
changelog_top_heading() {
    grep -m1 '^## ' "$CHANGELOG_FILE" || true
}

# changelog_top_section prints the body of that section, up to the next "## " heading.
changelog_top_section() {
    awk '/^## /{n++; next} n==1' "$CHANGELOG_FILE"
}

# changelog_type prints the bump the top section calls for: BREAKING CHANGES is a major bump
# from 1.0.0 on and a minor one before (0.x), FEATURES is a minor bump, anything else a patch.
changelog_type() {
    local current_major=$1 section
    section=$(changelog_top_section)
    if grep -q '^BREAKING CHANGES:' <<<"$section"; then
        if ((current_major == 0)); then echo minor; else echo major; fi
    elif grep -q '^FEATURES:' <<<"$section"; then
        echo minor
    else
        echo patch
    fi
}

check_git() {
    local new_version=$1 branch
    log_info "Fetching main, develop and the tags from ${REMOTE}..."
    git fetch --quiet "$REMOTE" main develop --tags

    branch=$(git branch --show-current)
    if [[ "$branch" != "main" ]]; then
        fail "Current branch is '${branch:-detached HEAD}': releases are tagged from main (git checkout main)."
    fi
    if [[ -n $(git status --porcelain) ]]; then
        fail "Working directory is not clean. Commit or stash the changes first."
    fi
    if ! git merge-base --is-ancestor "${REMOTE}/main" HEAD; then
        fail "HEAD doesn't contain ${REMOTE}/main: update main first (git pull ${REMOTE} main)."
    fi
    if ! git merge-base --is-ancestor "${REMOTE}/develop" HEAD; then
        fail "HEAD doesn't contain ${REMOTE}/develop: merge it first (git merge ${REMOTE}/develop)."
    fi
    if git rev-parse --quiet --verify "refs/tags/${new_version}" >/dev/null; then
        fail "Tag ${new_version} already exists locally."
    fi
    if git ls-remote --exit-code --tags "$REMOTE" "refs/tags/${new_version}" >/dev/null; then
        fail "Tag ${new_version} already exists on ${REMOTE}."
    fi
}

# check_changelog checks that the top section of CHANGELOG.md is the release: "## X.Y.Z (Month DD,
# YYYY)". A release candidate may also be released from "## Unreleased" or "## X.Y.Z-rcN (...)".
check_changelog() {
    local new_version=$1 final_version=$2 heading today expected version_re
    heading=$(changelog_top_heading)
    version_re=${final_version#v}
    version_re=${version_re//./\\.}
    today=$(LC_ALL=C date '+%B %d, %Y')
    expected="## ${final_version#v} (${today})"

    if [[ "$new_version" =~ $RC_RE ]]; then
        if [[ "$heading" == "## Unreleased" ]]; then
            log_warning "CHANGELOG.md: releasing ${new_version} from '## Unreleased'."
            return
        fi
        if [[ "$heading" =~ ^##\ ${version_re}(-rc[0-9]+)?\ \( ]]; then
            return
        fi
        fail "CHANGELOG.md: the top section is '${heading}', expected '## Unreleased' or '## ${final_version#v} (...)' for ${new_version}."
        return
    fi

    if [[ "$heading" == "## Unreleased" ]]; then
        fail "CHANGELOG.md isn't finalized: rename '## Unreleased' to '${expected}' through a pull request on develop, then merge develop into main."
        return
    fi
    if [[ ! "$heading" =~ ^##\ ${version_re}\ \(([A-Z][a-z]+\ [0-9]{2},\ [0-9]{4})\)$ ]]; then
        fail "CHANGELOG.md: the top section is '${heading}', expected '${expected}' (or another release date)."
        return
    fi
    if [[ "$heading" != "$expected" ]]; then
        log_warning "CHANGELOG.md: the release date is '${BASH_REMATCH[1]}', not today (${today})."
    fi
}

check_toolchain() {
    local wanted current
    wanted=$(awk '$1 == "toolchain" {print $2}' go.mod)
    current=$(go env GOVERSION)
    current=${current%%-X:*} # drop the GOEXPERIMENT suffix, e.g. go1.26.9-X:nodwarf5
    if [[ -n "$wanted" && "$current" != "$wanted" ]]; then
        log_warning "The checks run with ${current}, go.mod asks for ${wanted} (check GOTOOLCHAIN)."
    else
        log_info "Go: ${current}"
    fi
}

check_docs() {
    if ! command -v tfplugindocs >/dev/null 2>&1 || ! command -v npx >/dev/null 2>&1; then
        log_warning "tfplugindocs or npx not found, skipping the docs check (the CI runs it)."
        return
    fi
    log_info "Checking that docs/ matches a fresh tfplugindocs generation..."
    # The provider name comes from the directory name by default: set it, so that a clone under
    # another name gives the same page titles.
    tfplugindocs generate --provider-name terraform-provider-wallix-bastion >/dev/null
    npx --yes markdownlint-cli2 --fix "docs/**/*.md" >/dev/null 2>&1 || true
    if [[ -n $(git status --porcelain -- docs/) ]]; then
        git status --short -- docs/ >&2
        git checkout --quiet -- docs/
        git clean --quiet -fd -- docs/
        fail "docs/ is out of date: run 'tfplugindocs generate' and markdownlint on develop, and merge it."
        return
    fi
    log_success "Docs are up to date"
}

run_checks() {
    log_info "Running the linters, tests and build on $(git rev-parse --short HEAD)..."
    check_toolchain

    if command -v golangci-lint &>/dev/null; then
        # .golangci.yml uses the v2 config schema (required by golangci-lint-action@v9);
        # a v1 binary fails with a cryptic error, so check up front and point at the fix.
        local lint_version lint_major
        lint_version=$(golangci-lint --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        lint_major=${lint_version%%.*}
        if [[ -n "$lint_major" && "$lint_major" != "2" ]]; then
            log_error "golangci-lint v${lint_version} found, but .golangci.yml requires v2."
            log_error "Fix with: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
            exit 1
        fi
        golangci-lint run
        log_success "Linting passed"
    else
        log_warning "golangci-lint not found, skipping linting"
    fi

    go vet ./...
    go test ./...
    log_success "Tests passed"

    make build
    log_success "Build successful"

    go run "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}" ./...
    log_success "No vulnerability found"

    check_docs

    if [[ -n $(git status --porcelain) ]]; then
        git status --short >&2
        fail "The checks left changes in the working directory."
    fi
}

confirm() {
    local reply
    read -r -p "$1 (y/N): " reply
    [[ $reply =~ ^[Yy]$ ]]
}

main() {
    local release_type="" rc=false force_version="" skip_tests=false ignore_type=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --major | --minor | --patch)
                release_type=${1#--}
                shift
                ;;
            --rc)
                rc=true
                shift
                ;;
            --version)
                [[ $# -ge 2 ]] || { log_error "--version needs a value"; exit 1; }
                force_version=$2
                shift 2
                ;;
            --ignore-changelog-type)
                ignore_type=true
                shift
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --help | -h)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    if [[ ! -f "$CHANGELOG_FILE" ]]; then
        log_error "${CHANGELOG_FILE} not found: run the script from the root of the repository."
        exit 1
    fi

    # Tags first: the version depends on them.
    git fetch --quiet "$REMOTE" --tags

    local current_tag current_major suggested new_version final_version
    current_tag=$(latest_final_tag)
    [[ "$current_tag" =~ $FINAL_RE ]] && current_major=${BASH_REMATCH[1]}
    suggested=$(changelog_type "$current_major")
    log_info "Latest release: ${current_tag}"
    log_info "CHANGELOG.md top section: '$(changelog_top_heading)' (calls for a ${suggested} release)"

    if [[ -n "$force_version" ]]; then
        if [[ -n "$release_type" || "$rc" == true ]]; then
            log_error "--version can't be combined with --major, --minor, --patch or --rc."
            exit 1
        fi
        if [[ ! "$force_version" =~ $FINAL_RE && ! "$force_version" =~ $RC_RE ]]; then
            log_error "Invalid version format: ${force_version} (expected vX.Y.Z or vX.Y.Z-rcN)"
            exit 1
        fi
        new_version=$force_version
    else
        if [[ -z "$release_type" ]]; then
            release_type=$suggested
            log_info "No bump requested: using the ${release_type} release CHANGELOG.md calls for."
        fi
        new_version=$(bump_version "$current_tag" "$release_type")
        if [[ "$rc" == true ]]; then
            new_version=$(next_rc "$new_version")
        fi
    fi
    final_version=${new_version%-rc*}

    if [[ "$final_version" == "$current_tag" ]] || [[ $(printf '%s\n' "$current_tag" "$final_version" | sort -V | tail -1) != "$final_version" ]]; then
        log_error "Version ${new_version} isn't after the latest release ${current_tag}."
        exit 1
    fi

    local actual
    actual=$(bump_type "$current_tag" "$final_version")
    if (($(type_rank "$actual") < $(type_rank "$suggested"))) && [[ "$ignore_type" == false ]]; then
        fail "${new_version} is a ${actual} release, but CHANGELOG.md calls for a ${suggested} one (FEATURES or BREAKING CHANGES): use --${suggested}, or --ignore-changelog-type."
    fi

    log_info "Next version: ${new_version}"

    check_git "$new_version"
    check_changelog "$new_version" "$final_version"

    local ahead
    ahead=$(git rev-list --count "${REMOTE}/main..HEAD" 2>/dev/null || echo "?")

    if [[ "$DRY_RUN" == true ]]; then
        log_info "DRY RUN - the release would:"
        echo "  1. Run the linters, unit tests, build, govulncheck and the docs check"
        echo "  2. Push main to ${REMOTE} (${ahead} commit(s) ahead)"
        echo "  3. Create and push the annotated tag ${new_version}, which triggers the release workflow"
        echo ""
        log_info "Commits since ${current_tag}:"
        git log "${current_tag}..HEAD" --oneline --no-merges | sed 's/^/  /'
        if ((PROBLEMS > 0)); then
            log_error "${PROBLEMS} problem(s) to fix before releasing ${new_version}."
            exit 1
        fi
        log_success "Ready to release ${new_version}"
        exit 0
    fi

    if [[ "$skip_tests" == false ]]; then
        run_checks
    else
        log_warning "Skipping the checks as requested: nothing else runs them before the tag is pushed."
    fi

    log_info "Commits since ${current_tag}:"
    git log "${current_tag}..HEAD" --oneline --no-merges | sed 's/^/  /'

    # main first: the tag must point to a commit that is on the main branch of the remote.
    if [[ "$ahead" != "0" ]]; then
        confirm "Push main (${ahead} commit(s) ahead of ${REMOTE}/main) to ${REMOTE}?" || {
            log_info "Release cancelled: nothing was pushed."
            exit 0
        }
        git push "$REMOTE" main
        log_success "main pushed"
    fi

    log_warning "Pushing tag ${new_version} triggers the release workflow and publishes to the Terraform Registry."
    confirm "Create and push tag ${new_version} on $(git rev-parse --short HEAD)?" || {
        log_info "Release cancelled: no tag created."
        exit 0
    }
    git tag -a "$new_version" -m "Release ${new_version}"
    git push "$REMOTE" "refs/tags/${new_version}"
    log_success "Tag ${new_version} pushed"

    log_info "Next steps:"
    echo "  1. Follow the release workflow (gh run watch), which publishes the GitHub release and its notes"
    echo "  2. Check the release on the Terraform Registry"
    if [[ ! "$new_version" =~ $RC_RE ]]; then
        echo "  3. Add a '## Unreleased' section to CHANGELOG.md on develop for the next changes"
    fi
}

main "$@"
