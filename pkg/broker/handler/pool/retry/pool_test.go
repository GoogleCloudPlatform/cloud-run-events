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

package retry

import (
	"context"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	cepubsub "github.com/cloudevents/sdk-go/v2/protocol/pubsub"

	"github.com/google/knative-gcp/pkg/broker/config"
	"github.com/google/knative-gcp/pkg/broker/config/memory"
	"github.com/google/knative-gcp/pkg/broker/handler/pool"
	"github.com/google/knative-gcp/test/e2e/lib"
)

func TestWatchAndSync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	testProject := "test-project"
	ps, psclose := lib.CreateTestPubsubClient(ctx, t, testProject)
	defer psclose()
	signal := make(chan struct{})
	targets := memory.NewEmptyTargets()
	syncPool, err := NewSyncPool(targets,
		pool.WithPubsubClient(ps),
		pool.WithProjectID(testProject),
		pool.WithSyncSignal(signal))
	if err != nil {
		t.Errorf("unexpected error from getting sync pool: %v", err)
	}
	p, err := pool.StartSyncPool(ctx, syncPool)
	if err != nil {
		t.Errorf("unexpected error from starting sync pool: %v", err)
	}
	lib.AssertHandlers(t, p, lib.RetrySyncPool, targets)
	bs := make([]*config.Broker, 0, 4)
	ts := map[string]*config.Target{}

	t.Run("adding some brokers with their targets", func(t *testing.T) {
		// Add some brokers with their targets.
		for i := 0; i < 4; i++ {
			b := lib.GenTestBroker(ctx, t, ps)
			t := lib.GenTestTarget(ctx, t, ps, map[string]string{})
			bs = append(bs, b)
			targets.MutateBroker(b.Namespace, b.Name, func(bm config.BrokerMutation) {
				bm.SetDecoupleQueue(b.DecoupleQueue)
				bm.UpsertTargets(t)
			})
			ts[b.Name] = t
		}
		signal <- struct{}{}
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		lib.AssertHandlers(t, p, lib.RetrySyncPool, targets)
	})

	t.Run("delete and adding targets in brokers", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			t := lib.GenTestTarget(ctx, t, ps, map[string]string{})
			targets.MutateBroker(bs[i].Namespace, bs[i].Name, func(bm config.BrokerMutation) {
				bm.DeleteTargets(ts[bs[i].Name])
				bm.UpsertTargets(t)
			})
			ts[bs[i].Name] = t
		}
		signal <- struct{}{}
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		lib.AssertHandlers(t, p, lib.RetrySyncPool, targets)
	})

	t.Run("deleting all brokers with their targets", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			targets.MutateBroker(bs[i].Namespace, bs[i].Name, func(bm config.BrokerMutation) {
				bm.DeleteTargets(ts[bs[i].Name])
				bm.Delete()
			})
		}
		signal <- struct{}{}
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		lib.AssertHandlers(t, p, lib.RetrySyncPool, targets)
	})
}

func TestRetrySyncPoolE2E(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	testProject := "test-project"

	ps, psclose := lib.CreateTestPubsubClient(ctx, t, testProject)
	defer psclose()
	ceps, err := cepubsub.New(ctx, cepubsub.WithClient(ps))
	if err != nil {
		t.Fatalf("failed to create cloudevents pubsub protocol: %v", err)
	}

	// Create two brokers.
	b1 := lib.GenTestBroker(ctx, t, ps)
	b2 := lib.GenTestBroker(ctx, t, ps)
	targets := memory.NewTargets(&config.TargetsConfig{
		Brokers: map[string]*config.Broker{
			b1.Key(): b1,
			b2.Key(): b2,
		},
	})

	t1 := lib.GenTestTarget(ctx, t, ps, nil)
	t2 := lib.GenTestTarget(ctx, t, ps, map[string]string{"subject": "foo"})
	t3 := lib.GenTestTarget(ctx, t, ps, nil)

	b1t1, b1t1Client, b1t1close := lib.AddTestTargetToBroker(t, targets, t1, b1.Name)
	defer b1t1close()
	b1t2, b1t2Client, b1t2close := lib.AddTestTargetToBroker(t, targets, t2, b1.Name)
	defer b1t2close()
	b2t3, b2t3Client, b2t3close := lib.AddTestTargetToBroker(t, targets, t3, b2.Name)
	defer b2t3close()

	signal := make(chan struct{})
	syncPool, err := NewSyncPool(targets,
		pool.WithPubsubClient(ps),
		pool.WithProjectID(testProject),
		pool.WithSyncSignal(signal))
	if err != nil {
		t.Errorf("unexpected error from getting sync pool: %v", err)
	}

	if _, err := pool.StartSyncPool(ctx, syncPool); err != nil {
		t.Errorf("unexpected error from starting sync pool: %v", err)
	}

	e1 := lib.GenTestEvent("foo1", "bar1", "id1", "source1")
	e2 := lib.GenTestEvent("foo2", "bar2", "id2", "source2")
	e3 := lib.GenTestEvent("foo3", "bar3", "id3", "source3")

	t.Run("target with same broker but different trigger did't receive retry events", func(t *testing.T) {
		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// Target1 for broker1 should receive the event e1.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e1, 1)
		// Target2 for broker1 should't receive the event e2.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e1, 0)

		// Only send an event to trigger topic 1.
		sendEventToTriggerTopic(ctx, t, ceps, t1, &e1)
		<-vctx.Done()
	})

	t.Run("target with different broker did't receive retry events", func(t *testing.T) {
		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// Target1 for broker1 should't  receive the event e3.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e3, 0)
		// Target2 for broker1 should't receive the event e2.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e3, 0)
		// Target3 for broker2 should receive the event e3.
		go lib.VerifyNextReceivedEvent(vctx, t, b2t3, b2t3Client, &e3, 1)

		// Only send an event to trigger topic 3.
		sendEventToTriggerTopic(ctx, t, ceps, t3, &e3)
		<-vctx.Done()
	})

	t.Run("broker's target receive correct retry events", func(t *testing.T) {
		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// Target1 for broker1 should receive the event e1.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e1, 1)
		// Target2 for broker1 should receive the event e2.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e2, 1)
		// Target3 for broker2 should receive the event e3.
		go lib.VerifyNextReceivedEvent(vctx, t, b2t3, b2t3Client, &e3, 1)

		// Send different event to different trigger topic.
		sendEventToTriggerTopic(ctx, t, ceps, t1, &e1)
		sendEventToTriggerTopic(ctx, t, ceps, t2, &e2)
		sendEventToTriggerTopic(ctx, t, ceps, t3, &e3)
		<-vctx.Done()
	})
}

func sendEventToTriggerTopic(ctx context.Context, t *testing.T, ceps *cepubsub.Protocol, ta *config.Target, e *event.Event) {
	t.Helper()
	lib.SentEventToTopic(ctx, t, ceps, ta.RetryQueue.Topic, e)
}
