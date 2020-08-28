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

package v1alpha1

import (
	"context"

	resourceutil "github.com/google/knative-gcp/pkg/utils/resource"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/ptr"
)

const (
	avgCPUUtilizationFanout  int32 = 95
	avgCPUUtilizationIngress int32 = 95
	avgCPUUtilizationRetry   int32 = 95
	// The limit we set (for Fanout and Retry) is 3000Mi which is mostly used
	// to prevent surging memory usage causing OOM.
	// Here we only set half of the limit so that in case of surging memory
	// usage, HPA could have enough time to kick in.
	// See: https://github.com/google/knative-gcp/issues/1265
	avgMemoryUsageFanout                   string  = "1500Mi"
	avgMemoryUsageIngress                  string  = "700Mi"
	avgMemoryUsageRetry                    string  = "1500Mi"
	cpuRequestFanout                       string  = "1500m"
	cpuRequestIngress                      string  = "1000m"
	cpuRequestRetry                        string  = "1000m"
	cpuLimitFanout                         string  = ""
	cpuLimitIngress                        string  = ""
	cpuLimitRetry                          string  = ""
	memoryRequestFanout                    string  = "500Mi"
	memoryRequestIngress                   string  = "500Mi"
	memoryRequestRetry                     string  = "500Mi"
	memoryLimitToRequestCoefficientFanout  float64 = 6.0
	memoryLimitToRequestCoefficientIngress float64 = 2.0
	memoryLimitToRequestCoefficientRetry   float64 = 6.0
	targetMemoryUsageCoefficientFanout     float64 = 0.5
	targetMemoryUsageCoefficientIngress    float64 = 0.7
	targetMemoryUsageCoefficientRetry      float64 = 0.5
	minReplicas                            int32   = 1
	maxReplicas                            int32   = 10
)

// SetDefaults sets the default field values for a BrokerCell.
func (bc *BrokerCell) SetDefaults(ctx context.Context) {
	// Set defaults for the Spec.Components values.
	bc.Spec.SetDefaults(ctx)
}

// SetDefaults sets the default field values for a BrokerCellSpec.
func (bcs *BrokerCellSpec) SetDefaults(ctx context.Context) {
	// Fanout defaults
	bcs.Components.Fanout.SetResourceDefaults(cpuRequestFanout, cpuLimitFanout, memoryRequestFanout, memoryLimitToRequestCoefficientFanout)
	bcs.Components.Fanout.SetAutoScalingDefaults(targetMemoryUsageCoefficientFanout, avgCPUUtilizationFanout)
	// Retry defaults
	bcs.Components.Retry.SetResourceDefaults(cpuRequestRetry, cpuLimitRetry, memoryRequestRetry, memoryLimitToRequestCoefficientRetry)
	bcs.Components.Retry.SetAutoScalingDefaults(targetMemoryUsageCoefficientRetry, avgCPUUtilizationRetry)
	// Ingress defaults
	bcs.Components.Ingress.SetResourceDefaults(cpuRequestIngress, cpuLimitIngress, memoryRequestIngress, memoryLimitToRequestCoefficientIngress)
	bcs.Components.Ingress.SetAutoScalingDefaults(targetMemoryUsageCoefficientIngress, avgCPUUtilizationIngress)
}

func (componentParams *ComponentParameters) SetResourceDefaults(defaultCPURequest, defaultCPULimit, defaultMemoryRequest string, memoryLimitToRequestCoefficient float64) {
	componentParams.SetCPUDefaults(defaultCPURequest, defaultCPULimit)
	componentParams.SetMemoryDefaults(memoryLimitToRequestCoefficient, defaultMemoryRequest)
}

// SetCPUDefaults sets the CPU consumption related default field values for ComponentParameters.
func (componentParams *ComponentParameters) SetCPUDefaults(defaultCPURequest, defaultCPULimit string) {
	isRequestSpecified := componentParams.Resources.Requests.CPU != nil
	isLimitSpecified := componentParams.Resources.Limits.CPU != nil
	if !isRequestSpecified && !isLimitSpecified {
		componentParams.Resources.Requests.CPU = ptr.String(defaultCPURequest)
		componentParams.Resources.Limits.CPU = ptr.String(defaultCPULimit)
	} else if !isRequestSpecified {
		componentParams.Resources.Requests.CPU = ptr.String(*componentParams.Resources.Limits.CPU)
	}
}

// SetMemoryDefaults sets the memory consumption related default field values for ComponentParameters.
func (componentParams *ComponentParameters) SetMemoryDefaults(memoryLimitToRequestCoefficient float64, defaultMemoryRequest string) {
	isRequestSpecified := componentParams.Resources.Requests.Memory != nil
	isLimitSpecified := componentParams.Resources.Limits.Memory != nil
	autoSelectLimit := false
	if !isRequestSpecified && !isLimitSpecified {
		componentParams.Resources.Requests.Memory = ptr.String(defaultMemoryRequest)
		autoSelectLimit = true
	} else if !isRequestSpecified {
		componentParams.Resources.Requests.Memory = ptr.String(*componentParams.Resources.Limits.Memory)
	} else if !isLimitSpecified {
		autoSelectLimit = true
	}
	if autoSelectLimit {
		requestedMemoryQuantity, err := resource.ParseQuantity(*componentParams.Resources.Requests.Memory)
		if err == nil {
			autoSelectedLimit := resourceutil.MultiplyQuantity(requestedMemoryQuantity, memoryLimitToRequestCoefficient)
			componentParams.Resources.Limits.Memory = ptr.String(autoSelectedLimit.String())
		}
	}
}

// SetAutoScalingDefaults sets the autoscaling-related default field values for ComponentParameters.
func (componentParams *ComponentParameters) SetAutoScalingDefaults(targetMemoryUsageCoefficient float64, avgCPUUtilization int32) {
	if componentParams.MinReplicas == nil {
		componentParams.MinReplicas = ptr.Int32(minReplicas)
	}
	if componentParams.MaxReplicas == nil {
		componentParams.MaxReplicas = ptr.Int32(maxReplicas)
	}
	if componentParams.AvgCPUUtilization == nil {
		componentParams.AvgCPUUtilization = ptr.Int32(avgCPUUtilization)
	}
	// If target average consumption for the auto-scaler is not explicitly specified, default it
	if componentParams.AvgMemoryUsage == nil {
		anchorValue := ""
		if componentParams.Resources.Limits.Memory != nil {
			anchorValue = *componentParams.Resources.Limits.Memory
		} else if componentParams.Resources.Requests.Memory != nil {
			anchorValue = *componentParams.Resources.Requests.Memory
		}
		if anchorValue != "" {
			memoryAnchorQuantity, err := resource.ParseQuantity(anchorValue)
			if err == nil {
				autoSelectedAvgMemoryUsage := resourceutil.MultiplyQuantity(memoryAnchorQuantity, targetMemoryUsageCoefficient)
				componentParams.AvgMemoryUsage = ptr.String(autoSelectedAvgMemoryUsage.String())
			}
		}
	}
}
