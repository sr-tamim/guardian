#!/bin/bash
# Version management script for Guardian

CURRENT_VERSION="0.0.1"
VERSION_FILE="pkg/version/version.go"

function show_help() {
    echo "Guardian Version Management"
    echo "========================="
    echo ""
    echo "Usage: $0 [command] [version]"
    echo ""
    echo "Commands:"
    echo "  current              Show current version"
    echo "  bump <version>       Update version (e.g., 0.0.2)"
    echo "  build               Build with current version"
    echo "  release <version>   Bump version and build"
    echo ""
    echo "Examples:"
    echo "  $0 current"
    echo "  $0 bump 0.0.2"
    echo "  $0 release 0.1.0"
}

function get_current_version() {
    grep -o 'Version.*=.*"[^"]*"' $VERSION_FILE | grep -o '"[^"]*"' | head -1 | tr -d '"'
}

function update_version() {
    local new_version=$1
    echo "📦 Updating version from $(get_current_version) to $new_version"
    
    # Update version.go
    sed -i.bak "s/Version.*=.*/Version   = \"$new_version\"/" $VERSION_FILE
    
    # Update Makefile  
    sed -i.bak "s/VERSION=.*/VERSION=$new_version/" Makefile
    
    # Update build script
    sed -i.bak "s/VERSION=\".*\"/VERSION=\"$new_version\"/" scripts/build-server.sh
    
    echo "✅ Version updated to $new_version"
}

function build_version() {
    echo "🏗️  Building Guardian v$(get_current_version)..."
    ./scripts/build-server.sh
}

function release_version() {
    local new_version=$1
    echo "🚀 Creating release $new_version"
    
    update_version $new_version
    build_version
    
    echo ""
    echo "📦 Release $new_version ready!"
    echo "   Binary: bin/guardian.exe"
    echo "   Package: deploy/"
    echo ""
    echo "Next steps:"
    echo "1. Test the binary: ./bin/guardian.exe version"
    echo "2. Update CHANGELOG.md"
    echo "3. Commit changes: git add . && git commit -m 'chore: bump version to $new_version'"
    echo "4. Create tag: git tag v$new_version"
}

# Main execution
case "$1" in
    "current")
        echo "Current version: $(get_current_version)"
        ;;
    "bump")
        if [ -z "$2" ]; then
            echo "Error: Version number required"
            echo "Usage: $0 bump <version>"
            exit 1
        fi
        update_version $2
        ;;
    "build")
        build_version
        ;;
    "release")
        if [ -z "$2" ]; then
            echo "Error: Version number required"
            echo "Usage: $0 release <version>"
            exit 1
        fi
        release_version $2
        ;;
    *)
        show_help
        ;;
esac
