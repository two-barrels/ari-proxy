module github.com/CyCoreSystems/ari-proxy/v5

go 1.23.0

toolchain go1.23.8

replace github.com/CyCoreSystems/ari/v5 v5.3.1 => github.com/two-barrels/ari/v5 v5.3.2-0.20250512190349-08b289189f38

require (
	github.com/CyCoreSystems/ari/v5 v5.3.1
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/inconshreveable/log15 v2.16.0+incompatible
	github.com/nats-io/nats.go v1.42.0
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/rotisserie/eris v0.5.4
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/subosito/gotenv v1.6.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/cast v1.8.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
