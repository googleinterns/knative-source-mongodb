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
	"errors"
	"fmt"
	"io/ioutil"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/googleinterns/knative-source-mongodb/pkg/utils"
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
	EventSource            string `envconfig:"EVENT_SOURCE" required:"true"`
}

type mongoDbAdapter struct {
	namespace       string
	ceClient        cloudevents.Client
	ceSource        string
	database        string
	collection      string
	credentialsPath string
	logger          *zap.SugaredLogger
}

// dataSource interface to interact with either a mongo.database or a mongo.collection.
type dataSource interface {
	Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
}

// NewEnvConfig creates an empty environement variables configuration.
func NewEnvConfig() adapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter creates an adapter to convert incoming MongoDb changes events to CloudEvents and
// then sends them to the specified Sink.
func NewAdapter(ctx context.Context, processed adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	logger := logging.FromContext(ctx)
	env := processed.(*envConfig)

	return &mongoDbAdapter{
		namespace:       env.Namespace,
		ceClient:        ceClient,
		ceSource:        env.EventSource,
		database:        env.Database,
		collection:      env.Collection,
		credentialsPath: env.MongoDbCredentialsPath,
		logger:          logger,
	}
}

// Start connects to the database and creates the watch stream that will watch for dataSource changes.
func (a *mongoDbAdapter) Start(ctx context.Context) error {
	// Read the Credentials.
	rawURI, err := ioutil.ReadFile(a.credentialsPath + "/URI")
	if err != nil {
		return fmt.Errorf("unable to get MongoDb URI field: secretPath %s/URI : %w", a.credentialsPath, err)
	}
	URI := string(rawURI)
	if URI == "" {
		return errors.New("MongoDb URI field is empty")
	}

	// Create new Client.
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		return fmt.Errorf("error creating mongo client: %w", err)
	}

	// Get dataSource: either a mongo.Collection or a mongo.Database.
	var dataSource dataSource
	if a.collection != "" {
		dataSource = client.Database(a.database).Collection(a.collection)
	} else {
		dataSource = client.Database(a.database)
	}

	// Connect to Client.
	err = client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	defer client.Disconnect(ctx)

	// Create a watch stream for either the database or collection.
	stream, err := dataSource.Watch(ctx, mongo.Pipeline{})
	if err != nil {
		return fmt.Errorf("error setting up changeStream: %w", err)
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
			a.logger.Desugar().Error("Error decoding the change stream", zap.Error(err))
			continue
		}
		// Create corresponding event.
		event, err := a.makeCloudEvent(data)
		if err != nil {
			a.logger.Desugar().Error("Failed to create event", zap.Error(err))
			continue
		}

		// Send that Event.
		if result := a.ceClient.Send(ctx, *event); cloudevents.IsUndelivered(result) {
			a.logger.Desugar().Error("Failed to send event", zap.Any("result", result))
		}
	}
	return
}

// makeCloudEvent makes a cloud event out of the change object recevied.
func (a *mongoDbAdapter) makeCloudEvent(data bson.M) (*cloudevents.Event, error) {
	// Create Event.
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	// Set cloud event specs and attributes. TODO: issue #43
	// 		ID     -> id of mongo change object
	// 		Source -> Source environment variable.
	// 		Type   -> type of change either insert, delete or update.
	//		Data   -> data payload containing either id only for
	//                deletion or full object for other changes.
	change, err := utils.DecodeChangeBson(data)
	if err != nil {
		return nil, fmt.Errorf("error decoding bson change object: %w", err)
	}
	event.SetID(change.ID)
	event.SetType(change.OperationType)
	event.SetSource(a.ceSource)
	event.SetData(cloudevents.ApplicationJSON, change.Payload)

	return &event, nil
}
