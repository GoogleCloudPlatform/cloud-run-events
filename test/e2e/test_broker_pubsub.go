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
	"encoding/json"
	"testing"
	"time"

	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/eventing/test/base"
	eventingtestresources "knative.dev/eventing/test/base/resources"
	eventingCommon "knative.dev/eventing/test/common"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	pkgTest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"

	// The following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/google/knative-gcp/test/e2e/lib"
	"github.com/google/knative-gcp/test/e2e/lib/resources"
)

/*
PubSubWithBrokerTestImpl tests the following scenario:

                          5                 4
                    ------------------   --------------------
                    |                 | |                    |
              1     v	    2         | v        3           |
(Sender) ---> Broker(PubSub) ---> dummyTrigger -------> Knative Service(Receiver)
                    |
                    |    6                   7
                    |-------> respTrigger -------> Service(Target)

Note: the number denotes the sequence of the event that flows in this test case.
*/

func BrokerWithPubSubChannelTestImpl(t *testing.T, packages map[string]string) {
	brokerName := helpers.AppendRandomString("pubsub")
	dummyTriggerName := "dummy-broker-" + brokerName
	respTriggerName := "resp-broker-" + brokerName
	kserviceName := helpers.AppendRandomString("kservice")
	senderName := helpers.AppendRandomString("sender")
	targetName := helpers.AppendRandomString("target")

	client := lib.Setup(t, true)
	defer lib.TearDown(client)

	// Create a new broker.
	// TODO(chizhg): maybe we don't need to create these RBAC resources as they will now be automatically created?
	client.Core.CreateRBACResourcesForBrokers()
	client.Core.CreateBrokerOrFail(brokerName, lib.ChannelTypeMeta)

	client.Core.CreateTriggerOrFail(
		dummyTriggerName,
		eventingtestresources.WithBroker(brokerName),
		eventingtestresources.WithAttributesTriggerFilter(
			"", "",
			map[string]interface{}{"type": "e2e-testing-dummy"}),
		withSubscriberKServiceRefForTrigger(kserviceName),
	)

	// Create a target Job to receive the events.
	job := resources.TargetJob(targetName)
	client.CreateJobOrFail(job, lib.WithService(targetName))

	client.Core.CreateTriggerOrFail(
		respTriggerName,
		eventingtestresources.WithBroker(brokerName),
		eventingtestresources.WithAttributesTriggerFilter(
			"", "",
			map[string]interface{}{"type": "e2e-testing-resp"}),
		eventingtestresources.WithSubscriberRefForTrigger(targetName),
	)

	config := map[string]string{
		"namespace":        client.Namespace,
		"kserviceName":     kserviceName,
	}
	for k, v := range packages {
		config[k] = v
	}

	// Create resources.
	brokerInstaller := createResource(client, config, []string{"pubsub_broker", "istio"}, t)
	defer deleteResource(brokerInstaller, t)

	// Wait for broker, trigger, ksvc ready.
	if err := client.Core.WaitForResourceReady(brokerName, eventingCommon.BrokerTypeMeta); err != nil {
		t.Error(err)
	}

	if err := client.Core.WaitForResourcesReady(eventingCommon.TriggerTypeMeta); err != nil {
		t.Error(err)
	}

	if err := client.Core.WaitForResourceReady(kserviceName, lib.KsvcTypeMeta); err != nil {
		t.Error(err)
	}

	// Get broker URL.
	metaAddressable := eventingtestresources.NewMetaResource(brokerName, client.Namespace, eventingCommon.BrokerTypeMeta)
	u, err := base.GetAddressableURI(client.Core.Dynamic, metaAddressable)
	if err != nil {
		t.Error(err.Error())
	}

	// Just to make sure all resources are ready.
	time.Sleep(5 * time.Second)

	// Create a sender Job to sender the event.
	senderJob := resources.SenderJob(senderName, u.String())
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

// TODO(chizhg): move this to eventing
// withSubscriberKServiceRefForTrigger returns an option that adds a Subscriber Knative Service Ref for the given Trigger.
func withSubscriberKServiceRefForTrigger(name string) eventingtestresources.TriggerOption {
	return func(t *eventingv1alpha1.Trigger) {
		if name != "" {
			t.Spec.Subscriber = duckv1.Destination{
				Ref: pkgTest.CoreV1ObjectReference("Service", "serving.knative.dev/v1", name),
			}
		}
	}
}

func createResource(client *lib.Client, config map[string]string, folders []string, t *testing.T) *lib.Installer {
	installer := lib.NewInstaller(client.Core.Dynamic, config,
		lib.EndToEndConfigYaml(folders)...)
	if err := installer.Do("create"); err != nil {
		t.Errorf("failed to create, %s", err)
		return nil
	}
	return installer
}

func deleteResource(installer *lib.Installer, t *testing.T) {
	if err := installer.Do("delete"); err != nil {
		t.Errorf("failed to delete, %s", err)
	}
	// Wait for resources to be deleted.
	time.Sleep(15 * time.Second)
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
