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
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	listers "github.com/google/knative-gcp/pkg/client/listers/events/v1alpha1"
	gscheduler "github.com/google/knative-gcp/pkg/gclient/scheduler"
	"github.com/google/knative-gcp/pkg/pubsub/adapter/converters"
	"github.com/google/knative-gcp/pkg/reconciler/events/scheduler/resources"
	"github.com/google/knative-gcp/pkg/reconciler/pubsub"
	psresources "github.com/google/knative-gcp/pkg/reconciler/pubsub/resources"
	"github.com/google/knative-gcp/pkg/utils"
	schedulerpb "google.golang.org/genproto/googleapis/cloud/scheduler/v1"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	finalizerName = controllerAgentName

	resourceGroup = "cloudschedulersources.events.cloud.google.com"
)

// Reconciler is the controller implementation for Google Cloud Scheduler Jobs.
type Reconciler struct {
	*pubsub.PubSubBase

	// schedulerLister for reading schedulers.
	schedulerLister listers.CloudSchedulerSourceLister

	createClientFn gscheduler.CreateFn

	// serviceAccountLister for reading serviceAccounts.
	serviceAccountLister corev1listers.ServiceAccountLister
}

// Check that we implement the controller.Reconciler interface.
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Scheduler resource
// with the current status of the resource.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Invalid resource key")
		return nil
	}

	// Get the CloudSchedulerSource resource with this namespace/name
	original, err := r.schedulerLister.CloudSchedulerSources(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The CloudSchedulerSource resource may no longer exist, in which case we stop processing.
		logging.FromContext(ctx).Desugar().Error("CloudSchedulerSource in work queue no longer exists")
		return nil
	} else if err != nil {
		return err
	}

	// Don't modify the informers copy
	scheduler := original.DeepCopy()

	reconcileErr := r.reconcile(ctx, scheduler)

	// If no error is returned, mark the observed generation.
	if reconcileErr == nil {
		scheduler.Status.ObservedGeneration = scheduler.Generation
	}

	if equality.Semantic.DeepEqual(original.Finalizers, scheduler.Finalizers) {
		// If we didn't change finalizers then don't call updateFinalizers.

	} else if _, updated, fErr := r.updateFinalizers(ctx, scheduler); fErr != nil {
		logging.FromContext(ctx).Desugar().Warn("Failed to update CloudSchedulerSource finalizers", zap.Error(fErr))
		r.Recorder.Eventf(scheduler, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update finalizers for CloudSchedulerSource %q: %v", scheduler.Name, fErr)
		return fErr
	} else if updated {
		// There was a difference and updateFinalizers said it updated and did not return an error.
		r.Recorder.Eventf(scheduler, corev1.EventTypeNormal, "Updated", "Updated CloudSchedulerSource %q finalizers", scheduler.Name)
	}

	if equality.Semantic.DeepEqual(original.Status, scheduler.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.

	} else if uErr := r.updateStatus(ctx, original, scheduler); uErr != nil {
		logging.FromContext(ctx).Desugar().Warn("Failed to update CloudSchedulerSource status", zap.Error(uErr))
		r.Recorder.Eventf(scheduler, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for CloudSchedulerSource %q: %v", scheduler.Name, uErr)
		return uErr
	} else if reconcileErr == nil {
		// There was a difference and updateStatus did not return an error.
		r.Recorder.Eventf(scheduler, corev1.EventTypeNormal, "Updated", "Updated CloudSchedulerSource %q", scheduler.Name)
	}
	if reconcileErr != nil {
		r.Recorder.Event(scheduler, corev1.EventTypeWarning, "InternalError", reconcileErr.Error())
	}
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, scheduler *v1alpha1.CloudSchedulerSource) error {
	ctx = logging.WithLogger(ctx, r.Logger.With(zap.Any("scheduler", scheduler)))

	scheduler.Status.InitializeConditions()

	// If GCP ServiceAccount is provided, get the corresponding k8s ServiceAccount.
	// kServiceAccount will be nil if GCP ServiceAccount is not provided or there is no corresponding k8s ServiceAccount.
	var kServiceAccount *corev1.ServiceAccount
	if scheduler.Spec.ServiceAccount != nil {
		kServiceAccountName := psresources.GenerateServiceAccountName(scheduler.Spec.ServiceAccount)
		ksa, err := r.serviceAccountLister.ServiceAccounts(scheduler.Namespace).Get(kServiceAccountName)
		if err != nil {
			if !apierrs.IsNotFound(err) {
				logging.FromContext(ctx).Desugar().Error("Failed to get k8s service account", zap.Error(err))
				return err
			}
		} else {
			kServiceAccount = ksa
		}
	}

	// See if the source has been deleted.
	if scheduler.DeletionTimestamp != nil {
		// If k8s ServiceAccount exists and it only has one ownerReference, remove the corresponding GCP ServiceAccount iam policy binding.
		// No need to delete k8s ServiceAccount, it will be automatically handled by k8s Garbage Collection.
		if kServiceAccount != nil && len(kServiceAccount.OwnerReferences) == 1 {
			logging.FromContext(ctx).Desugar().Debug("Removing iam policy binding.")
			psresources.RemoveIamPolicyBinding(ctx, *scheduler.Spec.ServiceAccount, kServiceAccount)
		}

		logging.FromContext(ctx).Desugar().Debug("Deleting CloudSchedulerSource job")
		if err := r.deleteJob(ctx, scheduler); err != nil {
			scheduler.Status.MarkJobNotReady("JobDeleteFailed", "Failed to delete CloudSchedulerSource job: %s", err.Error())
			return err
		}
		scheduler.Status.MarkJobNotReady("JobDeleted", "Successfully deleted CloudSchedulerSource job: %s", scheduler.Status.JobName)

		if err := r.PubSubBase.DeletePubSub(ctx, scheduler); err != nil {
			return err
		}

		// Only set the jobName to empty after we successfully deleted the PubSub resources.
		// Otherwise, we may leak them.
		scheduler.Status.JobName = ""
		removeFinalizer(scheduler)
		return nil
	}

	// Ensure that there's finalizer there, since we're about to attempt to
	// change external state with the topic, so we need to clean it up.
	addFinalizer(scheduler)

	// If GCP ServiceAccount is provided, configure workload identity.
	if scheduler.Spec.ServiceAccount != nil {
		gServiceAccount := *scheduler.Spec.ServiceAccount
		// Create corresponding k8s ServiceAccount if doesn't exist, and add ownerReference to it.
		if err := r.PubSubBase.CreateServiceAccount(ctx, scheduler, kServiceAccount); err != nil {
			return err
		}
		// Add iam policy binding to GCP ServiceAccount.
		if err := psresources.AddIamPolicyBinding(ctx, gServiceAccount, kServiceAccount); err != nil {
			return err
		}
	}

	topic := resources.GenerateTopicName(scheduler)
	_, _, err := r.PubSubBase.ReconcilePubSub(ctx, scheduler, topic, resourceGroup)
	if err != nil {
		return err
	}

	jobName := resources.GenerateJobName(scheduler)
	err = r.reconcileJob(ctx, scheduler, topic, jobName)
	if err != nil {
		scheduler.Status.MarkJobNotReady("JobReconcileFailed", "Failed to reconcile CloudSchedulerSource job: %s", err.Error())
		return err
	}
	scheduler.Status.MarkJobReady(jobName)
	return nil
}

func (r *Reconciler) reconcileJob(ctx context.Context, scheduler *v1alpha1.CloudSchedulerSource, topic, jobName string) error {
	if scheduler.Status.ProjectID == "" {
		projectID, err := utils.ProjectID(scheduler.Spec.Project)
		if err != nil {
			logging.FromContext(ctx).Desugar().Error("Failed to find project id", zap.Error(err))
			return err
		}
		// Set the projectID in the status.
		scheduler.Status.ProjectID = projectID
	}

	client, err := r.createClientFn(ctx)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Failed to create CloudSchedulerSource client", zap.Error(err))
		return err
	}
	defer client.Close()

	// Check if the job exists.
	_, err = client.GetJob(ctx, &schedulerpb.GetJobRequest{Name: jobName})
	if err != nil {
		if st, ok := gstatus.FromError(err); !ok {
			logging.FromContext(ctx).Desugar().Error("Failed from CloudSchedulerSource client while retrieving CloudSchedulerSource job", zap.String("jobName", jobName), zap.Error(err))
			return err
		} else if st.Code() == codes.NotFound {
			// Create the job as it does not exist. For creation, we need a parent, extract it from the jobName.
			parent := resources.ExtractParentName(jobName)
			// Add our own converter type, jobName, and schedulerName as customAttributes.
			customAttributes := map[string]string{
				converters.KnativeGCPConverter:       converters.CloudSchedulerConverter,
				v1alpha1.CloudSchedulerSourceJobName: jobName,
				v1alpha1.CloudSchedulerSourceName:    scheduler.GetName(),
			}
			_, err = client.CreateJob(ctx, &schedulerpb.CreateJobRequest{
				Parent: parent,
				Job: &schedulerpb.Job{
					Name: jobName,
					Target: &schedulerpb.Job_PubsubTarget{
						PubsubTarget: &schedulerpb.PubsubTarget{
							TopicName:  resources.GeneratePubSubTargetTopic(scheduler, topic),
							Data:       []byte(scheduler.Spec.Data),
							Attributes: customAttributes,
						},
					},
					Schedule: scheduler.Spec.Schedule,
				},
			})
			if err != nil {
				logging.FromContext(ctx).Desugar().Error("Failed to create CloudSchedulerSource job", zap.String("jobName", jobName), zap.Error(err))
				return err
			}
		} else {
			logging.FromContext(ctx).Desugar().Error("Failed from CloudSchedulerSource client while retrieving CloudSchedulerSource job", zap.String("jobName", jobName), zap.Any("errorCode", st.Code()), zap.Error(err))
			return err
		}
	}
	return nil
}

// deleteJob looks at the status.JobName and if non-empty,
// hence indicating that we have created a job successfully
// in the Scheduler, remove it.
func (r *Reconciler) deleteJob(ctx context.Context, scheduler *v1alpha1.CloudSchedulerSource) error {
	if scheduler.Status.JobName == "" {
		return nil
	}

	client, err := r.createClientFn(ctx)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Failed to create CloudSchedulerSource client", zap.Error(err))
		return err
	}
	defer client.Close()

	err = client.DeleteJob(ctx, &schedulerpb.DeleteJobRequest{Name: scheduler.Status.JobName})
	if err == nil {
		logging.FromContext(ctx).Desugar().Debug("Deleted CloudSchedulerSource job", zap.String("jobName", scheduler.Status.JobName))
		return nil
	}
	if st, ok := gstatus.FromError(err); !ok {
		logging.FromContext(ctx).Desugar().Error("Failed from CloudSchedulerSource client while deleting CloudSchedulerSource job", zap.String("jobName", scheduler.Status.JobName), zap.Error(err))
		return err
	} else if st.Code() != codes.NotFound {
		logging.FromContext(ctx).Desugar().Error("Failed to delete CloudSchedulerSource job", zap.String("jobName", scheduler.Status.JobName), zap.Error(err))
		return err
	}
	return nil
}

func addFinalizer(s *v1alpha1.CloudSchedulerSource) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Insert(finalizerName)
	s.Finalizers = finalizers.List()
}

func removeFinalizer(s *v1alpha1.CloudSchedulerSource) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Delete(finalizerName)
	s.Finalizers = finalizers.List()
}

func (r *Reconciler) updateStatus(ctx context.Context, original *v1alpha1.CloudSchedulerSource, desired *v1alpha1.CloudSchedulerSource) error {
	existing := original.DeepCopy()
	return reconciler.RetryUpdateConflicts(func(attempts int) (err error) {
		// The first iteration tries to use the informer's state, subsequent attempts fetch the latest state via API.
		if attempts > 0 {
			existing, err = r.RunClientSet.EventsV1alpha1().CloudSchedulerSources(desired.Namespace).Get(desired.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
		}
		// If there's nothing to update, just return.
		if equality.Semantic.DeepEqual(existing.Status, desired.Status) {
			return nil
		}
		becomesReady := desired.Status.IsReady() && !existing.Status.IsReady()

		existing.Status = desired.Status
		_, err = r.RunClientSet.EventsV1alpha1().CloudSchedulerSources(desired.Namespace).UpdateStatus(existing)

		if err == nil && becomesReady {
			// TODO compute duration since last non-ready. See https://github.com/google/knative-gcp/issues/455.
			duration := time.Since(existing.ObjectMeta.CreationTimestamp.Time)
			logging.FromContext(ctx).Desugar().Info("CloudSchedulerSource became ready", zap.Any("after", duration))
			r.Recorder.Event(existing, corev1.EventTypeNormal, "ReadinessChanged", fmt.Sprintf("CloudSchedulerSource %q became ready", existing.Name))
			if metricErr := r.StatsReporter.ReportReady("CloudSchedulerSource", existing.Namespace, existing.Name, duration); metricErr != nil {
				logging.FromContext(ctx).Desugar().Error("Failed to record ready for CloudSchedulerSource", zap.Error(metricErr))
			}
		}
		return err
	})
}

// updateFinalizers is a generic method for future compatibility with a
// reconciler SDK.
func (r *Reconciler) updateFinalizers(ctx context.Context, desired *v1alpha1.CloudSchedulerSource) (*v1alpha1.CloudSchedulerSource, bool, error) {
	scheduler, err := r.schedulerLister.CloudSchedulerSources(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, false, err
	}

	// Don't modify the informers copy.
	existing := scheduler.DeepCopy()

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

	update, err := r.RunClientSet.EventsV1alpha1().CloudSchedulerSources(existing.Namespace).Patch(existing.Name, types.MergePatchType, patch)
	return update, true, err
}
