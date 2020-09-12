#!/bin/bash

set -e

TAG="$1"

# Release tag takes precedence
if [ -n "$TRAVIS_TAG" ]; then
    TAG="${TRAVIS_TAG#v}"
fi

# If tag is still empty, attempt to use the branch name
if [ -z "$TAG" ]; then
    # Special case: "master" becomes "latest"
    if [ "$TRAVIS_BRANCH" = "master" ]; then
        TAG="latest"
    elif [ -n "$TRAVIS_BRANCH" ]; then
        TAG="${TRAVIS_BRANCH}"
    else
        echo "No tag set, aborting..."
        exit 2
    fi
fi

export TAG="${TAG}"

docker build -t "moapis/shop-server:${TAG}" -f server.Dockerfile .
docker build -t "moapis/shop-migrations:${TAG}" migrations

docker push "moapis/shop-server:${TAG}"
docker push "moapis/shop-migrations:${TAG}"