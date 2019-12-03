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

package scheduler

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	"github.com/google/knative-gcp/pkg/reconciler/pubsub"

	schedulerinformers "github.com/google/knative-gcp/pkg/client/injection/informers/events/v1alpha1/scheduler"
	pullsubscriptioninformers "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/pullsubscription"
	topicinformers "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/topic"
)

const (
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "cloud-run-events-scheduler-source-controller"

	// receiveAdapterName is the string used as name for the receive adapter pod.
	receiveAdapterName = "scheduler.events.cloud.google.com"
)

type envConfig struct {
	// NotificationOps is the image for operating on notifications. Required.
	SchedulerJobOpsImage string `envconfig:"SCHEDULER_JOB_IMAGE" required:"true"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	pullsubscriptionInformer := pullsubscriptioninformers.Get(ctx)
	topicInformer := topicinformers.Get(ctx)
	schedulerInformer := schedulerinformers.Get(ctx)

	logger := logging.FromContext(ctx).Named(controllerAgentName)
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		logger.Fatal("Failed to process env var", zap.Error(err))
	}

	c := &Reconciler{
		SchedulerOpsImage: env.SchedulerJobOpsImage,
		PubSubBase:        pubsub.NewPubSubBase(ctx, controllerAgentName, receiveAdapterName, cmw),
		schedulerLister:   schedulerInformer.Lister(),
	}
	impl := controller.NewImpl(c, c.Logger, ReconcilerName)

	c.Logger.Info("Setting up event handlers")
	schedulerInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	topicInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Scheduler")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	pullsubscriptionInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Scheduler")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	//	c.tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))

	return impl
}
