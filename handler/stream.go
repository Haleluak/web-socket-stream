package handler

import (
	"encoding/json"
	"fmt"
	"github.com/micro/go-micro/config"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/util/log"
	"gitlab.com/brazncorp/brazn-websocket/pubsub"
	"gitlab.com/brazncorp/brazn-websocket/util"
	"github.com/golang/protobuf/proto"
	proto_bk "github.com/micro/go-micro/broker/service/proto"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var ps = &pubsub.PubSub{}

var (
	Broker broker.Broker
	ch = make(chan *proto_bk.Message, 200)
)

func Init( br broker.Broker)  {
	Broker = br
}

// Stream function
func Stream(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Upgrade: ", err)

		return
	}
	defer conn.Close()

	cli := pubsub.Client{
		ID:         generateID(),
		Connection: conn,
	}

	ps.AddClient(cli)
	log.Info("Client has connected: ", util.Beautify(map[string]interface{}{
		"id":           cli.ID,
		"client_count": len(ps.Clients),
	}))

	_, p, _ := conn.ReadMessage()
	req := &pubsub.Request{}
	if err = json.Unmarshal(p, req); err != nil {
		// do nothing because client send wrong message
		return
	}
	req.GenTopicName()

	go func() {
		if _, _, err := conn.ReadMessage(); err != nil {
			ps.RemoveClient(cli)
			log.Error("Something went wrong. Client has been disconnected: ", util.Beautify(map[string]interface{}{
				"topic":              req.Topic,
				"client_count":       len(ps.Clients),
				"subscription_count": len(ps.Subscriptions),
			}))

			return
		}
	}()

	ps.HandleReceiveMessage(cli, req)

	for {
		select {
		case msg := <-ch:
			var rsp interface{}
			err := json.Unmarshal(msg.Body, &rsp)
			if err != nil {
				log.Error("Failed to decode message: ", util.Beautify(map[string]interface{}{
					"error": err.Error(),
					"data":  err,
					"topic": req.Topic,
				}))

				return
			}
			topic := fmt.Sprintf("%s.%s.%s", config.Get("broker_topic").String(""), msg.Header["alias"], msg.Header["name"])
			cnt := ps.Send(topic, rsp)
			log.Debug("Publish message: ", util.Beautify(map[string]interface{}{
				"topic":                       topic,
				"message":                     rsp,
				"relevant_subscription_count": len(cnt),
			}))

		}
	}
}

func Subsubcriber() error {
	brokerTopic := config.Get("broker_topic").String("")
	err := onMessage(brokerTopic)
	if err != nil {
		return err
	}
	return nil
}

func onMessage(brokerTopic string) ( error) {
	_, err := Broker.Subscribe(brokerTopic, func(p broker.Event) error {
		pb := &proto_bk.Message{}
		err := proto.Unmarshal(p.Message().Body, pb)
		if err != nil{
			log.Error(err.Error())
		}

		topic := fmt.Sprintf("%s.%s.%s", brokerTopic, pb.Header["alias"], pb.Header["name"])
		if ps.IsSubscribed(topic) {
			ch <- pb
		}

		return nil
	})

	if err != nil {
		log.Error("Failed to connect to streamer: ", util.Beautify(map[string]interface{}{
			"message": err.Error(),
			"error":   err,
			"topic":   brokerTopic,
		}))
	}

	return err
}

func generateID() string {
	return uuid.Must(uuid.NewUUID()).String()
}