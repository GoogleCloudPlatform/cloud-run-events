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

package brokercell

import (
	"context"
	"time"

	"github.com/google/knative-gcp/pkg/apis/messaging/v1beta1"

	channelinformer "github.com/google/knative-gcp/pkg/client/injection/informers/messaging/v1beta1/channel"

	"go.uber.org/zap"

	brokerv1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	brokerinformer "github.com/google/knative-gcp/pkg/client/injection/informers/broker/v1beta1/broker"
	triggerinformer "github.com/google/knative-gcp/pkg/client/injection/informers/broker/v1beta1/trigger"
	brokercellinformer "github.com/google/knative-gcp/pkg/client/injection/informers/intevents/v1alpha1/brokercell"
	hpainformer "github.com/google/knative-gcp/pkg/client/injection/kube/informers/autoscaling/v2beta2/horizontalpodautoscaler"
	v1alpha1brokercell "github.com/google/knative-gcp/pkg/client/injection/reconciler/intevents/v1alpha1/brokercell"
	"github.com/google/knative-gcp/pkg/logging"
	"github.com/google/knative-gcp/pkg/metrics"
	"github.com/google/knative-gcp/pkg/reconciler"
	brokerresources "github.com/google/knative-gcp/pkg/reconciler/broker/resources"
	"github.com/google/knative-gcp/pkg/reconciler/brokercell/resources"
	"github.com/google/knative-gcp/pkg/utils/authcheck"
	customresourceutil "github.com/google/knative-gcp/pkg/utils/customresource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	configmapinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/configmap"
	endpointsinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/endpoints"
	podinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/pod"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	serviceaccountinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	systemnamespacesecretinformer "knative.dev/pkg/injection/clients/namespacedkube/informers/core/v1/secret"
	"knative.dev/pkg/system"
)

const (
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "brokercell-controller"
)

type Constructor injection.ControllerConstructor

// NewConstructor creates a constructor to make a BrokerCell controller.
func NewConstructor() Constructor {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		return NewController(ctx, cmw)
	}
}

// NewController creates a Reconciler for BrokerCell and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	brokerCellInformer := brokercellinformer.Get(ctx)

	logger := logging.FromContext(ctx)

	ls := listers{
		brokerLister:         brokerinformer.Get(ctx).Lister(),
		hpaLister:            hpainformer.Get(ctx).Lister(),
		triggerLister:        triggerinformer.Get(ctx).Lister(),
		configMapLister:      configmapinformer.Get(ctx).Lister(),
		secretLister:         systemnamespacesecretinformer.Get(ctx).Lister(),
		serviceAccountLister: serviceaccountinformer.Get(ctx).Lister(),
		serviceLister:        serviceinformer.Get(ctx).Lister(),
		endpointsLister:      endpointsinformer.Get(ctx).Lister(),
		deploymentLister:     deploymentinformer.Get(ctx).Lister(),
		podLister:            podinformer.Get(ctx).Lister(),

		channelLister: channelinformer.Get(ctx).Lister(),
	}

	base := reconciler.NewBase(ctx, controllerAgentName, cmw)
	r, err := NewReconciler(base, ls)
	if err != nil {
		logger.Fatal("Failed to create BrokerCell reconciler", zap.Error(err))
	}
	impl := v1alpha1brokercell.NewImpl(ctx, r)

	var latencyReporter *metrics.BrokerCellLatencyReporter
	if r.env.InternalMetricsEnabled {
		latencyReporter, err = metrics.NewBrokerCellLatencyReporter()
		if err != nil {
			logger.Error("Failed to create latency reporter", zap.Error(err))
		}
	}

	logger.Info("Setting up event handlers.")

	brokerCellInformer.Informer().AddEventHandlerWithResyncPeriod(controller.HandleAll(impl.Enqueue), reconciler.DefaultResyncPeriod)
	brokerCellLister := brokerCellInformer.Lister()

	// Watch brokers and triggers to invoke configmap update immediately.
	brokerinformer.Get(ctx).Informer().AddEventHandler(controller.HandleAll(
		func(obj interface{}) {
			if b, ok := obj.(*brokerv1beta1.Broker); ok {
				// TODO(#866) Select the brokercell that's associated with the given broker.
				impl.EnqueueKey(types.NamespacedName{Namespace: system.Namespace(), Name: brokerresources.DefaultBrokerCellName})
				reportLatency(ctx, b, latencyReporter, "Broker", b.Name, b.Namespace)
			}
		},
	))
	triggerinformer.Get(ctx).Informer().AddEventHandler(controller.HandleAll(
		func(obj interface{}) {
			if t, ok := obj.(*brokerv1beta1.Trigger); ok {
				// TODO(#866) Select the brokercell that's associated with the given broker.
				impl.EnqueueKey(types.NamespacedName{Namespace: system.Namespace(), Name: brokerresources.DefaultBrokerCellName})
				reportLatency(ctx, t, latencyReporter, "Trigger", t.Name, t.Namespace)
			}
		},
	))

	// Watch GCP Channels and subscriptions on those channels to invoke configmap update immediately.
	channelinformer.Get(ctx).Informer().AddEventHandler(controller.HandleAll(
		func(obj interface{}) {
			if c, ok := obj.(*v1beta1.Channel); ok {
				// TODO(#866) Select the brokercell that's associated with the given broker.
				impl.EnqueueKey(types.NamespacedName{Namespace: system.Namespace(), Name: brokerresources.DefaultBrokerCellName})
				reportLatency(ctx, c, latencyReporter, "Channel", c.Name, c.Namespace)
			}
		},
	))

	// Watch data plane components created by brokercell so we can update brokercell status immediately.
	// 1. Watch deployments for ingress, fanout and retry
	deploymentinformer.Get(ctx).Informer().AddEventHandler(handleResourceUpdate(impl))
	// 2. Watch ingress endpoints
	endpointsinformer.Get(ctx).Informer().AddEventHandler(handleResourceUpdate(impl))
	// 3. Watch hpa for ingress, fanout and retry deployments
	hpainformer.Get(ctx).Informer().AddEventHandler(handleResourceUpdate(impl))
	// 4. Watch the broker targets configmap.
	configmapinformer.Get(ctx).Informer().AddEventHandler(handleResourceUpdate(impl))

	// Watch componets which are not created by brokercell, but affect broker data plane.
	// 1. Watch broker data plane's secret,
	// if the filtered secret resource changes, enqueue brokercells from the same namespace.
	systemnamespacesecretinformer.Get(ctx).Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterWithNamespace(authcheck.ControlPlaneNamespace),
		Handler:    authcheck.EnqueueBrokerCell(impl, brokerCellLister),
	})
	// 2. Watch broker data plane's k8s service account,
	// if the filtered k8s service account resource changes, enqueue brokercells from the same namespace.
	serviceaccountinformer.Get(ctx).Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterWithNamespace(authcheck.ControlPlaneNamespace),
		Handler:    authcheck.EnqueueBrokerCell(impl, brokerCellLister),
	})
	return impl
}

// handleResourceUpdate returns an event handler for resources created by brokercell such as the ingress deployment.
func handleResourceUpdate(impl *controller.Impl) cache.ResourceEventHandler {
	// Since resources created by brokercell live in the same namespace as the brokercell, we use an
	// empty namespaceLabel so that the same namespace of the given object is used to enqueue.
	namespaceLabel := ""
	// Resources created by the brokercell, including the indirectly created ingress service endpoints,
	// have such a label resources.BrokerCellLabelKey=<brokercellName>. Resources without this label
	// will be skipped by the function.
	return controller.HandleAll(impl.EnqueueLabelOfNamespaceScopedResource(namespaceLabel, resources.BrokerCellLabelKey))
}

// reportLatency estimates the time spent since the last update of the resource object and records it to the latency metric
func reportLatency(ctx context.Context, resourceObj metav1.ObjectMetaAccessor, latencyReporter *metrics.BrokerCellLatencyReporter, resourceKind, resourceName, namespace string) {
	if latencyReporter == nil {
		return
	}
	if latestUpdateTime, err := customresourceutil.RetrieveLatestUpdateTime(resourceObj); err == nil {
		if err := latencyReporter.ReportLatency(ctx, time.Now().Sub(latestUpdateTime), resourceKind, resourceName, namespace); err != nil {
			logging.FromContext(ctx).Error("Failed to report latency", zap.Error(err))
		}
	} else {
		logging.FromContext(ctx).Error("Failed to retrieve the resource update time", zap.Error(err))
	}
}

// filterWithNamespace filters object based on a namespace.
func filterWithNamespace(namespace string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		if object, ok := obj.(metav1.Object); ok {
			return namespace == object.GetNamespace()
		}
		return false
	}
}
