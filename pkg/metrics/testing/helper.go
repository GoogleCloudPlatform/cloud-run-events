package testing

import (
	"knative.dev/pkg/metrics/metricstest"
	"testing"
)

func ResetIngressMetrics() {
	// OpenCensus metrics carry global state that need to be reset between unit tests.
	metricstest.Unregister("event_count", "event_dispatch_latencies")
}

func ResetDeliveryMetrics() {
	// OpenCensus metrics carry global state that need to be reset between unit tests.
	metricstest.Unregister("event_count", "event_dispatch_latencies", "event_processing_latencies")
}

func ExpectMetrics(t *testing.T, f func() error) {
	t.Helper()
	if err := f(); err != nil {
		t.Errorf("Reporter expected success but got error: %v", err)
	}
}
