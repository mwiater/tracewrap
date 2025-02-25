#!/bin/bash
set -e

# Disable Git pager for non-interactive output.
export GIT_PAGER=cat

# Check if a commit message argument is provided.
if [ -z "$1" ]; then
  echo "Error: Commit message is required."
  echo "Usage: $0 <commit_message>"
  exit 1
fi

COMMIT_MESSAGE="$1"

echo "Adding changes and committing to development branch..."
git add .
git commit -m "$COMMIT_MESSAGE"
git push origin development

echo "Checking out main and merging development..."
git checkout main
git merge development

echo "Determining new tag version..."
echo "Existing tags:"
git fetch --tags origin

# Get the latest tag that matches a strict semantic version format: v0.<minor>.0
LATEST=$(git tag --list "v0.[0-9]*.0" | sort -V | tail -n 1)
echo "DEBUG: Latest semantic tag: '$LATEST'"

if [ -z "$LATEST" ]; then
  NEW_TAG="v0.1.0"
  echo "No valid semantic version tag found. Setting new tag to $NEW_TAG"
else
  # Extract the minor version number from tag v0.X.0.
  MINOR=$(echo "$LATEST" | sed -E 's/^v0\.([0-9]+)\.0$/\1/')
  echo "DEBUG: Extracted minor version: $MINOR"
  NEW_MINOR=$((MINOR + 1))
  NEW_TAG="v0.${NEW_MINOR}.0"
  echo "New tag will be $NEW_TAG"
fi

echo "Tagging commit with $NEW_TAG"
git tag "$NEW_TAG"

echo "Pushing main branch and tags to origin..."
git push origin main
git push origin "$NEW_TAG"

echo "Switching back to development branch..."
git checkout development

echo "Release process completed. New tag: $NEW_TAG"

git ls-remote --tags origin | awk '/refs\/tags\/v0\.[0-9]*\.0$/ {sub(/^refs\/tags\//, ""); print}' | sort -V | tail -n 1