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

package build

import (
	"context"
	"github.com/google/knative-gcp/pkg/pubsub/adapter/converters"

	"github.com/google/knative-gcp/pkg/reconciler/pubsub"
	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	cloudbuildsourceinformers "github.com/google/knative-gcp/pkg/client/injection/informers/events/v1alpha1/cloudbuildsource"
	pullsubscriptioninformers "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/pullsubscription"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "CloudBuildSource"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "cloud-run-events-build-source-controller"

	// receiveAdapterName is the string used as name for the receive adapter pod.
	receiveAdapterName = "cloudbuildsource.events.cloud.google.com"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	pullsubscriptionInformer := pullsubscriptioninformers.Get(ctx)
	cloudbuildsourceInformer := cloudbuildsourceinformers.Get(ctx)

	r := &Reconciler{
		PubSubBase:             pubsub.NewPubSubBaseWithAdapter(ctx, controllerAgentName, receiveAdapterName, converters.CloudBuildConverter,cmw),
		buildLister:            cloudbuildsourceInformer.Lister(),
		pullsubscriptionLister: pullsubscriptionInformer.Lister(),
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	r.Logger.Info("Setting up event handlers")
	cloudbuildsourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	pullsubscriptionInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("CloudPubSubSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}

