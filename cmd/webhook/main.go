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

	mongodbv1alpha1 "github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/configmaps"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	mongodbv1alpha1.SchemeGroupVersion.WithKind("MongoDbSource"): &mongodbv1alpha1.MongoDbSource{},
}

var callbacks = map[schema.GroupVersionKind]validation.Callback{}

const admissionWebhookName = "webhook"

// NewDefaultingAdmissionController sets up mutating webhook.
func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return defaulting.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"defaulting.webhook.sources.google.com",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

// NewValidationAdmissionController sets up validation webhook.
func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"validation.webhook.sources.google.com",

		// The path on which to serve the webhook.
		"/validation",

		// The resources to validate.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,

		// Extra validating callbacks to be applied to resources.
		callbacks,
	)
}

// NewConfigValidationController sets up ConfigMap validation webhook.
func NewConfigValidationController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return configmaps.NewAdmissionController(ctx,

		// Name of the configmap webhook.
		"config.webhook.sources.google.com",

		// The path on which to serve the webhook.
		"/config-validation",

		// The configmaps to validate.
		configmap.Constructors{
			logging.ConfigMapName(): logging.NewConfigFromConfigMap,
			metrics.ConfigMapName(): metrics.NewObservabilityConfigFromConfigMap,
		},
	)
}

func main() {
	// Set up a signal context with our webhook options
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: admissionWebhookName,
		Port:        8443,
		SecretName:  "webhooks-certs",
	})

	sharedmain.WebhookMainWithContext(ctx, admissionWebhookName,
		certificates.NewController,
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
		NewConfigValidationController,
	)
}
