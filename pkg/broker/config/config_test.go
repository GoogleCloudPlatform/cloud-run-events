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
	"testing"
)

func TestBrokerKey(t *testing.T) {
	want := "namespace/broker"
	got := BrokerKey("namespace", "broker")
	if got != want {
		t.Errorf("unexpected readiness: want %v, got %v", want, got)
	}
}

func TestTriggerKeyAndSplitTriggerKey(t *testing.T) {
	want := "namespace/broker/target"
	got := TriggerKey(SplitTriggerKey(want))
	if got != want {
		t.Errorf("unexpected readiness: want %v, got %v", want, got)
	}
}

func TestBrokerPath(t *testing.T) {
	want := "/namespace/broker"
	got := BrokerPath("namespace", "broker")
	if got != want {
		t.Errorf("unexpected readiness: want %v, got %v", want, got)
	}
}
