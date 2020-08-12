module github.com/googleinterns/knative-source-mongodb

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/go-cmp v0.5.1
	go.mongodb.org/mongo-driver v1.1.2
	go.uber.org/zap v1.15.0
	k8s.io/api v0.18.7-rc.0
	k8s.io/apimachinery v0.18.7-rc.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.16.1-0.20200812122105-4dd7183c99ad
	knative.dev/pkg v0.0.0-20200812183506-c4576fd38ec2
	knative.dev/test-infra v0.0.0-20200812164905-ea529e42205b
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
)

replace k8s.io/code-generator => k8s.io/code-generator v0.17.6
