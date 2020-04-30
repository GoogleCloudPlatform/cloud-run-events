#!/usr/bin/env bash

# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Usage: ./init_control_plane.sh
# The current project set in gcloud MUST be the same as where the cluster is running.
source $(dirname $0)/lib.sh

readonly CONTROL_PLANE_SERVICE_ACCOUNT_KEY_TEMP="google-cloud-key.json"

init_control_plane_service_account ${PROJECT_ID} ${CONTROL_PLANE_SERVICE_ACCOUNT}

# Download a JSON key for the service account.
gcloud iam service-accounts keys create ${CONTROL_PLANE_SERVICE_ACCOUNT_KEY_TEMP} \
  --iam-account=${CONTROL_PLANE_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com

# Create/Patch the secret with the download JSON key in the control plane namespace
kubectl --namespace ${CONTROL_PLANE_NAMESPACE} create secret generic ${CONTROL_PLANE_SECRET_NAME} \
  --from-file=key.json=${CONTROL_PLANE_SERVICE_ACCOUNT_KEY_TEMP} --dry-run -o yaml | kubectl apply --filename -

# Delete the controller pod in the control plane namespace to refresh the created/patched secret
kubectl delete pod -n ${CONTROL_PLANE_NAMESPACE} --selector role=controller

# Remove the tmp file.
rm ${CONTROL_PLANE_SERVICE_ACCOUNT_KEY_TEMP}
