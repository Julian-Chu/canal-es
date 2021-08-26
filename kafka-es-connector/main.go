package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	/// es
	var (
		r map[string]interface{}
	)
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	log.Printf("Client: %s", elasticsearch.Version)
	log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
	log.Println(strings.Repeat("~", 37))

	/// kafka
	// to consume messages
	topic := "my-topic"
	partition := 0
	kafkaAddr := os.Getenv("kafkaAddr")
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaAddr+":9092", topic, partition)
	if err != nil {
		log.Fatal("kafka: failed to dial leader:", err)
	}
	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		batch := conn.ReadBatch(1e3, 1e6) // fetch 1kB min, 1MB max

		b := make([]byte, 1e3) // 10KB max per message
		fmt.Println("kafka read message")
		type loginDataDoc struct {
			Id        string `json:"id"`
			Name      string `json:"name"`
			Timestamp string `json:"timestamp"`
		}
		for {
			msgLen, err := batch.Read(b)
			if err != nil {
				break
			}
			b = b[:msgLen]
			fmt.Println(string(b))
			var doc loginDataDoc
			err = json.Unmarshal(b, &doc)
			if err != nil {
				log.Printf("err: %v\n", err)
				continue
			}
			fmt.Printf("json: %v\n", doc)
			/// es
			req := esapi.IndexRequest{
				Index:      "test",
				DocumentID: doc.Id,
				Body:       strings.NewReader(string(b)),
				Refresh:    "true",
			}

			// Perform the request with the client.
			res, err := req.Do(context.Background(), es)
			if err != nil {
				log.Fatalf("Error getting response: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				log.Printf("[%s] Error indexing document ID=%d", res.Status(), doc.Id)
			} else {
				// Deserialize the response into a map.
				var r map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					log.Printf("Error parsing the response body: %s", err)
				} else {
					// Print the response status and indexed document version.
					log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
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
