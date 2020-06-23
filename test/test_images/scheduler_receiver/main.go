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

package main

import (
	"context"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/knative-gcp/pkg/apis/events/v1beta1"
	"github.com/google/knative-gcp/test/e2e/lib"
)

type Receiver struct {
	client cloudevents.Client
}

func main() {
	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		panic(err)
	}
	r := &Receiver{
		client: client,
	}
	if err := r.client.StartReceiver(context.Background(), r.Receive); err != nil {
		log.Fatal(err)
	}
}

func (r *Receiver) Receive(event cloudevents.Event) (*cloudevents.Event, error) {
	// Check if the received event is the event sent by CloudSchedulerSource.
	// If it is, send back a response CloudEvent.
	// Print out event received to log
	fmt.Printf("scheduler receiver received event\n")
	fmt.Printf("context of event is: %v\n", event.Context.String())

	if event.Type() != v1beta1.CloudSchedulerSourceExecute {
		return nil, fmt.Errorf("unexpected cloud event type got=%s, want=%s", event.Type(), v1beta1.CloudSchedulerSourceExecute)
	}
	respEvent := cloudevents.NewEvent(cloudevents.VersionV1)
	respEvent.SetID(lib.E2ESchedulerRespEventID)
	respEvent.SetType(lib.E2ESchedulerRespType)
	respEvent.SetSource(event.Source())
	respEvent.SetSubject(event.Subject())
	respEvent.SetData(event.DataContentType(), event.Data)
	fmt.Printf("context of respEvent is: %v\n", respEvent.Context.String())
	return &respEvent, nil
}
