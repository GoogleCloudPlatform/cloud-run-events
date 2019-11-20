/*
Copyright 2019 The Knative Authors

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

package main

import (
	"context"
	"fmt"

	eventsv1alpha1 "github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	messagingv1alpha1 "github.com/google/knative-gcp/pkg/apis/messaging/v1alpha1"
	pubsubv1alpha1 "github.com/google/knative-gcp/pkg/apis/pubsub/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/eventing/pkg/logconfig"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/system"
	tracingconfig "knative.dev/pkg/tracing/config"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	// For group messaging.cloud.google.com.
	messagingv1alpha1.SchemeGroupVersion.WithKind("Channel"):   &messagingv1alpha1.Channel{},
	messagingv1alpha1.SchemeGroupVersion.WithKind("Decorator"): &messagingv1alpha1.Decorator{},

	// For group events.cloud.google.com.
	eventsv1alpha1.SchemeGroupVersion.WithKind("Storage"):   &eventsv1alpha1.Storage{},
	eventsv1alpha1.SchemeGroupVersion.WithKind("Scheduler"): &eventsv1alpha1.Scheduler{},
	eventsv1alpha1.SchemeGroupVersion.WithKind("PubSub"):    &eventsv1alpha1.PubSub{},

	// For group pubsub.cloud.google.com.
	pubsubv1alpha1.SchemeGroupVersion.WithKind("PullSubscription"): &pubsubv1alpha1.PullSubscription{},
	pubsubv1alpha1.SchemeGroupVersion.WithKind("Topic"):            &pubsubv1alpha1.Topic{},
}

func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	// store := defaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	// store.WatchConfigs(cmw)
	ctxFunc := func(ctx context.Context) context.Context {
		// return v1.WithUpgradeViaDefaulting(store.ToContext(ctx))
		return ctx
	}

	return defaulting.NewAdmissionController(ctx,

		// Name of the default webhook.
		// Keeping the same name as 0.10 release to avoid upgrade problems.
		fmt.Sprintf("webhook.%s.events.cloud.google.com", system.Namespace()),

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		ctxFunc,

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,

		// Name of the validation webhook.
		"validation.webhook.events.cloud.google.com",

		// The path on which to serve the webhook.
		"/validation",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			// return v1.WithUpgradeViaDefaulting(store.ToContext(ctx))
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.webhook.events.cloud.google.com",

		// The path on which to serve the webhook.
		"/config-validation",

		// The configmaps to validate.
		configmap.Constructors{
			tracingconfig.ConfigName: tracingconfig.NewTracingConfigFromConfigMap,
			// metrics.ConfigMapName():   metricsconfig.NewObservabilityConfigFromConfigMap,
			logging.ConfigMapName(): logging.NewConfigFromConfigMap,
		},
	)
}

func main() {
	// Set up a signal context with our webhook options
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: logconfig.WebhookName(),
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	sharedmain.MainWithContext(ctx, logconfig.WebhookName(),
		certificates.NewController,
		NewConfigValidationController,
		NewValidationAdmissionController,
		NewDefaultingAdmissionController,
	)
}
