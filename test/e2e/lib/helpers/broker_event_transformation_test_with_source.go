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

package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	"github.com/google/knative-gcp/pkg/apis/events/v1beta1"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	v1 "k8s.io/api/core/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	eventingtestlib "knative.dev/eventing/test/lib"
	eventingtestresources "knative.dev/eventing/test/lib/resources"
	"knative.dev/pkg/test/helpers"

	// The following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/google/knative-gcp/test/e2e/lib"
	"github.com/google/knative-gcp/test/e2e/lib/resources"
)

/*
 BrokerEventTransformationTestHelper provides the helper methods which test the following scenario:

                              5                   4
                    ------------------   --------------------
                    |                 | |                    |
          1         v	      2       | v         3          |
(Sender or Source) --->   Broker ---> trigger -------> Knative Service(Receiver)
                    |
                    |    6                   7
                    |-------> respTrigger -------> Service(Target)

Note: the number denotes the sequence of the event that flows in this test case.
*/

func BrokerEventTransformationTestHelper(client *lib.Client, brokerURL url.URL, brokerName string) {
	client.T.Helper()
	senderName := helpers.AppendRandomString("sender")
	targetName := helpers.AppendRandomString("target")

	// Create a target Job to receive the events.
	makeTargetJobOrDie(client, targetName)

	// Create the Knative Service.
	kserviceName := CreateKService(client, "receiver")

	// Create a Trigger with the Knative Service subscriber.
	triggerFilter:=eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter, eventingv1alpha1.TriggerAnyFilter,
		map[string]interface{}{"type": lib.E2EDummyEventType})
	createTriggerWithKServiceSubscriber(client, brokerName, kserviceName, triggerFilter)

	// Create a Trigger with the target Service subscriber.
	respTriggerFilter:=eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter, eventingv1alpha1.TriggerAnyFilter,
		map[string]interface{}{"type": 	lib.E2EDummyRespEventType})
	createTriggerWithTargetServiceSubscriber(client, brokerName, targetName, respTriggerFilter)

	// Wait for ksvc, trigger ready.
	client.Core.WaitForResourceReadyOrFail(kserviceName, lib.KsvcTypeMeta)
	client.Core.WaitForResourcesReadyOrFail(eventingtestlib.TriggerTypeMeta)

	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create a sender Job to sender the event.
	senderJob := resources.SenderJob(senderName, []v1.EnvVar{{
		Name:  "BROKER_URL",
		Value: brokerURL.String(),
	}})
	client.CreateJobOrFail(senderJob)

	// Check if dummy CloudEvent is sent out.
	if done := jobDone(client, senderName); !done {
		client.T.Error("dummy event wasn't sent to broker")
		client.T.Failed()
	}
	// Check if resp CloudEvent hits the target Service.
	if done := jobDone(client, targetName); !done {
		client.T.Error("resp event didn't hit the target pod")
		client.T.Failed()
	}
}

func BrokerEventTransformationTestWithPubSubSourceHelper(client *lib.Client, authConfig lib.AuthConfig, brokerURL url.URL, brokerName string) {
	client.T.Helper()
	project := os.Getenv(lib.ProwProjectKey)
	topicName, deleteTopic := lib.MakeTopicOrDie(client.T)
	defer deleteTopic()

	psName := helpers.AppendRandomString(topicName + "-pubsub")
	targetName := helpers.AppendRandomString(topicName + "-target")
	data := fmt.Sprintf(`{"topic":%s}`, topicName)
	source := v1alpha1.CloudPubSubSourceEventSource(project, topicName)

	// Create a target PubSub Job to receive the events.
	lib.MakePubSubTargetJobOrDie(client, source, targetName, lib.E2EPubSubRespEventType)
	// Create the Knative Service.
	kserviceName := CreateKService(client, "pubsub_receiver")

	// Create a Trigger with the Knative Service subscriber.
	triggerFilter := eventingtestresources.WithAttributesTriggerFilter(
			eventingv1alpha1.TriggerAnyFilter,
			v1beta1.CloudPubSubSourcePublish,
			map[string]interface{}{})
	createTriggerWithKServiceSubscriber(client, brokerName, kserviceName, triggerFilter)

	// Create a Trigger with the target Service subscriber.
	respTriggerFilter := eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter,
		lib.E2EPubSubRespEventType,
		map[string]interface{}{})
	createTriggerWithTargetServiceSubscriber(client, brokerName, targetName, respTriggerFilter)

	// Wait for ksvc, trigger ready.
	client.Core.WaitForResourceReadyOrFail(kserviceName, lib.KsvcTypeMeta)
	client.Core.WaitForResourcesReadyOrFail(eventingtestlib.TriggerTypeMeta)

	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create the PubSub source.
	lib.MakePubSubOrDie(client,
		lib.BrokerGVK,
		psName,
		brokerName,
		topicName,
		authConfig.PubsubServiceAccount,
	)

	topic := lib.GetTopic(client.T, topicName)

	r := topic.Publish(context.TODO(), &pubsub.Message{
		Data: []byte(data),
	})

	_, err := r.Get(context.TODO())
	if err != nil {
		client.T.Logf("%s", err)
	}

	// Check if resp CloudEvent hits the target Service.
	if done := jobDone(client, targetName); !done {
		client.T.Error("resp event didn't hit the target pod")
		client.T.Failed()
	}
}

func BrokerEventTransformationTestWithStorageSourceHelper(client *lib.Client, authConfig lib.AuthConfig, brokerURL url.URL, brokerName string) {
	client.T.Helper()
	ctx := context.Background()
	project := os.Getenv(lib.ProwProjectKey)

	bucketName :=  lib.MakeBucket(ctx, client.T, project)
	storageName := helpers.AppendRandomString(bucketName + "-storage")
	targetName := helpers.AppendRandomString(bucketName + "-target")
	source := v1alpha1.CloudStorageSourceEventSource(bucketName)
	fileName := helpers.AppendRandomString("test-file-for-storage")
	// Create a target StorageJob to receive the events.
	lib.MakeStorageJobOrDie(client, source, fileName, targetName, lib.E2EStorageRespEventType)
	// Create the Knative Service.
	kserviceName := CreateKService(client, "storage_receiver")

	// Create a Trigger with the Knative Service subscriber.
	triggerFilter := eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter,
		v1beta1.CloudStorageSourceFinalize,
		map[string]interface{}{})
	createTriggerWithKServiceSubscriber(client, brokerName, kserviceName, triggerFilter)

	// Create a Trigger with the target Service subscriber.
	respTriggerFilter := eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter,
		lib.E2EStorageRespEventType,
		map[string]interface{}{})
	createTriggerWithTargetServiceSubscriber(client, brokerName, targetName, respTriggerFilter)

	// Wait for ksvc, trigger ready.
	client.Core.WaitForResourceReadyOrFail(kserviceName, lib.KsvcTypeMeta)
	client.Core.WaitForResourcesReadyOrFail(eventingtestlib.TriggerTypeMeta)

	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create the Storage source.
	lib.MakeStorageOrDie(
		client,
		lib.BrokerGVK,
		bucketName,
		storageName,
		brokerName,
		authConfig.PubsubServiceAccount,
	)

	// Add a random name file in the bucket
	lib.AddRandomFile(ctx, client.T, bucketName, fileName, project)

	// Check if resp CloudEvent hits the target Service.
	if done := jobDone(client, targetName); !done {
		client.T.Error("resp event didn't hit the target pod")
	}
}

func BrokerEventTransformationTestWithAuditLogsSourceHelper(client *lib.Client, authConfig lib.AuthConfig, brokerURL url.URL, brokerName string) {
	client.T.Helper()
	project := os.Getenv(lib.ProwProjectKey)

	auditlogsName := helpers.AppendRandomString("auditlogs-e2e-test")
	targetName := helpers.AppendRandomString(auditlogsName + "-target")
	topicName := helpers.AppendRandomString(auditlogsName + "-topic")
	resourceName := fmt.Sprintf("projects/%s/topics/%s", project, topicName)
	// Create a target Job to receive the events.
	lib.MakeAuditLogsJobOrDie(client, lib.PubSubCreateTopicMethodName, project, resourceName, lib.PubSubServiceName, targetName, lib.E2EAuditLogsRespType)
	// Create the Knative Service.
	kserviceName := CreateKService(client, "auditlogs_receiver")

	// Create a Trigger with the Knative Service subscriber.
	triggerFilter := eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter,
		v1beta1.CloudAuditLogsSourceEvent,
		map[string]interface{}{})
	createTriggerWithKServiceSubscriber(client, brokerName, kserviceName, triggerFilter)

	// Create a Trigger with the target Service subscriber.
	respTriggerFilter := eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter,
		lib.E2EAuditLogsRespType,
		map[string]interface{}{})
	createTriggerWithTargetServiceSubscriber(client, brokerName, targetName, respTriggerFilter)

	// Wait for ksvc, trigger ready.
	client.Core.WaitForResourceReadyOrFail(kserviceName, lib.KsvcTypeMeta)
	client.Core.WaitForResourcesReadyOrFail(eventingtestlib.TriggerTypeMeta)
	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create the CloudAuditLogsSource.
	lib.MakeAuditLogsOrDie(client,
		lib.BrokerGVK,
		auditlogsName,
		lib.PubSubCreateTopicMethodName,
		project,
		resourceName,
		lib.PubSubServiceName,
		brokerName,
		authConfig.PubsubServiceAccount,
	)

	client.Core.WaitForResourceReadyOrFail(auditlogsName, lib.CloudAuditLogsSourceTypeMeta)

	// Audit logs source misses the topic which gets created shortly after the source becomes ready. Need to wait for a few seconds.
	// Tried with 45 seconds but the test has been quite flaky.
	time.Sleep(90 * time.Second)
	topicName, deleteTopic := lib.MakeTopicWithNameOrDie(client.T, topicName)
	defer deleteTopic()

	// Check if resp CloudEvent hits the target Service.
	if done := jobDone(client, targetName); !done {
		client.T.Error("resp event didn't hit the target pod")
		client.T.Failed()
	}
}

func BrokerEventTransformationTestWithSchedulerSourceHelper(client *lib.Client, authConfig lib.AuthConfig, brokerURL url.URL, brokerName string) {
	client.T.Helper()
	data := helpers.AppendRandomString("scheduler-source-with-broker")
	schedulerName := helpers.AppendRandomString("scheduler-e2e-test")
	targetName := helpers.AppendRandomString(schedulerName + "-target")

	lib.MakeSchedulerJobOrDie(client, data, targetName, lib.E2ESchedulerRespType)
	// Create the Knative Service.
	kserviceName := CreateKService(client, "scheduler_receiver")

	// Create a Trigger with the Knative Service subscriber.
	triggerFilter := eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter,
		v1beta1.CloudSchedulerSourceExecute,
		map[string]interface{}{})
	createTriggerWithKServiceSubscriber(client, brokerName, kserviceName, triggerFilter)

	// Create a Trigger with the target Service subscriber.
	respTriggerFilter := eventingtestresources.WithAttributesTriggerFilter(
		eventingv1alpha1.TriggerAnyFilter,
		lib.E2ESchedulerRespType,
		map[string]interface{}{})
	createTriggerWithTargetServiceSubscriber(client, brokerName, targetName, respTriggerFilter)

	// Wait for ksvc, trigger ready.
	client.Core.WaitForResourceReadyOrFail(kserviceName, lib.KsvcTypeMeta)
	client.Core.WaitForResourcesReadyOrFail(eventingtestlib.TriggerTypeMeta)
	// Just to make sure all resources are ready.

	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create the CloudSchedulerSource.
	lib.MakeSchedulerOrDie(client, lib.BrokerGVK, schedulerName, data, brokerName, authConfig.PubsubServiceAccount,
	)

	// Check if resp CloudEvent hits the target Service.
	if done := jobDone(client, targetName); !done {
		client.T.Error("resp event didn't hit the target pod")
		client.T.Failed()
	}
}

func CreateKService(client *lib.Client, imageName string) string {
	client.T.Helper()
	kserviceName := helpers.AppendRandomString("kservice")
	// Create the Knative Service.
	kservice := resources.ReceiverKService(
		kserviceName, client.Namespace, imageName)
	client.CreateUnstructuredObjOrFail(kservice)
	return kserviceName

}

func createTriggerWithKServiceSubscriber(client *lib.Client,
	brokerName, kserviceName string,
	triggerFilter eventingtestresources.TriggerOption) {
	client.T.Helper()
	// Please refer to the graph in the file to check what dummy trigger is used for.
	triggerName := "trigger-broker-" + brokerName
	client.Core.CreateTriggerOrFail(
		triggerName,
		eventingtestresources.WithBroker(brokerName),
		triggerFilter,
		eventingtestresources.WithSubscriberServiceRefForTrigger(kserviceName),
	)
}

func createTriggerWithTargetServiceSubscriber(client *lib.Client,
	brokerName, targetName string,
	triggerFilter eventingtestresources.TriggerOption) {
	client.T.Helper()
	respTriggerName := "resp-broker-" + brokerName
	client.Core.CreateTriggerOrFail(
		respTriggerName,
		eventingtestresources.WithBroker(brokerName),
		triggerFilter,
		eventingtestresources.WithSubscriberServiceRefForTrigger(targetName),
	)
}

func makeTargetJobOrDie(client *lib.Client, targetName string) {
	client.T.Helper()
	job := resources.TargetJob(targetName, []v1.EnvVar{{
		Name:  "TIME",
		Value: "2m",
	}})
	client.CreateJobOrFail(job, lib.WithServiceForJob(targetName))
}

func jobDone(client *lib.Client, podName string) bool {
	client.T.Helper()
	msg, err := client.WaitUntilJobDone(client.Namespace, podName)
	if err != nil {
		client.T.Error(err)
		return false
	}
	if msg == "" {
		client.T.Error("No terminating message from the pod")
		return false
	}

	out := &lib.TargetOutput{}
	if err := json.Unmarshal([]byte(msg), out); err != nil {
		client.T.Error(err)
		return false
	}
	if !out.Success {
		if logs, err := client.LogsFor(client.Namespace, podName, lib.JobTypeMeta); err != nil {
			client.T.Error(err)
		} else {
			client.T.Logf("job: %s\n", logs)
		}
		return false
	}
	return true
}
