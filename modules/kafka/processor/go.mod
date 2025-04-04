module github.com/seongpil0948/otel-kafka-pg/modules/kafka/processor

go 1.24

require (
	github.com/golang/protobuf v1.5.3
	github.com/golang/snappy v0.0.4
	github.com/seongpil0948/otel-kafka-pg/modules/common v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/modules/log v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/modules/trace v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/proto v0.0.0-00010101000000-000000000000
)

require (
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	golang.org/x/sys v0.18.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace (
	github.com/seongpil0948/otel-kafka-pg/modules/common => ../../common
	github.com/seongpil0948/otel-kafka-pg/modules/log => ../../log
	github.com/seongpil0948/otel-kafka-pg/modules/trace => ../../trace
	github.com/seongpil0948/otel-kafka-pg/proto => ../../../proto
)