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
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	availableDeployment = &appsv1.Deployment{
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	condReady = apis.Condition{
		Type:   MongoDbConditionReady,
		Status: corev1.ConditionTrue,
	}
)

func TestMongoDbSourceGetConditionSet(t *testing.T) {
	r := &MongoDbSource{}

	if got, want := r.GetConditionSet().GetTopLevelConditionType(), apis.ConditionReady; got != want {
		t.Errorf("GetTopLevelCondition=%v, want=%v", got, want)
	}
}

func TestMongoDbGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		ms        *MongoDbSourceStatus
		condQuery apis.ConditionType
		want      *apis.Condition
	}{{
		name: "single condition",
		ms: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						condReady,
					},
				},
			},
		},
		condQuery: apis.ConditionReady,
		want:      &condReady,
	}, {
		name: "unknown condition",
		ms: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{
						condReady,
					},
				},
			},
		},
		condQuery: apis.ConditionType("foo"),
		want:      nil,
	}, {
		name: "mark deployed",
		ms: func() *MongoDbSourceStatus {
			m := &MongoDbSourceStatus{}
			m.InitializeConditions()
			m.PropagateDeploymentAvailability(availableDeployment)
			return m
		}(),
		condQuery: MongoDbConditionReady,
		want: &apis.Condition{
			Type:   MongoDbConditionReady,
			Status: corev1.ConditionUnknown,
		},
	}, {
		name: "mark sink and deployed",
		ms: func() *MongoDbSourceStatus {
			m := &MongoDbSourceStatus{}
			m.InitializeConditions()
			m.MarkSink(apis.HTTP("example"))
			m.MarkConnectionSuccess()
			m.PropagateDeploymentAvailability(availableDeployment)
			return m
		}(),
		condQuery: MongoDbConditionReady,
		want: &apis.Condition{
			Type:   MongoDbConditionReady,
			Status: corev1.ConditionTrue,
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ms.GetCondition(test.condQuery)
			ignoreTime := cmpopts.IgnoreFields(apis.Condition{},
				"LastTransitionTime", "Severity")
			if diff := cmp.Diff(test.want, got, ignoreTime); diff != "" {
				t.Errorf("unexpected condition (-want, +got) = %v", diff)
			}
		})
	}
}

func TestMongoDbInitializeConditions(t *testing.T) {
	tests := []struct {
		name string
		ms   *MongoDbSourceStatus
		want *MongoDbSourceStatus
	}{{
		name: "empty",
		ms:   &MongoDbSourceStatus{},
		want: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   MongoDbConditionConnectionEstablished,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionDeployed,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionSinkProvided,
						Status: corev1.ConditionUnknown,
					}},
				},
			},
		},
	}, {
		name: "one false",
		ms: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   MongoDbConditionSinkProvided,
						Status: corev1.ConditionFalse,
					}},
				},
			},
		},
		want: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   MongoDbConditionConnectionEstablished,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionDeployed,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionSinkProvided,
						Status: corev1.ConditionFalse,
					}},
				},
			},
		},
	}, {
		name: "one true",
		ms: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   MongoDbConditionSinkProvided,
						Status: corev1.ConditionTrue,
					}},
				},
			},
		},
		want: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   MongoDbConditionConnectionEstablished,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionDeployed,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionSinkProvided,
						Status: corev1.ConditionTrue,
					}},
				},
			},
		},
	}, {
		name: "marksink",
		ms: func() *MongoDbSourceStatus {
			status := MongoDbSourceStatus{}
			status.MarkSink(apis.HTTP("sink"))
			return &status
		}(),
		want: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   MongoDbConditionConnectionEstablished,
						Status: corev1.ConditionUnknown,
					},{
						Type:   MongoDbConditionDeployed,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionReady,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionSinkProvided,
						Status: corev1.ConditionTrue,
					}},
				},
				SinkURI: apis.HTTP("sink"),
			},
		},
	}, {
		name: "marknosink",
		ms: func() *MongoDbSourceStatus {
			status := MongoDbSourceStatus{}
			status.MarkNoSink("nothere", "")
			return &status
		}(),
		want: &MongoDbSourceStatus{
			SourceStatus: duckv1.SourceStatus{
				Status: duckv1.Status{
					Conditions: []apis.Condition{{
						Type:   MongoDbConditionConnectionEstablished,
						Status: corev1.ConditionUnknown,
					},{
						Type:   MongoDbConditionDeployed,
						Status: corev1.ConditionUnknown,
					}, {
						Type:   MongoDbConditionReady,
						Status: corev1.ConditionFalse,
					}, {
						Type:   MongoDbConditionSinkProvided,
						Status: corev1.ConditionFalse,
					}},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.ms.InitializeConditions()
			ignore := cmpopts.IgnoreFields(
				apis.Condition{},
				"LastTransitionTime", "Message", "Reason", "Severity")
			if diff := cmp.Diff(test.want, test.ms, ignore); diff != "" {
				t.Errorf("unexpected conditions (-want, +got) = %v", diff)
			}
		})
	}
}
