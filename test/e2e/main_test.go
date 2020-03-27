// +build e2e

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
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/knative-gcp/test"
	"github.com/google/knative-gcp/test/e2e/lib"
	"github.com/google/knative-gcp/test/e2e/lib/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingtest "knative.dev/eventing/test"
	eventingtestlib "knative.dev/eventing/test/lib"
	"knative.dev/pkg/test/zipkin"
)

var channelTestRunner eventingtestlib.ChannelTestRunner
var authConfig lib.AuthConfig

func TestMain(m *testing.M) {
	test.InitializeFlags()
	eventingtest.InitializeEventingFlags()
	channelTestRunner = eventingtestlib.ChannelTestRunner{
		// ChannelFeatureMap saves the channel-features mapping.
		// Each pair means the channel support the given list of features.
		ChannelFeatureMap: map[metav1.TypeMeta][]eventingtestlib.Feature{
			{
				APIVersion: resources.MessagingAPIVersion,
				Kind:       "Channel",
			}: {
				eventingtestlib.FeatureBasic,
				eventingtestlib.FeatureRedelivery,
				eventingtestlib.FeaturePersistence,
			},
		},
		ChannelsToTest: eventingtest.EventingFlags.Channels,
	}
	authConfig.WorkloadIdentityEnabled = test.Flags.WorkloadIdentityEnabled
	// The format of a Google Cloud Service Account is: service-account-name@project-id.iam.gserviceaccount.com.
	authConfig.PubsubServiceAccount = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", strings.TrimSpace(test.Flags.PubsubServiceAccount), os.Getenv(lib.ProwProjectKey))
	// Any tests may SetupZipkinTracing, it will only actually be done once. This should be the ONLY
	// place that cleans it up. If an individual test calls this instead, then it will break other
	// tests that need the tracing in place.
	defer zipkin.CleanupZipkinTracingSetup(log.Printf)

	os.Exit(m.Run())
}
