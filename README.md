**This is not an officially supported Google product.**

[![GoDoc](https://godoc.org/github.com/googleinterns/knative-source-mongodb?status.svg)](https://pkg.go.dev/mod/github.com/googleinterns/knative-source-mongodb)
[![Go Report Card](https://goreportcard.com/badge/googleinterns/knative-source-mongodb)](https://goreportcard.com/report/googleinterns/knative-source-mongodb)
[![LICENSE](https://img.shields.io/github/license/googleinterns/knative-source-mongodb.svg)](https://github.com/googleinterns/knative-source-mongodb/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/googleinterns/knative-source-mongodb.svg?branch=master)](https://travis-ci.org/googleinterns/knative-source-mongodb)

# Knative Source - MongoDB

The MongoDb Event Source adds support of MongoDB resources to Knative Eventing.

## Prerequisites

1. Install [Knative Eventing](https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-eventing-component) in your Kubernetes Cluster.

2. Either:

    * Install MongoDb on your Kubernetes Cluster and create a Replica Set. Instructions [available here](https://www.mongodb.com/blog/post/running-mongodb-ops-manager-in-kubernetes#:~:text=The%20MongoDB%20Enterprise%20Kubernetes%20Operator%2C%20or%20simply%20the%20Operator%2C%20manages,changing%20these%20settings%20as%20needed).

    * Create a MongoDb Cluster on Atlas, through GCP for example. Link [available here](https://console.cloud.google.com/marketplace/details/gc-launcher-for-mongodb-atlas/mongodb-atlas).

3. Install [ko](https://github.com/google/ko) and then execute:

    ```
    ko apply -f ./config
    ```

## Usage

1. Create a secret containing the data needed to access your MongoDb service.
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
   The URI is the connection string of your Mongo Database or Cluster. USERDB is the database your user account pertains to (can be `admin`).

2. Create the `MongoDbSource` custom object: provide the required `database` field, provide the `collection` field (optional), and reference the `secret` just created as well as the destination `sink`.
   For example, with a Knative Service as a sink:

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
   ```
