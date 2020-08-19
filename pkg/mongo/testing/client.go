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

package testing

import (
	"context"

	mongo "github.com/googleinterns/knative-source-mongodb/pkg/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestClientCreator returns a mongodb.CreateFn used to construct the mongo test client.
func TestClientCreator(value interface{}) mongo.CreateFn {
	var data TestClientData
	var ok bool
	if data, ok = value.(TestClientData); !ok {
		data = TestClientData{}
	}
	if data.CreateClientErr != nil {
		return func(opts ...*options.ClientOptions) (mongo.Client, error) {
			return nil, data.CreateClientErr
		}
	}

	return func(opts ...*options.ClientOptions) (mongo.Client, error) {
		return &testClient{
			data: data,
		}, nil
	}
}

// testClient is the fake test mongo client that uses data to know what error to return.
type testClient struct {
	data TestClientData
}

// TestClientData is the data used to configure the test MongoDb client.
type TestClientData struct {
	CreateClientErr error
	ConnectErr      error
	DisconnectErr   error
	CloseErr        error
	ListDbErr       error
	Databases       []string
}

// Verify that it satisfies the mongo.Client interface.
var _ mongo.Client = &testClient{}

// Connect implements mongo.Client.Connect.
func (tc *testClient) Connect(ctx context.Context) error {
	return tc.data.ConnectErr
}

// Database implements mongo.Client.Database.
func (tc *testClient) Database(name string, opts ...*options.DatabaseOptions) mongo.Database {
	return &testDatabase{
		data: TestDbData{},
	}
}

// Disconnect implements mongo.Client.Disconnect.
func (tc *testClient) Disconnect(ctx context.Context) error {
	return tc.data.DisconnectErr
}

// ListDatabaseNames implements mongo.Client.ListDatabaseNames.
func (tc *testClient) ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) ([]string, error) {
	if tc.data.ListDbErr != nil {
		return nil, tc.data.ListDbErr
	}
	return tc.data.Databases, nil
}
