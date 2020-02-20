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

package keda

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/knative-gcp/pkg/apis/pubsub/v1alpha1"
	psreconciler "github.com/google/knative-gcp/pkg/reconciler/pubsub/pullsubscription"
	"github.com/google/knative-gcp/pkg/reconciler/pubsub/pullsubscription/keda/resources"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	eventingduck "knative.dev/eventing/pkg/duck"

	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
)

type DiscoverFunc func(discovery.DiscoveryInterface, schema.GroupVersion) error

// Reconciler implements controller.Reconciler for PullSubscription resources.
type Reconciler struct {
	*psreconciler.Base

	// scaledObjectTracker is used to notify us that a Keda ScaledObject has changed so that we can reconcile.
	scaledObjectTracker eventingduck.ListableTracker

	// discoveryFn is the function used to discover whether Keda is installed or not. Needed for UTs purposes.
	discoveryFn DiscoverFunc
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	return r.Base.Reconcile(ctx, key)
}

// TODO upstream to pkg
func (r *Reconciler) ReconcileScaledObject(ctx context.Context, ra *appsv1.Deployment, src *v1alpha1.PullSubscription) error {
	// Check whether KEDA is installed, if not, error out.
	// Ideally this should be done in the webhook, thus not even allowing the creation of the object.
	if err := r.discoveryFn(r.KubeClientSet.Discovery(), resources.KedaSchemeGroupVersion); err != nil {
		if strings.Contains(err.Error(), "server does not support API version") {
			logging.FromContext(ctx).Desugar().Error("KEDA not installed, failed to check API version", zap.Any("GroupVersion", resources.KedaSchemeGroupVersion))
			return err
		}
	}

	existing, err := r.Base.GetOrCreateReceiveAdapter(ctx, ra, src)
	if err != nil {
		return err
	}
	// Given than the Deployment replicas will be controlled by Keda, we assume
	// the replica count from the existing one is the correct one.
	ra.Spec.Replicas = existing.Spec.Replicas
	if !equality.Semantic.DeepDerivative(ra.Spec, existing.Spec) {
		existing.Spec = ra.Spec
		_, err = r.KubeClientSet.AppsV1().Deployments(src.Namespace).Update(existing)
		if err != nil {
			logging.FromContext(ctx).Desugar().Error("Error updating Receive Adapter", zap.Error(err))
			return err
		}
	}
	// Now we reconcile the ScaledObject.
	gvr, _ := meta.UnsafeGuessKindToResource(resources.ScaledObjectGVK)
	scaledObjectResourceInterface := r.DynamicClientSet.Resource(gvr).Namespace(src.Namespace)
	if scaledObjectResourceInterface == nil {
		return fmt.Errorf("unable to create dynamic client for ScaledObject")
	}

	so, err := resources.MakeScaledObject(ctx, existing, src)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Error creating ScaledObject", zap.Error(err))
		return err
	}

	apiVersion, kind := resources.ScaledObjectGVK.ToAPIVersionAndKind()
	ref := tracker.Reference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       resources.GenerateScaledObjectName(src),
		Namespace:  src.Namespace,
	}
	track := r.scaledObjectTracker.TrackInNamespace(src)

	objRef := ref.ObjectReference()
	// Track changes in the ScaledObject.
	if err = track(objRef); err != nil {
		logging.FromContext(ctx).Desugar().Error("Unable to track changes to ScaledObject", zap.Error(err))
		return err
	}

	lister, err := r.scaledObjectTracker.ListerFor(objRef)
	if err != nil {
		logging.FromContext(ctx).Error("Error getting lister for ScaledObject", zap.Any("so", objRef), zap.Error(err))
		return err
	}

	_, err = lister.ByNamespace(so.GetNamespace()).Get(so.GetName())
	if err != nil {
		if apierrs.IsNotFound(err) {
			_, err = scaledObjectResourceInterface.Create(so, metav1.CreateOptions{})
			if err != nil {
				logging.FromContext(ctx).Desugar().Error("Failed to create ScaledObject", zap.Any("so", so), zap.Error(err))
				return err
			}
		} else {
			logging.FromContext(ctx).Desugar().Error("Failed to get ScaledObject", zap.Any("so", so), zap.Error(err))
			return err
		}
	}
	return nil
}
