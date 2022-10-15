package client

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/utils"
)

func (client Subscriber) RestartService(callback func(string) error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.RestartService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(errorOrOk(err))
	})
}
func (client Publisher) RestartService(service string) {
	client.nc.Publish(shell.RestartService, utils.EncodeAny(service))
}

func (client Subscriber) StopService(callback func(string) error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.StopService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(errorOrOk(err))
	})
}

func (client Publisher) StopService(service string) {
	client.nc.Publish(shell.StopService, utils.EncodeAny(service))
}

func (client Subscriber) StartService(callback func(string) error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.StartService, func(msg *nats.Msg) {
		err := callback(utils.DecodeAny[string](msg.Data))
		msg.Respond(errorOrOk(err))
	})
}

func (client Publisher) StartService(service string) {
	client.nc.Publish(shell.StartService, utils.EncodeAny(service))
}

func (client Subscriber) RestartShell(callback func() error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.RestartShell, func(msg *nats.Msg) {
		err := callback()
		msg.Respond(errorOrOk(err))
	})
}

func (client Publisher) RestartShell() {
	client.nc.Publish(shell.RestartShell, []byte{})
}

func (client Subscriber) QuitShell(callback func() error) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.QuitShell, func(msg *nats.Msg) {
		err := callback()
		msg.Respond(errorOrOk(err))
	})
}

func (client Requester) QuitShell() error {
	msg, _ := client.nc.Request(shell.QuitShell, nil, client.timeout)
	return utils.DecodeAny[error](msg.Data)
}

func (client Subscriber) Config(callback func(string) *shell.Service) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.GetService, func(msg *nats.Msg) {
		key := utils.DecodeAny[string](msg.Data)
		config := callback(key)
		msg.Respond(utils.EncodeAny(config))
	})
}

// Get the request for the named service.
// If the empty string is specified, this function attempts to
// find the currently executing service based on the environment
// variable
func (client Requester) Config(name string) (service shell.Service, err error) {
	if name == "" {
		name = os.Getenv(shell.SERVICE_ENV_KEY)
	}
	if name == "" {
		fmt.Println("No such config: ", name)
		return
	}
	msg, err := client.nc.Request(shell.GetService, []byte(name), client.timeout)
	if err != nil {
		return
	}
	service = utils.DecodeAny[shell.Service](msg.Data)
	return
}

func (client Subscriber) ShellConfig(callback func() shell.Configuration) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.ShellConfig, func(msg *nats.Msg) {
		config := callback()
		msg.Respond(utils.EncodeAny(config))
	})
}

func (client Publisher) WH_SHELL(evt shell.ShellEventInfo) {
	client.nc.Publish(shell.ShellEvent, utils.EncodeAny(evt))
}

func (client Subscriber) WH_SHELL(callback func(shell.ShellEventInfo)) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.ShellEvent, func(msg *nats.Msg) {
		evt := utils.DecodeAny[shell.ShellEventInfo](msg.Data)
		callback(evt)
	})
}

func (client Publisher) WH_KEYBOARD(evt shell.KeyboardEventInfo) {
	client.nc.Publish(shell.KeyboardEvent, utils.EncodeAny(evt))
}

func (client Requester) WH_KEYBOARD(evt shell.KeyboardEventInfo) bool {
	response, err := client.nc.Request(shell.KeyboardEvent, utils.EncodeAny(evt), time.Millisecond)
	if err != nil {
		return false
	}
	return utils.DecodeAny[bool](response.Data)
}

func (client Subscriber) WH_KEYBOARD(callback func(shell.KeyboardEventInfo) bool) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.KeyboardEvent, func(msg *nats.Msg) {
		evt := utils.DecodeAny[shell.KeyboardEventInfo](msg.Data)
		callback(evt)
	})
}
