
kafka-setup:
	docker rm -f kafka 2>/dev/null || true
	docker run -d --name kafka \
	  -p 9092:9092 \
	  -e KAFKA_NODE_ID=1 \
	  -e KAFKA_PROCESS_ROLES=broker,controller \
	  -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093 \
	  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
	  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
	  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
	  -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
	  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
	  -e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
	  -e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
	  apache/kafka && \
	sleep 5 && \
	docker exec kafka /opt/kafka/bin/kafka-topics.sh \
	  --create --topic foo \
	  --bootstrap-server localhost:9092 \
	  --partitions 1 \
	  --replication-factor 1

run:
	go run kafka.go

start: kafka-setup run
