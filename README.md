**This is not an officially supported Google product.**

[![GoDoc](https://godoc.org/github.com/googleinterns/knative-source-mongodb?status.svg)](https://godoc.org/github.com/googleinterns/knative-source-mongodb)
[![Go Report Card](https://goreportcard.com/badge/googleinterns/knative-source-mongodb)](https://goreportcard.com/report/googleinterns/knative-source-mongodb)
[![LICENSE](https://img.shields.io/github/license/googleinterns/knative-source-mongodb.svg)](https://github.com/googleinterns/knative-source-mongodb/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/googleinterns/knative-source-mongodb.svg?branch=master)](https://travis-ci.org/googleinterns/knative-source-mongodb)

# Knative Source - MongoDB

The MongoDb Event source adds support of MongoDb ressources to Knative Eventing.

## Prerequisites

1. Install [Knative Eventing](https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-eventing-component) in your Kubernetes Cluster.

2. Install MongoDb and create a Replica Set using the following [instructions](https://www.mongodb.com/blog/post/running-mongodb-ops-manager-in-kubernetes#:~:text=The%20MongoDB%20Enterprise%20Kubernetes%20Operator%2C%20or%20simply%20the%20Operator%2C%20manages,changing%20these%20settings%20as%20needed).

## Usage

3. Create a secret containing the data needed to access your MongoDb service.
   For example:

   ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
    name: my-mongo-secret
    namespace: default
    stringData:
            URI: mongodb://USERNAME:PASSWORD@IP:PORT/USERDB
   ```
   where the IP is the IP of the main/principal pod of your replica set. USERDB is the database your user account partains to (can be `admin`).

4. Create the `MongoDbSource` custom objects, by configuring the required
   `database` has to be provided, but `collection` is optional.
   For example, with an Event-Display as a sink:

   ```yaml
    apiVersion: sources.google.com/v1alpha1
    kind: MongoDbSource
    metadata:
    name: mongodb-example-source
    namespace: default
    spec:
    database: db1
    collection: coll1  # optional
    secret:
        name: my-mongo-secret
    sink:
        ref:
          apiVersion: serving.knative.dev/v1
          kind: Service
          name: event-display
          namespace: mongodb
   ```
