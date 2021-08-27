package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/segmentio/kafka-go"
	"github.com/withlin/canal-go/client"
	pbe "github.com/withlin/canal-go/protocol/entry"
)

func main() {
	connector := client.NewSimpleCanalConnector(os.Getenv("canalAddr"), 11111, "", "", "example", 60000, 60*60*1000)
	err := connector.Connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = connector.Subscribe(".*\\..*")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	topic := "my-topic"
	partition := 0
	kafkaAddr := os.Getenv("kafkaAddr")
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaAddr+":9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	for {
		message, err := connector.Get(100, nil, nil)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		batchId := message.Id
		if batchId == -1 || len(message.Entries) <= 0 {
			time.Sleep(300 * time.Millisecond)
			fmt.Println("===no data===")
			continue
		}
		printEntry(message.Entries)
		entries := extractEntries(message.Entries)

		// convert to kafka-go message
		messages := make([]kafka.Message, 0, len(entries))
		for _, m := range entries {
			jsonStr, err := json.Marshal(m)
			fmt.Println(m)
			fmt.Println(jsonStr)
			if err != nil {
				fmt.Printf("failed: marshal to json : %+v", err)
				continue
			}
			messages = append(messages, kafka.Message{Value: jsonStr})
		}
		// kafka
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err = conn.WriteMessages(
			messages...,
		)
		log.Println("Send message to kafka")
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
	}
}

func extractEntries(entries []pbe.Entry) []map[string]interface{} {
	var res []map[string]interface{}
	for _, entry := range entries {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}
		rowChange := new(pbe.RowChange)

		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		checkError(err)
		if rowChange != nil {
			eventType := rowChange.GetEventType()
			header := entry.GetHeader()
			fmt.Println(fmt.Sprintf("================> binlog[%s : %d],name[%s,%s], eventType: %s", header.GetLogfileName(), header.GetLogfileOffset(), header.GetSchemaName(), header.GetTableName(), header.GetEventType()))

			for _, rowData := range rowChange.GetRowDatas() {
				if eventType == pbe.EventType_INSERT {
					printColumn(rowData.GetAfterColumns())
					columns := rowData.GetAfterColumns()
					m := make(map[string]interface{})
					for _, col := range columns {
						fmt.Println("col:", col.GetName())
						m[col.GetName()] = col.GetValue()
					}
					res = append(res, m)
				}
			}
		}
	}
	return res
}
func printEntry(entries []pbe.Entry) {
	for _, entry := range entries {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}
		rowChange := new(pbe.RowChange)

		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		checkError(err)
		if rowChange != nil {
			eventType := rowChange.GetEventType()
			header := entry.GetHeader()
			fmt.Println(fmt.Sprintf("================> binlog[%s : %d],name[%s,%s], eventType: %s", header.GetLogfileName(), header.GetLogfileOffset(), header.GetSchemaName(), header.GetTableName(), header.GetEventType()))

			for _, rowData := range rowChange.GetRowDatas() {
				if eventType == pbe.EventType_DELETE {
					printColumn(rowData.GetBeforeColumns())
				} else if eventType == pbe.EventType_INSERT {
					printColumn(rowData.GetAfterColumns())
				} else {
					fmt.Println("-------> before")
					printColumn(rowData.GetBeforeColumns())
					fmt.Println("-------> after")
					printColumn(rowData.GetAfterColumns())
				}
			}
		}
	}
}

func printColumn(columns []*pbe.Column) {
	for _, col := range columns {
		fmt.Println(fmt.Sprintf("%s : %s  update= %t", col.GetName(), col.GetValue(), col.GetUpdated()))
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
