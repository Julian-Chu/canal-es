package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
)

type loginDataDoc struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

func main() {
	es := NewESProducer(elasticsearch.Config{})

	topic := "my-topic"
	partition := 0
	kafkaAddr := os.Getenv("kafkaAddr")
	kafkaEndpoint := kafkaAddr + ":9092"
	kafkaConsumer := NewKafkaConsumer("tcp", kafkaEndpoint, topic, partition)

	for {
		_ = kafkaConsumer.SetReadDeadline(time.Now().Add(10 * time.Second))
		batch := kafkaConsumer.ReadBatch(1e3, 1e6) // fetch 1B min, 1MB max

		b := make([]byte, 1e3) // 1KB max per message
		fmt.Println("kafka reads message")
		for {
			msgLen, err := batch.Read(b)
			if err != nil {
				break
			}
			b = b[:msgLen]
			var doc loginDataDoc
			err = json.Unmarshal(b, &doc)
			if err != nil {
				log.Printf("err: %v\n", err)
				continue
			}
			fmt.Printf("json: %v\n", doc)
			/// es
			es.Send("test", doc.Id, string(b), "true")
		}

		if err := batch.Close(); err != nil {
			log.Println("failed to close batch:", err)
			log.Println("retry after 10s")
			time.Sleep(10 * time.Second)
		}

	}
}
