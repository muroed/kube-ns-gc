#!/bin/bash

# Release script for kube-ns-gc
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh 1.0.0

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format (semantic versioning)
if [[ ! $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.0.0)"
    exit 1
fi

echo "🚀 Creating release for version $VERSION"

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Error: You must be on the main branch to create a release"
    echo "Current branch: $CURRENT_BRANCH"
    exit 1
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory is not clean. Please commit or stash changes first."
    exit 1
fi

# Update Chart.yaml version
echo "📝 Updating Chart.yaml version to $VERSION"

# Detect OS and use appropriate sed command
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/^version: .*/version: $VERSION/" deploy/kube-ns-gc/Chart.yaml
    sed -i '' "s/^appVersion: .*/appVersion: \"$VERSION\"/" deploy/kube-ns-gc/Chart.yaml
else
    # Linux and other systems
    sed -i "s/^version: .*/version: $VERSION/" deploy/kube-ns-gc/Chart.yaml
    sed -i "s/^appVersion: .*/appVersion: \"$VERSION\"/" deploy/kube-ns-gc/Chart.yaml
fi

# Commit version update
echo "💾 Committing version update"
git add deploy/kube-ns-gc/Chart.yaml
git commit -m "Bump version to $VERSION"

# Create and push tag
echo "🏷️  Creating tag v$VERSION"
git tag -a "v$VERSION" -m "Release v$VERSION"

echo "📤 Pushing changes and tag"
git push origin main
git push origin "v$VERSION"

echo "✅ Release v$VERSION created successfully!"
echo ""
echo "📋 What happens next:"
echo "1. GitHub Actions will build and push Docker image with tag $VERSION"
echo "2. Helm chart will be packaged and released"
echo "3. Security scan will run on the new image"
echo ""
echo "🔍 You can monitor the progress at:"
echo "https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"
