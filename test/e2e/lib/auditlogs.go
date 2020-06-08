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
	"context"
	"fmt"
	"os"
	"testing"

	"google.golang.org/grpc/status"

	"cloud.google.com/go/logging/logadmin"
	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	kngcptesting "github.com/google/knative-gcp/pkg/reconciler/testing"
	"github.com/google/knative-gcp/test/e2e/lib/resources"
	"google.golang.org/grpc/codes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PubSubServiceName           = "pubsub.googleapis.com"
	PubSubCreateTopicMethodName = "google.pubsub.v1.Publisher.CreateTopic"
)

type AuditLogsConfig struct {
	SinkGVK              metav1.GroupVersionKind
	SinkName             string
	AuditlogsName        string
	MethodName           string
	Project              string
	ResourceName         string
	ServiceName          string
	SourceServiceAccount string
	Options              []kngcptesting.CloudAuditLogsSourceOption
}

func MakeAuditLogsOrDie(client *Client, config AuditLogsConfig) {
	client.T.Helper()
	so := config.Options
	so = append(so, kngcptesting.WithCloudAuditLogsSourceServiceName(config.ServiceName))
	so = append(so, kngcptesting.WithCloudAuditLogsSourceMethodName(config.MethodName))
	so = append(so, kngcptesting.WithCloudAuditLogsSourceProject(config.Project))
	so = append(so, kngcptesting.WithCloudAuditLogsSourceResourceName(config.ResourceName))
	so = append(so, kngcptesting.WithCloudAuditLogsSourceSink(config.SinkGVK, config.SinkName))
	so = append(so, kngcptesting.WithCloudAuditLogsSourceServiceAccount(config.SourceServiceAccount))
	eventsAuditLogs := kngcptesting.NewCloudAuditLogsSource(config.AuditlogsName, client.Namespace, so...)
	client.CreateAuditLogsOrFail(eventsAuditLogs)

	client.Core.WaitForResourceReadyOrFail(config.AuditlogsName, CloudAuditLogsSourceTypeMeta)
}

func MakeAuditLogsJobOrDie(client *Client, methodName, project, resourceName, serviceName, targetName, eventType string) {
	client.T.Helper()
	job := resources.AuditLogsTargetJob(targetName, []v1.EnvVar{{
		Name:  "SERVICENAME",
		Value: serviceName,
	}, {
		Name:  "METHODNAME",
		Value: methodName,
	}, {
		Name:  "RESOURCENAME",
		Value: resourceName,
	}, {
		Name:  "TYPE",
		Value: eventType,
	}, {
		Name:  "SOURCE",
		Value: v1alpha1.CloudAuditLogsSourceEventSource(serviceName, fmt.Sprintf("projects/%s", project)),
	}, {
		Name:  "SUBJECT",
		Value: resourceName,
	}, {
		Name:  "TIME",
		Value: "6m",
	}})
	client.CreateJobOrFail(job, WithServiceForJob(targetName))
}

func StackdriverSinkExists(t *testing.T, sinkID string) bool {
	t.Helper()
	ctx := context.Background()
	project := os.Getenv(ProwProjectKey)
	client, err := logadmin.NewClient(ctx, project)
	if err != nil {
		t.Fatalf("failed to create LogAdmin client, %s", err.Error())
	}
	defer client.Close()

	_, err = client.Sink(ctx, sinkID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false
		}

		t.Fatalf("Failed from LogAdmin client while retrieving StackdriverSink %s with error %s", sinkID, err.Error())
	}
	return true
}
