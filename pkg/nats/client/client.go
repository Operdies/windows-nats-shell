package client

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/internal/api"
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

func (c Client) WindowsUpdated(callback func(Windows)) {
	c.nc.Subscribe(api.WindowsUpdated, func(m *nats.Msg) {
		windows := utils.DecodeAny[Windows](m.Data)
		callback(windows)
	})

}

func (c Client) SetFocus(window uint64) {
	c.nc.Request(api.SetFocus, utils.EncodeAny(window), time.Second)
}

func Create(url string) (c Client, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return
	}
	c.nc = nc
	return
}
