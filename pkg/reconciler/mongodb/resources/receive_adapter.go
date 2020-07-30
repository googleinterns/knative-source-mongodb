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

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
)

// ReceiveAdapterArgs are the arguments needed to create a MongoDbSource Receive Adapter.
// Every field is required.
type ReceiveAdapterArgs struct {
	Image       string
	Labels      map[string]string
	Source      *v1alpha1.MongoDbSource
	EventSource string
	SinkURL     string
}

// MakeReceiveAdapter generates (but does not insert into K8s) the Receive Adapter Deployment for
// MongoDb sources.
func MakeReceiveAdapter(args *ReceiveAdapterArgs) *v1.Deployment {
	replicas := int32(1)
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: args.Source.Namespace,
			Name:      kmeta.ChildName(fmt.Sprintf("mongodbsource-%s-", args.Source.Name), string(args.Source.GetUID())),
			Labels:    args.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Source),
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: args.Labels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: args.Labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: args.Source.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: args.Image,
							Env:   makeEnv(args.EventSource, args.SinkURL, &args.Source.Spec),
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
								Secret: &corev1.SecretVolumeSource{
									SecretName: args.Source.Spec.Secret.Name,
								},
							},
						},
					},
				},
			},
		},
	}
}

func makeEnv(eventSource string, sinkURI string, spec *v1alpha1.MongoDbSourceSpec) []corev1.EnvVar {
	return []corev1.EnvVar{{
		Name:  "EVENT_SOURCE",
		Value: eventSource,
	}, {
		Name: "NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	}, {
		Name: "NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	}, {
		Name:  "K_SINK",
		Value: sinkURI,
	}, {
		Name:  "METRICS_DOMAIN",
		Value: "sources.google.com",
	}, {
		Name:  "K_METRICS_CONFIG",
		Value: "",
	}, {
		Name:  "K_LOGGING_CONFIG",
		Value: "",
	}, {
		Name:  "MONGODB_DATABASE",
		Value: spec.Database,
	}, {
		Name:  "MONGODB_COLLECTION",
		Value: spec.Collection,
	}, {
		Name:  "MONGODB_CREDENTIALS",
		Value: "/etc/mongodb-credentials",
	},
	}
}
