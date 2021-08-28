package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type ESProducer struct {
	client *elasticsearch.Client
}

func NewESProducer(cfg elasticsearch.Config) *ESProducer {
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("es: failed to init client: %+v\n", err)
	}
	var (
		esInfo map[string]interface{}
	)
	// es container can be ready but server is still starting
	cnt := 5
	waitTime := 5 * time.Second
	var res *esapi.Response
	for retried := 0; retried <= cnt; retried++ {
		res, err = es.Info()
		if err != nil {
			retried++
			log.Printf("retry %d of %d\n", retried, cnt)
			time.Sleep(waitTime)
			waitTime *= 2
			continue
		}
	}
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&esInfo); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	log.Printf("Client: %s", elasticsearch.Version)
	log.Printf("Server: %s", esInfo["version"].(map[string]interface{})["number"])
	log.Println(strings.Repeat("~", 37))
	return &ESProducer{client: es}
}

func (e ESProducer) Send(index, documentID, messageBody, refresh string) {
	/// es
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       strings.NewReader(messageBody),
		Refresh:    refresh,
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%s", res.Status(), documentID)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Printf("document:%s [%s] %s; version=%d", documentID, res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}
}
