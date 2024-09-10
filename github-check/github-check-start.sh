#!/bin/bash
#
# Updates GitHub Check Runs
#
set -o errexit
set -o nounset
set -o pipefail


cd "$(dirname "$0")"
source ./generate-github-installation-access-token.sh


#
# Parameters
#

# GitHub App ID
APP_ID="$1"

# Path to the private key of the GitHub App
KEY_PATH="$2"

# GitHub Installation ID
INSTALLATION_ID="$3"

# The "org/name" portion of the repo url. e.g. "osoriano/deploy-steps"
REPO_SHORT="$4"

# Url to the Workflow UI
DETAILS_URL="$5"

# Event type. Used in the GitHub check description
EVENT_TYPE="$6"

# The head commit sha
HEAD_SHA="$7"

# The output file to contain the id of the started GitHub Check
OUTPUT_FILE="$8"

echo "Deploying with parameters:"
echo "  APP_ID=${APP_ID}"
echo "  KEY_PATH=${KEY_PATH}"
echo "  INSTALLATION_ID=${INSTALLATION_ID}"
echo "  REPO_SHORT=${REPO_SHORT}"
echo "  DETAILS_URL=${DETAILS_URL}"
echo "  EVENT_TYPE=${EVENT_TYPE}"
echo "  HEAD_SHA=${HEAD_SHA}"
echo "  OUTPUT_FILE=${OUTPUT_FILE}"

# Fetch GitHub Access Token
echo "Fetching GitHub Access Token"
GH_ACCESS_TOKEN="$(generate-installation-access-token "${APP_ID}" "${KEY_PATH}" "${INSTALLATION_ID}")"

echo "Starting GitHub Status Check"
STATUS_CHECK_DATA='{
  "name": "Argo Workflow '"${EVENT_TYPE}"' Build",
  "head_sha": "'"${HEAD_SHA}"'",
  "details_url": "'"${DETAILS_URL}"'",
  "status": "in_progress",
  "output":{
    "title": "Argo Workflow '"${EVENT_TYPE}"' Build for CI/CD",
    "summary": "Runs the '"${EVENT_TYPE}"' build in an Argo workflow. See the [workflow link]('"${DETAILS_URL}"') for more details"
  }
}'

curl \
  --silent \
  --show-error \
  --fail \
  --location \
  --request POST \
  --header "Accept: application/vnd.github+json" \
  --header "Authorization: Bearer $GH_ACCESS_TOKEN" \
  --header "X-GitHub-Api-Version: 2022-11-28" \
  --data-binary "${STATUS_CHECK_DATA}" \
  "https://api.github.com/repos/${REPO_SHORT}/check-runs" \
  | jq '.id' \
  > "${OUTPUT_FILE}"
