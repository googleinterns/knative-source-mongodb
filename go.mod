module github.com/googleinterns/knative-source-mongodb

go 1.14

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.1 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.1-0.20191009090205-6c0755d89d1e // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	go.mongodb.org/mongo-driver v1.1.2
	go.uber.org/zap v1.14.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/pkg v0.0.0-20200702222342-ea4d6e985ba0
	knative.dev/test-infra v0.0.0-20200630141629-15f40fe97047
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
)

replace k8s.io/code-generator => k8s.io/code-generator v0.17.6
