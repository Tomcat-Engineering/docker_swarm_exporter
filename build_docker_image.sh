#!/bin/bash
set -ue

# Use the git version tag if one exists, or failing that the short hash of this commit
VERSION=$(git tag --list 'v*')
if [ -z "$VERSION" ]
then
	VERSION=$(git rev-parse --short HEAD)
fi

# Check to see whether we are on the master branch - if not, append the branch name to the version ID
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ $BRANCH != "master" ]]
then
	VERSION="${VERSION}-${BRANCH}"
fi

# Check to see whether there are un-committed git changes, in which case append "-dev" to the version ID.
# We ignore untracked files for now - not strictly correct but more convenient.
if output=$(git status --untracked-files=no --porcelain) && [ ! -z "$output" ]
then
	VERSION="${VERSION}-dev"
fi

echo "** Building version ${VERSION}"

# Build the binary
GOOS=linux GOARCH=amd64 go build

# Build the docker image
docker build -t "tomcatengineering/docker_swarm_exporter:${VERSION}" .

# Push to docker hub
docker push "tomcatengineering/docker_swarm_exporter:${VERSION}"
