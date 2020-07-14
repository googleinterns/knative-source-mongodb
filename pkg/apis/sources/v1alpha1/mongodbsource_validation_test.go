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
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/webhook/resourcesemantics"

	"knative.dev/pkg/apis"
)

func TestMongoDbSourceValidation(t *testing.T) {
	testCases := map[string]struct {
		cr   resourcesemantics.GenericCRD
		want *apis.FieldError
	}{
		"missing all": {
			cr: &MongoDbSource{
				Spec: MongoDbSourceSpec{},
			},
			want: func() *apis.FieldError {
				var errs *apis.FieldError
				fe := apis.ErrMissingField("spec.database, spec.secret, spec.serviceAccountName, spec.sink")
				errs = errs.Also(fe)
				return errs
			}(),
		},
		"Incorrect Sink": {
			cr: &MongoDbSource{
				Spec: MongoDbSourceSpec{
					ServiceAccountName: "google",
					Secret: corev1.LocalObjectReference{
						Name: "pwd",
					},
					Database:   "db",
					Collection: "col1",
					SourceSpec: duckv1.SourceSpec{
						Sink: duckv1.Destination{
							Ref: &duckv1.KReference{
								APIVersion: "foo",
								// Kind:       "bar",
								Namespace: "baz",
								Name:      "qux",
							},
						},
					},
				},
			},
			want: func() *apis.FieldError {
				var errs *apis.FieldError
				fe := apis.ErrMissingField("spec.sink.ref.kind")
				errs = errs.Also(fe)
				return errs
			}(),
		},
		"No Secret": {
			cr: &MongoDbSource{
				Spec: MongoDbSourceSpec{
					ServiceAccountName: "google",
					Secret: corev1.LocalObjectReference{
						Name: "",
					},
					Database:   "db",
					Collection: "col1",
					SourceSpec: duckv1.SourceSpec{
						Sink: duckv1.Destination{
							Ref: &duckv1.KReference{
								APIVersion: "foo",
								Kind:       "bar",
								Namespace:  "baz",
								Name:       "qux",
							},
						},
					},
				},
			},
			want: func() *apis.FieldError {
				var errs *apis.FieldError
				fe := apis.ErrMissingField("spec.secret")
				errs = errs.Also(fe)
				return errs
			}(),
		},
		"All fields present": {
			cr: &MongoDbSource{
				Spec: MongoDbSourceSpec{
					ServiceAccountName: "google",
					Secret: corev1.LocalObjectReference{
						Name: "pwd",
					},
					Database:   "db",
					Collection: "col1",
					SourceSpec: duckv1.SourceSpec{
						Sink: duckv1.Destination{
							Ref: &duckv1.KReference{
								APIVersion: "foo",
								Kind:       "bar",
								Namespace:  "baz",
								Name:       "qux",
							},
						},
					},
				},
			},
			want: nil,
		},
	}

	for n, test := range testCases {
		t.Run(n, func(t *testing.T) {
			got := test.cr.Validate(context.Background())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("%s: validate (-want, +got) = %v", n, diff)
			}
		})
	}
}
