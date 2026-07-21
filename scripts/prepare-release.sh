#!/bin/bash

# Terraform Provider Wallix-Bastion Release Preparation Script
# This script prepares a new release by:
# 1. Determining the next version
# 2. Updating go.mod dependencies
# 3. Running tests and linters
# 4. Creating and pushing a new tag
# 5. Generating release notes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to get the latest tag
get_latest_tag() {
    git tag --sort=-version:refname | head -1
}

# Function to increment version
increment_version() {
    local version=$1
    local type=$2
    
    # Remove 'v' prefix if present
    version=${version#v}
    
    IFS='.' read -ra VERSION_PARTS <<< "$version"
    major=${VERSION_PARTS[0]}
    minor=${VERSION_PARTS[1]}
    patch=${VERSION_PARTS[2]}
    
    case $type in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch"|*)
            patch=$((patch + 1))
            ;;
    esac
    
    echo "v${major}.${minor}.${patch}"
}

# Function to check if working directory is clean
check_git_status() {
    if [[ -n $(git status --porcelain) ]]; then
        log_error "Working directory is not clean. Please commit or stash changes."
        exit 1
    fi
}

# Function to ensure we're on the develop branch
check_branch() {
    local current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "develop" ]]; then
        log_warning "Current branch is '$current_branch', not 'develop'"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Function to update dependencies
update_dependencies() {
    log_info "Updating Go dependencies..."
    
    # Update only direct dependencies to avoid breaking changes in indirect dependencies
    # such as golang.org/x/tools v0.38.0 which has breaking changes
    mapfile -t deps < <(go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all | grep -v '^$')
    if [ ${#deps[@]} -gt 0 ]; then
        go get -u "${deps[@]}"
    fi
    go mod tidy
    
    # Verify dependencies
    go mod verify
    
    log_success "Dependencies updated successfully"
}

# Function to run tests
run_tests() {
    log_info "Running tests and linters..."
    
    # Run golangci-lint
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
        log_success "Linting passed"
    else
        log_warning "golangci-lint not found, skipping linting"
    fi
    
    # Run unit tests
    go test ./...
    log_success "Tests passed"
    
    # Build to ensure compilation works
    make build
    log_success "Build successful"
}

# Function to update changelog version
update_changelog() {
    local new_version=$1
    local changelog_file="CHANGELOG.md"
    
    if [[ -f "$changelog_file" ]]; then
        # Get current date
        local current_date=$(date '+%B %d, %Y')
        
        # Check if the changelog already has the new version
        if grep -q "^## ${new_version#v}" "$changelog_file"; then
            log_info "Changelog already contains version ${new_version#v}"
        else
            log_info "Updating changelog with version ${new_version#v}..."
            # This would need manual intervention or more sophisticated parsing
            log_warning "Please manually update the changelog with the final release date"
        fi
    else
        log_warning "CHANGELOG.md not found"
    fi
}

# Function to create and push tag
create_tag() {
    local new_version=$1
    local tag_message="Release ${new_version}"
    
    log_info "Creating tag ${new_version}..."
    
    # Create annotated tag
    git tag -a "$new_version" -m "$tag_message"
    
    log_success "Tag ${new_version} created"
    
    # Ask if we should push the tag
    read -p "Push tag to remote? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git push origin "$new_version"
        log_success "Tag ${new_version} pushed to remote"
    else
        log_info "Tag created locally only. Push manually with: git push origin ${new_version}"
    fi
}

# Function to generate release notes
generate_release_notes() {
    local previous_tag=$1
    local new_version=$2
    
    log_info "Generating release notes..."
    
    echo "## Release Notes for ${new_version}"
    echo ""
    echo "### Changes since ${previous_tag}:"
    echo ""
    
    # Get commit messages since last tag
    git log "${previous_tag}..HEAD" --oneline --no-merges
    
    echo ""
    echo "### Files changed:"
    git diff --name-only "${previous_tag}..HEAD" | sort
}

# Main function
main() {
    local release_type="patch"
    local force_version=""
    local skip_tests=false
    local dry_run=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --major)
                release_type="major"
                shift
                ;;
            --minor)
                release_type="minor"
                shift
                ;;
            --patch)
                release_type="patch"
                shift
                ;;
            --version)
                force_version="$2"
                shift 2
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --major       Increment major version"
                echo "  --minor       Increment minor version"
                echo "  --patch       Increment patch version (default)"
                echo "  --version V   Use specific version V"
                echo "  --skip-tests  Skip running tests and linters"
                echo "  --dry-run     Show what would be done without making changes"
                echo "  --help, -h    Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    log_info "Starting release preparation..."
    
    # Check prerequisites
    if [[ "$dry_run" == false ]]; then
        check_git_status
        check_branch
    fi
    
    # Get current version
    local current_tag=$(get_latest_tag)
    log_info "Current version: $current_tag"
    
    # Determine next version
    local new_version
    if [[ -n "$force_version" ]]; then
        new_version="$force_version"
        if [[ ! "$new_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            log_error "Invalid version format: $new_version (expected: vX.Y.Z)"
            exit 1
        fi
    else
        new_version=$(increment_version "$current_tag" "$release_type")
    fi
    
    log_info "Next version: $new_version"
    
    if [[ "$dry_run" == true ]]; then
        log_info "DRY RUN MODE - Would perform the following actions:"
        echo "  1. Update dependencies"
        echo "  2. Run tests and linters"
        echo "  3. Update changelog (manual step)"
        echo "  4. Create tag: $new_version"
        echo "  5. Generate release notes"
        exit 0
    fi
    
    # Confirm before proceeding
    read -p "Proceed with release $new_version? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Release preparation cancelled"
        exit 0
    fi
    
    # Update dependencies
    update_dependencies
    
    # Run tests (unless skipped)
    if [[ "$skip_tests" == false ]]; then
        run_tests
    else
        log_warning "Skipping tests as requested"
    fi
    
    # Update changelog
    update_changelog "$new_version"
    
    # If there are changes from dependency updates, commit them
    if [[ -n $(git status --porcelain) ]]; then
        log_info "Committing dependency updates..."
        git add go.mod go.sum
        git commit -m "chore: update dependencies for release $new_version"
    fi
    
    # Create tag
    create_tag "$new_version"
    
    # Generate release notes
    generate_release_notes "$current_tag" "$new_version"
    
    log_success "Release preparation completed for $new_version"
    log_info "Next steps:"
    echo "  1. Review and finalize CHANGELOG.md"
    echo "  2. Create a GitHub release from the tag"
    echo "  3. Monitor the release workflow"
}

# Run main function with all arguments
main "$@"