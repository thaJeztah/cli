module github.com/docker/cli

// 'vendor.mod' enables use of 'go mod vendor' to managed 'vendor/' directory.
// There is no 'go.mod' file, as that would imply opting in for all the rules
// around SemVer, which this repo cannot abide by as it uses CalVer.

go 1.18

require (
	github.com/containerd/containerd v1.6.10
	github.com/creack/pty v1.1.11
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v23.0.0-beta.1+incompatible
	github.com/docker/docker-credential-helpers v0.7.0
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.5.0
	github.com/fvbommel/sortorder v1.0.2
	github.com/gogo/protobuf v1.3.2
	github.com/google/go-cmp v0.5.9
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/imdario/mergo v0.3.12
	github.com/manifoldco/promptui v0.6.0
	github.com/mattn/go-runewidth v0.0.13
	github.com/mitchellh/mapstructure v1.3.2
	github.com/moby/buildkit v0.10.6
	github.com/moby/patternmatcher v0.5.0
	github.com/moby/swarmkit/v2 v2.0.0-20221123162438-b17f02f0a054
	github.com/moby/sys/sequential v0.5.0
	github.com/moby/sys/signal v0.7.0
	github.com/moby/term v0.0.0-20221128092401-c43b287e0e0f
	github.com/morikuni/aec v1.0.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.3-0.20220303224323-02efb9a75ee1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/theupdateframework/notary v0.7.1-0.20210315103452-bf96a202a09a
	github.com/tonistiigi/go-rosetta v0.0.0-20200727161949-f79598599c5d
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/sys v0.2.0
	golang.org/x/term v0.1.0
	golang.org/x/text v0.4.0
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools/v3 v3.4.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/alecthomas/gometalinter v3.0.0+incompatible // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/client9/misspell v0.3.4 // indirect
	github.com/docker/go v1.5.1-1.0.20160303222718-d30aec9fd63c // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20180909121442-1003c8bd00dc // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/moby/sys/symlink v0.2.0 // indirect
	github.com/nicksnyder/go-i18n v1.10.1 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/tsenart/deadcode v0.0.0-20160724212837-210d2dc333e9 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.6 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/genproto v0.0.0-20220706185917-7780775163c4 // indirect
	google.golang.org/grpc v1.48.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/alecthomas/kingpin.v3-unstable v3.0.0-20171010053543-63abe20a23e2 // indirect
)
