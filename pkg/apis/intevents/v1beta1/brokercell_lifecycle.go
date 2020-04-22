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

package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
)

var brokerCellCondSet = apis.NewLivingConditionSet(
	BrokerCellConditionIngress,
	BrokerCellConditionFanout,
	BrokerCellConditionRetry,
	BrokerCellConditionTargetsConfig,
)

const (
	// BrokerCellConditionReady has status true when all subconditions below
	// have been set to True.
	BrokerCellConditionReady apis.ConditionType = apis.ConditionReady

	// BrokerCellConditionIngress reports the availability of the
	// BrokerCell's ingress service.
	BrokerCellConditionIngress apis.ConditionType = "IngressReady"

	// BrokerCellConditionFanout reports the readiness of the BrokerCell's
	// fanout service.
	BrokerCellConditionFanout apis.ConditionType = "FanoutReady"

	// BrokerCellConditionRetry reports the readiness of the BrokerCell's retry
	// service.
	BrokerCellConditionRetry apis.ConditionType = "RetryReady"

	// BrokerCellConditionTargetsConfig reports the readiness of the
	// BrokerCell's targets configmap.
	BrokerCellConditionTargetsConfig apis.ConditionType = "TargetsConfigReady"
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (bs *BrokerCellStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return brokerCellCondSet.Manage(bs).GetCondition(t)
}

// GetTopLevelCondition returns the top level Condition.
func (bs *BrokerCellStatus) GetTopLevelCondition() *apis.Condition {
	return brokerCellCondSet.Manage(bs).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (bs *BrokerCellStatus) IsReady() bool {
	return brokerCellCondSet.Manage(bs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (bs *BrokerCellStatus) InitializeConditions() {
	brokerCellCondSet.Manage(bs).InitializeConditions()
}

// PropagateIngressAvailability uses the availability of the provided Endpoints
// to determine if BrokerCellConditionIngress should be marked as true or
// false.
func (bs *BrokerCellStatus) PropagateIngressAvailability(ep *corev1.Endpoints) {
	if duck.EndpointsAreAvailable(ep) {
		brokerCellCondSet.Manage(bs).MarkTrue(BrokerCellConditionIngress)
	} else {
		brokerCellCondSet.Manage(bs).MarkFalse(BrokerCellConditionIngress, "EndpointsUnavailable", "Endpoints %q is unavailable.", ep.Name)
	}
}

// PropagateFanoutAvailability uses the availability of the provided Deployment
// to determine if BrokerCellConditionFanout should be marked as true or
// false.
func (bs *BrokerCellStatus) PropagateFanoutAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		brokerCellCondSet.Manage(bs).MarkTrue(BrokerCellConditionFanout)
	} else {
		brokerCellCondSet.Manage(bs).MarkFalse(BrokerCellConditionFanout, "DeploymentUnavailable", "Deployment %q is unavailable.", d.Name)
	}
}

// PropagateRetryAvailability uses the availability of the provided Deployment
// to determine if BrokerCellConditionRetry should be marked as true or
// unknown.
func (bs *BrokerCellStatus) PropagateRetryAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		brokerCellCondSet.Manage(bs).MarkTrue(BrokerCellConditionRetry)
	} else {
		brokerCellCondSet.Manage(bs).MarkFalse(BrokerCellConditionRetry, "DeploymentUnavailable", "Deployment %q is unavailable.", d.Name)
	}
}

func (bs *BrokerCellStatus) MarkTargetsConfigReady() {
	brokerCellCondSet.Manage(bs).MarkTrue(BrokerCellConditionTargetsConfig)
}

func (bs *BrokerCellStatus) MarkTargetsConfigFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(BrokerCellConditionTargetsConfig, reason, format, args...)
}
