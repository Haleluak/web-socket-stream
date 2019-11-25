package pubsub

import (
	"fmt"
	"github.com/micro/go-micro/util/log"

	"github.com/gorilla/websocket"
	"github.com/micro/go-micro/config"
)

// PubSub struct
type PubSub struct {
	Clients       []Client
	Subscriptions []Subscription
}

// Client struct
type Client struct {
	ID         string
	Connection *websocket.Conn
}

// Request struct
type Request struct {
	Topic    string `json:"topic"`
	Alias    string `json:"alias"`
	Resource string `json:"resource"`
}

// Subscription struct
type Subscription struct {
	Topic       string
	BrokerTopic string
	Client      *Client
}

// AddClient function
func (ps *PubSub) AddClient(client Client) *PubSub {

	ps.Clients = append(ps.Clients, client)

	return ps
}

// RemoveClient function
func (ps *PubSub) RemoveClient(client Client) *PubSub {

	// firstly, remove all subscriptions by this client
	for index, sub := range ps.Subscriptions {
		if client.ID == sub.Client.ID {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	// remove client from the list
	for index, c := range ps.Clients {
		if c.ID == client.ID {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}
	}

	return ps
}

func (ps *PubSub) getSubscriptions(topic string, client *Client) []Subscription {

	var ls []Subscription

	for _, sub := range ps.Subscriptions {
		if topic == sub.Topic {
			if client == nil || client != nil && sub.Client.ID == client.ID {
				ls = append(ls, sub)
			}
		}
	}

	return ls
}

func (ps *PubSub) subscribe(client *Client, topic string) *PubSub {

	cliSubs := ps.getSubscriptions(topic, client)
	if len(cliSubs) > 0 {
		// client subscribed the topic already
		return ps
	}

	sub := Subscription{
		Topic:       topic,
		BrokerTopic: topic,
		Client:      client,
	}
	ps.Subscriptions = append(ps.Subscriptions, sub)

	return ps
}

// Send function
func (ps *PubSub) Send(topic string, rsp interface{}) []Subscription {
	ls := ps.getSubscriptions(topic, nil)
	for _, sub := range ls {
		go sub.Client.Connection.WriteJSON(rsp)
	}

	return ls
}

// Unsubscribe function
func (ps *PubSub) Unsubscribe(client *Client, topic string) *PubSub {

	for index, sub := range ps.Subscriptions {
		if topic == sub.Topic && sub.Client.ID == client.ID {
			// found this subscription from client and we need remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	return ps
}

// HandleReceiveMessage function
func (ps *PubSub) HandleReceiveMessage(client Client, req *Request) *PubSub {
	log.Info(req.Topic)
	ps.subscribe(&client, req.Topic)

	return ps
}

// IsSubscribed function
func (ps *PubSub) IsSubscribed(topic string) bool {

	for _, sub := range ps.Subscriptions {
		if topic == sub.BrokerTopic {
			return true
		}
	}

	return false
}

// GenTopicName function
func (req *Request) GenTopicName() {
	req.Topic = fmt.Sprintf("%s.%s", config.Get("broker_topic").String(""), req.Alias)
	if len(req.Resource) > 0 {
		req.Topic += fmt.Sprintf(".%s", req.Resource)
	}
}
