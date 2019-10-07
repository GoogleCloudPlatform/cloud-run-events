/*
Copyright 2019 The Knative Authors

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
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/util/uuid"
	"knative.dev/eventing/test/base/resources"
	"knative.dev/eventing/test/common"
)

/*
singleEvent tests the following scenario:

EventSource ---> Channel ---> Subscription ---> Service(Logger)

*/

// SingleEventHelperForChannelTestHelper is the helper function for header_test
func SingleEventHelperForChannelTestHelper(t *testing.T, encoding string, channelTestRunner common.ChannelTestRunner) {
	channelName := "conformance-headers-channel-" + encoding
	senderName := "conformance-headers-sender-" + encoding
	subscriptionName := "conformance-headers-subscription-" + encoding
	loggerPodName := "conformance-headers-logger-pod-" + encoding

	channelTestRunner.RunTests(t, common.FeatureBasic, func(st *testing.T, channel string) {
		st.Logf("Running header conformance test with channel %q", channel)
		client := common.Setup(st, true)
		defer common.TearDown(client)

		// create channel
		st.Logf("Creating channel")
		channelTypeMeta := common.GetChannelTypeMeta(channel)
		client.CreateChannelOrFail(channelName, channelTypeMeta)

		// create logger service as the subscriber
		pod := resources.EventDetailsPod(loggerPodName)
		client.CreatePodOrFail(pod, common.WithService(loggerPodName))

		// create subscription to subscribe the channel, and forward the received events to the logger service
		client.CreateSubscriptionOrFail(
			subscriptionName,
			channelName,
			channelTypeMeta,
			resources.WithSubscriberForSubscription(loggerPodName),
		)

		// wait for all test resources to be ready, so that we can start sending events
		if err := client.WaitForAllTestResourcesReady(); err != nil {
			st.Fatalf("Failed to get all test resources ready: %v", err)
		}

		// send fake CloudEvent to the channel
		eventID := fmt.Sprintf("%s", uuid.NewUUID())
		body := fmt.Sprintf("TestSingleHeaderEvent %s", eventID)
		event := &resources.CloudEvent{
			ID:       eventID,
			Source:   senderName,
			Type:     resources.CloudEventDefaultType,
			Data:     fmt.Sprintf(`{"msg":%q}`, body),
			Encoding: encoding,
		}

		st.Logf("Sending event with tracing headers to %s", senderName)
		if err := client.SendFakeEventWithTracingToAddressable(senderName, channelName, channelTypeMeta, event); err != nil {
			st.Fatalf("Failed to send fake CloudEvent to the channel %q", channelName)
		}

		// verify the logger service receives the event
		st.Logf("Logging for event with body %s", body)

		if err := client.CheckLog(loggerPodName, common.CheckerContains(body)); err != nil {
			st.Fatalf("String %q not found in logs of logger pod %q: %v", body, loggerPodName, err)
		}

		//verify that required x-b3-spani and x-b3-traceid are set
		requiredHeaderNameList := []string{"X-B3-Traceid", "X-B3-Spanid", "X-B3-Sampled"}
		for _, headerName := range requiredHeaderNameList {
			expectedHeaderLog := fmt.Sprintf("Got Header %s:", headerName)
			if err := client.CheckLog(loggerPodName, common.CheckerContains(expectedHeaderLog)); err != nil {
				st.Fatalf("String %q not found in logs of logger pod %q: %v", expectedHeaderLog, loggerPodName, err)
			}
		}

		//TODO report on optional x-b3-parentspanid and x-b3-sampled if present?
		//TODO report x-custom-header

	})
}
