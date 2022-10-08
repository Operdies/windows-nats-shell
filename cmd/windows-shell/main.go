//go:build windows && amd64
// +build windows,amd64

package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/cmd/windows-shell/service"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"gopkg.in/yaml.v3"
)

func parseCfg(path string) (config *shell.Configuration, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg shell.Configuration
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return
	}
	config = &cfg
	config.Path = path
	return
}

func loadConfig() *string {
	fileExists := func(f string) bool {
		_, err := os.Stat(f)
		return err == nil
	}

	for _, cand := range getConfigPaths() {
		if fileExists(cand) {
			return &cand
		}
	}
	return nil
}

func getExeDir() string {
	thisExe := os.Args[0]
	for i := len(thisExe) - 1; i >= 0; i = i - 1 {
		if thisExe[i] == '\\' || thisExe[i] == '/' {
			thisDir := thisExe[:i]
			return strings.ReplaceAll(thisDir, "\\", "/")
		}
	}
	return ""
}

func getConfigPaths() []string {
	result := make([]string, 0)
	exeDir := getExeDir()
	if exeDir != "" {
		result = append(result, path.Join(exeDir, "config.yml"))
		result = append(result, path.Join(path.Dir(exeDir), "config.yml"))
	}

	wd, _ := os.Getwd()
	result = append(result, path.Join(wd, "config.yml"))

	return result
}

func truther() *bool {
	b := true
	return &b
}
func falser() *bool {
	b := false
	return &b
}

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
			cfg2, err := parseCfg(config.Path)
			if err != nil {
				fmt.Printf("Error in config: %v", err)
			} else {
				if newCfg, ok := cfg2.Services[s]; ok {
					config.Services[s] = newCfg
				}
			}
			job.Stop()
			job = service.NewProcessJob(s, config.Services[s])
			jobs[s] = job
			go job.Start()
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
			if ser.AutoRestart == nil {
				ser.AutoRestart = falser()
			}
			if ser.Enabled == nil {
				ser.Enabled = truther()
			}
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
	configFile := loadConfig()
	if configFile == nil {
		panic("No config file found")
	}

	config, err := parseCfg(*configFile)
	if err != nil {
		panic(err.Error())
	}

	home, _ := os.UserHomeDir()
	os.Chdir(home)

	for start(config) {
		config2, err := parseCfg(*configFile)
		if err != nil {
			fmt.Println("Error in reloaded config:", err.Error())
			fmt.Println("Services were restarted, but no changes were made.")
		} else {
			fmt.Println("Loaded new config file.")
			config = config2
		}
	}
}
