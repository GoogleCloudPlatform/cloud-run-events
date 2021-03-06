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
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	pkgerrors "github.com/pkg/errors"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/eventing/test/lib"
	"knative.dev/eventing/test/lib/duck"
	"knative.dev/eventing/test/lib/resources"
	pkgTest "knative.dev/pkg/test"

	duckv1 "github.com/google/knative-gcp/pkg/apis/duck/v1"
	knativegcp "github.com/google/knative-gcp/pkg/client/clientset/versioned"
	"github.com/google/knative-gcp/test/lib/metrics"
	"github.com/google/knative-gcp/test/lib/operations"
)

// Setup runs the Setup in the common eventing test framework.
func Setup(ctx context.Context, t *testing.T, runInParallel, workloadIdentity bool) *Client {
	t.Helper()
	client, err := newClient(pkgTest.Flags.Kubeconfig, pkgTest.Flags.Cluster)
	if err != nil {
		t.Fatalf("Failed to initialize client for Knative GCP: %v", err)
	}

	coreClient := lib.Setup(t, runInParallel)
	client.Core = coreClient
	client.Namespace = coreClient.Namespace
	client.Tracker = coreClient.Tracker
	client.T = t
	GetCredential(ctx, coreClient, workloadIdentity)
	return client
}

func newClient(configPath string, clusterName string) (*Client, error) {
	config, err := pkgTest.BuildClientConfig(configPath, clusterName)
	if err != nil {
		return nil, err
	}

	kgc, err := knativegcp.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{KnativeGCP: kgc}, nil
}

// TearDown runs the TearDown in the common eventing test framework.
func TearDown(ctx context.Context, client *Client) {
	client.T.Helper()

	printAllPodMetricsIfTestFailed(ctx, client)
	lib.TearDown(client.Core)
}

// Client holds instances of interfaces for making requests to Knative.
type Client struct {
	Core *lib.Client

	KnativeGCP *knativegcp.Clientset
	Namespace  string
	T          *testing.T
	Tracker    *lib.Tracker
}

var setStackDriverConfigOnce = sync.Once{}

func (c *Client) SetupStackDriverMetrics(ctx context.Context, t *testing.T) {
	t.Helper()
	setStackDriverConfigOnce.Do(func() {
		err := pkgTest.UpdateConfigMap(ctx, c.Core.Kube, "cloud-run-events", "config-observability", map[string]string{
			"metrics.allow-stackdriver-custom-metrics":     "false",
			"metrics.backend-destination":                  "stackdriver",
			"metrics.stackdriver-custom-metrics-subdomain": "cloud.google.com",
			"metrics.reporting-period-seconds":             "60",
		})
		if err != nil {
			t.Fatalf("Unable to set the ConfigMap: %v", err)
		}
	})
}

func (c *Client) SetupStackDriverMetricsInNamespace(ctx context.Context, t *testing.T) {
	t.Helper()
	c.SetupStackDriverMetrics(ctx, t)
	_ = c.Core.CreateConfigMapOrFail("eventing-config-observability", c.Namespace, map[string]string{
		"metrics.allow-stackdriver-custom-metrics":     "true",
		"metrics.backend-destination":                  "stackdriver",
		"metrics.stackdriver-custom-metrics-subdomain": "cloud.google.com",
		"metrics.reporting-period-seconds":             "60",
	})
}

const (
	interval = 1 * time.Second
	timeout  = 5 * time.Minute
)

// TODO(chizhg): move this function to knative/pkg/test or knative/eventing/test
// WaitForResourceReady waits until the specified resource in the given namespace are ready.
func (c *Client) WaitUntilJobDone(ctx context.Context, namespace, name string) (string, error) {
	cc := c.Core
	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		job, err := cc.Kube.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Println(namespace, name, "not found", err)
				// keep polling
				return false, nil
			}
			return false, err
		}
		return operations.IsJobComplete(job), nil
	})
	if err != nil {
		return "", err
	}

	// poll until the pod is terminated.
	err = wait.PollImmediate(interval, timeout, func() (bool, error) {
		pod, err := operations.GetJobPodByJobName(context.TODO(), cc.Kube, namespace, name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Println(namespace, name, "not found", err)
				// keep polling
				return false, nil
			}
			return false, err
		}
		if pod != nil {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated != nil {
					return true, nil
				}
			}
		}
		return false, nil
	})

	if err != nil {
		return "", err
	}
	pod, err := operations.GetJobPodByJobName(context.TODO(), cc.Kube, namespace, name)
	if err != nil {
		return "", err
	}
	return operations.GetFirstTerminationMessage(pod), nil
}

// TODO(chizhg): move this function to knative/pkg/test or knative/eventing/test
func (c *Client) LogsFor(ctx context.Context, namespace, name string, tm *metav1.TypeMeta) (string, error) {
	cc := c.Core
	// Get all pods in this namespace.
	pods, err := cc.Kube.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	logs := make([]string, 0)

	// Look for a pod with the name that was passed in inside the pod name.
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, name) {
			// Collect all the logs from all the containers for this pod.
			if l, err := cc.Kube.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).DoRaw(ctx); err != nil {
				logs = append(logs, err.Error())
			} else {
				logs = append(logs, string(l))
			}
		}
	}

	// Did we find a match like the given name?
	if len(logs) == 0 {
		return "", fmt.Errorf(`pod for "%s/%s" [%s] not found`, namespace, name, tm.String())
	}

	return strings.Join(logs, "\n"), nil
}

// TODO make this function more generic.
func (c *Client) StackDriverEventCountMetricFor(_, projectID, filter string) (int64, error) {
	metricClient, err := monitoring.NewMetricClient(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to create stackdriver metric client: %v", err)
	}

	// TODO make times configurable if needed.
	metricRequest := &monitoringpb.ListTimeSeriesRequest{
		Name:   fmt.Sprintf("projects/%s", projectID),
		Filter: filter,
		Interval: &monitoringpb.TimeInterval{
			// Starting 5 minutes back until now.
			StartTime: &timestamp.Timestamp{Seconds: time.Now().Add(-5 * time.Minute).Unix()},
			EndTime:   &timestamp.Timestamp{Seconds: time.Now().Unix()},
		},
		// Delta counts aggregated every 2 minutes.
		// We aggregate for count as other aggregations will give higher values.
		// The reason is that PubSub upon an error, will retry, thus we will be recording multiple events.
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod:    &duration.Duration{Seconds: 120},
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_DELTA,
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_COUNT,
		},
	}

	res, err := metrics.ListTimeSeries(context.TODO(), metricClient, metricRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to iterate over result: %v", err)
	}
	if len(res) == 0 {
		return 0, errors.New("no metric reported")
	}
	if len(res[0].GetPoints()) == 0 {
		return 0, errors.New("no metric points reported")
	}
	return res[0].GetPoints()[0].GetValue().GetInt64Value(), nil
}

// WaitForSourceAuthCheckPendingOrFail waits for the GCP Sources
// to have authenticationCheckPending condition reason for PullSubscriptionConditionReady condition or fail,
// and checks if the condition contains wanted message.
// To use this function, the given resource must have implemented the PubSub Status duck-type,
// and have PullSubscriptionConditionReady condition.
func (c *Client) WaitForSourceAuthCheckPendingOrFail(name string, typemeta *metav1.TypeMeta, wantMessage string) {
	namespace := c.Namespace
	metaResource := resources.NewMetaResource(name, namespace, typemeta)
	waitErr := WaitForSourceAuthCheckPending(c.Core.Dynamic, metaResource, wantMessage)
	// Get the real-time object right after it is running into the desired status.
	untyped, err := duck.GetGenericObject(c.Core.Dynamic, metaResource, &duckv1.PubSub{})
	if err != nil {
		c.T.Fatalf("Failed to get the object %v-%s: %v", *typemeta, name, err)
	}
	if waitErr != nil {
		if untyped != nil {
			c.T.Errorf("Object that did not run into authenticationCheckPending or did not have wanted message %v-%s when dumping error state: %+v", *typemeta, name, untyped)
		}
		c.T.Fatalf("Failed to get %s-%s with authenticationCheckPending reason in type PullSubscriptionConditionReady : %+v", typemeta, name, pkgerrors.WithStack(err))
	}
}
