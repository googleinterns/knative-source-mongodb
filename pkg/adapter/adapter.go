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
	"io/ioutil"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

type envConfig struct {
	adapter.EnvConfig

	MongoDbCredentialsPath string `envconfig:"MONGODB_CREDENTIALS" required:"true"`
	Database               string `envconfig:"MONGODB_DATABASE" required:"true"`
	Collection             string `envconfig:"MONGODB_COLLECTION" required:"false"`
	EventSource            string `envconfig:"EVENT_SOURCE" required:"false"`
	SinkURL                string `envconfig:"K_SINK" required:"false"`
}

type mongoDbAdapter struct {
	namespace string
	ce        cloudevents.Client
	client    *mongo.Client
	source    string
	db        *mongo.Database
	coll      *mongo.Collection
	sinkurl   string
}

// NewEnvConfig creates an empty environement variables configuration.
func NewEnvConfig() adapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter creates an adapter to convert incoming MongoDb changes events to CloudEvents and
// then sends them to the specified Sink.
func NewAdapter(ctx context.Context, processed adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := processed.(*envConfig)

	rawURI, err := ioutil.ReadFile(env.MongoDbCredentialsPath + "/URI")
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Unable to get MongoDb URI field", zap.Any("secretPath", env.MongoDbCredentialsPath+"/URI"))
	}
	URI := string(rawURI)

	return newAdapter(ctx, env, ceClient, URI)
}

func newAdapter(ctx context.Context, env *envConfig, ceClient cloudevents.Client, URI string) adapter.Adapter {
	// Create new Client.
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Error creating mongo client", zap.Error(err))
	}

	var collection *mongo.Collection = nil
	if env.Collection != "" {
		collection = client.Database(env.Database).Collection(env.Collection)
	}

	return &mongoDbAdapter{
		namespace: env.Namespace,
		ce:        ceClient,
		client:    client,
		db:        client.Database(env.Database),
		coll:      collection,
		source:    env.EventSource,
		sinkurl:   env.SinkURL,
	}
}

// Start creates the watch stream that will watch for dataSource changes.
func (a *mongoDbAdapter) Start(ctx context.Context) error {
	// Connect to Client.
	err := a.client.Connect(ctx)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Error connecting to database", zap.Error(err))
	}
	defer a.client.Disconnect(ctx)

	// Create a watch stream for either the database or collection.
	var stream *mongo.ChangeStream = nil
	if a.coll != nil {
		stream, err = a.coll.Watch(ctx, mongo.Pipeline{})
	} else {
		stream, err = a.db.Watch(ctx, mongo.Pipeline{})
	}
	if err != nil {
		return err
	}
	defer stream.Close(ctx)

	// Watch and process changes.
	a.processChanges(ctx, stream)
	return nil
}

// processChanges processes the new incoming change, creates a cloud event and sends it.
func (a *mongoDbAdapter) processChanges(ctx context.Context, stream *mongo.ChangeStream) {
	// For each new change recorded.
	for stream.Next(ctx) {
		var data bson.M
		if err := stream.Decode(&data); err != nil {
			panic(err)
		}
		// Create corresponding event.
		event, err := makeCloudEvent(data)
		if err != nil {
			logging.FromContext(ctx).Desugar().Error("Failed to create event", zap.Error(err))
		}

		// Send event using client.
		ctx := cloudevents.ContextWithTarget(ctx, a.sinkurl)

		// Send that Event.
		if result := a.ce.Send(ctx, *event); cloudevents.IsUndelivered(result) {
			logging.FromContext(ctx).Desugar().Error("Failed to send event")
		}
	}
}

// makeCloudEvent makes a cloud event out of the change object recevied.
func makeCloudEvent(data bson.M) (*cloudevents.Event, error) {
	// Create Event.
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	// Set cloud event specs and attributes.
	event.SetID(data["_id"].(bson.M)["_data"].(string))
	event.SetSource(data["ns"].(bson.M)["db"].(string))
	event.SetType(data["operationType"].(string))

	// Add payload if replace or insert, else add document key.
	if data["operationType"].(string) == "delete" {
		event.SetData(cloudevents.ApplicationJSON, data["documentKey"].(bson.M))
	} else {
		event.SetData(cloudevents.ApplicationJSON, data["fullDocument"].(bson.M))
	}
	return &event, nil
}
