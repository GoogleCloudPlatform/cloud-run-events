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

// Code generated by injection-gen. DO NOT EDIT.

package pullsubscription

import (
	context "context"
	json "encoding/json"
	fmt "fmt"
	reflect "reflect"

	v1beta1 "github.com/google/knative-gcp/pkg/apis/intevents/v1beta1"
	versioned "github.com/google/knative-gcp/pkg/client/clientset/versioned"
	inteventsv1beta1 "github.com/google/knative-gcp/pkg/client/listers/intevents/v1beta1"
	zap "go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	equality "k8s.io/apimachinery/pkg/api/equality"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	sets "k8s.io/apimachinery/pkg/util/sets"
	cache "k8s.io/client-go/tools/cache"
	record "k8s.io/client-go/tools/record"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"
	reconciler "knative.dev/pkg/reconciler"
)

// Interface defines the strongly typed interfaces to be implemented by a
// controller reconciling v1beta1.PullSubscription.
type Interface interface {
	// ReconcileKind implements custom logic to reconcile v1beta1.PullSubscription. Any changes
	// to the objects .Status or .Finalizers will be propagated to the stored
	// object. It is recommended that implementors do not call any update calls
	// for the Kind inside of ReconcileKind, it is the responsibility of the calling
	// controller to propagate those properties. The resource passed to ReconcileKind
	// will always have an empty deletion timestamp.
	ReconcileKind(ctx context.Context, o *v1beta1.PullSubscription) reconciler.Event
}

// Finalizer defines the strongly typed interfaces to be implemented by a
// controller finalizing v1beta1.PullSubscription.
type Finalizer interface {
	// FinalizeKind implements custom logic to finalize v1beta1.PullSubscription. Any changes
	// to the objects .Status or .Finalizers will be ignored. Returning a nil or
	// Normal type reconciler.Event will allow the finalizer to be deleted on
	// the resource. The resource passed to FinalizeKind will always have a set
	// deletion timestamp.
	FinalizeKind(ctx context.Context, o *v1beta1.PullSubscription) reconciler.Event
}

// ReadOnlyInterface defines the strongly typed interfaces to be implemented by a
// controller reconciling v1beta1.PullSubscription if they want to process resources for which
// they are not the leader.
type ReadOnlyInterface interface {
	// ObserveKind implements logic to observe v1beta1.PullSubscription.
	// This method should not write to the API.
	ObserveKind(ctx context.Context, o *v1beta1.PullSubscription) reconciler.Event
}

// ReadOnlyFinalizer defines the strongly typed interfaces to be implemented by a
// controller finalizing v1beta1.PullSubscription if they want to process tombstoned resources
// even when they are not the leader.  Due to the nature of how finalizers are handled
// there are no guarantees that this will be called.
type ReadOnlyFinalizer interface {
	// ObserveFinalizeKind implements custom logic to observe the final state of v1beta1.PullSubscription.
	// This method should not write to the API.
	ObserveFinalizeKind(ctx context.Context, o *v1beta1.PullSubscription) reconciler.Event
}

// reconcilerImpl implements controller.Reconciler for v1beta1.PullSubscription resources.
type reconcilerImpl struct {
	// LeaderAwareFuncs is inlined to help us implement reconciler.LeaderAware
	reconciler.LeaderAwareFuncs

	// Client is used to write back status updates.
	Client versioned.Interface

	// Listers index properties about resources
	Lister inteventsv1beta1.PullSubscriptionLister

	// Recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	Recorder record.EventRecorder

	// configStore allows for decorating a context with config maps.
	// +optional
	configStore reconciler.ConfigStore

	// reconciler is the implementation of the business logic of the resource.
	reconciler Interface

	// finalizerName is the name of the finalizer to reconcile.
	finalizerName string
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*reconcilerImpl)(nil)

// Check that our generated Reconciler is always LeaderAware.
var _ reconciler.LeaderAware = (*reconcilerImpl)(nil)

func NewReconciler(ctx context.Context, logger *zap.SugaredLogger, client versioned.Interface, lister inteventsv1beta1.PullSubscriptionLister, recorder record.EventRecorder, r Interface, options ...controller.Options) controller.Reconciler {
	// Check the options function input. It should be 0 or 1.
	if len(options) > 1 {
		logger.Fatalf("up to one options struct is supported, found %d", len(options))
	}

	// Fail fast when users inadvertently implement the other LeaderAware interface.
	// For the typed reconcilers, Promote shouldn't take any arguments.
	if _, ok := r.(reconciler.LeaderAware); ok {
		logger.Fatalf("%T implements the incorrect LeaderAware interface.  Promote() should not take an argument as genreconciler handles the enqueuing automatically.", r)
	}
	// TODO: Consider validating when folks implement ReadOnlyFinalizer, but not Finalizer.

	rec := &reconcilerImpl{
		LeaderAwareFuncs: reconciler.LeaderAwareFuncs{
			PromoteFunc: func(bkt reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
				all, err := lister.List(labels.Everything())
				if err != nil {
					return err
				}
				for _, elt := range all {
					// TODO: Consider letting users specify a filter in options.
					enq(bkt, types.NamespacedName{
						Namespace: elt.GetNamespace(),
						Name:      elt.GetName(),
					})
				}
				return nil
			},
		},
		Client:        client,
		Lister:        lister,
		Recorder:      recorder,
		reconciler:    r,
		finalizerName: defaultFinalizerName,
	}

	for _, opts := range options {
		if opts.ConfigStore != nil {
			rec.configStore = opts.ConfigStore
		}
		if opts.FinalizerName != "" {
			rec.finalizerName = opts.FinalizerName
		}
	}

	return rec
}

// Reconcile implements controller.Reconciler
func (r *reconcilerImpl) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}
	// Establish whether we are the leader for use below.
	isLeader := r.IsLeaderFor(types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	})
	roi, isROI := r.reconciler.(ReadOnlyInterface)
	rof, isROF := r.reconciler.(ReadOnlyFinalizer)
	if !isLeader && !isROI && !isROF {
		// If we are not the leader, and we don't implement either ReadOnly
		// interface, then take a fast-path out.
		return nil
	}

	// If configStore is set, attach the frozen configuration to the context.
	if r.configStore != nil {
		ctx = r.configStore.ToContext(ctx)
	}

	// Add the recorder to context.
	ctx = controller.WithEventRecorder(ctx, r.Recorder)

	// Get the resource with this namespace/name.

	getter := r.Lister.PullSubscriptions(namespace)

	original, err := getter.Get(name)

	if errors.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logger.Debugf("resource %q no longer exists", key)
		return nil
	} else if err != nil {
		return err
	}

	// Don't modify the informers copy.
	resource := original.DeepCopy()

	var reconcileEvent reconciler.Event
	if resource.GetDeletionTimestamp().IsZero() {
		if isLeader {
			// Append the target method to the logger.
			logger = logger.With(zap.String("targetMethod", "ReconcileKind"))

			// Set and update the finalizer on resource if r.reconciler
			// implements Finalizer.
			if resource, err = r.setFinalizerIfFinalizer(ctx, resource); err != nil {
				return fmt.Errorf("failed to set finalizers: %w", err)
			}

			reconciler.PreProcessReconcile(ctx, resource)

			// Reconcile this copy of the resource and then write back any status
			// updates regardless of whether the reconciliation errored out.
			reconcileEvent = r.reconciler.ReconcileKind(ctx, resource)

			reconciler.PostProcessReconcile(ctx, resource, original)

		} else if isROI {
			// Append the target method to the logger.
			logger = logger.With(zap.String("targetMethod", "ObserveKind"))

			// Observe any changes to this resource, since we are not the leader.
			reconcileEvent = roi.ObserveKind(ctx, resource)
		}
	} else if fin, ok := r.reconciler.(Finalizer); isLeader && ok {
		// Append the target method to the logger.
		logger = logger.With(zap.String("targetMethod", "FinalizeKind"))

		// For finalizing reconcilers, if this resource being marked for deletion
		// and reconciled cleanly (nil or normal event), remove the finalizer.
		reconcileEvent = fin.FinalizeKind(ctx, resource)
		if resource, err = r.clearFinalizer(ctx, resource, reconcileEvent); err != nil {
			return fmt.Errorf("failed to clear finalizers: %w", err)
		}
	} else if !isLeader && isROF {
		// Append the target method to the logger.
		logger = logger.With(zap.String("targetMethod", "ObserveFinalizeKind"))

		// For finalizing reconcilers, just observe when we aren't the leader.
		reconcileEvent = rof.ObserveFinalizeKind(ctx, resource)
	}

	// Synchronize the status.
	if equality.Semantic.DeepEqual(original.Status, resource.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the injectionInformer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if !isLeader {
		logger.Warn("Saw status changes when we aren't the leader!")
		// TODO: Consider logging the diff at Debug?
	} else if err = r.updateStatus(original, resource); err != nil {
		logger.Warnw("Failed to update resource status", zap.Error(err))
		r.Recorder.Eventf(resource, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for %q: %v", resource.Name, err)
		return err
	}

	// Report the reconciler event, if any.
	if reconcileEvent != nil {
		var event *reconciler.ReconcilerEvent
		if reconciler.EventAs(reconcileEvent, &event) {
			logger.Infow("Returned an event", zap.Any("event", reconcileEvent))
			r.Recorder.Eventf(resource, event.EventType, event.Reason, event.Format, event.Args...)

			// the event was wrapped inside an error, consider the reconciliation as failed
			if _, isEvent := reconcileEvent.(*reconciler.ReconcilerEvent); !isEvent {
				return reconcileEvent
			}
			return nil
		}

		logger.Errorw("Returned an error", zap.Error(reconcileEvent))
		r.Recorder.Event(resource, v1.EventTypeWarning, "InternalError", reconcileEvent.Error())
		return reconcileEvent
	}

	return nil
}

func (r *reconcilerImpl) updateStatus(existing *v1beta1.PullSubscription, desired *v1beta1.PullSubscription) error {
	existing = existing.DeepCopy()
	return reconciler.RetryUpdateConflicts(func(attempts int) (err error) {
		// The first iteration tries to use the injectionInformer's state, subsequent attempts fetch the latest state via API.
		if attempts > 0 {

			getter := r.Client.InternalV1beta1().PullSubscriptions(desired.Namespace)

			existing, err = getter.Get(desired.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
		}

		// If there's nothing to update, just return.
		if reflect.DeepEqual(existing.Status, desired.Status) {
			return nil
		}

		existing.Status = desired.Status

		updater := r.Client.InternalV1beta1().PullSubscriptions(existing.Namespace)

		_, err = updater.UpdateStatus(existing)
		return err
	})
}

// updateFinalizersFiltered will update the Finalizers of the resource.
// TODO: this method could be generic and sync all finalizers. For now it only
// updates defaultFinalizerName or its override.
func (r *reconcilerImpl) updateFinalizersFiltered(ctx context.Context, resource *v1beta1.PullSubscription) (*v1beta1.PullSubscription, error) {

	getter := r.Lister.PullSubscriptions(resource.Namespace)

	actual, err := getter.Get(resource.Name)
	if err != nil {
		return resource, err
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	var finalizers []string

	// If there's nothing to update, just return.
	existingFinalizers := sets.NewString(existing.Finalizers...)
	desiredFinalizers := sets.NewString(resource.Finalizers...)

	if desiredFinalizers.Has(r.finalizerName) {
		if existingFinalizers.Has(r.finalizerName) {
			// Nothing to do.
			return resource, nil
		}
		// Add the finalizer.
		finalizers = append(existing.Finalizers, r.finalizerName)
	} else {
		if !existingFinalizers.Has(r.finalizerName) {
			// Nothing to do.
			return resource, nil
		}
		// Remove the finalizer.
		existingFinalizers.Delete(r.finalizerName)
		finalizers = existingFinalizers.List()
	}

	mergePatch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"finalizers":      finalizers,
			"resourceVersion": existing.ResourceVersion,
		},
	}

	patch, err := json.Marshal(mergePatch)
	if err != nil {
		return resource, err
	}

	patcher := r.Client.InternalV1beta1().PullSubscriptions(resource.Namespace)

	resourceName := resource.Name
	resource, err = patcher.Patch(resourceName, types.MergePatchType, patch)
	if err != nil {
		r.Recorder.Eventf(resource, v1.EventTypeWarning, "FinalizerUpdateFailed",
			"Failed to update finalizers for %q: %v", resourceName, err)
	} else {
		r.Recorder.Eventf(resource, v1.EventTypeNormal, "FinalizerUpdate",
			"Updated %q finalizers", resource.GetName())
	}
	return resource, err
}

func (r *reconcilerImpl) setFinalizerIfFinalizer(ctx context.Context, resource *v1beta1.PullSubscription) (*v1beta1.PullSubscription, error) {
	if _, ok := r.reconciler.(Finalizer); !ok {
		return resource, nil
	}

	finalizers := sets.NewString(resource.Finalizers...)

	// If this resource is not being deleted, mark the finalizer.
	if resource.GetDeletionTimestamp().IsZero() {
		finalizers.Insert(r.finalizerName)
	}

	resource.Finalizers = finalizers.List()

	// Synchronize the finalizers filtered by r.finalizerName.
	return r.updateFinalizersFiltered(ctx, resource)
}

func (r *reconcilerImpl) clearFinalizer(ctx context.Context, resource *v1beta1.PullSubscription, reconcileEvent reconciler.Event) (*v1beta1.PullSubscription, error) {
	if _, ok := r.reconciler.(Finalizer); !ok {
		return resource, nil
	}
	if resource.GetDeletionTimestamp().IsZero() {
		return resource, nil
	}

	finalizers := sets.NewString(resource.Finalizers...)

	if reconcileEvent != nil {
		var event *reconciler.ReconcilerEvent
		if reconciler.EventAs(reconcileEvent, &event) {
			if event.EventType == v1.EventTypeNormal {
				finalizers.Delete(r.finalizerName)
			}
		}
	} else {
		finalizers.Delete(r.finalizerName)
	}

	resource.Finalizers = finalizers.List()

	// Synchronize the finalizers filtered by r.finalizerName.
	return r.updateFinalizersFiltered(ctx, resource)
}
