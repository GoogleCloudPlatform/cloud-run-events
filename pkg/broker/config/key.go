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

package config

import (
	"fmt"
	"strings"

	"go.opencensus.io/trace"

	brokerv1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	"go.opencensus.io/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	kntracing "knative.dev/eventing/pkg/tracing"
	"knative.dev/pkg/metrics/metricskey"
)

// CellTenantKey uniquely identifies a single CellTenant, at a given point in time.
type CellTenantKey struct {
	cellTenantType CellTenantType
	namespace      string
	name           string
}

// String creates a human readable version of this key. It is for debug purposes only. It is free to
// change at any time.
func (k *CellTenantKey) String() string {
	// Note that this is explicitly different than the PersistenceString, so that we don't
	// accidentally use String(), rather than PersistenceString().
	return fmt.Sprintf("%s:%s//%s", k.cellTenantType, k.namespace, k.name)
}

// PersistenceString is the string that is persisted as the key for this Broker in the protobuf. It
// is stable and can only change if all existing usage locations are made backwards compatible,
// supporting _both_ the old and the new format, for at least one release.
func (k *CellTenantKey) PersistenceString() string {
	// TODO Remove this.
	if k.cellTenantType == CellTenantType_BROKER {
		// For backwards compatibility from when the only type was Broker, Brokers do not embed
		// their type into the string.
		return k.namespace + "/" + k.name
	}
	return fmt.Sprintf("%s/%s/%s", k.cellTenantType, k.namespace, k.name)
}

func CellTenantKeyFromPersistenceString(s string) (*CellTenantKey, error) {
	pieces := strings.Split(s, "/")
	if len(pieces) <= 2 || len(pieces) >= 5 {
		return nil, fmt.Errorf(
			"malformed request path; expect format '/<ns>/<broker>' or '/<type>/<ns>/<broker>', actually %q", s)
	}
	var blank, ns, brokerName string
	var t CellTenantType
	if len(pieces) == 3 {
		// Broker's persistence strings are in the form "/<ns>/<brokerName>".
		blank, ns, brokerName = pieces[0], pieces[1], pieces[2]
		t = CellTenantType_BROKER
	} else {
		// len(pieces) must be 4, so this is the standard form of the persistence string,
		// '/<type>/<ns>/<name>'.
		var ts string
		blank, ts, ns, brokerName = pieces[0], pieces[1], pieces[2], pieces[3]
		if ctt, err := validateCellTenantTypeFromString(ts); err != nil {
			return nil, err
		} else {
			t = ctt
		}
	}

	if blank != "" {
		return nil, fmt.Errorf(
			"malformed request path; expect format '/<ns>/<broker>' or '/<type>/<ns>/<broker>', actually %q", s)
	}
	if err := validateNamespace(ns); err != nil {
		return nil, err
	}
	if err := validateBrokerName(brokerName); err != nil {
		return nil, err
	}
	return &CellTenantKey{
		cellTenantType: t,
		namespace:      ns,
		name:           brokerName,
	}, nil
}

// MetricsResource generates the Resource object that metrics will be associated with.
func (k *CellTenantKey) MetricsResource() resource.Resource {
	return resource.Resource{
		Type: metricskey.ResourceTypeKnativeBroker,
		Labels: map[string]string{
			metricskey.LabelNamespaceName: k.namespace,
			metricskey.LabelBrokerName:    k.name,
		},
	}
}

// CreateEmptyCellTenant creates an empty CellTenant that corresponds to this CellTenantKey. It is
// empty except for the portions known about by the CellTenantKey.
func (k *CellTenantKey) CreateEmptyCellTenant() *CellTenant {
	return &CellTenant{
		Type:      k.cellTenantType,
		Namespace: k.namespace,
		Name:      k.name,
	}
}

// SpanMessagingDestination is the Messaging Destination of requests sent to this CellTenantKey.
func (k *CellTenantKey) SpanMessagingDestination() string {
	return kntracing.BrokerMessagingDestination(k.namespacedName())
}

// SpanMessagingDestinationAttribute is the Messaging Destination attribute that should be attached
// to the tracing Span.
func (k *CellTenantKey) SpanMessagingDestinationAttribute() trace.Attribute {
	return kntracing.BrokerMessagingDestinationAttribute(k.namespacedName())
}

func (k *CellTenantKey) namespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: k.namespace,
		Name:      k.name,
	}
}

// Key returns the CellTenantKey for this Broker.
func (x *CellTenant) Key() *CellTenantKey {
	return &CellTenantKey{
		cellTenantType: x.Type,
		namespace:      x.Namespace,
		name:           x.Name,
	}
}

// TargetKey uniquely identifies a single Target, at a given point in time.
type TargetKey struct {
	cellTenantKey CellTenantKey
	name          string
}

// ParentKey is the key of the parent this Target corresponds to.
func (k *TargetKey) ParentKey() *CellTenantKey {
	return &k.cellTenantKey
}

// String creates a human readable version of this key. It is for debug purposes only. It is free to
// change at any time.
func (k *TargetKey) String() string {
	// Note that this is explicitly different than the PersistenceString, so that we don't
	// accidentally use String(), rather than PersistenceString().
	return k.cellTenantKey.String() + "//" + k.name
}

// Key returns the TargetKey for this Target.
func (x *Target) Key() *TargetKey {
	return &TargetKey{
		cellTenantKey: CellTenantKey{
			cellTenantType: x.CellTenantType,
			namespace:      x.Namespace,
			name:           x.CellTenantName,
		},
		name: x.Name,
	}
}

// KeyFromBroker creates a CellTenantKey from a K8s Broker object.
func KeyFromBroker(b *brokerv1beta1.Broker) *CellTenantKey {
	return &CellTenantKey{
		cellTenantType: CellTenantType_BROKER,
		namespace:      b.Namespace,
		name:           b.Name,
	}
}

// TestOnlyBrokerKey returns the key of a broker. This method exists to make tests that need a
// CellTenantKey, but do not need an actual Broker, easier to write.
func TestOnlyBrokerKey(namespace, name string) *CellTenantKey {
	return &CellTenantKey{
		cellTenantType: CellTenantType_BROKER,
		namespace:      namespace,
		name:           name,
	}
}

// validateNamespace validates that the given string is a valid K8s namespace.
func validateNamespace(ns string) error {
	errs := validation.IsDNS1123Label(ns)
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("invalid namespace %q, %v", ns, errs)
}

// validateBrokerName validates that the given string is a valid name for a Broker in K8s.
func validateBrokerName(name string) error {
	errs := validation.IsDNS1123Label(name)
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("invalid name %q, %v", name, errs)
}

func validateCellTenantTypeFromString(s string) (CellTenantType, error) {
	i, present := CellTenantType_value[s]
	if !present {
		return CellTenantType_UNKNOWN_CELL_TENANT_TYPE, fmt.Errorf("unknown CellTenantType %q", s)
	}
	return CellTenantType(i), nil
}
