package notifications

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func SendNewMsgNotifications(client *websocket.Conn) {
	err := sendSingleMessage(client, newInitialMessage())
	if err != nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	for {
		err := sendSingleMessage(client, newMessage())
		if err != nil {
			return
		}

		<-ticker.C
	}
}

func sendSingleMessage(client *websocket.Conn, msg []byte) error {
	w, err := client.NextWriter(websocket.TextMessage)
	if err != nil {
		return fmt.Errorf("Could not start writing")
	}

	log.Println(string(msg))

	w.Write(msg)
	w.Close()
	return nil
}

func newMessage() []byte {
	data, _ := json.Marshal(map[string]interface{}{
		"type": "notification",
		"notification": map[string]interface{}{
			"id":       123,
			"title":    "123title",
			"category": "some_category_idk",
			"text":     "Hello darkness my old friend",
			"isRead":   false,
		},
	})
	return data
}

func newInitialMessage() []byte {
	data, _ := json.Marshal(map[string]interface{}{
		"type": "all-notifications",
		"notification": []map[string]interface{}{
			{
				"id":       123,
				"title":    "123title",
				"category": "some_category_idk",
				"text":     "Hello darkness my old friend",
				"isRead":   false,
			},
		},
	})
	return data
}
