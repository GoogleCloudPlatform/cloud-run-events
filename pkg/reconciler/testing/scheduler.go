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

package testing

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
)

// SchedulerOption enables further configuration of a Scheduler.
type SchedulerOption func(*v1alpha1.Scheduler)

// NewScheduler creates a Scheduler with SchedulerOptions
func NewScheduler(name, namespace string, so ...SchedulerOption) *v1alpha1.Scheduler {
	s := &v1alpha1.Scheduler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       "test-scheduler-uid",
		},
	}
	for _, opt := range so {
		opt(s)
	}
	s.SetDefaults(context.Background())
	return s
}

func WithSchedulerSink(gvk metav1.GroupVersionKind, name string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Spec.Sink = duckv1.Destination{
			Ref: &corev1.ObjectReference{
				APIVersion: apiVersion(gvk),
				Kind:       gvk.Kind,
				Name:       name,
			},
		}
	}
}

func WithSchedulerLocation(location string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Spec.Location = location
	}
}

func WithSchedulerProject(project string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Spec.Project = project
	}
}

func WithSchedulerSchedule(schedule string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Spec.Schedule = schedule
	}
}

func WithSchedulerData(data string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Spec.Data = data
	}
}

func WithSchedulerDeletionTimestamp(s *v1alpha1.Scheduler) {
	t := metav1.NewTime(time.Unix(1e9, 0))
	s.ObjectMeta.SetDeletionTimestamp(&t)
}

// WithInitSchedulerConditions initializes the Schedulers's conditions.
func WithInitSchedulerConditions(s *v1alpha1.Scheduler) {
	s.Status.InitializeConditions()
}

// WithSchedulerTopicFalse marks the condition that the
// status of topic is False.
func WithSchedulerTopicFalse(reason, message string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkTopicFalse(reason, message)
	}
}

// WithSchedulerTopicUnknown marks the condition that the
// status of topic is Unknown.
func WithSchedulerTopicUnknown(reason, message string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkTopicUnknown(reason, message)
	}
}

// WithSchedulerTopicNotReady marks the condition that the
// topic is not ready.
func WithSchedulerTopicReady(topicID, projectID string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkTopicReady(topicID, projectID)
	}
}

// WithSchedulerPullSubscriptionFalse marks the condition that the
// topic is False.
func WithSchedulerPullSubscriptionFalse(reason, message string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkPullSubscriptionFalse(reason, message)
	}
}

// WithSchedulerPullSubscriptionUnknown marks the condition that the
// topic is Unknown.
func WithSchedulerPullSubscriptionUnknown(reason, message string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkPullSubscriptionUnknown(reason, message)
	}
}

// WithSchedulerPullSubscriptionReady marks the condition that the
// topic is ready.
func WithSchedulerPullSubscriptionReady() SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkPullSubscriptionReady()
	}
}

// WithSchedulerJobNotReady marks the condition that the
// Scheduler Job is not ready.
func WithSchedulerJobNotReady(reason, message string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkJobNotReady(reason, message)
	}
}

// WithSchedulerJobReady marks the condition that the
// Scheduler Job is ready and sets Status.JobName to jobName.
func WithSchedulerJobReady(jobName string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.MarkJobReady(jobName)
	}
}

// WithSchedulerSinkURI sets the status for sink URI
func WithSchedulerSinkURI(url *apis.URL) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.SinkURI = url
	}
}

// WithSchedulerJobName sets the status for job Name
func WithSchedulerJobName(jobName string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Status.JobName = jobName
	}
}

func WithSchedulerFinalizers(finalizers ...string) SchedulerOption {
	return func(s *v1alpha1.Scheduler) {
		s.Finalizers = finalizers
	}
}
