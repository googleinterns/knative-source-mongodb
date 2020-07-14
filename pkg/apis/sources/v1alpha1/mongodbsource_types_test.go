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

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestGroupVersionKind(t *testing.T) {
	src := MongoDbSource{}
	gvk := src.GetGroupVersionKind()

	if gvk.Kind != "MongoDbSource" {
		t.Errorf("Should be MongoDbSource.")
	}
}

func TestMongoDbSourceGetStatus(t *testing.T) {
	status := &duckv1.Status{}
	config := MongoDbSource{
		Status: MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{Status: *status},
		},
	}

	if !cmp.Equal(config.GetStatus(), status) {
		t.Errorf("GetStatus did not retrieve status. Got=%v Want=%v", config.GetStatus(), status)
	}
}
