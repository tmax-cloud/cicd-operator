module github.com/tmax-cloud/cicd-operator

go 1.13

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/go-logr/logr v0.1.0
	github.com/gorilla/mux v1.7.4
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/operator-framework/operator-lib v0.1.0
	github.com/sourcegraph/go-diff v0.5.3
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tektoncd/pipeline v0.19.0
	gopkg.in/robfig/cron.v2 v2.0.0-20150107220207-be2e0b0deed5
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.18.8
	k8s.io/kubernetes v1.13.0
	knative.dev/pkg v0.0.0-20200922164940-4bf40ad82aab
	sigs.k8s.io/controller-runtime v0.6.4
	sigs.k8s.io/controller-tools v0.4.1 // indirect
)

replace (
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200922164940-4bf40ad82aab
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.4
)

// Pin k8s deps to v0.18.8
replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
)
