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
	"fmt"

	mongoclient "github.com/googleinterns/knative-source-mongodb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// TestChangeStream wraps the fake mongo.ChangeStream.
type TestChangeStream struct {
	Data TestCSData
}

// TestCSData is the data used to configure the test MongoDb ChangeStream.
type TestCSData struct {
	DecodeErr error
	NewChange bson.M
	// nextFnNotExecuted is set so that Next returns true only once, then false, for each test.
	nextFnExecuted bool
}

// Verify that it satisfies the mongo.ChangeStream interface.
var _ mongoclient.ChangeStream = &TestChangeStream{}

// Next implements mongo.Client.ChangeStream.Next.
func (tCS *TestChangeStream) Next(ctx context.Context) bool {
	if tCS.Data.nextFnExecuted {
		return false
	}
	tCS.Data.nextFnExecuted = true
	return true
}

// Decode implements mongo.Client.ChangeStream.Decode.
func (tCS *TestChangeStream) Decode(val interface{}) error {
	if tCS.Data.DecodeErr != nil {
		return tCS.Data.DecodeErr
	}
	switch v := val.(type) {
	case *bson.M:
		*v = tCS.Data.NewChange
		return nil
	default:
		return fmt.Errorf("unknown type %T", val)
	}
}
