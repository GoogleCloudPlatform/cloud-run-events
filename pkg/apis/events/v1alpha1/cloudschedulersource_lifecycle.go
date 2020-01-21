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

package v1alpha1

import (
	"knative.dev/pkg/apis"

	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *CloudSchedulerSourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return schedulerCondSet.Manage(s).GetCondition(t)
}

// GetTopLevelCondition returns the top level condition.
func (s *CloudSchedulerSourceStatus) GetTopLevelCondition() *apis.Condition {
	return schedulerCondSet.Manage(s).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (s *CloudSchedulerSourceStatus) IsReady() bool {
	return schedulerCondSet.Manage(s).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *CloudSchedulerSourceStatus) InitializeConditions() {
	schedulerCondSet.Manage(s).InitializeConditions()
}

// MarkPullSubscriptionFailed sets the condition that the underlying PullSubscription
// is False and why.
func (s *CloudSchedulerSourceStatus) MarkPullSubscriptionFailed(reason, messageFormat string, messageA ...interface{}) {
	schedulerCondSet.Manage(s).MarkFalse(duckv1alpha1.PullSubscriptionReady, reason, messageFormat, messageA...)
}

// MarkPullSubscriptionUnknown sets the condition that the underlying PullSubscription
// is Unknown and why.
func (s *CloudSchedulerSourceStatus) MarkPullSubscriptionUnknown(reason, messageFormat string, messageA ...interface{}) {
	schedulerCondSet.Manage(s).MarkUnknown(duckv1alpha1.PullSubscriptionReady, reason, messageFormat, messageA...)
}

// MarkPullSubscriptionReady sets the condition that the underlying PullSubscription is ready.
func (s *CloudSchedulerSourceStatus) MarkPullSubscriptionReady() {
	schedulerCondSet.Manage(s).MarkTrue(duckv1alpha1.PullSubscriptionReady)
}

// MarkTopicFailed sets the condition that the Topic was not created and why.
func (s *CloudSchedulerSourceStatus) MarkTopicFailed(reason, messageFormat string, messageA ...interface{}) {
	schedulerCondSet.Manage(s).MarkFalse(duckv1alpha1.TopicReady, reason, messageFormat, messageA...)
}

// MarkTopicUnknown sets the condition that the status of Topic is Unknown and why.
func (s *CloudSchedulerSourceStatus) MarkTopicUnknown(reason, messageFormat string, messageA ...interface{}) {
	schedulerCondSet.Manage(s).MarkUnknown(duckv1alpha1.TopicReady, reason, messageFormat, messageA...)
}

// MarkTopicReady sets the condition that the underlying Topic was created
// successfully and sets the Status.TopicID to the specified topic
// and Status.ProjectID to the specified project.
func (s *CloudSchedulerSourceStatus) MarkTopicReady(topicID, projectID string) {
	schedulerCondSet.Manage(s).MarkTrue(duckv1alpha1.TopicReady)
	s.TopicID = topicID
	s.ProjectID = projectID
}

// MarkJobNotReady sets the condition that the CloudSchedulerSource Job has not been
// successfully created.
func (s *CloudSchedulerSourceStatus) MarkJobNotReady(reason, messageFormat string, messageA ...interface{}) {
	schedulerCondSet.Manage(s).MarkFalse(JobReady, reason, messageFormat, messageA...)
}

// MarkJobReady sets the condition for CloudSchedulerSource Job as Read and sets the
// Status.JobName to jobName
func (s *CloudSchedulerSourceStatus) MarkJobReady(jobName string) {
	schedulerCondSet.Manage(s).MarkTrue(JobReady)
	s.JobName = jobName
}
