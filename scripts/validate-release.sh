#!/usr/bin/env bash

# Validates the published assets of a release: the SHA256SUMS signature by the release key, which
# must be one of the signing keys the Terraform Registry holds for the namespace, then the checksum
# of every file it lists and the provider binary in each archive.
#
# Usage: scripts/validate-release.sh vX.Y.Z [dir]
#   dir: keeps the downloads there and reuses the files already present (default: a temp dir).
# Needs curl, jq, gpg, gpgv, GNU sha256sum and unzip, all present on the GitHub-hosted Ubuntu runners.

set -euo pipefail

TAG="${1:?usage: $0 vX.Y.Z [dir]}"
VERSION="${TAG#v}"
NAMESPACE="wallix"
NAME="terraform-provider-wallix-bastion"
BASE_URL="https://github.com/wallix/${NAME}/releases/download/${TAG}"
KEYS_URL="https://registry.terraform.io/v2/gpg-keys?filter%5Bnamespace%5D=${NAMESPACE}"
# Primary key fingerprint of the release key: only a signature by this key passes.
RELEASE_KEY_FPR="2A456F10F190998B8C1149C9A5B491EF453DFE8C"
SUMS="${NAME}_${VERSION}_SHA256SUMS"
REQUIRED_PLATFORMS="darwin_amd64 darwin_arm64 linux_amd64 linux_arm64 windows_amd64"

fail() {
    echo "ERROR: $*" >&2
    exit 1
}

for cmd in curl jq gpg gpgv sha256sum unzip; do
    command -v "$cmd" >/dev/null || fail "${cmd} is required"
done

# -f fails on an HTTP error instead of saving the error page as the asset.
fetch() {
    [ -s "$2" ] && return 0
    curl -fsSL --proto '=https' --retry 5 --retry-all-errors --retry-delay 10 -o "$2.part" "$1" ||
        fail "cannot download $1"
    mv "$2.part" "$2"
}

TMP_DIR="$(mktemp -d)"
trap 'gpgconf --homedir "$TMP_DIR/gnupg" --kill all >/dev/null 2>&1 || true; rm -rf "$TMP_DIR"' EXIT
WORK_DIR="${2:-$TMP_DIR/release}"
mkdir -p "$WORK_DIR" "$TMP_DIR/gnupg"
chmod 700 "$TMP_DIR/gnupg"
cd "$WORK_DIR"

echo "== ${TAG}: signature of ${SUMS}"
fetch "${BASE_URL}/${SUMS}" "$SUMS"
fetch "${BASE_URL}/${SUMS}.sig" "${SUMS}.sig"
fetch "$KEYS_URL" "$TMP_DIR/keys.json"
jq -r '.data[].attributes["ascii-armor"]' "$TMP_DIR/keys.json" >"$TMP_DIR/keys.asc"
[ -s "$TMP_DIR/keys.asc" ] || fail "the Terraform Registry has no signing key for the ${NAMESPACE} namespace"
gpg --homedir "$TMP_DIR/gnupg" --batch --quiet --import "$TMP_DIR/keys.asc"
gpg --homedir "$TMP_DIR/gnupg" --batch --export >"$TMP_DIR/keyring.gpg"
gpgv --homedir "$TMP_DIR/gnupg" --keyring "$TMP_DIR/keyring.gpg" --status-fd 3 "${SUMS}.sig" "$SUMS" \
    3>"$TMP_DIR/gpgv.status" ||
    fail "${SUMS}.sig is not a valid signature by a key the Terraform Registry holds for ${NAMESPACE}"
# The last field of VALIDSIG is the primary key fingerprint, also when a subkey signed.
awk -v fpr="$RELEASE_KEY_FPR" '$2 == "VALIDSIG" && $NF == fpr { ok = 1 } END { exit !ok }' \
    "$TMP_DIR/gpgv.status" || fail "${SUMS}.sig is not signed by the release key ${RELEASE_KEY_FPR}"

echo "== ${TAG}: files listed in ${SUMS}"
# File names end up in URLs and paths: only accept this release's own files.
line_re="^[0-9a-f]{64}  ${NAME}_${VERSION//./\\.}_[A-Za-z0-9_.]+$"
if grep -Evq "$line_re" "$SUMS"; then
    fail "${SUMS} has a line other than '<sha256>  ${NAME}_${VERSION}_<file>'"
fi
for platform in $REQUIRED_PLATFORMS; do
    grep -qF "  ${NAME}_${VERSION}_${platform}.zip" "$SUMS" || fail "${SUMS} lists no ${platform} archive"
done
while read -r _ file || [ -n "$file" ]; do
    fetch "${BASE_URL}/${file}" "$file"
done <"$SUMS"
sha256sum --strict -c "$SUMS" || fail "checksum mismatch"

while read -r _ file || [ -n "$file" ]; do
    case "$file" in
    *_windows_*.zip) binary="${NAME}_v${VERSION}.exe" ;;
    *.zip) binary="${NAME}_v${VERSION}" ;;
    *) continue ;;
    esac
    unzip -tq "$file" >/dev/null || fail "${file} is not a valid zip archive"
    entries="$(unzip -Z1 "$file")"
    grep -qxF "$binary" <<<"$entries" || fail "${file} does not contain ${binary}"
done <"$SUMS"

echo "OK: ${TAG}: ${SUMS} signed by the release key, $(wc -l <"$SUMS") files verified"
