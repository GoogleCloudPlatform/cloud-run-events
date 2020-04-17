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

package fanout

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
	lib.AssertHandlers(t, p, lib.FanoutSyncPool, targets)
	bs := make([]*config.Broker, 0, 4)

	t.Run("adding new brokers creates new handlers", func(t *testing.T) {
		// First add some brokers.
		for i := 0; i < 4; i++ {
			b := lib.GenTestBroker(ctx, t, ps)
			bs = append(bs, b)
			targets.MutateBroker(b.Namespace, b.Name, func(bm config.BrokerMutation) {
				bm.SetDecoupleQueue(b.DecoupleQueue)
			})
		}
		signal <- struct{}{}
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		lib.AssertHandlers(t, p, lib.FanoutSyncPool, targets)
	})

	t.Run("adding and deleting brokers changes handlers", func(t *testing.T) {
		// Delete old and add new.
		for i := 0; i < 2; i++ {
			targets.MutateBroker(bs[i].Namespace, bs[i].Name, func(bm config.BrokerMutation) {
				bm.Delete()
			})
			b := lib.GenTestBroker(ctx, t, ps)
			targets.MutateBroker(b.Namespace, b.Name, func(bm config.BrokerMutation) {
				bm.SetDecoupleQueue(b.DecoupleQueue)
			})
		}
		signal <- struct{}{}
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		lib.AssertHandlers(t, p, lib.FanoutSyncPool, targets)
	})

	t.Run("deleting all brokers deletes all handlers", func(t *testing.T) {
		// clean up all brokers
		targets.RangeBrokers(func(b *config.Broker) bool {
			targets.MutateBroker(b.Namespace, b.Name, func(bm config.BrokerMutation) {
				bm.Delete()
			})
			return true
		})
		signal <- struct{}{}
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		lib.AssertHandlers(t, p, lib.FanoutSyncPool, targets)
	})
}

func TestFanoutSyncPoolE2E(t *testing.T) {
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

	e := lib.GenTestEvent("foo", "type", "id", "source")

	t.Run("broker's targets receive fanout events", func(t *testing.T) {
		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// Targets for broker1 should both receive the event.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e, 1)
		go lib.VerifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e, 1)
		// Target for broker2 shouldn't receive any event.
		go lib.VerifyNextReceivedEvent(vctx, t, b2t3, b2t3Client, &e, 0)

		// Only send an event to broker1.
		sendEventToBrokerTopic(ctx, t, ceps, b1, &e)
		<-vctx.Done()
	})

	t.Run("target with unmatching filter didn't receive event", func(t *testing.T) {
		t4 := lib.GenTestTarget(ctx, t, ps, map[string]string{"subject": "bar"})
		b1t4, b1t4Client, b1t4close := lib.AddTestTargetToBroker(t, targets, t4, b1.Name)
		defer b1t4close()
		signal <- struct{}{}

		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// The old targets for broker1 should still receive the event.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e, 1)
		go lib.VerifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e, 1)
		// The new target for broker1 shouldn't receive the event
		// because the event doesn't match its filter.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t4, b1t4Client, &e, 0)
		// Target for broker2 still shouldn't receive any event.
		go lib.VerifyNextReceivedEvent(vctx, t, b2t3, b2t3Client, &e, 0)

		// Only send an event to broker1.
		sendEventToBrokerTopic(ctx, t, ceps, b1, &e)
		<-vctx.Done()
	})

	t.Run("event sent to a broker didn't reach another broker's targets", func(t *testing.T) {
		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// This time targets for broker1 shouldn't receive any event.
		go lib.VerifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e, 0)
		go lib.VerifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e, 0)
		// Target for broker2 should receive the event.
		go lib.VerifyNextReceivedEvent(vctx, t, b2t3, b2t3Client, &e, 1)

		// Only send an event to broker2.
		sendEventToBrokerTopic(ctx, t, ceps, b2, &e)
		<-vctx.Done()
	})
}

func sendEventToBrokerTopic(ctx context.Context, t *testing.T, ceps *cepubsub.Protocol, b *config.Broker, e *event.Event) {
	t.Helper()
	lib.SentEventToTopic(ctx, t, ceps, b.DecoupleQueue.Topic, e)
}
