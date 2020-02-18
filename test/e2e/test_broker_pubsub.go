/*
Copyright 2019 Google LLC

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

package e2e

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	v1 "k8s.io/api/core/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	eventingtestlib "knative.dev/eventing/test/lib"
	"knative.dev/eventing/test/lib/duck"
	eventingtestresources "knative.dev/eventing/test/lib/resources"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/test/helpers"

	// The following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	kngcptesting "github.com/google/knative-gcp/pkg/reconciler/testing"
	"github.com/google/knative-gcp/test/e2e/lib"
	"github.com/google/knative-gcp/test/e2e/lib/resources"
)

/*
PubSubWithBrokerTestImpl tests the following scenario:

                              5                   4
                    ------------------   --------------------
                    |                 | |                    |
          1         v	      2       | v         3          |
(Sender) ---> Broker(PubSub) ---> dummyTrigger -------> Knative Service(Receiver)
                    |
                    |    6                   7
                    |-------> respTrigger -------> Service(Target)

Note: the number denotes the sequence of the event that flows in this test case.
*/

func BrokerWithPubSubChannelTestImpl(t *testing.T) {
	brokerName := helpers.AppendRandomString("pubsub")
	dummyTriggerName := "dummy-broker-" + brokerName
	respTriggerName := "resp-broker-" + brokerName
	kserviceName := helpers.AppendRandomString("kservice")
	senderName := helpers.AppendRandomString("sender")
	targetName := helpers.AppendRandomString("target")

	client := lib.Setup(t, true)
	defer lib.TearDown(client)

	u := createBrokerWithPubSubChannel(t, client, brokerName, dummyTriggerName, kserviceName, respTriggerName, targetName)

	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create a sender Job to sender the event.
	senderJob := resources.SenderJob(senderName, []v1.EnvVar{{
		Name:  "BROKER_URL",
		Value: u.String(),
	}})
	client.CreateJobOrFail(senderJob)

	// Check if dummy CloudEvent is sent out.
	if done := jobDone(client, senderName, t); !done {
		t.Error("dummy event wasn't sent to broker")
		t.Failed()
	}
	// Check if resp CloudEvent hits the target Service.
	if done := jobDone(client, targetName, t); !done {
		t.Error("resp event didn't hit the target pod")
		t.Failed()
	}
}

func PubSubSourceBrokerWithPubSubChannelTestImpl(t *testing.T) {
	topicName, deleteTopic := lib.MakeTopicOrDie(t)
	defer deleteTopic()

	brokerName := helpers.AppendRandomString("pubsub")
	dummyTriggerName := "dummy-broker-" + brokerName
	psName := helpers.AppendRandomString(topicName + "-pubsub")
	respTriggerName := "resp-broker-" + brokerName
	kserviceName := helpers.AppendRandomString("kservice")
	targetName := helpers.AppendRandomString(topicName + "-target")

	client := lib.Setup(t, true)
	defer lib.TearDown(client)

	u := createBrokerWithPubSubChannel(t, client, brokerName, dummyTriggerName, kserviceName, respTriggerName, targetName)
	var url apis.URL = apis.URL(u)
	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create the PubSub source.
	lib.MakePubSubOrDie(t, client, lib.ServiceGVK, psName, targetName, topicName, kngcptesting.WithCloudPubSubSourceSinkURI(&url))

	topic := lib.GetTopic(t, topicName)

	r := topic.Publish(context.TODO(), &pubsub.Message{
		Attributes: map[string]string{
			"target": "falldown",
		},
		Data: []byte(`{"foo":bar}`),
	})

	_, err := r.Get(context.TODO())
	if err != nil {
		t.Logf("%s", err)
	}

	// Check if resp CloudEvent hits the target Service.
	if done := jobDone(client, targetName, t); !done {
		t.Error("resp event didn't hit the target pod")
		t.Failed()
	}

	// TODO(nlopezgi): define if we want to assert StackDriver metrics here. CloudPubSubSourceWithTargetTestImpl can do so, but the test is currently disabled, so its unclear whether we want to replicate that functionality here.
}

func createBrokerWithPubSubChannel(t *testing.T, client *lib.Client, brokerName, dummyTriggerName, kserviceName, respTriggerName, targetName string) url.URL {
	// Create a new Broker.
	// TODO(chizhg): maybe we don't need to create these RBAC resources as they will now be automatically created?
	client.Core.CreateRBACResourcesForBrokers()
	client.Core.CreateBrokerOrFail(brokerName, eventingtestresources.WithChannelTemplateForBroker(lib.ChannelTypeMeta))

	// Create the Knative Service.
	kservice := resources.ReceiverKService(
		kserviceName, client.Namespace)
	client.CreateUnstructuredObjOrFail(kservice)

	// Create a Trigger with the Knative Service subscriber.
	client.Core.CreateTriggerOrFail(
		dummyTriggerName,
		eventingtestresources.WithBroker(brokerName),
		eventingtestresources.WithAttributesTriggerFilter(
			eventingv1alpha1.TriggerAnyFilter, eventingv1alpha1.TriggerAnyFilter,
			map[string]interface{}{"type": "e2e-testing-dummy"}),
		eventingtestresources.WithSubscriberKServiceRefForTrigger(kserviceName),
	)

	// Create a target Job to receive the events.
	job := resources.TargetJob(targetName, []v1.EnvVar{{
		Name:  "TARGET",
		Value: "falldown",
	}})
	client.CreateJobOrFail(job, lib.WithServiceForJob(targetName))

	// Create a Trigger with the target Service subscriber.
	client.Core.CreateTriggerOrFail(
		respTriggerName,
		eventingtestresources.WithBroker(brokerName),
		eventingtestresources.WithAttributesTriggerFilter(
			eventingv1alpha1.TriggerAnyFilter, eventingv1alpha1.TriggerAnyFilter,
			map[string]interface{}{"type": "e2e-testing-resp"}),
		eventingtestresources.WithSubscriberServiceRefForTrigger(targetName),
	)

	// Wait for broker, trigger, ksvc ready.
	client.Core.WaitForResourceReadyOrFail(brokerName, eventingtestlib.BrokerTypeMeta)
	client.Core.WaitForResourcesReadyOrFail(eventingtestlib.TriggerTypeMeta)
	client.Core.WaitForResourceReadyOrFail(kserviceName, lib.KsvcTypeMeta)

	// Get broker URL.
	metaAddressable := eventingtestresources.NewMetaResource(brokerName, client.Namespace, eventingtestlib.BrokerTypeMeta)
	u, err := duck.GetAddressableURI(client.Core.Dynamic, metaAddressable)
	if err != nil {
		t.Error(err.Error())
	}
	return u
}

func jobDone(client *lib.Client, podName string, t *testing.T) bool {
	msg, err := client.WaitUntilJobDone(client.Namespace, podName)
	if err != nil {
		t.Error(err)
		return false
	}
	if msg == "" {
		t.Error("No terminating message from the pod")
		return false
	} else {
		out := &lib.TargetOutput{}
		if err := json.Unmarshal([]byte(msg), out); err != nil {
			t.Error(err)
			return false
		}
		if !out.Success {
			if logs, err := client.LogsFor(client.Namespace, podName, lib.JobTypeMeta); err != nil {
				t.Error(err)
			} else {
				t.Logf("job: %s\n", logs)
			}
			return false
		}
	}
	return true
}
