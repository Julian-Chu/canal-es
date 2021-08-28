package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	*kafka.Conn
}

func NewKafkaProducer(network, kafkaEndpoint, topic string, partition int) *KafkaProducer {
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaEndpoint, topic, partition)
	if err != nil {
		log.Fatal("kafka: failed to dial leader:", err)
	}
	return &KafkaProducer{Conn: conn}
}
