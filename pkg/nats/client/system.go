package client

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/system"
	"github.com/operdies/windows-nats-shell/pkg/utils"
)

func (client Requester) GetPrograms() []string {
	msg, _ := client.nc.Request(system.GetPrograms, nil, client.timeout)
	programs := utils.DecodeAny[[]string](msg.Data)
	return programs
}

func (client Requester) LaunchProgramAsAdmin(program string) error {
	msg, _ := client.nc.Request(system.LaunchProgramAsAdmin, utils.EncodeAny(program), client.timeout)
	status := utils.DecodeAny[string](msg.Data)
	if status == "Ok" {
		return nil
	}
	return errors.New(string(msg.Data))
}

func (client Subscriber) LaunchProgramAsAdmin(callback func(string) string) (*nats.Subscription, error) {
	return client.nc.Subscribe(system.LaunchProgramAsAdmin, func(msg *nats.Msg) {
		program := utils.DecodeAny[string](msg.Data)
		response := callback(program)
		msg.Respond(utils.EncodeAny(response))
	})
}

func (client Requester) LaunchProgram(program string) error {
	msg, _ := client.nc.Request(system.LaunchProgram, utils.EncodeAny(program), client.timeout)
	status := utils.DecodeAny[string](msg.Data)
	if status == "Ok" {
		return nil
	}
	return errors.New(string(msg.Data))
}

func (client Subscriber) LaunchProgram(callback func(string) string) (*nats.Subscription, error) {
	return client.nc.Subscribe(system.LaunchProgram, func(msg *nats.Msg) {
		program := utils.DecodeAny[string](msg.Data)
		response := callback(program)
		msg.Respond(utils.EncodeAny(response))
	})
}

func (client Subscriber) GetPrograms(callback func() []string) (*nats.Subscription, error) {
	return client.nc.Subscribe(system.GetPrograms, func(m *nats.Msg) {
		windows := callback()
		m.Respond(utils.EncodeAny(windows))
	})
}
