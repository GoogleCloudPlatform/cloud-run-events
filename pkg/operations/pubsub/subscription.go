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

package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"go.uber.org/zap"
	"knative.dev/pkg/logging"

	"github.com/google/knative-gcp/pkg/operations"
	"github.com/google/knative-gcp/pkg/pubsub"

	corev1 "k8s.io/api/core/v1"
)

// TODO: This is currently only used on success to communicate the
// project status. If there's something else that could be useful
// to communicate to the controller, add them here.
type SubActionResult struct {
	// Project is the project id that we used (this might have
	// been defaulted, so we'll expose it so that controller can
	// reflect this in the Status).
	ProjectId string `json:"projectId,omitempty"`
}

// SubArgs are the configuration required to make a NewSubscriptionOps.
type SubArgs struct {
	PubSubArgs
	TopicID        string
	SubscriptionID string
}

func (s SubArgs) OperationSubgroup() string {
	return "s"
}

func (s SubArgs) LabelKey() string {
	return "subscription"
}

func SubEnv(s SubArgs) []corev1.EnvVar {
	return append(PubSubEnv(s.PubSubArgs),
		corev1.EnvVar{
			Name:  "PUBSUB_TOPIC_ID",
			Value: s.TopicID,
		}, corev1.EnvVar{
			Name:  "PUBSUB_SUBSCRIPTION_ID",
			Value: s.SubscriptionID,
		})
}

type SubExistsArgs struct {
	SubArgs
}

var _ operations.JobArgs = SubExistsArgs{}

func (s SubExistsArgs) Action() string {
	return operations.ActionExists
}

func (s SubExistsArgs) Env() []corev1.EnvVar {
	return SubEnv(s.SubArgs)
}

//TODO: Add validation
func (s SubExistsArgs) Validate() error {
	return nil
}

type SubCreateArgs struct {
	SubArgs
	AckDeadline         time.Duration
	RetainAckedMessages bool
	RetentionDuration   time.Duration
}

var _ operations.JobArgs = SubCreateArgs{}

func (s SubCreateArgs) Action() string {
	return operations.ActionCreate
}

func (s SubCreateArgs) Env() []corev1.EnvVar {
	return append(SubEnv(s.SubArgs),
		corev1.EnvVar{
			Name:  "PUBSUB_SUBSCRIPTION_CONFIG_ACK_DEAD",
			Value: s.AckDeadline.String(),
		},
		corev1.EnvVar{
			Name:  "PUBSUB_SUBSCRIPTION_CONFIG_RET_ACKED",
			Value: strconv.FormatBool(s.RetainAckedMessages),
		},
		corev1.EnvVar{
			Name:  "PUBSUB_SUBSCRIPTION_CONFIG_RET_DUR",
			Value: s.RetentionDuration.String(),
		})
}

//TODO: Add validation
func (s SubCreateArgs) Validate() error {
	return nil
}

type SubDeleteArgs struct {
	SubArgs
}

var _ operations.JobArgs = SubDeleteArgs{}

func (s SubDeleteArgs) Action() string {
	return operations.ActionDelete
}

func (s SubDeleteArgs) Env() []corev1.EnvVar {
	return SubEnv(s.SubArgs)
}

//TODO: Add validation
func (s SubDeleteArgs) Validate() error {
	return nil
}

// TODO: the job could output the resolved projectID.

// SubscriptionOps defines the configuration to use for this operation.
type SubscriptionOps struct {
	PubSubOps

	// Action is the operation the job should run.
	// Options: [exists, create, delete]
	Action string `envconfig:"ACTION" required:"true"`

	// Topic is the environment variable containing the PubSub Topic being
	// subscribed to's name. In the form that is unique within the project.
	// E.g. 'laconia', not 'projects/my-gcp-project/topics/laconia'.
	Topic string `envconfig:"PUBSUB_TOPIC_ID" required:"true"`

	// Subscription is the environment variable containing the name of the
	// subscription to use.
	Subscription string `envconfig:"PUBSUB_SUBSCRIPTION_ID" required:"true"`

	// AckDeadline is the default maximum time after a subscriber receives a
	// message before the subscriber should acknowledge the message. Defaults
	// to 30 seconds.
	AckDeadline time.Duration `envconfig:"PUBSUB_SUBSCRIPTION_CONFIG_ACK_DEAD" required:"true" default:"30s"`

	// RetainAckedMessages defines whether to retain acknowledged messages. If
	// true, acknowledged messages will not be expunged until they fall out of
	// the RetentionDuration window.
	RetainAckedMessages bool `envconfig:"PUBSUB_SUBSCRIPTION_CONFIG_RET_ACKED" required:"true" default:"false"`

	// RetentionDuration defines how long to retain messages in backlog, from
	// the time of publish. If RetainAckedMessages is true, this duration
	// affects the retention of acknowledged messages, otherwise only
	// unacknowledged messages are retained. Defaults to 7 days. Cannot be
	// longer than 7 days or shorter than 10 minutes.
	RetentionDuration time.Duration `envconfig:"PUBSUB_SUBSCRIPTION_CONFIG_RET_DUR" required:"true" default:"168h"`
}

var (
	ignoreSubConfig = cmpopts.IgnoreFields(pubsub.SubscriptionConfig{}, "Topic", "Labels")
)

// Run will perform the action configured upon a subscription.
func (s *SubscriptionOps) Run(ctx context.Context) error {
	if s.Client == nil {
		return errors.New("pub/sub client is nil")
	}
	logger := logging.FromContext(ctx)

	logger = logger.With(
		zap.String("action", s.Action),
		zap.String("project", s.Project),
		zap.String("topic", s.Topic),
		zap.String("subscription", s.Subscription),
	)

	logger.Info("Pub/Sub Subscription Job.")

	// Load the subscription.
	sub := s.Client.Subscription(s.Subscription)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify topic exists: %s", err)
	}

	switch s.Action {
	case operations.ActionExists:
		// If subscription doesn't exist, that is an error.
		if !exists {
			return errors.New("subscription does not exist")
		}
		logger.Info("Previously created.")

	case operations.ActionCreate:
		// Load the topic.
		topic, err := s.getTopic(ctx)
		if err != nil {
			return fmt.Errorf("failed to get topic, %s", err)
		}
		// subConfig is the wanted config based on settings.
		subConfig := pubsub.SubscriptionConfig{
			Topic:               topic,
			AckDeadline:         s.AckDeadline,
			RetainAckedMessages: s.RetainAckedMessages,
			RetentionDuration:   s.RetentionDuration,
		}

		// If topic doesn't exist, create it.
		if !exists {
			// Create a new subscription to the previous topic with the given name.
			sub, err = s.Client.CreateSubscription(ctx, s.Subscription, subConfig)
			if err != nil {
				return fmt.Errorf("failed to create subscription, %s", err)
			}
			logger.Info("Successfully created.")
		} else {
			// TODO: here is where we could update config.
			logger.Info("Previously created.")
			// Get current config.
			currentConfig, err := sub.Config(ctx)
			if err != nil {
				return fmt.Errorf("failed to get subscription config, %s", err)
			}
			// Compare the current config to the expected config. Update if different.
			if diff := cmp.Diff(subConfig, currentConfig, ignoreSubConfig); diff != "" {
				_, err := sub.Update(ctx, pubsub.SubscriptionConfig{
					AckDeadline:         s.AckDeadline,
					RetainAckedMessages: s.RetainAckedMessages,
					RetentionDuration:   s.RetentionDuration,
					Labels:              currentConfig.Labels,
				})
				if err != nil {
					return fmt.Errorf("failed to update subscription config, %s", err)
				}
				logger.Info("Updated subscription config.", zap.String("diff", diff))

			}
		}

	case operations.ActionDelete:
		if exists {
			if err := sub.Delete(ctx); err != nil {
				return fmt.Errorf("failed to delete subscription, %s", err)
			}
			logger.Info("Successfully deleted.")
		} else {
			logger.Info("Previously deleted.")
		}

	default:
		return fmt.Errorf("unknown action value %v", s.Action)
	}

	s.writeTerminationMessage(&SubActionResult{})
	logger.Info("Done.")
	return nil
}

func (s *SubscriptionOps) getTopic(ctx context.Context) (pubsub.Topic, error) {
	// Load the topic.
	topic := s.Client.Topic(s.Topic)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		return topic, err
	}
	return nil, errors.New("topic does not exist")
}

func (s *SubscriptionOps) writeTerminationMessage(result *SubActionResult) error {
	// Always add the project regardless of what we did.
	result.ProjectId = s.Project
	m, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("/dev/termination-log", m, 0644)
}
