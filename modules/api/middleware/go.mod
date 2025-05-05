module github.com/seongpil0948/otel-kafka-pg/modules/api/middleware

go 1.24.2

replace (
	github.com/seongpil0948/otel-kafka-pg/modules/common/cache => ../../common/cache
	github.com/seongpil0948/otel-kafka-pg/modules/common/config => ../../common/config
	github.com/seongpil0948/otel-kafka-pg/modules/common/logger => ../../common/logger
	github.com/seongpil0948/otel-kafka-pg/modules/common/redis => ../../common/redis
)
