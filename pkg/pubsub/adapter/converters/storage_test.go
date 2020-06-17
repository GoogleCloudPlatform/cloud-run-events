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

package converters

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/knative-gcp/pkg/apis/events/v1beta1"

	cev2 "github.com/cloudevents/sdk-go/v2"
)

const (
	bucket    = "my-bucket"
	objectId  = "myfile.jpg"
	eventType = "OBJECT_FINALIZE"
)

var (
	storagePublishTime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
)

func TestConvertCloudStorageSource(t *testing.T) {

	tests := []struct {
		name        string
		message     *pubsub.Message
		wantEventFn func() *cev2.Event
		wantErr     bool
	}{{
		name: "no attributes",
		message: &pubsub.Message{
			Data: []byte("test data"),
			Attributes: map[string]string{
				"knative-gcp": "com.google.cloud.storage",
			},
		},
		wantEventFn: func() *cev2.Event {
			return storageCloudEvent()
		},
		wantErr: true,
	}, {
		name: "no bucketId attribute",
		message: &pubsub.Message{
			Data: []byte("test data"),
			Attributes: map[string]string{
				"knative-gcp": "com.google.cloud.storage",
				"eventType":   eventType,
				"attribute1":  "value1",
				"attribute2":  "value2",
			},
		},
		wantEventFn: func() *cev2.Event {
			return storageCloudEvent()
		},
		wantErr: true,
	}, {
		name: "no eventType attribute",
		message: &pubsub.Message{
			Data: []byte("test data"),
			Attributes: map[string]string{
				"knative-gcp": "com.google.cloud.storage",
				"bucketId":    bucket,
				"objectId":    objectId,
			},
		},
		wantEventFn: func() *cev2.Event {
			return storageCloudEvent()
		},
		wantErr: true,
	}, {
		name: "unkown eventType attribute",
		message: &pubsub.Message{
			Data: []byte("test data"),
			Attributes: map[string]string{
				"knative-gcp": "com.google.cloud.storage",
				"eventType":   "RANDOM_EVENT",
				"bucketId":    bucket,
				"objectId":    objectId,
			},
		},
		wantEventFn: func() *cev2.Event {
			return storageCloudEvent()
		},
		wantErr: true,
	}, {
		name: "no objectId attribute",
		message: &pubsub.Message{
			Data: []byte("test data"),
			Attributes: map[string]string{
				"knative-gcp": "com.google.cloud.storage",
				"bucketId":    bucket,
				"eventType":   eventType,
			},
		},
		wantEventFn: func() *cev2.Event {
			return storageCloudEvent()
		},
		wantErr: true,
	}, {
		name: "valid message",
		message: &pubsub.Message{
			ID:   "id",
			PublishTime: storagePublishTime,
			Data: []byte("test data"),
			Attributes: map[string]string{
				"knative-gcp": "com.google.cloud.storage",
				"bucketId":    bucket,
				"eventType":   eventType,
				"objectId":    objectId,
			},
		},
		wantEventFn: func() *cev2.Event {
			return storageCloudEvent()
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotEvent, err := NewPubSubConverter().Convert(context.Background(), test.message, "")

			if err != nil {
				if !test.wantErr {
					t.Fatalf("converters.convertCloudStorage got error %v want error=%v", err, test.wantErr)
				}
			} else {
				if gotEvent.ID() != "id" {
					t.Errorf("ID '%s' != '%s'", gotEvent.ID(), "id")
				}
				if !gotEvent.Time().Equal(storagePublishTime) {
					t.Errorf("Time '%v' != '%v'", gotEvent.Time(), storagePublishTime)
				}
				if want := v1beta1.CloudStorageSourceEventSource("my-bucket"); gotEvent.Source() != want {
					t.Errorf("Source %q != %q", gotEvent.Source(), want)
				}
				if gotEvent.Type() != v1beta1.CloudStorageSourceFinalize {
					t.Errorf(`Type %q != %q`, gotEvent.Type(), v1beta1.CloudStorageSourceFinalize)
				}
				if gotEvent.Subject() != objectId {
					t.Errorf("Subject %q != %q", gotEvent.Subject(), objectId)
				}
				if gotEvent.DataSchema() != storageSchemaUrl {
					t.Errorf("DataSchema %q != %q", gotEvent.DataSchema(), storageSchemaUrl)
				}
			}
		})
	}
}

func storageCloudEvent() *cev2.Event {
	e := cev2.NewEvent(cev2.VersionV1)
	e.SetID("id")
	e.SetTime(storagePublishTime)
	e.SetData(cev2.ApplicationJSON, []byte("test data"))
	e.SetDataSchema(storageSchemaUrl)
	e.SetSource(v1beta1.CloudStorageSourceEventSource(bucket))
	e.SetType(v1beta1.CloudStorageSourceFinalize)
	e.SetSubject(objectId)
	return &e
}
