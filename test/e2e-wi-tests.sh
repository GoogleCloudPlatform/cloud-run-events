#!/usr/bin/env bash
#chmod +x e2e-wi-tests.sh

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

source $(dirname $0)/e2e-common.sh

# Override the setup and teardown functions to install wi-enabled control plane components.

readonly K8S_CONTROLLER_SERVICE_ACCOUNT="controller"
readonly AUTHENTICATED_SERVICE_ACCOUNT=$(gcloud config list account --format "value(core.account)")
readonly A=$(gcloud config get-value accoun)
echo ${AUTHENTICATED_SERVICE_ACCOUNT}
echo ${A}
echo ${E2E_PROJECT_ID}

# Create resources required for the Control Plane setup.
function knative_setup() {
  control_plane_setup || return 1
  start_knative_gcp "workloadIdentityEnabled" || return 1
}

function control_plane_setup() {
  # When not running on Prow we need to set up a service account for managing resources.
  if (( ! IS_PROW )); then
    echo "Set up ServiceAccount used by the Control Plane"
    gcloud iam service-accounts create ${CONTROL_PLANE_SERVICE_ACCOUNT}
    AUTHENTICATED_SERVICE_ACCOUNT=${CONTROL_PLANE_SERVICE_ACCOUNT}
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/pubsub.admin
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/pubsub.editor
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/storage.admin
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/cloudscheduler.admin
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/logging.configWriter
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/logging.privateLogViewer
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/cloudscheduler.admin
        # Give iam.serviceAccountAdmin role to the Google service account.
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/iam.serviceAccountAdmin
  fi
  # Allow the Kubernetes service account to use Google service account.
  MEMBER="serviceAccount:${E2E_PROJECT_ID}.svc.id.goog[${CONTROL_PLANE_NAMESPACE}/${K8S_CONTROLLER_SERVICE_ACCOUNT}]"
  echo ${MEMBER}
  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member $MEMBER ${AUTHENTICATED_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com
}

# Create resources required for Pub/Sub Admin setup.
function pubsub_setup() {
  # If the tests are run on Prow, clean up the topics and subscriptions before running them.
  # See https://github.com/google/knative-gcp/issues/494
  if (( IS_PROW )); then
    subs=$(gcloud pubsub subscriptions list --format="value(name)")
    while read -r sub_name
    do
      gcloud pubsub subscriptions delete "${sub_name}"
    done <<<"$subs"
    topics=$(gcloud pubsub topics list --format="value(name)")
    while read -r topic_name
    do
      gcloud pubsub topics delete "${topic_name}"
    done <<<"$topics"
  fi

  # When not running on Prow we need to set up a service account for PubSub.
  if (( ! IS_PROW )); then
    # Enable monitoring
    gcloud services enable monitoring
    echo "Set up ServiceAccount for Pub/Sub Admin"
    gcloud services enable pubsub.googleapis.com
    gcloud iam service-accounts create ${PUBSUB_SERVICE_ACCOUNT}
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${PUBSUB_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/pubsub.editor
    gcloud projects add-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${PUBSUB_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/monitoring.editor
    gcloud iam service-accounts keys create ${PUBSUB_SERVICE_ACCOUNT_KEY} \
      --iam-account=${PUBSUB_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com
    service_account_key="${PUBSUB_SERVICE_ACCOUNT_KEY}"
  fi
}

# Tear down resources required for Control Plane setup.
function control_plane_teardown() {
  # When not running on Prow we need to delete the service accounts and namespaces created.
  if (( ! IS_PROW )); then
    echo "Tear down ServiceAccount for Control Plane"
    gcloud iam service-accounts keys delete -q ${CONTROL_PLANE_SERVICE_ACCOUNT_KEY} \
      --iam-account=${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/pubsub.admin
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/pubsub.editor
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/storage.admin
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/cloudscheduler.admin
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/logging.configWriter
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/logging.privateLogViewer
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/cloudscheduler.admin
    # Remove iam.serviceAccountAdmin role to the Google service account.
    gcloud projects remove-iam-policy-binding ${E2E_PROJECT_ID} \
      --member=serviceAccount:${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
      --role roles/iam.serviceAccountAdmin
    gcloud iam service-accounts delete -q ${CONTROL_PLANE_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com
  fi
}

#create a cluster with Workload Identity enabled.
initialize $@ --cluster-creation-flag "--workload-pool=\${PROJECT}.svc.id.goog"

# Add annotation to Kubernetes service account.
kubectl annotate serviceaccount ${K8S_CONTROLLER_SERVICE_ACCOUNT} iam.gke.io/gcp-service-account=${AUTHENTICATED_SERVICE_ACCOUNT}@${E2E_PROJECT_ID}.iam.gserviceaccount.com \
  --namespace ${CONTROL_PLANE_NAMESPACE}

# Channel related e2e tests we have in Eventing is not running here.
go_test_e2e -timeout=20m -parallel=12 ./test/e2e -workloadIndentity=true -pubsubServiceAccount=${PUBSUB_SERVICE_ACCOUNT} || fail_test

success
