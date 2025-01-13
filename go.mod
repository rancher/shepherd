module github.com/rancher/shepherd

go 1.22.0

toolchain go1.22.7

replace (
	github.com/docker/distribution => github.com/docker/distribution v2.8.2+incompatible // rancher-machine requires a replace is set
	k8s.io/api => k8s.io/api v0.31.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.31.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.31.1
	k8s.io/apiserver => k8s.io/apiserver v0.31.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.31.1
	k8s.io/client-go => k8s.io/client-go v0.31.1
)

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/aws/aws-sdk-go v1.50.38
	github.com/creasty/defaults v1.5.2
	github.com/gorilla/websocket v1.5.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hc-install v0.6.2
	github.com/hashicorp/terraform-exec v0.20.0
	github.com/hashicorp/terraform-json v0.20.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.13.5
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.52.0
	github.com/rancher/aks-operator v1.10.0
	github.com/rancher/apiserver v0.0.0-20240604183424-8c448886365e
	github.com/rancher/eks-operator v1.10.0
	github.com/rancher/fleet/pkg/apis v0.11.0
	github.com/rancher/gke-operator v1.10.0
	github.com/rancher/lasso v0.0.0-20240924233157-8f384efc8813
	github.com/rancher/norman v0.0.0-20241001183610-78a520c160ab
	github.com/rancher/rancher/pkg/apis v0.0.0-20241127174121-c051d99dcded
	github.com/rancher/rke v1.7.0-rc.5
	github.com/rancher/system-upgrade-controller/pkg/apis v0.0.0-20210727200656-10b094e30007
	github.com/rancher/wrangler v1.1.2
	github.com/rancher/wrangler/v3 v3.1.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	golang.org/x/crypto v0.28.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.31.1
	k8s.io/apimachinery v0.31.1
	k8s.io/apiserver v0.31.1
	k8s.io/cli-runtime v0.31.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-aggregator v0.31.1
	k8s.io/kubectl v0.31.1
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8
	sigs.k8s.io/cluster-api v1.8.3
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.9.10 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371 // indirect
	github.com/apparentlymart/go-cidr v1.0.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/containerd v1.6.27 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v20.10.25+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-ini/ini v1.37.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matryer/moq v0.3.4 // indirect
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.4.0 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.7.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b // indirect
	github.com/opencontainers/runc v1.1.14 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/zclconf/go-cty v1.14.1 // indirect
	go.etcd.io/etcd/api/v3 v3.5.15 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.15 // indirect
	go.etcd.io/etcd/client/v2 v2.305.13 // indirect
	go.etcd.io/etcd/client/v3 v3.5.15 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.starlark.net v0.0.0-20230525235612-a134d8f9ddca // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/term v0.25.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240930140551-af27646dc61f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.31.1 // indirect
	k8s.io/code-generator v0.31.1 // indirect
	k8s.io/component-base v0.31.1 // indirect
	k8s.io/gengo v0.0.0-20240826214909-a7b603a56eb7 // indirect
	k8s.io/gengo/v2 v2.0.0-20240228010128-51d4e06bde70 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	k8s.io/kubernetes v1.31.1 // indirect
	k8s.io/pod-security-admission v0.31.1 // indirect
	sigs.k8s.io/cli-utils v0.37.2 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.17.2 // indirect
	sigs.k8s.io/kustomize/kyaml v0.17.1 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)
