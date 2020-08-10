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
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/adapter/v2"
	reconcilersource "knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/system"

	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
)

// ReceiveAdapterArgs are the arguments needed to create a MongoDbSource Receive Adapter.
// Every field is required.
type ReceiveAdapterArgs struct {
	Image          string
	Labels         map[string]string
	Source         *v1alpha1.MongoDbSource
	CeSourcePrefix string
	SinkURL        string
	Configs        reconcilersource.ConfigAccessor
}

// MakeReceiveAdapter generates (but does not insert into K8s) the Receive Adapter Deployment for
// MongoDb sources.
func MakeReceiveAdapter(args *ReceiveAdapterArgs) (*v1.Deployment, error) {
	replicas := int32(1)

	env, err := makeEnv(args)
	if err != nil {
		return nil, fmt.Errorf("error generating env vars: %w", err)
	}

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
							Env:   env,
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
	}, nil
}

func makeEnv(args *ReceiveAdapterArgs) ([]corev1.EnvVar, error) {
	envs := []corev1.EnvVar{{
		Name:  adapter.EnvConfigSink,
		Value: args.SinkURL,
	}, {
		Name:  "CE_SOURCE_PREFIX",
		Value: args.CeSourcePrefix,
	}, {
		Name:  "SYSTEM_NAMESPACE",
		Value: system.Namespace(),
	}, {
		Name: adapter.EnvConfigNamespace,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	}, {
		Name:  adapter.EnvConfigName,
		Value: args.Source.Name,
	}, {
		Name:  "METRICS_DOMAIN",
		Value: "sources.google.com",
	}, {
		Name:  "MONGODB_DATABASE",
		Value: args.Source.Spec.Database,
	}, {
		Name:  "MONGODB_COLLECTION",
		Value: args.Source.Spec.Collection,
	}, {
		Name:  "MONGODB_CREDENTIALS",
		Value: "/etc/mongodb-credentials",
	}}

	// envs = append(envs, args.Configs.ToEnvVars()...)

	if args.Source.Spec.CloudEventOverrides != nil && args.Source.Spec.CloudEventOverrides.Extensions != nil {
		ceJSON, err := json.Marshal(args.Source.Spec.CloudEventOverrides.Extensions)
		if err != nil {
			return nil, fmt.Errorf("failure to marshal cloud event overrides %v: %v", args.Source.Spec.CloudEventOverrides.Extensions, err)
		}
		envs = append(envs, corev1.EnvVar{Name: adapter.EnvConfigCEOverrides, Value: string(ceJSON)})
	}
	return envs, nil

}
