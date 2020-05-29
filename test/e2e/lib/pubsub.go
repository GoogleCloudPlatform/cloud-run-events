/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	v1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	kngcptesting "github.com/google/knative-gcp/pkg/reconciler/testing"
	"github.com/google/knative-gcp/test/e2e/lib/metrics"
	"github.com/google/knative-gcp/test/e2e/lib/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgmetrics "knative.dev/pkg/metrics"
)

func MakePubSubOrDie(client *Client,
	sinkGVK metav1.GroupVersionKind,
	psName, sinkName, topicName, pubsubServiceAccount string,
	so ...kngcptesting.CloudPubSubSourceOption,
) {
	client.T.Helper()
	eventsPubSub := makePubSub(client, sinkGVK, psName, sinkName, topicName, pubsubServiceAccount, so...)
	client.CreatePubSubOrFail(eventsPubSub)

	client.Core.WaitForResourceReadyOrFail(psName, CloudPubSubSourceTypeMeta)
}

func MakePubSubOrDieWithoutOwnerRef(client *Client,
	sinkGVK metav1.GroupVersionKind,
	psName, sinkName, topicName, pubsubServiceAccount string,
	so ...kngcptesting.CloudPubSubSourceOption,
) {
	client.T.Helper()
	eventsPubSub := makePubSub(client, sinkGVK, psName, sinkName, topicName, pubsubServiceAccount, so...)
	client.CreatePubSubOrFailWithoutOwnerRef(eventsPubSub)

	client.Core.WaitForResourceReadyOrFail(psName, CloudPubSubSourceTypeMeta)
}

func makePubSub(client *Client,
	sinkGVK metav1.GroupVersionKind,
	psName, sinkName, topicName, pubsubServiceAccount string,
	so ...kngcptesting.CloudPubSubSourceOption) *v1alpha1.CloudPubSubSource {
	client.T.Helper()
	so = append(so, kngcptesting.WithCloudPubSubSourceSink(sinkGVK, sinkName))
	so = append(so, kngcptesting.WithCloudPubSubSourceTopic(topicName))
	so = append(so, kngcptesting.WithCloudPubSubSourceGCPServiceAccount(pubsubServiceAccount))
	pubSub := kngcptesting.NewCloudPubSubSource(psName, client.Namespace, so...)
	return pubSub
}

func MakePubSubTargetJobOrDie(client *Client, source, targetName, eventType string) {
	client.T.Helper()
	job := resources.PubSubTargetJob(targetName, []v1.EnvVar{
		{
			Name:  "TYPE",
			Value: eventType,
		},
		{
			Name:  "SOURCE",
			Value: source,
		}, {
			Name:  "TIME",
			Value: "6m",
		}})
	client.CreateJobOrFail(job, WithServiceForJob(targetName))
}

func AssertMetrics(t *testing.T, client *Client, topicName, psName string) {
	t.Helper()
	sleepTime := 1 * time.Minute
	t.Logf("Sleeping %s to make sure metrics were pushed to stackdriver", sleepTime.String())
	time.Sleep(sleepTime)

	// If we reach this point, the projectID should have been set.
	projectID := os.Getenv(ProwProjectKey)
	f := map[string]interface{}{
		"metric.type":                 EventCountMetricType,
		"resource.type":               GlobalMetricResourceType,
		"metric.label.resource_group": PubsubResourceGroup,
		"metric.label.event_type":     v1alpha1.CloudPubSubSourcePublish,
		"metric.label.event_source":   v1alpha1.CloudPubSubSourceEventSource(projectID, topicName),
		"metric.label.namespace_name": client.Namespace,
		"metric.label.name":           psName,
		// We exit the target image before sending a response, thus check for 500.
		"metric.label.response_code":       http.StatusInternalServerError,
		"metric.label.response_code_class": pkgmetrics.ResponseCodeClass(http.StatusInternalServerError),
	}

	filter := metrics.StringifyStackDriverFilter(f)
	t.Logf("Filter expression: %s", filter)

	actualCount, err := client.StackDriverEventCountMetricFor(client.Namespace, projectID, filter)
	if err != nil {
		t.Errorf("failed to get stackdriver event count metric: %v", err)
		t.Fail()
	}
	expectedCount := int64(1)
	if *actualCount != expectedCount {
		t.Errorf("Actual count different than expected count, actual: %d, expected: %d", actualCount, expectedCount)
		t.Fail()
	}
}
