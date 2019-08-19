package handler

import (
	"fmt"
	"github.com/micro/go-micro/broker"
	"os"
)

func Sub(Publisher broker.Broker) (chan []byte, chan bool, error) {
	che := make(chan []byte, 20)
	exit := make(chan bool)
	topic := os.Getenv("BROKER_TOPIC")

	sub, err := Publisher.Subscribe(topic, func(p broker.Publication) error {
		fmt.Println("[sub] received message:", string(p.Message().Body))
		che <- p.Message().Body
		return nil
	},broker.Queue("web.socket"))
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		<-exit
		sub.Unsubscribe()
	}()

	return che, exit, nil
}
