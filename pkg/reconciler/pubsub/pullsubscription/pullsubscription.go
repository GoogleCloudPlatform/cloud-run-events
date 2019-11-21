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

package pullsubscription

import (
	"context"
	"encoding/json"
	"time"

	duckv1 "knative.dev/pkg/apis/duck/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/resolver"

	tracingconfig "knative.dev/pkg/tracing/config"

	"knative.dev/pkg/metrics"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/google/go-cmp/cmp"
	"github.com/google/knative-gcp/pkg/apis/pubsub/v1alpha1"
	listers "github.com/google/knative-gcp/pkg/client/listers/pubsub/v1alpha1"
	ops "github.com/google/knative-gcp/pkg/operations"
	pubsubOps "github.com/google/knative-gcp/pkg/operations/pubsub"
	"github.com/google/knative-gcp/pkg/reconciler/events/pubsub"
	"github.com/google/knative-gcp/pkg/reconciler/pubsub/pullsubscription/resources"
	"github.com/google/knative-gcp/pkg/tracing"
	"go.uber.org/zap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const (
	// ReconcilerName is the name of the reconciler
	ReconcilerName = "PullSubscriptions"

	// Component names for metrics.
	sourceComponent  = "source"
	channelComponent = "channel"

	finalizerName = controllerAgentName

	// Custom secret finalizer requires at least one slash
	secretFinalizerName = controllerAgentName + "/secret"
)

// Reconciler implements controller.Reconciler for PullSubscription resources.
type Reconciler struct {
	*pubsub.PubSubBase

	// Listers index properties about resources.
	deploymentLister appsv1listers.DeploymentLister
	sourceLister     listers.PullSubscriptionLister

	uriResolver *resolver.URIResolver

	receiveAdapterImage string

	loggingConfig *logging.Config
	metricsConfig *metrics.ExporterOptions
	tracingConfig *tracingconfig.Config

	//	eventTypeReconciler eventtype.Reconciler // TODO: event types.
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the PullSubscription resource
// with the current status of the resource.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Invalid resource key")
		return nil
	}
	// Get the PullSubscription resource with this namespace/name
	original, err := r.sourceLister.PullSubscriptions(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logging.FromContext(ctx).Desugar().Error("PullSubscription in work queue no longer exists")
		return nil
	} else if err != nil {
		return err
	}

	// Don't modify the informers copy
	source := original.DeepCopy()

	// Reconcile this copy of the source and then write back any status
	// updates regardless of whether the reconciliation errored out.
	var reconcileErr = r.reconcile(ctx, source)

	// If no error is returned, mark the observed generation.
	// This has to be done before updateStatus is called.
	if reconcileErr == nil {
		source.Status.ObservedGeneration = source.Generation
	}

	if equality.Semantic.DeepEqual(original.Finalizers, source.Finalizers) {
		// If we didn't change finalizers then don't call updateFinalizers.

	} else if _, updated, fErr := r.updateFinalizers(ctx, source); fErr != nil {
		logging.FromContext(ctx).Desugar().Warn("Failed to update PullSubscription finalizers", zap.Error(fErr))
		r.Recorder.Eventf(source, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update finalizers for PullSubscription %q: %v", source.Name, fErr)
		return fErr
	} else if updated {
		// There was a difference and updateFinalizers said it updated and did not return an error.
		r.Recorder.Eventf(source, corev1.EventTypeNormal, "Updated", "Updated PullSubscription %q finalizers", source.GetName())
	}

	if equality.Semantic.DeepEqual(original.Status, source.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if _, uErr := r.updateStatus(ctx, source); uErr != nil {
		logging.FromContext(ctx).Desugar().Warn("Failed to update source status", zap.Error(uErr))
		r.Recorder.Eventf(source, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for PullSubscription %q: %v", source.Name, uErr)
		return uErr
	} else if reconcileErr == nil {
		// There was a difference and updateStatus did not return an error.
		r.Recorder.Eventf(source, corev1.EventTypeNormal, "Updated", "Updated PullSubscription %q", source.GetName())
	}
	if reconcileErr != nil {
		r.Recorder.Event(source, corev1.EventTypeWarning, "InternalError", reconcileErr.Error())
	}

	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, ps *v1alpha1.PullSubscription) error {
	ctx = logging.WithLogger(ctx, r.Logger.With(zap.Any("pullsubscription", ps)))

	ps.Status.ObservedGeneration = ps.Generation
	ps.Status.InitializeConditions()

	if ps.GetDeletionTimestamp() != nil {
		logging.FromContext(ctx).Desugar().Debug("Deleting Pub/Sub subscription")
		if err := r.deleteSubscription(ctx, ps); err != nil {
			ps.Status.MarkNoSubscription("SubscriptionDeleteFailed", "Failed to delete Pub/Sub subscription: %s", err.Error())
			logging.FromContext(ctx).Desugar().Error("Failed to delete Pub/Sub subscription", zap.Error(err))
			return err
		}
		ps.Status.MarkNoSubscription("SubscriptionDeleted", "Successfully deleted Pub/Sub subscription %q", ps.Status.SubscriptionID)
		ps.Status.SubscriptionID = ""
		removeFinalizer(ps)
		return nil
	}

	// Sink is required.
	sinkURI, err := r.resolveDestination(ctx, ps.Spec.Sink, ps)
	if err != nil {
		ps.Status.MarkNoSink("InvalidSink", err.Error())
		return err
	} else {
		ps.Status.MarkSink(sinkURI)
	}

	ps.Status.SubscriptionID = resources.GenerateSubscriptionName(ps)

	state, err := r.EnsureSubscriptionCreated(ctx, ps, *ps.Spec.Secret, ps.Spec.Project, ps.Spec.Topic,
		ps.Status.SubscriptionID, ps.Spec.GetAckDeadline(), ps.Spec.RetainAckedMessages, ps.Spec.GetRetentionDuration())
	switch state {

	case ops.OpsJobCompleteSuccessful:
		ps.Status.MarkSubscribed()

	case ops.OpsJobCreateFailed, ops.OpsJobCompleteFailed:
		logging.FromContext(ctx).Desugar().Error("Failed to create subscription.", zap.Any("state", state), zap.Error(err))

		msg := "unknown"
		if err != nil {
			msg = err.Error()
		}
		ps.Status.MarkNoSubscription(
			"CreateFailed",
			"Failed to create Subscription: %q.",
			msg)
		return err
	}

	_, err = r.createOrUpdateReceiveAdapter(ctx, ps)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Unable to create the receive adapter", zap.Error(err))
		return err
	}
	ps.Status.MarkDeployed()

	return nil
}

func subscriptionExists(sub *v1alpha1.PullSubscription) bool {
	for _, c := range sub.Status.Conditions {
		if c.Type == v1alpha1.PullSubscriptionConditionSubscribed && !c.IsFalse() {
			return true
		}
	}
	return false
}

func (r *Reconciler) deleteSubscription(ctx context.Context, ps v1alpha1.PullSubscription) error {
	// TODO nacho
	return nil
}

func (r *Reconciler) resolveDestination(ctx context.Context, destination duckv1.Destination, ps *v1alpha1.PullSubscription) (string, error) {
	dest := duckv1beta1.Destination{
		Ref: destination.GetRef(),
		URI: destination.URI,
	}
	if dest.Ref != nil {
		dest.Ref.Namespace = ps.Namespace
	}
	return r.uriResolver.URIFromDestination(dest, ps)
}

func (r *Reconciler) updateStatus(ctx context.Context, desired *v1alpha1.PullSubscription) (*v1alpha1.PullSubscription, error) {
	source, err := r.sourceLister.PullSubscriptions(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if equality.Semantic.DeepEqual(source.Status, desired.Status) {
		return source, nil
	}
	becomesReady := desired.Status.IsReady() && !source.Status.IsReady()
	// Don't modify the informers copy.
	existing := source.DeepCopy()
	existing.Status = desired.Status

	src, err := r.RunClientSet.PubsubV1alpha1().PullSubscriptions(desired.Namespace).UpdateStatus(existing)
	if err == nil && becomesReady {
		duration := time.Since(src.ObjectMeta.CreationTimestamp.Time)
		r.Logger.Infof("PullSubscription %q became ready after %v", source.Name, duration)

		if err := r.StatsReporter.ReportReady("PullSubscription", source.Namespace, source.Name, duration); err != nil {
			logging.FromContext(ctx).Infof("failed to record ready for PullSubscription, %v", err)
		}
	}

	return src, err
}

// updateSecretFinalizer adds or deletes the finalizer on the secret used by the PullSubscription.
func (r *Reconciler) updateSecretFinalizer(ctx context.Context, desired *v1alpha1.PullSubscription, ensureFinalizer bool) error {
	psl, err := r.sourceLister.PullSubscriptions(desired.Namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	// Only delete the finalizer if this PullSubscription is the last one
	// references the Secret.
	if !ensureFinalizer && !(len(psl) == 1 && psl[0].Name == desired.Name) {
		return nil
	}

	secret, err := r.KubeClientSet.CoreV1().Secrets(desired.Namespace).Get(desired.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	existing := secret.DeepCopy()
	existingFinalizers := sets.NewString(existing.Finalizers...)
	hasFinalizer := existingFinalizers.Has(secretFinalizerName)

	if ensureFinalizer == hasFinalizer {
		return nil
	}

	var desiredFinalizers []string
	if ensureFinalizer {
		desiredFinalizers = append(existing.Finalizers, secretFinalizerName)
	} else {
		existingFinalizers.Delete(secretFinalizerName)
		desiredFinalizers = existingFinalizers.List()
	}

	mergePatch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"finalizers":      desiredFinalizers,
			"resourceVersion": existing.ResourceVersion,
		},
	}

	patch, err := json.Marshal(mergePatch)
	if err != nil {
		return err
	}

	_, err = r.KubeClientSet.CoreV1().Secrets(existing.GetNamespace()).Patch(existing.GetName(), types.MergePatchType, patch)
	if err != nil {
		logging.FromContext(ctx).Errorf("Failed to update PullSubscription Secret's finalizers", zap.Error(err))
	}
	return err
}

// updateFinalizers is a generic method for future compatibility with a
// reconciler SDK.
func (r *Reconciler) updateFinalizers(ctx context.Context, desired *v1alpha1.PullSubscription) (*v1alpha1.PullSubscription, bool, error) {
	source, err := r.sourceLister.PullSubscriptions(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, false, err
	}

	// Don't modify the informers copy.
	existing := source.DeepCopy()

	var finalizers []string

	// If there's nothing to update, just return.
	existingFinalizers := sets.NewString(existing.Finalizers...)
	desiredFinalizers := sets.NewString(desired.Finalizers...)

	if desiredFinalizers.Has(finalizerName) {
		if existingFinalizers.Has(finalizerName) {
			// Nothing to do.
			return desired, false, nil
		}
		// Add the finalizer.
		finalizers = append(existing.Finalizers, finalizerName)
	} else {
		if !existingFinalizers.Has(finalizerName) {
			// Nothing to do.
			return desired, false, nil
		}
		// Remove the finalizer.
		existingFinalizers.Delete(finalizerName)
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
		return desired, false, err
	}

	update, err := r.RunClientSet.PubsubV1alpha1().PullSubscriptions(existing.Namespace).Patch(existing.Name, types.MergePatchType, patch)
	return update, true, err
}

func addFinalizer(s *v1alpha1.PullSubscription) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Insert(finalizerName)
	s.Finalizers = finalizers.List()
}

func removeFinalizer(s *v1alpha1.PullSubscription) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Delete(finalizerName)
	s.Finalizers = finalizers.List()
}

func (r *Reconciler) createOrUpdateReceiveAdapter(ctx context.Context, src *v1alpha1.PullSubscription) (*appsv1.Deployment, error) {
	existing, err := r.getReceiveAdapter(ctx, src)
	if err != nil && !apierrors.IsNotFound(err) {
		logging.FromContext(ctx).Error("Unable to get an existing receive adapter", zap.Error(err))
		return nil, err
	}

	loggingConfig, err := logging.LoggingConfigToJson(r.loggingConfig)
	if err != nil {
		logging.FromContext(ctx).Error("Error serializing existing logging config", zap.Error(err))
	}

	if r.metricsConfig != nil {
		component := sourceComponent
		// Set the metric component based on the channel label.
		if _, ok := src.Labels["events.cloud.google.com/channel"]; ok {
			component = channelComponent
		}
		r.metricsConfig.Component = component
	}

	metricsConfig, err := metrics.MetricsOptionsToJson(r.metricsConfig)
	if err != nil {
		logging.FromContext(ctx).Errorw("Error serializing metrics config", zap.Error(err))
	}

	tracingConfig, err := tracing.ConfigToJSON(r.tracingConfig)
	if err != nil {
		logging.FromContext(ctx).Errorw("Error serializing tracing config", zap.Error(err))
	}

	desired := resources.MakeReceiveAdapter(ctx, &resources.ReceiveAdapterArgs{
		Image:          r.receiveAdapterImage,
		Source:         src,
		Labels:         resources.GetLabels(controllerAgentName, src.Name),
		SubscriptionID: src.Status.SubscriptionID,
		SinkURI:        src.Status.SinkURI,
		LoggingConfig:  loggingConfig,
		MetricsConfig:  metricsConfig,
		TracingConfig:  tracingConfig,
	})

	if existing == nil {
		ra, err := r.KubeClientSet.AppsV1().Deployments(src.Namespace).Create(desired)
		logging.FromContext(ctx).Desugar().Info("Receive Adapter created.", zap.Error(err), zap.Any("receiveAdapter", ra))
		return ra, err
	}
	if diff := cmp.Diff(desired.Spec, existing.Spec); diff != "" {
		existing.Spec = desired.Spec
		ra, err := r.KubeClientSet.AppsV1().Deployments(src.Namespace).Update(existing)
		logging.FromContext(ctx).Desugar().Info("Receive Adapter updated.",
			zap.Error(err), zap.Any("receiveAdapter", ra), zap.String("diff", diff))
		return ra, err
	}
	logging.FromContext(ctx).Desugar().Info("Reusing existing Receive Adapter", zap.Any("receiveAdapter", existing))
	return existing, nil
}

func (r *Reconciler) getReceiveAdapter(ctx context.Context, src *v1alpha1.PullSubscription) (*appsv1.Deployment, error) {
	dl, err := r.KubeClientSet.AppsV1().Deployments(src.Namespace).List(metav1.ListOptions{
		LabelSelector: resources.GetLabelSelector(controllerAgentName, src.Name).String(),
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
	})

	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Unable to list deployments: %v", zap.Error(err))
		return nil, err
	}
	for _, dep := range dl.Items {
		if metav1.IsControlledBy(&dep, src) {
			return &dep, nil
		}
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func (r *Reconciler) UpdateFromLoggingConfigMap(cfg *corev1.ConfigMap) {
	if cfg != nil {
		delete(cfg.Data, "_example")
	}

	logcfg, err := logging.NewConfigFromConfigMap(cfg)
	if err != nil {
		r.Logger.Warnw("Failed to create logging config from configmap", zap.String("cfg.Name", cfg.Name))
		return
	}
	r.loggingConfig = logcfg
	r.Logger.Debugw("Update from logging ConfigMap", zap.Any("loggingCfg", cfg))
	// TODO: requeue all pullsubscriptions
}

func (r *Reconciler) UpdateFromMetricsConfigMap(cfg *corev1.ConfigMap) {
	if cfg != nil {
		delete(cfg.Data, "_example")
	}

	// Cannot set the component here as we don't know if its a source or a channel.
	// Will set that up dynamically before creating the receive adapter.
	// Won't be able to requeue the PullSubscriptions.
	r.metricsConfig = &metrics.ExporterOptions{
		Domain:    metrics.Domain(),
		ConfigMap: cfg.Data,
	}
	r.Logger.Debugw("Update from metrics ConfigMap", zap.Any("metricsCfg", cfg))
}

func (r *Reconciler) UpdateFromTracingConfigMap(cfg *corev1.ConfigMap) {
	if cfg == nil {
		r.Logger.Error("Tracing ConfigMap is nil")
		return
	}
	delete(cfg.Data, "_example")

	tracingCfg, err := tracingconfig.NewTracingConfigFromConfigMap(cfg)
	if err != nil {
		r.Logger.Warnw("Failed to create tracing config from configmap", zap.String("cfg.Name", cfg.Name))
		return
	}
	r.tracingConfig = tracingCfg
	r.Logger.Debugw("Updated Tracing config", zap.Any("tracingCfg", r.tracingConfig))
	// TODO: requeue all PullSubscriptions.
}
