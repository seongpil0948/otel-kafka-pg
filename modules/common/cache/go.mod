module github.com/seongpil0948/otel-kafka-pg/modules/common/cache

go 1.24.2

replace (
	github.com/seongpil0948/otel-kafka-pg/modules/common/config => ../config
	github.com/seongpil0948/otel-kafka-pg/modules/common/logger => ../logger
	github.com/seongpil0948/otel-kafka-pg/modules/common/redis => ../redis
)
