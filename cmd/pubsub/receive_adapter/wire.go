// +build wireinject

/*
Copyright 2020 Google LLC.

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

package main

import (
	"context"

	"github.com/google/knative-gcp/pkg/pubsub/adapter"
	"github.com/google/knative-gcp/pkg/pubsub/adapter/converters"
	"github.com/google/knative-gcp/pkg/utils/clients"

	"github.com/google/wire"
)

func InitializeAdapter(
	ctx context.Context,
	projectID clients.ProjectID,
	subscription adapter.SubscriptionID,
	maxConnsPerHost clients.MaxConnsPerHost,
	name adapter.Name,
	namespace adapter.Namespace,
	resourceGroup adapter.ResourceGroup,
	converterType converters.ConverterType,
	sinkURI adapter.SinkURI,
	transformerURI adapter.TransformerURI,
	extensions map[string]string,
) (*adapter.Adapter, error) {
	panic(wire.Build(
		adapter.AdapterSet,
	))
}
