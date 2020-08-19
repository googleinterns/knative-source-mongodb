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

package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateFn is a factory function to create a Mongo client.
type CreateFn func(opts ...*options.ClientOptions) (Client, error)

// NewClient creates a new wrapped Mongo client.
func NewClient(opts ...*options.ClientOptions) (Client, error) {
	client, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &mongoClient{
		client: client,
	}, nil
}

// mongoClient wraps mongo.Client. It is the client that will be used everywhere except unit tests.
type mongoClient struct {
	client *mongo.Client
}

// Verify that it satisfies the mongo.Client interface.
var _ Client = &mongoClient{}

// Connect implements mongo.Client.Connect.
func (mc *mongoClient) Connect(ctx context.Context) error {
	return mc.client.Connect(ctx)
}

// Database implements mongo.Client.Database.
func (mc *mongoClient) Database(name string, opts ...*options.DatabaseOptions) Database {
	return &database{
		database: mc.client.Database(name, opts...),
	}
}

// Disconnect implements mongo.Client.Disconnect.
func (mc *mongoClient) Disconnect(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

// ListDatabaseNames implements mongo.Client.ListDatabaseNames.
func (mc *mongoClient) ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) ([]string, error) {
	return mc.client.ListDatabaseNames(ctx, filter, opts...)
}
