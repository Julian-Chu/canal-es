package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/segmentio/kafka-go"
	"github.com/withlin/canal-go/client"
	pbe "github.com/withlin/canal-go/protocol/entry"
)

type CanalConsumer struct {
	*client.SimpleCanalConnector
}

func NewCanalConsumer(canalServer string, port int, username, password, destination string) *CanalConsumer {
	connector := client.NewSimpleCanalConnector(canalServer, port, username, password, destination, 60000, 60*60*1000)
	return &CanalConsumer{connector}
}

func (c CanalConsumer) ExtractEntries(entries []pbe.Entry) []map[string]interface{} {
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
					c.printColumn(rowData.GetAfterColumns())
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
func (c CanalConsumer) printEntry(entries []pbe.Entry) {
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
					c.printColumn(rowData.GetBeforeColumns())
				} else if eventType == pbe.EventType_INSERT {
					c.printColumn(rowData.GetAfterColumns())
				} else {
					fmt.Println("-------> before")
					c.printColumn(rowData.GetBeforeColumns())
					fmt.Println("-------> after")
					c.printColumn(rowData.GetAfterColumns())
				}
			}
		}
	}
}

func (c CanalConsumer) printColumn(columns []*pbe.Column) {
	for _, col := range columns {
		fmt.Println(fmt.Sprintf("%s : %s  update= %t", col.GetName(), col.GetValue(), col.GetUpdated()))
	}
}

func (k KafkaProducer) ConvertToKafkaMessages(entries []map[string]interface{}) []kafka.Message {
	messages := make([]kafka.Message, 0, len(entries))
	for _, m := range entries {
		jsonStr, err := json.Marshal(m)
		if err != nil {
			log.Printf("failed: marshal to json : %+v", err)
			continue
		}
		messages = append(messages, kafka.Message{Value: jsonStr})
	}
	return messages
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
