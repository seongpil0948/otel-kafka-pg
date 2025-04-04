module github.com/seongpil0948/otel-kafka-pg/modules/kafka/consumer

go 1.24

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
	github.com/seongpil0948/otel-kafka-pg/modules/common v0.0.0
	github.com/seongpil0948/otel-kafka-pg/modules/kafka/processor v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/modules/log v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/modules/trace v0.0.0-00010101000000-000000000000
)

require (
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250102185135-69823020774d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250102185135-69823020774d // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/seongpil0948/otel-kafka-pg/modules/common => ../../common
	github.com/seongpil0948/otel-kafka-pg/modules/kafka/processor => ../processor
	github.com/seongpil0948/otel-kafka-pg/modules/log => ../../log
	github.com/seongpil0948/otel-kafka-pg/modules/trace => ../../trace
	github.com/seongpil0948/otel-kafka-pg/proto => ../../../proto
)
