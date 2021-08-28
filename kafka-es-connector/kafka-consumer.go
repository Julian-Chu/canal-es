package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	*kafka.Conn
}

func NewKafkaConsumer(network, kafkaEndpoint, topic string, partition int) *KafkaConsumer {
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaEndpoint, topic, partition)
	if err != nil {
		log.Fatal("kafka: failed to dial leader:", err)
	}
	return &KafkaConsumer{Conn: conn}
}
