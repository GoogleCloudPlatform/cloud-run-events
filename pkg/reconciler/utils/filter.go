/*
Copyright 2021 Google LLC

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

package utils

import (
	brokerv1 "github.com/google/knative-gcp/pkg/apis/broker/v1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	reconciler "knative.dev/pkg/reconciler"
)

// BrokerClassFilter is the function to filter brokers with proper brokerclass.
var BrokerClassFilter = reconciler.AnnotationFilterFunc(eventingv1.BrokerClassAnnotationKey, brokerv1.BrokerClass, false /*allowUnset*/)
