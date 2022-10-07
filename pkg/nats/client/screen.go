package client

import (
	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/screen"
	"github.com/operdies/windows-nats-shell/pkg/utils"
)

func (client Requester) GetResolution() screen.Resolution {
	response, _ := client.nc.Request(screen.GetResolution, nil, client.timeout)
	return utils.DecodeAny[screen.Resolution](response.Data)
}

func (client Subscriber) GetResolution(callback func() screen.Resolution) (*nats.Subscription, error) {
	return client.nc.Subscribe(screen.GetResolution, func(msg *nats.Msg) {
		resolution := callback()
		msg.Respond(utils.EncodeAny(resolution))
	})
}

func (client Requester) SetResolution(r screen.Resolution) bool {
	response, _ := client.nc.Request(screen.SetResolution, utils.EncodeAny(r), client.timeout)
	return utils.DecodeAny[bool](response.Data)
}

func (client Subscriber) SetResolution(callback func(screen.Resolution) bool) (*nats.Subscription, error) {
	return client.nc.Subscribe(screen.SetResolution, func(msg *nats.Msg) {
		resolution := utils.DecodeAny[screen.Resolution](msg.Data)
		result := callback(resolution)
		msg.Respond(utils.EncodeAny(result))
	})
}
