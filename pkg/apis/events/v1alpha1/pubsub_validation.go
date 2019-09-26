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
	"context"
	"time"

	"knative.dev/pkg/apis/v1alpha1"

	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/google/go-cmp/cmp"
	"knative.dev/pkg/apis"
)

const (
	minRetentionDuration = 10 * time.Second   // 10 seconds.
	maxRetentionDuration = 7 * 24 * time.Hour // 7 days.

	minAckDeadline = 0 * time.Second  // 0 seconds.
	maxAckDeadline = 10 * time.Minute // 10 minutes.
)

func (current *PubSub) Validate(ctx context.Context) *apis.FieldError {
	return current.Spec.Validate(ctx).ViaField("spec")
}

func (current *PubSubSpec) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError
	// Topic [required]
	if current.Topic == "" {
		errs = errs.Also(apis.ErrMissingField("topic"))
	}
	// Sink [required]
	if equality.Semantic.DeepEqual(current.Sink, v1alpha1.Destination{}) {
		errs = errs.Also(apis.ErrMissingField("sink"))
	} else if err := validateDestination(current.Sink); err != nil {
		errs = errs.Also(err.ViaField("sink"))
	}

	if current.RetentionDuration != nil {
		// If set, RetentionDuration Cannot be longer than 7 days or shorter than 10 minutes.
		rd, err := time.ParseDuration(*current.RetentionDuration)
		if err != nil {
			errs = errs.Also(apis.ErrInvalidValue(*current.RetentionDuration, "retentionDuration"))
		} else if rd < minRetentionDuration || rd > maxRetentionDuration {
			errs = errs.Also(apis.ErrOutOfBoundsValue(*current.RetentionDuration, minRetentionDuration.String(), maxRetentionDuration.String(), "retentionDuration"))
		}
	}

	if current.AckDeadline != nil {
		// If set, AckDeadline needs to parse to a valid duration.
		ad, err := time.ParseDuration(*current.AckDeadline)
		if err != nil {
			errs = errs.Also(apis.ErrInvalidValue(*current.AckDeadline, "ackDeadline"))
		} else if ad < minAckDeadline || ad > maxAckDeadline {
			errs = errs.Also(apis.ErrOutOfBoundsValue(*current.AckDeadline, minAckDeadline.String(), maxAckDeadline.String(), "ackDeadline"))
		}
	}

	return errs
}

func validateDestination(dest v1alpha1.Destination) *apis.FieldError {
	if dest.URI != nil {
		if dest.ObjectReference != nil {
			return apis.ErrMultipleOneOf("uri", "name")
		}
		if dest.URI.Host == "" || dest.URI.Scheme == "" {
			return apis.ErrInvalidValue(dest.URI.String(), "uri")
		}
	} else {
		return validateRef(dest.ObjectReference)
	}
	return nil
}

func validateRef(ref *corev1.ObjectReference) *apis.FieldError {
	// nil check.
	if ref == nil {
		return apis.ErrMissingField(apis.CurrentField)
	}
	// Check the object.
	var errs *apis.FieldError
	// Required Fields
	if ref.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}
	if ref.APIVersion == "" {
		errs = errs.Also(apis.ErrMissingField("apiVersion"))
	}
	if ref.Kind == "" {
		errs = errs.Also(apis.ErrMissingField("kind"))
	}

	return errs
}

func (current *PubSub) CheckImmutableFields(ctx context.Context, og apis.Immutable) *apis.FieldError {
	original, ok := og.(*PubSub)
	if !ok {
		return &apis.FieldError{Message: "The provided original was not a PubSub"}
	}
	if original == nil {
		return nil
	}

	// Modification of Topic, Secret and Project are not allowed. Everything else is mutable.
	if diff := cmp.Diff(original.Spec, current.Spec,
		cmpopts.IgnoreFields(PubSubSpec{},
			"Sink", "AckDeadline", "RetainAckedMessages", "RetentionDuration", "CloudEventOverrides")); diff != "" {
		return &apis.FieldError{
			Message: "Immutable fields changed (-old +new)",
			Paths:   []string{"spec"},
			Details: diff,
		}
	}
	return nil
}
