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

package reconciler

import (
	"context"
	"errors"
	"fmt"
	"knative.dev/pkg/configmap"

	duckv1alpha1 "github.com/google/knative-gcp/pkg/apis/duck/v1alpha1"
	pubsubsourcev1alpha1 "github.com/google/knative-gcp/pkg/apis/pubsub/v1alpha1"
	pubsubsourceclientset "github.com/google/knative-gcp/pkg/client/clientset/versioned"
	pubsubClient "github.com/google/knative-gcp/pkg/client/injection/client"
	"github.com/google/knative-gcp/pkg/reconciler/resources"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
)

type PubSubBase struct {
	*Base

	// For dealing with Topics and Pullsubscriptions
	pubsubClient pubsubsourceclientset.Interface
}

func NewPubSubBase(ctx context.Context, controllerAgentName string, cmw configmap.Watcher) *PubSubBase {
	return &PubSubBase{
		Base:         NewBase(ctx, controllerAgentName, cmw),
		pubsubClient: pubsubClient.Get(ctx),
	}
}

// ReconcilePubSub reconciles Topic / PullSubscription given a PubSubSpec.
func (psb *PubSubBase) ReconcilePubSub(ctx context.Context, namespace, name string, spec *duckv1alpha1.PubSubSpec, status *duckv1alpha1.PubSubStatus, cs *apis.ConditionSet, owner kmeta.OwnerRefable, topic string) (*pubsubsourcev1alpha1.Topic, *pubsubsourcev1alpha1.PullSubscription, error) {
	status.MarkTopicNotReady(cs, "TopicNotReady", "Topic %s/%s not ready", namespace, name)

	topics := psb.pubsubClient.PubsubV1alpha1().Topics(namespace)
	t, err := topics.Get(name, v1.GetOptions{})

	if err != nil {
		if !apierrs.IsNotFound(err) {
			psb.Logger.Infof("Failed to get Topics: %s", err)
			return nil, nil, fmt.Errorf("failed to get topics: %s", err)
		}
		newTopic := resources.MakeTopic(namespace, name, spec, owner, topic)
		psb.Logger.Infof("Creating topic %+v", newTopic)
		t, err = topics.Create(newTopic)
		if err != nil {
			psb.Logger.Infof("Failed to create Topic: %s", err)
			return nil, nil, fmt.Errorf("failed to create topic: %s", err)
		}
	}

	if !t.Status.IsReady() {
		status.MarkTopicNotReady(cs, "TopicNotReady", "Topic %s/%s not ready", t.Namespace, t.Name)
		return t, nil, errors.New("topic not ready")
	}

	if t.Status.ProjectID == "" {
		status.MarkTopicNotReady(cs, "TopicNotReady", "Topic %s/%s did not expose projectid", t.Namespace, t.Name)
		return t, nil, errors.New("topic did not expose projectid")
	}

	if t.Status.TopicID == "" {
		status.MarkTopicNotReady(cs, "TopicNotReady", "Topic %s/%s did not expose topicid", t.Namespace, t.Name)
		return t, nil, errors.New("topic did not expose topicid")
	}

	if t.Status.TopicID != topic {
		status.MarkTopicNotReady(cs, "TopicNotReady", "Topic %s/%s topic mismatch expected %q got %q", t.Namespace, t.Name, topic, t.Status.TopicID)
		return t, nil, errors.New(fmt.Sprintf("topic did not match expected: %q got: %q", topic, t.Status.TopicID))
	}

	status.TopicID = t.Status.TopicID
	status.ProjectID = t.Status.ProjectID
	status.MarkTopicReady(cs)

	status.MarkPullSubscriptionNotReady(cs, "PullSubscriptionNotReady", "PullSubscription %s/%s not ready", namespace, name)

	// Ok, so the Topic is ready, let's reconcile PullSubscription.
	pullSubscriptions := psb.pubsubClient.PubsubV1alpha1().PullSubscriptions(namespace)
	ps, err := pullSubscriptions.Get(name, v1.GetOptions{})
	if err != nil {
		if !apierrs.IsNotFound(err) {
			psb.Logger.Infof("Failed to get PullSubscriptions: %s", err)
			return t, nil, fmt.Errorf("failed to get pullsubscriptions: %s", err)
		}
		newPS := resources.MakePullSubscription(namespace, name, spec, owner, topic)
		psb.Logger.Infof("Creating pullsubscription %+v", newPS)
		ps, err = pullSubscriptions.Create(newPS)
		if err != nil {
			psb.Logger.Infof("Failed to create PullSubscription: %s", err)
			return t, nil, fmt.Errorf("failed to create pullsubscription: %s", err)
		}
	}

	if !ps.Status.IsReady() {
		psb.Logger.Infof("PullSubscription is not ready yet")
		status.MarkPullSubscriptionNotReady(cs, "PullSubscriptionNotReady", "PullSubscription %s/%s not ready", ps.Namespace, ps.Name)
		return t, nil, errors.New("pullsubscription not ready")
	} else {
		status.MarkPullSubscriptionReady(cs)
	}
	psb.Logger.Infof("Using %q as a cluster internal sink", ps.Status.SinkURI)
	uri, err := apis.ParseURL(ps.Status.SinkURI)
	if err != nil {
		return t, ps, errors.New(fmt.Sprintf("failed to parse url %q : %q", ps.Status.SinkURI, err))
	}
	status.SinkURI = uri
	return t, ps, nil
}
