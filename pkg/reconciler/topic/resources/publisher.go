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
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/knative/pkg/kmeta"

	"github.com/GoogleCloudPlatform/cloud-run-events/pkg/apis/pubsub/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PublisherArgs are the arguments needed to create a Channel invoker.
// Every field is required.
type PublisherArgs struct {
	Image  string
	Topic  *v1alpha1.Topic
	Labels map[string]string
}

const (
	credsVolume    = "google-cloud-key"
	credsMountPath = "/var/secrets/google"
)

// DefaultSecretSelector is the default secret selector used to load the creds
// for the invoker to auth with Google Cloud.
func DefaultSecretSelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: "google-cloud-key",
		},
		Key: "key.json",
	}
}

// MakePublisher generates (but does not insert into K8s) the Invoker Deployment for
// Channels.
func MakePublisher(args *PublisherArgs) *v1.Deployment {

	secret := args.Topic.Spec.Secret
	if secret == nil {
		secret = DefaultSecretSelector()
	}

	credsFile := fmt.Sprintf("%s/%s", credsMountPath, secret.Key)
	replicas := int32(1)
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       args.Topic.Namespace,
			GenerateName:    fmt.Sprintf("pubsub-publisher-%s-", args.Topic.Name),
			Labels:          args.Labels, // TODO: not sure we should use labels like this.
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(args.Topic)},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: args.Labels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: args.Labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: args.Topic.Spec.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:  "invoker",
						Image: args.Image,
						Env: []corev1.EnvVar{{
							Name:  "GOOGLE_APPLICATION_CREDENTIALS",
							Value: credsFile,
						}, {
							Name:  "PROJECT_ID",
							Value: args.Topic.Spec.Project,
						}, {
							Name:  "TOPIC_ID",
							Value: args.Topic.Spec.Topic,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      credsVolume,
							MountPath: credsMountPath,
						}}},
					},
					Volumes: []corev1.Volume{{
						Name: credsVolume,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: secret.Name,
							},
						},
					}},
				},
			},
		},
	}
}

// MakePublisherService creates the in-memory representation of the Broker's ingress Service.
func MakePublisherService(args *PublisherArgs) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: args.Topic.Namespace,
			Name:      fmt.Sprintf("%s-topic", args.Topic.Name),
			Labels:    args.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Topic),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: args.Labels,
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			}, {
				Name: "metrics",
				Port: 9090,
			}},
		},
	}
}
