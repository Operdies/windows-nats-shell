//go:build windows && amd64
// +build windows,amd64

package main

import (
	"fmt"
	"io"
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
	fmt.Println(cfg)
	return
}

func loadConfig() *string {
	fileExists := func(f string) bool {
		_, err := os.Stat(f)
		return err == nil
	}

	for _, cand := range getConfigPaths() {
		fmt.Printf("Trying %s\n", cand)
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
		parent := path.Dir(exeDir)
		fmt.Printf("Path %s has parent %s\n", exeDir, parent)
		result = append(result, path.Join(exeDir, "config.yml"))
		result = append(result, path.Join(path.Dir(exeDir), "config.yml"))
		fmt.Println(result)
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
	client.Subscribe.QuitShell(func() error {
		quit <- false
		return nil
	})

	client.Subscribe.Config(func(key string) *shell.Service {
		if section, ok := config.Services[key]; ok {
			return &section
		}
		return nil
	})
	client.Subscribe.ShellConfig(func() shell.Configuration {
		return *config
	})

	return <-quit
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
