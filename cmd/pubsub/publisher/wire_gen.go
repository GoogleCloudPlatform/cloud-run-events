// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"
	"github.com/google/knative-gcp/pkg/pubsub/publisher"
	"github.com/google/knative-gcp/pkg/utils/clients"
)

// Injectors from wire.go:

func InitializePublisher(ctx context.Context, port clients.Port, projectID clients.ProjectID, topicID publisher.TopicID) (*publisher.Publisher, error) {
	httpMessageReceiver := clients.NewHTTPMessageReceiver(port)
	client, err := clients.NewPubsubClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	topic := publisher.NewPubSubTopic(ctx, client, topicID)
	publisherPublisher := publisher.NewPublisher(ctx, httpMessageReceiver, topic)
	return publisherPublisher, nil
}
