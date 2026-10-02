module github.com/two-barrels/ari-proxy/v6

go 1.25.0

require (
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/inconshreveable/log15 v2.16.0+incompatible
	github.com/nats-io/nats.go v1.49.0
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/rotisserie/eris v0.5.4
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/subosito/gotenv v1.6.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.18.4 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/two-barrels/ari/v6 v6.0.0-20251024161400-681c62bc07e7
	golang.org/x/term v0.40.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require golang.org/x/exp v0.0.0-20260218203240-3dfff04db8fa

require (
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
)

// Development-only: the required pseudo-version is not the release candidate.
// Replace it with the approved two-barrels ARI v6 tag and remove this override before publication.
replace github.com/two-barrels/ari/v6 => ../ari
