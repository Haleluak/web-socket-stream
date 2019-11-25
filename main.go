package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/web"
	"github.com/micro/go-plugins/broker/kafka"
	"gitlab.com/brazncorp/brazn-webSocketStream/handler"
	"gitlab.com/brazncorp/brazn-webSocketStream/pubsub"
	"log"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func autoId() (string) {
	return uuid.Must(uuid.NewUUID()).String()
}

var ps = &pubsub.PubSub{}


func main() {
	e := godotenv.Load()
	if e != nil {
		fmt.Print(e.Error())
	}

	service := web.NewService(web.Name("go.micro.web.socket"))
	if err := service.Init(); err != nil {
		log.Fatal("Init", err)
	}

	brkAddr := fmt.Sprintf("%s:%s", os.Getenv("BROKER_HOST"), os.Getenv("BROKER_PORT"))
	brokerKafka := kafka.NewBroker(func(options *broker.Options) {
		options.Addrs = []string{ brkAddr}
	})

	ch, exit, err := handler.Sub(brokerKafka)
	if err != nil {
		log.Fatal("Sub: ", err)
	}
	defer func() {
		close(exit)
	}()

	service.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade request to websocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal("Upgrade: ", err)
			return
		}
		defer conn.Close()

		client := pubsub.Client{
			Id:         autoId(),
			Connection: conn,
		}

		// add this client into the list
		ps.AddClient(client)

		fmt.Println("New Client is connected, total: ", len(ps.Clients))
		go func() {
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					ps.RemoveClient(client)
					log.Println("total clients and subscriptions ", len(ps.Clients), len(ps.Subscriptions))
					return
				}
			}
		}()
		ps.HandleReceiveMessage(client);

		for {
			select {
			case node := <-ch:
				var r interface{}
				err = json.Unmarshal(node, r)
				log.Println(r)
				ps.Publish("sub-client-suprema", r, nil)
			}
		}
	})
	// Run the server
	if err := service.Run(); err != nil {
		log.Println(err)
	}

}
