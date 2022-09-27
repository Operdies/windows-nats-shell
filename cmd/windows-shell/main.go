//go:build windows && amd64
// +build windows,amd64

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

type forwardedStdout struct {
	name string
}

type forwardedStdin struct {
	read func([]byte)
}

func valueOrDefault[T any](value *T, def T) T {
	if value == nil {
		return def
	}
	return *value
}

func (f forwardedStdout) Write(p []byte) (n int, err error) {
	fmt.Printf("%s: [%s]\n", f.name, string(p))
	n = len(p)
	err = nil
	return
}

func startProgram(job *job) {
	i := 0
	prog := job.service
	onStopped := make(chan bool)
	autoRestart := valueOrDefault(prog.AutoRestart, true)
	var runningCmd *exec.Cmd
	defer func() {
		if runningCmd != nil {
			runningCmd.Process.Kill()
		}
	}()

	start := func() {
		i += 1
		fmt.Printf("Starting Process(%d): %v %v in %v (%v)\n", i, prog.Executable, prog.Arguments, prog.WorkingDirectory, prog.Name)

		runningCmd = exec.Command(prog.Executable, prog.Arguments...)
		runningCmd.Env = append(runningCmd.Env, prog.Environment...)
		runningCmd.Dir = prog.WorkingDirectory
		if prog.ForwardStderror {
			runningCmd.Stderr = os.Stderr
		}
		if prog.ForwardStdin {
			runningCmd.Stdin = os.Stdin
		}

		if prog.ForwardStdout {
			var f forwardedStdout
			f.name = prog.Name
			runningCmd.Stdout = os.Stdout
		}
		err := runningCmd.Run()
		fmt.Printf("Process '%s' exited.\n", prog.Name)
		onStopped <- true
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	stopProg := func() bool {
		if runningCmd != nil {
			runningCmd.Process.Kill()
			runningCmd.Wait()
		}
		return true
	}

	go func() {
		for {
			<-job.stop
			fmt.Printf("Got signal to stop '%s'.\n", prog.Name)
			autoRestart = false
			stopProg()
		}
	}()

	go func() {
		for {
			<-job.start
			fmt.Printf("Got signal to start '%s'.\n", prog.Name)
			autoRestart = *job.service.AutoRestart == true
			if runningCmd != nil {
				if runningCmd.ProcessState != nil {
					if runningCmd.ProcessState.Exited() {
						go start()
					}
				}
			}
		}
	}()

	go func() {
		for {
			<-job.restart
			fmt.Printf("Got signal to restart '%s'\n.", prog.Name)
			autoRestart = *job.service.AutoRestart == true
			stopProg()
			if !autoRestart {
				go start()
			}
		}
	}()

	// auto-restart on crash / stop
	go func() {
		for {
			<-onStopped
			if autoRestart {
				go start()
			}
		}
	}()
	go start()

}

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

type job struct {
	service shell.Service
	stop    chan bool
	start   chan bool
	restart chan bool
}

func start(config *shell.Configuration) {
	fmt.Println("Starting shell!")

	home, _ := os.UserHomeDir()
	os.Chdir(home)

	var jobs map[string]*job

	stopJobs := func() {
		if jobs != nil {
			for _, job := range jobs {
				job.stop <- true
			}
		}
	}
	defer stopJobs()

	reloadConfig := func() error {
		stopJobs()

		jobs = map[string]*job{}

		for _, service := range config.Services {
			if _, ok := jobs[service.Name]; ok {
				return fmt.Errorf("Duplicate definition for service '%s'\n", service.Name)
			}
			jobs[service.Name] = &job{service: service, stop: make(chan bool), start: make(chan bool), restart: make(chan bool)}
		}

		for _, job := range jobs {
			startProgram(job)
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
			job.start <- true
			return nil
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})
	client.Subscribe.StopService(func(s string) error {
		job, ok := jobs[s]
		if ok {
			job.stop <- true
			return nil
		}
		return fmt.Errorf("Service '%s' is not configured.", s)
	})
	client.Subscribe.RestartService(func(s string) error {
		job, ok := jobs[s]
		if ok {
			job.restart <- true
			return nil
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

	go flushStdinPipeIndefinitely()
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
