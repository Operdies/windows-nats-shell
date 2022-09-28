//go:build windows && amd64
// +build windows,amd64

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/cmd/windows-shell/service"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

func parseCfg(path string) (config *shell.Configuration, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg shell.Configuration
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return
	}
	added := map[string]string{}
	for _, service := range cfg.Services {
		if _, exists := added[service.Name]; exists {
			err = fmt.Errorf("Multiple services with the name '%s' defined.", service.Name)
			return
		}
	}
	config = &cfg
	return &cfg, nil
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
			return thisDir
		}
	}
	return ""
}

func getConfigPaths() []string {
	result := make([]string, 0)
	exeDir := getExeDir()
	if exeDir != "" {
		result = append(result, path.Join(exeDir, "config.json"))
	}

	wd, _ := os.Getwd()
	result = append(result, path.Join(wd, "config.json"))

	return result
}

func start(config *shell.Configuration) {
	fmt.Println("Starting shell!")

	home, _ := os.UserHomeDir()
	os.Chdir(home)

	var jobs map[string]*service.ProcessJob

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

		for _, ser := range config.Services {
			if _, ok := jobs[ser.Name]; ok {
				return fmt.Errorf("Duplicate definition for service '%s'\n", ser.Name)
			}
			jobs[ser.Name] = service.NewProcessJob(ser)
		}

		for _, job := range jobs {
			job.Start()
		}
		return nil
	}

	err := reloadConfig()
	if err != nil {
		panic(err.Error())
	}

	client, _ := client.New(nats.DefaultURL, time.Second)
	defer client.Close()
	client.Subscribe.StartService(func(s string) error {
		job, ok := jobs[s]
		if ok {
      return job.Start()
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})
	client.Subscribe.StopService(func(s string) error {
		job, ok := jobs[s]
		if ok {
      return job.Stop()
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})
	client.Subscribe.RestartService(func(s string) error {
		job, ok := jobs[s]
		if ok {
      return job.Restart()
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})

	quit := make(chan bool)
	client.Subscribe.RestartShell(func() error {
		quit <- true
		return nil
	})

	client.Subscribe.Config(func() shell.Configuration {
		return *config
	})

	<-quit
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

	go flushStdinPipeIndefinitely()

	for {
		start(config)
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

func flushStdinPipeIndefinitely() {
	buf := make([]byte, 1)
	for {
		// We need to flush the stdin buffer in order to other processes to be able to read it
		_, eof := os.Stdin.Read(buf)
		if eof == io.EOF {
			return
		}
	}
}
