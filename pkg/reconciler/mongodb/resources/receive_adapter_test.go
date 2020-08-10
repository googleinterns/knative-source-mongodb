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

package resources

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	_ "knative.dev/pkg/metrics/testing"
	_ "knative.dev/pkg/system/testing"
)

func TestMakeReceiveAdapter(t *testing.T) {
	name := "source-name"
	src := &v1alpha1.MongoDbSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "source-namespace",
			UID:       "sampleUID",
		},
		Spec: v1alpha1.MongoDbSourceSpec{
			ServiceAccountName: "source-svc-acct",
			Secret:             corev1.LocalObjectReference{Name: "my-mongo-secret"},
			Database:           "db",
			Collection:         "coll",
		},
	}

	got, err := MakeReceiveAdapter(&ReceiveAdapterArgs{
		Image:  "test-image",
		Source: src,
		Labels: map[string]string{
			"test-key1": "test-value1",
			"test-key2": "test-value2",
		},
		SinkURL:        "sink-uri",
		CeSourcePrefix: "mongodb://",
	})
	if err != nil {
		t.Errorf("Couldn't make Receive Adapter %w", err)
	}
	if got != nil {
		fmt.Printf("yo")
	}
	one := int32(1)
	trueValue := true

	want := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "source-namespace",
			Name:      kmeta.ChildName(fmt.Sprintf("mongodbsource-%s-", name), "sampleUID"),
			Labels: map[string]string{
				"test-key1": "test-value1",
				"test-key2": "test-value2",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "sources.google.com/v1alpha1",
					Kind:               "MongoDbSource",
					Name:               name,
					UID:                "sampleUID",
					Controller:         &trueValue,
					BlockOwnerDeletion: &trueValue,
				},
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"test-key1": "test-value1",
					"test-key2": "test-value2",
				},
			},
			Replicas: &one,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"test-key1": "test-value1",
						"test-key2": "test-value2",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "source-svc-acct",
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: "test-image",
							Env: []corev1.EnvVar{
								{
									Name:  "K_SINK",
									Value: "sink-uri",
								}, {
									Name:  "CE_SOURCE_PREFIX",
									Value: "mongodb://",
								}, {
									Name:  "SYSTEM_NAMESPACE",
									Value: "knative-testing",
								}, {
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								}, {
									Name:  "NAME",
									Value: name,
								}, {
									Name:  "METRICS_DOMAIN",
									Value: "sources.google.com",
								}, {
									Name:  "MONGODB_DATABASE",
									Value: "db",
								}, {
									Name:  "MONGODB_COLLECTION",
									Value: "coll",
								}, {
									Name:  "MONGODB_CREDENTIALS",
									Value: "/etc/mongodb-credentials",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mongodb-credentials",
									MountPath: "/etc/mongodb-credentials",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "mongodb-credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{SecretName: "my-mongo-secret"},
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected deploy (-want, +got) = %v", diff)
	}
}
