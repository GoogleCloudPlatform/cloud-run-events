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

package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
	pubsubv1alpha1 "github.com/google/knative-gcp/pkg/apis/pubsub/v1alpha1"
)

// MakeTopic creates the spec for, but does not create, a GCP Topic
// for a given GCS.
func MakeTopic(namespace, name string, spec *duckv1alpha1.PubSubSpec, owner kmeta.OwnerRefable, topic, receiveAdapterName string) *pubsubv1alpha1.Topic {
	labels := map[string]string{
		"receive-adapter": receiveAdapterName,
	}

	pubsubSecret := spec.Secret
	if spec.PubSubSecret != nil {
		pubsubSecret = spec.PubSubSecret
	}

	return &pubsubv1alpha1.Topic{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(owner)},
		},
		Spec: pubsubv1alpha1.TopicSpec{
			Secret:            pubsubSecret,
			Project:           spec.Project,
			Topic:             topic,
			PropagationPolicy: pubsubv1alpha1.TopicPolicyCreateDelete,
		},
	}
}
