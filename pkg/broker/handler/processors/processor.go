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

package processors

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Processor is the interface to process an event.
type Processor interface {
	// Process processes an event. It may decide to terminate the processing early
	// or it can pass the event to the next Processor for further processing.
	Process(ctx context.Context, e *event.Event) error
	// Next returns the next Processor to process events.
	Next() Processor
}

// ChainableProcessor is the interface to chainable Processor.
type ChainableProcessor interface {
	Processor

	// WithNext sets the next Processor to pass the event.
	WithNext(ChainableProcessor) ChainableProcessor
}

// BaseProcessor provoides implementation to set and get
// next processor. It can gracefully handle the case where the next
// processor doesn't exist.
type BaseProcessor struct {
	N ChainableProcessor
}

// Next returns the next processor otherwise it will return a
// no-op processor so that caller doesn't need to worry about
// calling a nil processor.
func (p *BaseProcessor) Next() Processor {
	if p.N == nil {
		return noop
	}
	return p.N
}

// WithNext sets the next Processor to pass the event.
func (p *BaseProcessor) WithNext(n ChainableProcessor) ChainableProcessor {
	p.N = n
	return p.N
}

// ChainProcessors chains the given processors in order.
func ChainProcessors(first ChainableProcessor, rest ...ChainableProcessor) Processor {
	next := first
	for _, p := range rest {
		next = next.WithNext(p)
	}
	return first
}

var noop = &noOpProcessor{}

type noOpProcessor struct{}

func (p noOpProcessor) Process(_ context.Context, _ *event.Event) error {
	return nil
}

func (p noOpProcessor) Next() Processor {
	return noop
}
