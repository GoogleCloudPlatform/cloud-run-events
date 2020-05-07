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

package ingress

import (
	"context"
	"fmt"
	"strconv"
	"time"

	m "github.com/google/knative-gcp/pkg/broker/metrics"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"knative.dev/pkg/metrics"
)

// stats_exporter is adapted from knative.dev/eventing/pkg/broker/ingress/stats_reporter.go
// with the following changes:
// - Metric descriptions are updated to match GCP broker specifics.
// - Removed StatsReporter interface and directly use helper methods instead.

var (
	// dispatchTimeInMsecM records the time spent dispatching an event to
	// a decouple queue, in milliseconds.
	dispatchTimeInMsecM = stats.Float64(
		"event_dispatch_latencies",
		"The time spent dispatching an event to the decouple topic",
		stats.UnitMilliseconds,
	)
)

type PodName string
type ContainerName string

type reportArgs struct {
	namespace    string
	broker       string
	eventType    string
	responseCode int
}

func (r *StatsReporter) register() error {
	tagKeys := []tag.Key{
		m.ContainerNameKey,
		m.BrokerNameKey,
		m.EventTypeKey,
		m.ResponseCodeKey,
		m.ResponseCodeClassKey,
		m.PodNameKey,
		m.ContainerNameKey,
	}

	// Create view to see our measurements.
	return view.Register(
		&view.View{
			Name:        "event_count",
			Description: "Number of events received by a Broker",
			Measure:     r.dispatchTimeInMsecM,
			Aggregation: view.Count(),
			TagKeys:     tagKeys,
		},
		&view.View{
			Name:        r.dispatchTimeInMsecM.Name(),
			Description: r.dispatchTimeInMsecM.Description(),
			Measure:     r.dispatchTimeInMsecM,
			Aggregation: view.Distribution(metrics.Buckets125(1, 10000)...), // 1, 2, 5, 10, 20, 50, 100, 500, 1000, 5000, 10000
			TagKeys:     tagKeys,
		},
	)
}

// NewStatsReporter creates a new StatsReporter.
func NewStatsReporter(podName PodName, containerName ContainerName) (*StatsReporter, error) {
	r := &StatsReporter{
		podName:       podName,
		containerName: containerName,
		dispatchTimeInMsecM: stats.Float64(
			"event_dispatch_latencies",
			"The time spent dispatching an event to the decouple topic",
			stats.UnitMilliseconds,
		),
	}
	if err := r.register(); err != nil {
		return nil, fmt.Errorf("failed to register ingress stats: %w", err)
	}
	return r, nil
}

// StatsReporter reports ingress metrics.
type StatsReporter struct {
	podName       PodName
	containerName ContainerName
	// dispatchTimeInMsecM records the time spent dispatching an event to a decouple queue, in
	// milliseconds.
	dispatchTimeInMsecM *stats.Float64Measure
}

func (r *StatsReporter) reportEventDispatchTime(ctx context.Context, args reportArgs, d time.Duration) error {
	tag, err := tag.New(
		ctx,
		tag.Insert(m.PodNameKey, string(r.podName)),
		tag.Insert(m.ContainerNameKey, string(r.containerName)),
		tag.Insert(m.NamespaceNameKey, args.namespace),
		tag.Insert(m.BrokerNameKey, args.broker),
		tag.Insert(m.EventTypeKey, args.eventType),
		tag.Insert(m.ResponseCodeKey, strconv.Itoa(args.responseCode)),
		tag.Insert(m.ResponseCodeClassKey, metrics.ResponseCodeClass(args.responseCode)),
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics tag: %v", err)
	}
	// convert time.Duration in nanoseconds to milliseconds.
	metrics.Record(tag, r.dispatchTimeInMsecM.M(float64(d/time.Millisecond)))
	return nil
}
