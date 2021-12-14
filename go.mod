module github.com/trino-network/trino

go 1.16

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/briandowns/spinner v1.11.1
	github.com/cosmos/go-bip39 v1.0.0
	github.com/fatih/color v1.12.0
	github.com/goccy/go-yaml v1.9.2
	github.com/google/go-github/v37 v37.0.0
	github.com/gookit/color v1.4.2
	github.com/imdario/mergo v0.3.12
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/starport v0.18.6
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
