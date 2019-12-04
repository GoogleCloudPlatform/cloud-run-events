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

package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
)

// pubsubTopic wraps pubsub.Topic. Is the topic that will be used everywhere except unit tests.
type pubsubTopic struct {
	topic *pubsub.Topic
}

// Verify that it satisfies the pubsub.Topic interface.
var _ Topic = &pubsubTopic{}

// Exists implements pubsub.Topic.Exists
func (t *pubsubTopic) Exists(ctx context.Context) (bool, error) {
	return t.topic.Exists(ctx)
}
