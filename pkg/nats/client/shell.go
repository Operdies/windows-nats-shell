package client

import (
	"fmt"
	"math/rand"
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

func (client Subscriber) Config(callback func(string) any) (*nats.Subscription, error) {
	return client.nc.Subscribe(shell.GetService, func(msg *nats.Msg) {
		key := utils.DecodeAny[string](msg.Data)
		config := callback(key)
		msg.Respond(utils.EncodeAny(config))
	})
}

func GetConfig[T any](client *Requester) T {
	name := os.Getenv(shell.SERVICE_ENV_KEY)
	if name == "" {
		panic(fmt.Sprintf("Environment variable '%v' not set.", shell.SERVICE_ENV_KEY))
	}
	result, _ := GetServiceConfig[T](client, name)
	return result
}

func GetServiceConfig[T any](client *Requester, name string) (result T, err error) {
	msg, err := client.nc.Request(shell.GetService, []byte(name), client.timeout)
	if err != nil {
		return
	}
	result = utils.DecodeAny[T](msg.Data)
	return result, nil
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func RandStringBytesMask(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func (client Requester) RequestManyKeyboards(evt shell.KeyboardEventInfo) bool {
	replyTo := RandStringBytesMask(10)

	sub, err := client.nc.SubscribeSync(replyTo)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()
	client.nc.Flush()
	client.nc.PublishRequest(shell.KeyboardEvent, replyTo, utils.EncodeAny(evt))
	max := time.Millisecond
	start := time.Now()

	for time.Now().Sub(start) < max {
		rem := max - time.Now().Sub(start)
		msg, err := sub.NextMsg(rem)
		if err != nil {
			return false
		}
		response := utils.DecodeAny[bool](msg.Data)
		if response {
			return true
		}
	}
	return false
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
		handled := callback(evt)
		msg.Respond(utils.EncodeAny(handled))
	})
}
