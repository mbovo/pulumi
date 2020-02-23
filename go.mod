module github.com/pulumi/pulumi

go 1.12

require (
	cloud.google.com/go v0.45.1
	github.com/Microsoft/go-winio v0.4.14
	github.com/aws/aws-sdk-go v1.28.9
	github.com/blang/semver v3.5.1+incompatible
	github.com/cheggaaa/pb v1.0.27
	github.com/d5/tengo/v2 v2.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/djherbis/times v1.0.1
	github.com/docker/docker v1.13.1
	github.com/dustin/go-humanize v1.0.0
	github.com/gofrs/flock v0.7.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/google/go-querystring v1.0.0
	github.com/gorilla/mux v1.6.2
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20171105060200-01f8541d5372
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/hcl/v2 v2.3.0
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93
	github.com/hashicorp/terraform-plugin-sdk v1.6.0
	github.com/ijc/Gotty v0.0.0-20170406111628-a8b993ba6abd
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/go-ps v0.0.0-20190716172923-621e5597135b
	github.com/mxschmitt/golang-combinations v1.0.0
	github.com/nbutton23/zxcvbn-go v0.0.0-20180912185939-ae427f1e4c1d
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pkg/errors v0.8.1
	github.com/pulumi/pulumi-aws v1.21.0 // indirect
	github.com/pulumi/pulumi-terraform-bridge v1.6.4
	github.com/pulumi/tf2pulumi v0.6.0
	github.com/rjeczalik/notify v0.9.2
	github.com/satori/go.uuid v1.2.0
	github.com/sergi/go-diff v1.0.0
	github.com/skratchdot/open-golang v0.0.0-20160302144031-75fb7ed4208c
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.1-0.20191106224347-f1bd0923b832
	github.com/texttheater/golang-levenshtein v0.0.0-20180516184445-d188e65d659e
	github.com/uber/jaeger-client-go v2.15.0+incompatible
	github.com/zclconf/go-cty v1.2.1
	go.uber.org/multierr v1.1.0
	gocloud.dev v0.18.0
	gocloud.dev/secrets/hashivault v0.18.0
	golang.org/x/crypto v0.0.0-20190923035154-9ee001bba392
	golang.org/x/net v0.0.0-20191009170851-d66e71096ffb
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190922100055-0a153f010e69
	google.golang.org/api v0.9.0
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.24.0
	gopkg.in/AlecAivazis/survey.v1 v1.4.1
	gopkg.in/src-d/go-git.v4 v4.8.1
	gopkg.in/yaml.v2 v2.2.4
	sourcegraph.com/sourcegraph/appdash v0.0.0-20190731080439-ebfcffb1b5c0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.4.3+incompatible
	gocloud.dev => github.com/pulumi/go-cloud v0.18.1-0.20191119155701-6a8381d0793f
)
