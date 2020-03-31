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

package context

import (
	"context"

	"knative.dev/pkg/logging"

	"github.com/google/knative-gcp/pkg/broker/config"
)

type targetKey struct{}

// WithTarget sets a target in the context.
func WithTarget(ctx context.Context, t *config.Target) context.Context {
	return context.WithValue(ctx, targetKey{}, t)
}

// GetTarget gets a target from the context.
func GetTarget(ctx context.Context) *config.Target {
	untyped := ctx.Value(targetKey{})
	if untyped == nil {
		logging.FromContext(ctx).Panic("Unable to fetch Target from context.")
	}
	return untyped.(*config.Target)
}
