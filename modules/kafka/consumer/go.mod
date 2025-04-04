module github.com/seongpil0948/otel-kafka-pg/modules/kafka/consumer

go 1.24

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
	github.com/seongpil0948/otel-kafka-pg/modules/common v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/modules/kafka/processor v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/modules/log v0.0.0-00010101000000-000000000000
	github.com/seongpil0948/otel-kafka-pg/modules/trace v0.0.0-00010101000000-000000000000
)

replace (
	github.com/seongpil0948/otel-kafka-pg/modules/common => ../../common
	github.com/seongpil0948/otel-kafka-pg/modules/kafka/processor => ../processor
	github.com/seongpil0948/otel-kafka-pg/modules/log => ../../log
	github.com/seongpil0948/otel-kafka-pg/modules/trace => ../../trace
)