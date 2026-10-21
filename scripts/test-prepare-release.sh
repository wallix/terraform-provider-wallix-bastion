#!/bin/bash
# The checks read "conditions && ok ... || ko ...": ok only echoes and never fails (SC2015), and sed
# strips the ANSI colors, which a parameter expansion can't match (SC2001).
# shellcheck disable=SC2015,SC2001
# Tests scripts/prepare-release.sh on a throwaway repository, with a local bare origin whose
# pre-receive hook refuses a tag that isn't on main. Needs git and bash only, no network.
# Usage: scripts/test-prepare-release.sh [path to prepare-release.sh]
set -uo pipefail

SCRIPT=$(realpath "${1:-$(dirname "$0")/prepare-release.sh}")
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
export GIT_AUTHOR_NAME=t GIT_AUTHOR_EMAIL=t@t GIT_COMMITTER_NAME=t GIT_COMMITTER_EMAIL=t@t
TODAY=$(LC_ALL=C date '+%B %d, %Y')
PASS=0 FAILN=0

ok() { PASS=$((PASS + 1)); echo "  ok   $1"; }
ko() { FAILN=$((FAILN + 1)); echo "  FAIL $1"; echo "$OUT" | sed 's/\x1b\[[0-9;]*m//g' | sed 's/^/       | /' | tail -12; }
# run <stdin answers> <args...>: runs the script in the work clone, sets OUT and RC.
run() {
    local answers=$1
    shift
    OUT=$(cd "$T/work" && printf '%b' "$answers" | bash ./scripts/prepare-release.sh "$@" 2>&1)
    RC=$?
    OUT=$(echo "$OUT" | sed 's/\x1b\[[0-9;]*m//g')
}
expect_rc() { [[ $RC == "$1" ]]; }
has() { grep -qF -- "$1" <<<"$OUT"; }
remote_tag() { git -C "$T/origin.git" rev-parse -q --verify "refs/tags/$1^{commit}" 2>/dev/null; }
remote_main() { git -C "$T/origin.git" rev-parse refs/heads/main; }
work() { git -C "$T/work" "$@"; }
changelog() { # changelog <top heading> <sections...>: rewrites the top of CHANGELOG.md on develop
    local heading=$1
    shift
    work checkout -q develop
    { echo "# changelog"; echo; echo "$heading"; echo; for s in "$@"; do echo "$s:"; echo; echo "- an entry"; echo; done; cat "$T/old_changelog"; } >"$T/work/CHANGELOG.md"
}
on_develop_commit() { work commit -qam "$1" && work push -q origin develop; }
merge_to_main() { work checkout -q main && work merge -q --no-edit origin/develop; }

# --- fixture: origin with main, develop and v0.15.2 --------------------------------------
git init -q --bare -b main "$T/origin.git"
cat >"$T/origin.git/hooks/pre-receive" <<'EOF'
#!/bin/bash
# Refuse a tag whose commit isn't on main at the time it is received.
while read -r old new ref; do
    if [[ $ref == refs/tags/* ]]; then
        commit=$(git rev-parse "$new^{commit}")
        if ! git merge-base --is-ancestor "$commit" refs/heads/main 2>/dev/null; then
            echo "hook: tag ${ref#refs/tags/} points to a commit that isn't on main" >&2
            exit 1
        fi
    fi
done
EOF
chmod +x "$T/origin.git/hooks/pre-receive"
git clone -q "$T/origin.git" "$T/work" 2>/dev/null
mkdir -p "$T/work/scripts"
cp "$SCRIPT" "$T/work/scripts/prepare-release.sh"
printf 'module example.com/p\n\ngo 1.26.0\n\ntoolchain go1.26.9\n' >"$T/work/go.mod"
printf '## 0.15.2 (October 01, 2026)\n\nBUG FIXES:\n\n- a fix\n' >"$T/old_changelog"
{ echo "# changelog"; echo; cat "$T/old_changelog"; } >"$T/work/CHANGELOG.md"
work add -A && work commit -qm init && work tag -a v0.15.2 -m "Release v0.15.2" && work push -q origin main && work push -q origin --tags
work checkout -q -b develop && work push -q -u origin develop 2>/dev/null
changelog "## Unreleased" FEATURES "BUG FIXES"
on_develop_commit "feat: something"

echo "== version and checks"
work checkout -q develop
run "" --dry-run
expect_rc 1 && has "Next version: v0.16.0" && has "calls for a minor release" && has "Current branch is 'develop'" && has "CHANGELOG.md isn't finalized" && has "## 0.16.0 (${TODAY})" \
    && ok "dry-run on develop: suggests v0.16.0, reports the branch and the unfinalized CHANGELOG" || ko "dry-run on develop"

work checkout -q main
run "" --dry-run
expect_rc 1 && has "doesn't contain origin/develop" && ok "main without develop merged is refused" || ko "main without develop"

changelog "## 0.16.0 (${TODAY})" FEATURES "BUG FIXES"
on_develop_commit "docs(changelog): 0.16.0"
merge_to_main
run "" --patch --skip-tests
expect_rc 1 && has "is a patch release, but CHANGELOG.md calls for a minor one" && [[ -z $(remote_tag v0.15.3) ]] \
    && ok "--patch on a CHANGELOG with FEATURES is refused, nothing pushed" || ko "--patch with FEATURES"

run "" --patch --skip-tests --ignore-changelog-type
expect_rc 1 && has "the top section is '## 0.16.0" && ok "--ignore-changelog-type still checks the CHANGELOG version" || ko "--ignore-changelog-type"

run "" --version v1.0 --skip-tests
expect_rc 1 && has "Invalid version format" && ok "--version v1.0 is refused" || ko "invalid --version"
run "" --version v1.0.0-rc1 --rc --skip-tests
expect_rc 1 && has "can't be combined" && ok "--version with --rc is refused" || ko "--version with --rc"
run "" --version v0.15.2 --skip-tests
expect_rc 1 && has "isn't after the latest release" && ok "--version not after the latest release is refused" || ko "version not after latest"

echo "== release v0.16.0"
before=$(remote_main)
run "n\n" --minor --skip-tests
expect_rc 0 && has "nothing was pushed" && [[ $(remote_main) == "$before" ]] && [[ -z $(remote_tag v0.16.0) ]] \
    && ok "declining the main push pushes nothing" || ko "decline main push"
run "y\nn\n" --minor --skip-tests
expect_rc 0 && has "no tag created" && [[ $(remote_main) == $(work rev-parse HEAD) ]] && [[ -z $(remote_tag v0.16.0) ]] && ! work rev-parse -q --verify refs/tags/v0.16.0 >/dev/null \
    && ok "declining the tag pushes main and creates no tag, locally or remotely" || ko "decline tag"
run "y\n" --minor --skip-tests
expect_rc 0 && [[ $(remote_tag v0.16.0) == $(work rev-parse HEAD) ]] && [[ $(git -C "$T/origin.git" cat-file -t v0.16.0) == tag ]] \
    && ok "v0.16.0 pushed as an annotated tag on HEAD, accepted by the tag-on-main hook" || ko "release v0.16.0"
run "" --minor --skip-tests --dry-run
expect_rc 1 && has "Next version: v0.17.0" && has "the top section is '## 0.16.0" && ok "a second v0.16.x needs a new CHANGELOG section" || ko "rerun after release"

echo "== git state checks"
echo dirty >>"$T/work/go.mod"
run "" --dry-run
expect_rc 1 && has "Working directory is not clean" && ok "dirty working directory is refused" || ko "dirty tree"
work checkout -q -- go.mod
git clone -q "$T/origin.git" "$T/other" 2>/dev/null
git -C "$T/other" commit -q --allow-empty -m "pushed by someone else" && git -C "$T/other" push -q origin main
run "" --dry-run
expect_rc 1 && has "doesn't contain origin/main" && ok "main behind origin/main is refused" || ko "main behind origin"
work pull -q --no-edit origin main

echo "== release candidates"
changelog "## Unreleased" "BREAKING CHANGES" FEATURES
on_develop_commit "feat!: breaking"
merge_to_main
run "" --dry-run
expect_rc 1 && has "calls for a minor release" && has "Next version: v0.17.0" && ok "BREAKING CHANGES on 0.x calls for a minor release" || ko "breaking on 0.x"
run "y\ny\n" --major --rc --skip-tests
expect_rc 0 && has "releasing v1.0.0-rc1 from '## Unreleased'" && [[ -n $(remote_tag v1.0.0-rc1) ]] \
    && ok "--major --rc: v1.0.0-rc1 from '## Unreleased', with a warning" || ko "rc1"
run "y\n" --major --rc --skip-tests
expect_rc 0 && [[ -n $(remote_tag v1.0.0-rc2) ]] && ok "--major --rc again: v1.0.0-rc2" || ko "rc2"
run "" --major --skip-tests
expect_rc 1 && has "CHANGELOG.md isn't finalized" && has "## 1.0.0 (${TODAY})" && ok "--major refuses v1.0.0 with '## Unreleased'" || ko "final with Unreleased"
changelog "## 1.0.0 (${TODAY})" "BREAKING CHANGES" FEATURES
on_develop_commit "docs(changelog): 1.0.0"
merge_to_main
run "y\ny\n" --major --skip-tests
expect_rc 0 && has "Next version: v1.0.0" && [[ -n $(remote_tag v1.0.0) ]] && ok "--major after the rcs: v1.0.0 (rc tags ignored for the bump)" || ko "final 1.0.0"

echo "== after 1.0.0"
changelog "## Unreleased" "BREAKING CHANGES"
on_develop_commit "feat!: another"
merge_to_main
run "" --minor --skip-tests
expect_rc 1 && has "is a minor release, but CHANGELOG.md calls for a major one" && ok "BREAKING CHANGES on 1.x calls for a major release" || ko "breaking on 1.x"
changelog "## 1.0.1 (October 01, 2026)" "BUG FIXES"
on_develop_commit "docs(changelog): 1.0.1 with an old date"
merge_to_main
run "" --dry-run
expect_rc 0 && has "Next version: v1.0.1" && has "the release date is 'October 01, 2026', not today" && ok "BUG FIXES only: patch, and an old release date is a warning" || ko "patch with old date"

echo
echo "passed=$PASS failed=$FAILN"
((FAILN == 0))
