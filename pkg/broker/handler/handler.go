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

package handler

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol/pubsub"
	"github.com/google/knative-gcp/pkg/broker/handler/processors"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/logging"
)

// Handler pulls Pubsub messages as events and processes them
// with chain of processors.
type Handler struct {
	// PubsubEvents is the CloudEvents Pubsub protocol to pull
	// messages as events.
	PubsubEvents *pubsub.Protocol

	// Processor is the processor to process events.
	Processor processors.Interface

	// Timeout is the timeout for processing each individual event.
	Timeout time.Duration

	// Concurrency is the number of goroutines that will
	// concurrently process events. If not positive, will fall back
	// to 1.
	Concurrency int

	// cancel is function to stop pulling messages.
	cancel context.CancelFunc
	alive  atomic.Value
}

// Start starts the handler.
// done func will be called if the pubsub inbound is closed.
func (h *Handler) Start(ctx context.Context, done func(error)) {
	ctx, h.cancel = context.WithCancel(ctx)
	h.alive.Store(true)

	go func() {
		// For any reason if inbound is closed, mark alive as false.
		defer h.alive.Store(false)
		done(h.PubsubEvents.OpenInbound(ctx))
	}()

	curr := h.Concurrency
	if curr <= 0 {
		curr = 1
	}
	for i := 0; i < curr; i++ {
		go h.handle(ctx)
	}
}

// Stop stops the handlers.
func (h *Handler) Stop() {
	h.cancel()
}

// IsAlive indicates whether the handler is alive.
func (h *Handler) IsAlive() bool {
	return h.alive.Load().(bool)
}

func (h *Handler) handle(ctx context.Context) {
	for {
		msg, err := h.PubsubEvents.Receive(ctx)
		// It doesn't seem like that these errors will even happen.
		if err == io.EOF {
			logging.FromContext(ctx).Warn("handler goroutine no longer receiving messages from Pubsub")
			break
		} else if err != nil {
			logging.FromContext(ctx).Error("failed to receive the next message from Pubsub", zap.Error(err))
			continue
		}

		event, err := binding.ToEvent(ctx, msg)
		if err != nil {
			logging.FromContext(ctx).Error("failed to convert received message to an event", zap.Any("message", msg), zap.Error(err))
			continue
		}

		pctx := ctx
		if h.Timeout != 0 {
			var cancel context.CancelFunc
			pctx, cancel = context.WithTimeout(ctx, h.Timeout)
			defer cancel()
		}
		perr := h.Processor.Process(pctx, event)
		if perr != nil {
			logging.FromContext(pctx).Error("failed to process event", zap.Any("event", event), zap.Error(perr))
		}
		// This will ack/nack the message.
		if err := msg.Finish(perr); err != nil {
			logging.FromContext(ctx).Warn("failed to finish the message", zap.Any("message", msg), zap.Error(err))
		}
	}
}
