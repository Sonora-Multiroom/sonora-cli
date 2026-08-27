#!/bin/sh
# Tags a release and pushes the tag to trigger the release.yml workflow.
set -e

semver_re='^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'

git fetch origin --tags >/dev/null 2>&1 || true
current_tag=$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1)
current_tag=${current_tag#v}

if [ -n "$current_tag" ]; then
    major=$(echo "$current_tag" | cut -d. -f1)
    minor=$(echo "$current_tag" | cut -d. -f2)
    patch=$(echo "$current_tag" | cut -d. -f3)
    echo "Current version: v$current_tag"
    echo "Suggested next patch: v$major.$minor.$((patch + 1))"
    echo "Suggested next minor: v$major.$((minor + 1)).0"
else
    echo "Current version: none (no vX.Y.Z tags found)"
fi

printf 'Version to release (e.g. 0.1.0 or v0.1.0): '
read -r input

version=${input#v}

if ! printf '%s' "$version" | grep -Eq "$semver_re"; then
    echo "Error: '$input' is not a valid semantic version (expected X.Y.Z)." >&2
    exit 1
fi

tag="v$version"

if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
    echo "Error: tag $tag already exists." >&2
    exit 1
fi

echo "Fetching origin..."
git fetch origin main

current_branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$current_branch" = "main" ]; then
    if [ -n "$(git status --porcelain)" ]; then
        echo "Error: working tree has uncommitted changes. Commit or stash before releasing." >&2
        exit 1
    fi
    git merge --ff-only origin/main
else
    if ! git fetch origin main:main; then
        echo "Error: local main has diverged from origin/main and can't be fast-forwarded." >&2
        echo "Switch to main and resolve it manually, then re-run this script." >&2
        exit 1
    fi
fi

main_sha=$(git rev-parse main)
printf 'About to tag %s at main (%s) and push to origin. Continue? [y/N] ' "$tag" "$main_sha"
read -r confirm
case "$confirm" in
    y|Y|yes|YES) ;;
    *) echo "Aborted."; exit 1 ;;
esac

git tag -a "$tag" -m "Release $tag" main
git push origin "$tag"

echo "Pushed $tag. The release workflow will build and publish it."
