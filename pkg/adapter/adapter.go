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

package adapter

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/eventing/pkg/adapter/v2"
)

type envConfig struct {
	adapter.EnvConfig

	MongoDbCredentialsPath string `envconfig:"MONGODB_CREDENTIALS" required:"true"`
	Database               string `envconfig:"MONGODB_DATABASE" required:"true"`
	Collection             string `envconfig:"MONGODB_COLLECTION" required:"false"`
}

// NewEnvConfig creates an empty configuration
func NewEnvConfig() adapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter creates an adapter to convert incoming MongoDb changes events to CloudEvents and
// then sends them to the specified Sink
func NewAdapter(ctx context.Context, processed adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	fmt.Printf("Hey from mongodb adapter")
	time.Sleep(600 * time.Second)
	return nil
}
