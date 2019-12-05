/*
Copyright 2019 Google LLC

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

package testing

import (
	"context"

	"github.com/google/knative-gcp/pkg/gclient/pubsub"
	"google.golang.org/api/option"
)

// NewClient creates a new test Pub/Sub client.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (pubsub.Client, error) {
	return &TestClient{Project: projectID}, nil
}

// TestClient is a test Pub/Sub client.
type TestClient struct {
	Project string
}

// Verify that it satisfies the pubsub.Client interface.
var _ pubsub.Client = &TestClient{}

// Close implements client.Close
func (c *TestClient) Close() error {
	return nil
}

// Topic implements Client.Topic.
func (c *TestClient) Topic(id string) pubsub.Topic {
	return &TestTopic{id: id}
}

// Subscription implements Client.Subscription.
func (c *TestClient) Subscription(id string) pubsub.Subscription {
	return &TestSubscription{id: id}
}

// CreateSubscription implements Client.CreateSubscription.
func (c *TestClient) CreateSubscription(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (pubsub.Subscription, error) {
	return &TestSubscription{id: id}, nil
}

// CreateTopic implements pubsub.Client.CreateTopic
func (c *TestClient) CreateTopic(ctx context.Context, id string) (pubsub.Topic, error) {
	return &TestTopic{id: id}, nil
}
