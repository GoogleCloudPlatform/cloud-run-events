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

package broker

import (
	"context"
	"fmt"
	"github.com/google/knative-gcp/pkg/reconciler/broker/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/eventing/pkg/logging"
	"knative.dev/eventing/pkg/reconciler/names"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/system"

	inteventsv1alpha1 "github.com/google/knative-gcp/pkg/apis/intevents/v1alpha1"
	brokerv1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	brokercellresources "github.com/google/knative-gcp/pkg/reconciler/brokercell/resources"
)

// reconcileBrokerCell creates a BrokerCell if it doesn't exist, and update broker status based on brokercell status.
func (r *Reconciler) reconcileBrokerCell(ctx context.Context, b *brokerv1beta1.Broker) error {
	var bc *inteventsv1alpha1.BrokerCell
	var err error
	// TODO(#866) Get brokercell based on the label (or annotation) on the broker.
	bc, err = r.brokerCellLister.BrokerCells(system.Namespace()).Get(resources.DefaultBroekrCellName)

	if err != nil && !apierrs.IsNotFound(err) {
		logging.FromContext(ctx).Error("Error reconciling brokercell", zap.String("namespace", b.Namespace), zap.String("broker", b.Name), zap.Error(err))
		b.Status.MarkBrokerCelllUnknown("Unknown", "Failed to get brokercell %s/%s", bc.Namespace, bc.Name)
		return err
	}

	if apierrs.IsNotFound(err) {
		want := resources.CreateBrokerCell(b)
		bc, err = r.RunClientSet.InternalV1alpha1().BrokerCells(system.Namespace()).Create(want)
		if err != nil {
			logging.FromContext(ctx).Error("Error creating brokercell", zap.String("namespace", b.Namespace), zap.String("broker", b.Name), zap.Error(err))
			b.Status.MarkBrokerCelllFailed("CreationFailed", "Failed to create %s/%s", want.Namespace, want.Name)
			return err
		}
		r.Recorder.Eventf(b, corev1.EventTypeNormal, brokerCellCreated, "Created brokercell %s/%s", bc.Namespace, bc.Name)
	}

	if bc.Status.IsReady() {
		b.Status.MarkBrokerCellReady()
	} else {
		b.Status.MarkBrokerCelllUnknown("NotReady", "Brokercell %s/%s is not ready", bc.Namespace, bc.Name)
	}

	//TODO(#1019) Use the IngressTemplate of brokercell.
	ingressServiceName := brokercellresources.Name(bc.Name, brokercellresources.IngressName)
	b.Status.SetAddress(&apis.URL{
		Scheme: "http",
		Host:   names.ServiceHostName(ingressServiceName, bc.Namespace),
		Path:   fmt.Sprintf("/%s/%s", b.Namespace, b.Name),
	})

	return nil
}

