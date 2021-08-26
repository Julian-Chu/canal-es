package main

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"time"
)

func main() {
	// to consume messages
	topic := "my-topic"
	partition := 0
	kafkaAddr := os.Getenv("kafkaAddr")
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaAddr+":9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		batch := conn.ReadBatch(1e3, 1e6) // fetch 1kB min, 1MB max

		b := make([]byte, 1e3) // 10KB max per message
		fmt.Println("kafka read message")
		for {
			_, err := batch.Read(b)
			if err != nil {
				break
			}
			fmt.Println(string(b))
		}

		if err := batch.Close(); err != nil {
			log.Println("failed to close batch:", err)
			log.Println("retry after 10s")
			time.Sleep(10 * time.Second)
		}

	}
	//if err := conn.Close(); err != nil {
	//	log.Fatal("failed to close connection:", err)
	//}
}
