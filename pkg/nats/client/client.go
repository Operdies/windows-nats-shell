package client

import (
	"time"

	"github.com/nats-io/nats.go"
)

var subscriber interface {
}

type Client struct {
	nc        *nats.Conn
	timeout   time.Duration
	Subscribe *Subscriber
	Publish   *Publisher
	Request   *Requester
}

type Subscriber struct {
	nc      *nats.Conn
	timeout time.Duration
}

type Publisher struct {
	nc      *nats.Conn
	timeout time.Duration
}

type Requester struct {
	nc      *nats.Conn
	timeout time.Duration
}

func (client Client) Close() {
	client.nc.Close()
}

func Default() Client {
	c, err := New(nats.DefaultURL, time.Second)
	if err != nil {
		panic(err)
	}
	return c
}

func New(url string, timeout time.Duration) (c Client, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return
	}
	c.Publish = &Publisher{nc, timeout}
	c.Request = &Requester{nc, timeout}
	c.Subscribe = &Subscriber{nc, timeout}
	return
}

func errorOrOk(err error) []byte {
	if err == nil {
		return []byte("Ok")
	}
	return []byte(err.Error())
}
