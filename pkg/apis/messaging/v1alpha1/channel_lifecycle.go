/*
 * Copyright 2019 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// GetCondition returns the condition currently associated with the given type,
// or nil.
func (cs *ChannelStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return channelCondSet.Manage(cs).GetCondition(t)
}

// GetTopLevelCondition returns the top level condition.
func (cs *ChannelStatus) GetTopLevelCondition() *apis.Condition {
	return channelCondSet.Manage(cs).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (cs *ChannelStatus) IsReady() bool {
	return channelCondSet.Manage(cs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (cs *ChannelStatus) InitializeConditions() {
	channelCondSet.Manage(cs).InitializeConditions()
}

// SetAddress updates the Addressable status of the channel and propagates a
// url status to the Addressable status condition based on url.
func (cs *ChannelStatus) SetAddress(url *apis.URL) {
	if cs.Address == nil {
		cs.Address = &duckv1.Addressable{}
	}
	if url != nil {
		cs.Address.URL = url
		channelCondSet.Manage(cs).MarkTrue(ChannelConditionAddressable)
	} else {
		cs.Address.URL = nil
		channelCondSet.Manage(cs).MarkFalse(ChannelConditionAddressable, "emptyUrl", "url is empty")
	}
}

// MarkTopicReady sets the condition that the topic has been created and ready.
func (cs *ChannelStatus) MarkTopicReady() {
	channelCondSet.Manage(cs).MarkTrue(ChannelConditionTopicReady)
}

func (cs *ChannelStatus) PropagateTopicStatus(ready *apis.Condition) {
	if ready == nil {
		cs.MarkTopicNotConfigured()
		return
	}

	switch {
	case ready.Status == corev1.ConditionUnknown:
		cs.MarkTopicUnknown(ready.Reason, ready.Message)
	case ready.Status == corev1.ConditionTrue:
		cs.MarkTopicReady()
	case ready.Status == corev1.ConditionFalse:
		cs.MarkTopicFalse(ready.Reason, ready.Message)
	default:
		cs.MarkTopicUnknown("TopicUnknown", "The status of Topic is invalid: %v", ready.Status)
	}
}

// MarkTopicFalse sets the condition that signals there is not a topic for this
// Channel. This could be because of an error or the Channel is being deleted.
func (cs *ChannelStatus) MarkTopicFalse(reason, messageFormat string, messageA ...interface{}) {
	channelCondSet.Manage(cs).MarkFalse(ChannelConditionTopicReady, reason, messageFormat, messageA...)
}

func (cs *ChannelStatus) MarkTopicNotOwned(messageFormat string, messageA ...interface{}) {
	channelCondSet.Manage(cs).MarkFalse(ChannelConditionTopicReady, "NotOwned", messageFormat, messageA...)
}

func (cs *ChannelStatus) MarkTopicNotConfigured() {
	channelCondSet.Manage(cs).MarkUnknown(ChannelConditionTopicReady,
		"TopicNotConfigured", "Topic has not yet been reconciled")
}

func (cs *ChannelStatus) MarkTopicUnknown(reason, messageFormat string, messageA ...interface{}) {
	channelCondSet.Manage(cs).MarkUnknown(ChannelConditionTopicReady, reason, messageFormat, messageA...)
}
