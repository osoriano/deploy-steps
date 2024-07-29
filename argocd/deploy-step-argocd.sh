#!/bin/bash
#
# Update a Kubernetes resource by pushing to a git repo.
# ArgoCD is expected to synchronize with the git repo.
#
# Then, this script will wait for the deployment to complete
#
# Expects the repo to already be cloned in /repo
#
set -o errexit
set -o nounset
set -o pipefail


cd "$(dirname "$0")"
source ./generate-github-installation-access-token.sh


#
# Parameters
#

# Url of the git repo to clone
REPO_URL="$1"

# Branch of the git repo to clone
REPO_BRANCH="$2"

# GitHub App ID
APP_ID="$3"

# GitHub App User ID
APP_USER_ID="$4"

# GitHub App User Name
APP_USER_NAME="$5"

# Path to the private key of the GitHub App
KEY_PATH="$6"

# GitHub Installation ID
INSTALLATION_ID="$7"

# Path to the resource inside the repo. If a directory, runs all on files in directory
RESOURCE_PATH="$8"

# New deployment image to substute in the specified resources
NEW_IMAGE="$9"

echo "Deploying with parameters:"
echo "  REPO_URL=${REPO_URL}"
echo "  REPO_BRANCH=${REPO_BRANCH}"
echo "  APP_ID=${APP_ID}"
echo "  APP_USER_ID=${APP_USER_ID}"
echo "  APP_USER_NAME=${APP_USER_NAME}"
echo "  KEY_PATH=${KEY_PATH}"
echo "  INSTALLATION_ID=${INSTALLATION_ID}"
echo "  RESOURCE_PATH=${RESOURCE_PATH}"
echo "  NEW_IMAGE=${NEW_IMAGE}"

# Configure git
git config user.name "${APP_USER_NAME}"
git config user.email "${APP_USER_ID}+${APP_USER_NAME}@users.noreply.github.com"
GH_ACCESS_TOKEN="$(generate-installation-access-token "${APP_ID}" "${KEY_PATH}" "${INSTALLATION_ID}")"
git config user.password "${GH_ACCESS_TOKEN}"

# Clone the repo
git clone --depth 1 --branch "${REPO_BRANCH}" --single-branch "${REPO_URL}" /repo
cd /repo

# Perform the subtitution
IMAGE_WITHOUT_TAG="${NEW_IMAGE%%:*}"
if [[ -d "${RESOURCE_PATH}" ]]; then
  sed "s|${IMAGE_WITHOUT_TAG}:.*|${NEW_IMAGE}|g" -i "${RESOURCE_PATH}"/*
elif [[ -f "${RESOURCE_PATH}" ]]; then
  sed "s|${IMAGE_WITHOUT_TAG}:.*|${NEW_IMAGE}|g" -i "${RESOURCE_PATH}"
else
  echo "Resource path does not exist: ${RESOURCE_PATH}"
fi

# Commit and push to git
NEW_TAG="${NEW_IMAGE##*:}"
git commit -am "Bump ${RESOURCE_PATH} to \`${NEW_TAG:0:8}\`

Bump resource ${RESOURCE_PATH} to version:
${NEW_TAG}"

git push

# Wait for the resource to be available
# todo
