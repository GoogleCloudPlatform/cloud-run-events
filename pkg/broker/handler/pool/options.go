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

package pool

import (
	"fmt"
	"runtime"
	"time"

	"cloud.google.com/go/pubsub"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	"github.com/google/knative-gcp/pkg/utils"
)

var (
	defaultHandlerConcurrency     = runtime.NumCPU()
	defaultMaxConcurrencyPerEvent = 1
	defaultTimeout                = 10 * time.Minute
	defaultCeClientOpts           = []ceclient.Option{
		ceclient.WithUUIDs(),
		ceclient.WithTimeNow(),
		ceclient.WithTracePropagation(),
	}
)

// Options holds all the options for create handler pool.
type Options struct {
	// ProjectID is the project for pubsub.
	ProjectID string
	// HandlerConcurrency is the number of goroutines
	// will be spawned in each handler.
	HandlerConcurrency int
	// MaxConcurrencyPerEvent is the max number of goroutines
	// will be spawned to handle an event.
	MaxConcurrencyPerEvent int
	// TimeoutPerEvent is the timeout for handling an event.
	TimeoutPerEvent time.Duration
	// PubsubClient is the pubsub client used to receive pubsub messages.
	PubsubClient *pubsub.Client
	// PubsubReceiveSettings is the pubsub receive settings.
	PubsubReceiveSettings pubsub.ReceiveSettings
	// CeClientOptions is the options used to create cloudevents client.
	CeClientOptions []ceclient.Option
}

// NewOptions creates a Options.
func NewOptions(opts ...Option) (*Options, error) {
	opt := &Options{
		HandlerConcurrency:     defaultHandlerConcurrency,
		MaxConcurrencyPerEvent: defaultMaxConcurrencyPerEvent,
		TimeoutPerEvent:        defaultTimeout,
		PubsubReceiveSettings:  pubsub.DefaultReceiveSettings,
	}
	for _, o := range opts {
		o(opt)
	}
	if opt.ProjectID == "" {
		pid, err := utils.ProjectID(opt.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get default ProjectID: %w", err)
		}
		opt.ProjectID = pid
	}
	if opt.CeClientOptions == nil {
		opt.CeClientOptions = defaultCeClientOpts
	}
	return opt, nil
}

// Option is for providing individual option.
type Option func(*Options)

// WithProjectID sets project ID.
func WithProjectID(id string) Option {
	return func(o *Options) {
		o.ProjectID = id
	}
}

// WithHandlerConcurrency sets HandlerConcurrency.
func WithHandlerConcurrency(c int) Option {
	return func(o *Options) {
		o.HandlerConcurrency = c
	}
}

// WithMaxConcurrentPerEvent sets MaxConcurrencyPerEvent.
func WithMaxConcurrentPerEvent(c int) Option {
	return func(o *Options) {
		o.MaxConcurrencyPerEvent = c
	}
}

// WithTimeoutPerEvent sets TimeoutPerEvent.
func WithTimeoutPerEvent(t time.Duration) Option {
	return func(o *Options) {
		o.TimeoutPerEvent = t
	}
}

// WithPubsubClient sets the PubsubClient.
func WithPubsubClient(c *pubsub.Client) Option {
	return func(o *Options) {
		o.PubsubClient = c
	}
}

// WithPubsubReceiveSettings sets PubsubReceiveSettings.
func WithPubsubReceiveSettings(s pubsub.ReceiveSettings) Option {
	return func(o *Options) {
		o.PubsubReceiveSettings = s
	}
}
