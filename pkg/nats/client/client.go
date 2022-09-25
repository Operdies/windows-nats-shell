package client

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api"
	"github.com/operdies/windows-nats-shell/pkg/nats/utils"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

const (
	timeout = time.Second
)

type Client struct {
	nc *nats.Conn
}

type Windows = []wintypes.Window

func (c Client) Windows() Windows {
	response, _ := c.nc.Request(api.Windows, nil, timeout)
  return utils.DecodeAny[Windows](response.Data)
}

func CreateClient(url string) (c Client, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return
	}
	c.nc = nc
	return
}
