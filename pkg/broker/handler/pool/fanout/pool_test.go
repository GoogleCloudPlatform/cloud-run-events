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
	"net/http/httptest"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	cepubsub "github.com/cloudevents/sdk-go/v2/protocol/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/google/knative-gcp/pkg/broker/config"
	"github.com/google/knative-gcp/pkg/broker/handler/pool"
)

func TestWatchAndSync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	testProject := "test-project"
	ps, psclose := testPubsubClient(ctx, t, testProject)
	defer psclose()
	targets := make(chan *config.TargetsConfig)
	defer close(targets)
	emptyTargetConfig := new(config.TargetsConfig)
	p, err := StartSyncPool(ctx, config.NewTargetsWatcher(targets),
		pool.WithPubsubClient(ps),
		pool.WithProjectID(testProject),
	)
	if err != nil {
		t.Errorf("unexpected error from starting sync pool: %v", err)
	}
	assertHandlers(t, p, emptyTargetConfig)
	bs := make([]*config.Broker, 0, 4)

	targetConfig := &config.TargetsConfig{Brokers: make(map[string]*config.Broker)}
	t.Run("adding new brokers creates new handlers", func(t *testing.T) {
		// First add some brokers.
		for i := 0; i < 4; i++ {
			b := genTestBroker(ctx, t, ps)
			targetConfig.Brokers[config.BrokerKey(b.Namespace, b.Name)] = b
			bs = append(bs, b)
		}
		targets <- targetConfig
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		assertHandlers(t, p, targetConfig)
	})

	t.Run("adding and deleting brokers changes handlers", func(t *testing.T) {
		newTargetConfig := &config.TargetsConfig{Brokers: make(map[string]*config.Broker)}
		// Populate with new brokers.
		for i := 0; i < 2; i++ {
			b := genTestBroker(ctx, t, ps)
			targetConfig.Brokers[config.BrokerKey(b.Namespace, b.Name)] = b
			bs = append(bs, b)
		}
		targets <- newTargetConfig
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		assertHandlers(t, p, newTargetConfig)
	})

	t.Run("deleting all brokers deletes all handlers", func(t *testing.T) {
		// clean up all brokers
		targets <- emptyTargetConfig
		// Wait a short period for the handlers to be updated.
		<-time.After(time.Second)
		assertHandlers(t, p, emptyTargetConfig)
	})
}

func TestSyncPoolE2E(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	testProject := "test-project"

	ps, psclose := testPubsubClient(ctx, t, testProject)
	defer psclose()
	ceps, err := cepubsub.New(ctx, cepubsub.WithClient(ps))
	if err != nil {
		t.Fatalf("failed to create cloudevents pubsub protocol: %v", err)
	}

	// Create two brokers.
	b1 := genTestBroker(ctx, t, ps)
	b2 := genTestBroker(ctx, t, ps)
	targetsConfig := &config.TargetsConfig{
		Brokers: map[string]*config.Broker{
			config.BrokerKey(b1.Namespace, b1.Name): b1,
			config.BrokerKey(b2.Namespace, b2.Name): b2,
		},
	}
	b1t1, b1t1Client, b1t1close := addTestTargetToBroker(t, b1, nil)
	defer b1t1close()
	b1t2, b1t2Client, b1t2close := addTestTargetToBroker(t, b1, map[string]string{"subject": "foo"})
	defer b1t2close()
	b2t1, b2t1Client, b2t1close := addTestTargetToBroker(t, b2, nil)
	defer b2t1close()

	targetsCh := make(chan *config.TargetsConfig)
	watcher := config.NewTargetsWatcher(targetsCh)
	targetsCh <- targetsConfig
	defer close(targetsCh)
	if _, err := StartSyncPool(ctx, watcher,
		pool.WithProjectID(testProject),
		pool.WithPubsubClient(ps),
	); err != nil {
		t.Errorf("unexpected error from starting sync pool: %v", err)
	}

	e := event.New()
	e.SetSubject("foo")
	e.SetType("type")
	e.SetID("id")
	e.SetSource("source")

	t.Run("broker's targets receive fanout events", func(t *testing.T) {
		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// Targets for broker1 should both receive the event.
		go verifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e, 1)
		go verifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e, 1)
		// Target for broker2 shouldn't receive any event.
		go verifyNextReceivedEvent(vctx, t, b2t1, b2t1Client, &e, 0)

		// Only send an event to broker1.
		sendEventForBroker(ctx, t, ceps, b1, &e)
		<-vctx.Done()
	})

	t.Run("target with unmatching filter didn't receive event", func(t *testing.T) {
		newTargetsConfig := proto.Clone(targetsConfig).(*config.TargetsConfig)
		b1 = newTargetsConfig.Brokers[config.BrokerKey(b1.Namespace, b1.Name)]
		b1t3, b1t3Client, b1t3close := addTestTargetToBroker(t, b1, map[string]string{"subject": "bar"})
		defer b1t3close()
		targetsCh <- newTargetsConfig

		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// The old targets for broker1 should still receive the event.
		go verifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e, 1)
		go verifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e, 1)
		// The new target for broker1 shouldn't receive the event
		// because the event doesn't match its filter.
		go verifyNextReceivedEvent(vctx, t, b1t3, b1t3Client, &e, 0)
		// Target for broker2 still shouldn't receive any event.
		go verifyNextReceivedEvent(vctx, t, b2t1, b2t1Client, &e, 0)

		// Only send an event to broker1.
		sendEventForBroker(ctx, t, ceps, b1, &e)
		<-vctx.Done()
	})

	t.Run("event sent to a broker didn't reach another broker's targets", func(t *testing.T) {
		// Set timeout context so that verification can be done before
		// exiting test func.
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// This time targets for broker1 shouldn't receive any event.
		go verifyNextReceivedEvent(vctx, t, b1t1, b1t1Client, &e, 0)
		go verifyNextReceivedEvent(vctx, t, b1t2, b1t2Client, &e, 0)
		// Target for broker2 should receive the event.
		go verifyNextReceivedEvent(vctx, t, b2t1, b2t1Client, &e, 1)

		// Only send an event to broker2.
		sendEventForBroker(ctx, t, ceps, b2, &e)
		<-vctx.Done()
	})
}

func verifyNextReceivedEvent(ctx context.Context, t *testing.T, receiver string, client *cehttp.Protocol, wantEvent *event.Event, wantCnt int) {
	t.Helper()

	gotCnt := 0
	defer func() {
		if gotCnt != wantCnt {
			t.Errorf("[%s] event received got=%d, want=%d", receiver, gotCnt, wantCnt)
		}
	}()

	msg, err := client.Receive(ctx)
	if err != nil {
		// In case Receive is stopped.
		return
	}
	msg.Finish(nil)
	gotEvent, err := binding.ToEvent(ctx, msg)
	if err != nil {
		t.Errorf("[%s] ingress received message cannot be converted to an event: %v", receiver, err)
	}
	gotCnt++
	// Force the time to be the same so that we can compare easier.
	gotEvent.SetTime(wantEvent.Time())
	if diff := cmp.Diff(wantEvent, gotEvent); diff != "" {
		t.Errorf("[%s] target received event (-want,+got): %v", receiver, diff)
	}
}

func sendEventForBroker(ctx context.Context, t *testing.T, ceps *cepubsub.Protocol, b *config.Broker, e *event.Event) {
	t.Helper()
	ctx = cecontext.WithTopic(ctx, b.DecoupleQueue.Topic)
	if err := ceps.Send(ctx, binding.ToMessage(e)); err != nil {
		t.Errorf("failed to seed event to pubsub: %v", err)
	}
}

func testPubsubClient(ctx context.Context, t *testing.T, projectID string) (*pubsub.Client, func()) {
	t.Helper()
	srv := pstest.NewServer()
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial test pubsub connection: %v", err)
	}
	close := func() {
		srv.Close()
		conn.Close()
	}
	c, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("failed to create test pubsub client: %v", err)
	}
	return c, close
}

func genTestBroker(ctx context.Context, t *testing.T, ps *pubsub.Client) *config.Broker {
	t.Helper()
	tn := "topic-" + uuid.New().String()
	sn := "sub-" + uuid.New().String()

	tt, err := ps.CreateTopic(ctx, tn)
	if err != nil {
		t.Fatalf("failed to create test pubsub topic: %v", err)
	}
	if _, err := ps.CreateSubscription(ctx, sn, pubsub.SubscriptionConfig{Topic: tt}); err != nil {
		t.Fatalf("failed to create test subscription: %v", err)
	}

	return &config.Broker{
		Name:      "broker-" + uuid.New().String(),
		Namespace: "ns",
		DecoupleQueue: &config.Queue{
			Topic:        tn,
			Subscription: sn,
		},
		State: config.State_READY,
	}
}

func addTestTargetToBroker(t *testing.T, broker *config.Broker, filters map[string]string) (string, *cehttp.Protocol, func()) {
	t.Helper()

	targetClient, err := cehttp.New()
	if err != nil {
		t.Fatalf("failed to create target cloudevents client: %v", err)
	}
	targetSvr := httptest.NewServer(targetClient)

	tn := "target-" + uuid.New().String()
	if broker.Targets == nil {
		broker.Targets = make(map[string]*config.Target)
	}
	broker.Targets[tn] = &config.Target{
		Name:             tn,
		Address:          targetSvr.URL,
		FilterAttributes: filters,
		State:            config.State_READY,
		// TODO(yolocs): once retry is ready, we also need to set retry queue.
	}

	return broker.Name + "/" + tn, targetClient, targetSvr.Close
}

func assertHandlers(t *testing.T, p *SyncPool, c *config.TargetsConfig) {
	t.Helper()
	gotHandlers := make(map[string]bool)
	wantHandlers := make(map[string]bool)

	for key := range p.pool {
		gotHandlers[key] = true
	}

	for _, b := range c.GetBrokers() {
		wantHandlers[config.BrokerKey(b.Namespace, b.Name)] = true
	}

	if diff := cmp.Diff(wantHandlers, gotHandlers); diff != "" {
		t.Errorf("handlers map (-want,+got): %v", diff)
	}
}
