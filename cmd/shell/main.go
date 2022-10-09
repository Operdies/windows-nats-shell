//go:build windows && amd64
// +build windows,amd64

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/cmd/shell/service"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

func start(config *shell.Configuration) bool {
	var subs []*nats.Subscription
	var jobs map[string]*service.ProcessJob
	quit := make(chan bool)
	fmt.Println("Starting shell!")

	client, err := client.New(nats.DefaultURL, time.Second)
	if err != nil {
		panic(err)
	}

	defer func() {
		for _, s := range subs {
			s.Unsubscribe()
		}
		client.Close()
	}()

	s, _ := client.Subscribe.StartService(func(s string) error {
		job, ok := jobs[s]
		if ok {
			return job.Start()
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})
	subs = append(subs, s)
	s, _ = client.Subscribe.StopService(func(s string) error {
		job, ok := jobs[s]
		if ok {
			return job.Stop()
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})
	subs = append(subs, s)
	s, _ = client.Subscribe.RestartService(func(s string) error {
		job, ok := jobs[s]
		if ok {
			cfg2, err := config.Reload()
			if err != nil {
				fmt.Printf("Error in config: %v", err)
			} else {
				if newCfg, ok := cfg2.Services[s]; ok {
					config.Services[s] = newCfg
				}
			}
			job.Stop()
			job = service.NewProcessJob(s, config.Services[s])
			job.StartCount = jobs[s].StartCount
			jobs[s] = job
			go job.Start()
			return nil
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})
	subs = append(subs, s)
	s, _ = client.Subscribe.RestartShell(func() error {
		fmt.Println("Restart Shell!")
		quit <- true
		return nil
	})
	subs = append(subs, s)
	s, _ = client.Subscribe.QuitShell(func() error {
		quit <- false
		return nil
	})
	subs = append(subs, s)
	s, _ = client.Subscribe.Config(func(key string) *shell.Service {
		if section, ok := config.Services[key]; ok {
			return &section
		}
		return nil
	})
	subs = append(subs, s)
	s, _ = client.Subscribe.ShellConfig(func() shell.Configuration {
		return *config
	})

	stopJobs := func() {
		if jobs != nil {
			for _, job := range jobs {
				job.Stop()
			}
		}
	}
	defer stopJobs()

	reloadConfig := func() error {
		stopJobs()

		jobs = map[string]*service.ProcessJob{}

		for name, ser := range config.Services {
			jobs[name] = service.NewProcessJob(name, ser)
		}

		for _, job := range jobs {
			go job.Start()
		}
		return nil
	}

	err = reloadConfig()
	if err != nil {
		panic(err.Error())
	}

	restart := <-quit
	close(quit)
	return restart
}

func main() {
	config := shell.LoadDefault()

	exe := os.Args[0]
	here := filepath.Dir(exe)
	os.Chdir(here)

	for start(config) {
		config2, err := config.Reload()
		if err != nil {
			fmt.Println("Error in reloaded config:", err.Error())
			fmt.Println("Services were restarted, but no changes were made.")
		} else {
			fmt.Println("Loaded new config file.")
			config = config2
		}
	}
}
