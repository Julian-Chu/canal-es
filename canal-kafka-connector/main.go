package main

import (
	"log"
	"os"
	"time"
)

func main() {
	canalServer := os.Getenv("canalAddr")
	canalConsumer := NewCanalConsumer(canalServer, 11111, "", "", "example")
	err := canalConsumer.Connect()
	if err != nil {
		log.Fatalf("canal: failed to connect to canal: %+v\n", err)
	}

	err = canalConsumer.Subscribe(".*\\..*")
	if err != nil {
		log.Fatalf("canal: failed to subscribe to canal: %+v\n", err)
	}

	topic := "my-topic"
	partition := 0
	kafkaAddr := os.Getenv("kafkaAddr")
	kafkaEndpoint := kafkaAddr + ":9092"
	kafkaProducer := NewKafkaProducer("tcp", kafkaEndpoint, topic, partition)

	for {
		message, err := canalConsumer.Get(100, nil, nil)
		if err != nil {
			log.Fatalf("canal: failed to get message: %+v", err)
		}
		batchId := message.Id
		if batchId == -1 || len(message.Entries) <= 0 {
			time.Sleep(300 * time.Millisecond)
			log.Println("===no data===")
			continue
		}
		entries := canalConsumer.ExtractEntries(message.Entries)

		// convert to kafka-go message
		messages := kafkaProducer.ConvertToKafkaMessages(entries)
		_ = kafkaProducer.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err = kafkaProducer.WriteMessages(
			messages...,
		)
		log.Println("kafka: Send message to kafka")
		if err != nil {
			log.Fatal("kafka: failed to write messages:", err)
		}
	}
}
