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

package utils

import (
	"go.mongodb.org/mongo-driver/bson"
)

// ChangeObject gathers the information obtianed from the change object issued by
// the mongodb change stream.
type ChangeObject struct {
	ID            string
	OperationType string
	Payload       *bson.M
}

// DecodeChangeBson decodes Bson change object.
func DecodeChangeBson(data bson.M) (*ChangeObject, error) {
	id := data["_idd"].(bson.M)["_data"].(string)
	operationType := data["operationType"].(string)

	// Add payload if replace or insert, else add document key.
	var payload bson.M
	if operationType == "delete" {
		payload = data["documentKey"].(bson.M)
	} else {
		payload = data["fullDocument"].(bson.M)
	}

	return &ChangeObject{
		ID:            id,
		OperationType: operationType,
		Payload:       &payload,
	}, nil
}
