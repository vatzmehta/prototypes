package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

/* Run kafka locally in docker container using following command

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

*/

const TOTAL_MSGS = 10

// run kafka locally at 9092, create "foo" as a topic before starting the program
func main() {

	seeds := []string{"localhost:9092"}

	producer, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
	)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("my-consumer-group"),
		kgo.ConsumeTopics("foo"),
	)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	var wg sync.WaitGroup

	for i := 0; i < TOTAL_MSGS; i++ {
		msg := "test " + strconv.Itoa(i)
		wg.Add(1)
		go Producer(producer, &wg, []byte(msg))
	}

	wg.Wait()
	fmt.Println("Starting consuming")
	Consumer(consumer)

}

func Producer(c1 *kgo.Client, wg *sync.WaitGroup, message []byte) {
	ctx := context.Background()

	record := &kgo.Record{Topic: "foo", Value: []byte(message)}
	fmt.Println("message created: " + string(message))
	c1.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			fmt.Printf("record had a produce error %v\n", err)
		} else {
			fmt.Println("Produced message: " + string(message))
		}
	})

}

func Consumer(c1 *kgo.Client) {
	ctx := context.Background()
	count := 0

	for {
		currentTime := time.Now()
		fmt.Println("running consumer...")
		fetches := c1.PollFetches(ctx)
		fmt.Println("Waited for ", time.Since(currentTime).Milliseconds(), "ms")
		if errs := fetches.Errors(); len(errs) > 0 {
			panic(fmt.Sprint(errs))
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			count++
			record := iter.Next()
			fmt.Println(string(record.Value), "from an iterator!")
			fmt.Println("count" + fmt.Sprint(count))
		}

		if count == TOTAL_MSGS {
			return
		}
	}

}
