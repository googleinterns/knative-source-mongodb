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
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// Validate validates MongoDbSource.
func (m *MongoDbSource) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError

	//validation for "spec" field.
	errs = errs.Also(m.Spec.Validate(ctx).ViaField("spec"))

	//errs is nil if everything is fine.
	return errs
}

// Validate validates MongoDbSourceSpecs.
func (ms *MongoDbSourceSpec) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError

	// Validate sink.
	if equality.Semantic.DeepEqual(ms.Sink, duckv1.Destination{}) {
		errs = errs.Also(apis.ErrMissingField("sink"))
	} else if err := ms.Sink.Validate(ctx); err != nil {
		errs = errs.Also(err.ViaField("sink"))
	}

	//Validation for serviceAccountName field.
	if ms.ServiceAccountName == "" {
		errs = errs.Also(apis.ErrMissingField("serviceAccountName"))
	}
	return errs
}
