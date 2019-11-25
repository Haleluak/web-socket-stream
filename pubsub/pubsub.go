package pubsub

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/micro/go-log"
)

const (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

type PubSub struct {
	Clients       []Client
	Subscriptions []Subscription
}

type Client struct {
	Id         string
	Connection *websocket.Conn
}

type Message struct {
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic  string
	Client *Client
}

func (ps *PubSub) AddClient(client Client) (*PubSub) {

	ps.Clients = append(ps.Clients, client)

	payload := []byte("Hello Client ID:" +
		client.Id)

	log.Log("Socket: ", string(payload))

	return ps

}

func (ps *PubSub) RemoveClient(client Client) (*PubSub) {

	// first remove all subscriptions by this client

	for index, sub := range ps.Subscriptions {

		if client.Id == sub.Client.Id {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	// remove client from the list

	for index, c := range ps.Clients {

		if c.Id == client.Id {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}

	}

	return ps
}

func (ps *PubSub) GetSubscriptions(topic string, client *Client) ([]Subscription) {

	var subscriptionList []Subscription

	for _, subscription := range ps.Subscriptions {

		if client != nil {

			if subscription.Client.Id == client.Id && subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)

			}
		} else {

			if subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}

	return subscriptionList
}

func (ps *PubSub) Subscribe(client *Client, topic string) (*PubSub) {

	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs) > 0 {

		// client is subscribed this topic before

		return ps
	}

	newSubscription := Subscription{
		Topic:  topic,
		Client: client,
	}

	ps.Subscriptions = append(ps.Subscriptions, newSubscription)

	return ps
}

func (ps *PubSub) Publish(topic string, message interface{}, excludeClient *Client) {

	subscriptions := ps.GetSubscriptions(topic, nil)

	for _, sub := range subscriptions {
		fmt.Printf("Sending to client id %s message is %s \n", sub.Client.Id, message)
		sub.Client.Send(message)
	}

}

func (client *Client) Send(message interface{}) (error) {

	return client.Connection.WriteJSON(message)

}

func (ps *PubSub) Unsubscribe(client *Client, topic string) (*PubSub) {

	//clientSubscriptions := ps.GetSubscriptions(topic, client)
	for index, sub := range ps.Subscriptions {

		if sub.Client.Id == client.Id && sub.Topic == topic {
			// found this subscription from client and we do need remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}

	return ps

}

func (ps *PubSub) HandleReceiveMessage(client Client) (*PubSub) {
	topic := "sub-client-suprema"
	ps.Subscribe(&client, topic)
	fmt.Println("new subscriber to topic",topic, len(ps.Subscriptions), client.Id)
	return ps
}