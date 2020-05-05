# E2E Tests

Prow will run `./e2e-tests.sh` with authentication mechanism using Kubernetes Secrets and `./e2e-wi-tests.sh` with authentication mechanism using Workload Identity.

## Adding E2E Tests

E2E tests are tagged with `// +build e2e` but tagging a Go file this way will
prevent the compiler from compiling the test code. To work around this, for the
test code we separate them into different files:

```shell
test/e2e
├── e2e_test.go
└── test_xxx.go
```

- `e2e_test.go` is the testing file entry point (tagged with e2e).
- `test_xxx.go` are the test implementations (not tagged with e2e).

We leverage the
[test library in Eventing](https://github.com/knative/eventing/tree/master/test/lib)
as much as possible for implementing the e2e tests. Logic specific to
knative-gcp should be added under [knative-gcp e2e test lib](lib).

## Running E2E Tests on an existing cluster

### Prerequisites
There are two ways to set up authentication mechanism. 
- If you want to run E2E tests with authentication mechanism using **Workload Identity**, please configure the authentication mechanism with **Workload Identity**.
- If you want to run E2E tests with authentication mechanism using **Kubernetes Secrets**, please configure the authentication mechanism with **Kubernetes Secrets**.
1. A running Kubernetes cluster with [knative-gcp](../../docs/install/install-knative-gcp.md) installed and configured.
1. [Pub/Sub Enabled Service Account](../../docs/install/pubsub-service-account.md) installed.
1. [GCP Broker Deployment](../../docs/install/install-gcp-broker.md#deployment) and [GCP Broker Authentication Setup](../../docs/install/install-gcp-broker.md#authentication-setup-for-gcp-broker).
1. [Broker with Pub/Sub Channel](../../docs/install/install-broker-with-pubsub-channel.md) installed.
1. [CloudStorageSource Prerequisites](../../docs/examples/cloudstoragesource/README.md#prerequisites).
   Note that you only need to: 
   1. Create with an App Engine application in your project.
1. [CloudSchedulerSource Prerequisites](../../docs/examples/cloudschedulersource/README.md#prerequisites).
   Note that you only need to:
   1. Enable the Cloud Storage API on your project.
   1. Give Google Cloud Storage permissions to publish to GCP Pub/Sub.
1. A docker repo containing [the test images](#test-images). Remember to specify the build tag `e2e`.
1. (Optional) Note that if you plan on running metrics-related E2E tests using the StackDriver
backend, you need to give your
[Service Account](../../docs/install/pubsub-service-account.md) the
`Monitoring Editor` role on your Google Cloud project:

   ```shell
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member=serviceAccount:cloudrunevents-pullsub@$PROJECT_ID.iam.gserviceaccount.com \
     --role roles/monitoring.editor
   ```


### Running E2E tests
### Running E2E tests with authentication mechanism using Kubernetes Secrets
```shell
go test --tags=e2e ./test/e2e/...
```

And count is supported too:

```shell
go test --tags=e2e ./test/e2e/... --count=3
```

If you want to run a specific test:

```shell
E2E_PROJECT_ID=<project name> \
  go test --tags=e2e ./test/e2e/... -run NameOfTest
```

For example, to run TestPullSubscription:

```shell \
E2E_PROJECT_ID=<project name> \
  go test --tags=e2e ./test/e2e/... -run TestPullSubscription
```

### Running E2E tests with authentication mechanism using Workload Identity
```shell
go test --tags=e2e 
  -workloadIndentity=true \
  -pubsubServiceAccount=cre-pubsub@$PROJECT_ID.iam.gserviceaccount.com \
  ./test/e2e/...
```

And count is supported too:

```shell
go test --tags=e2e \
  -workloadIndentity=true \
  -pubsubServiceAccount=cre-pubsub@$PROJECT_ID.iam.gserviceaccount.com \
  ./test/e2e/... --count=3
```

If you want to run a specific test:

```shell
E2E_PROJECT_ID=<project name> go test --tags=e2e \
  -workloadIndentity=true \
  -pubsubServiceAccount=cre-pubsub@$PROJECT_ID.iam.gserviceaccount.com \
  ./test/e2e/... -run NameOfTest
```

For example, to run TestPullSubscription:

```shell \
E2E_PROJECT_ID=<project name> go test --tags=e2e \
  -workloadIndentity=true \
  -pubsubServiceAccount=cre-pubsub@$PROJECT_ID.iam.gserviceaccount.com \
  ./test/e2e/... -run TestPullSubscription
```
## Running E2E Tests on an new cluster

### Prerequisites
1. Enable necessary APIs.

```shell
gcloud services enable compute.googleapis.com
gcloud services enable container.googleapis.com 
```

2. Set the projcet you want to run E2E tests to be the default one with:

```shell
export PROJECT=<REPLACE_ME>
gcloud config set core/project $PROJECT
```

### Running E2E tests

If you want to run E2E tests with authentication mechanism using **Kubernetes Secrets**:
```shell
./test/e2e-tests.sh
```
If you want to run E2E tests with authentication mechanism using **Workload Identity**:
```shell
./e2e-wi-tests.sh
```

## Test images

### Building the test images

_Note: this is only required when you run e2e tests locally with `go test`
commands. Running tests through e2e-tests.sh will publish the images
automatically._

The [`upload-test-images.sh`](./../upload-test-images.sh) script can be used to
build and push the test images used by the e2e tests. It requires:

- [`KO_DOCKER_REPO`](https://github.com/knative/serving/blob/master/DEVELOPMENT.md#environment-setup)
  to be set
- You need to be
  [authenticated with your `KO_DOCKER_REPO`](https://github.com/knative/serving/blob/master/DEVELOPMENT.md#environment-setup)
- [`docker`](https://docs.docker.com/install/) to be installed

For images deployed in GCR, a docker tag is mandatory to avoid issues with using
`latest` tag:

```bash
./test/upload-test-images.sh ./test/test_images e2e
./test/upload-test-images.sh ./vendor/knative.dev/eventing/test/test_images/ e2e
```

To run the script for all end to end test images:

```bash
./test/upload-test-images.sh ./test/test_images
./test/upload-test-images.sh ./vendor/knative.dev/eventing/test/test_images/
```

### Adding new test images

New test images should be placed in `./test/test_images`. For each image create
a new sub-folder and include a Go file that will be an entry point to the
application. This Go file should use the package `main` and include the function
`main()`. It is a good practice to include a `README` file as well. When
uploading test images, `ko` will build an image from this folder and upload to
the Docker repository configured as
[`KO_DOCKER_REPO`](https://github.com/knative/serving/blob/master/DEVELOPMENT.md#environment-setup).
