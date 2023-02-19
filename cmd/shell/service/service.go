package service

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
)

type Jobber interface {
	Start() error
	Stop() error
}

type ProcessJob struct {
	restart    bool
	StartCount int
	service    *shell.Service
	cmd        *exec.Cmd
	name       string
}

func withTimeout[T any](f func() T, timeout time.Duration) (result T, err error) {
	r := make(chan T)
	go func() {
		r <- f()
	}()

	select {
	case <-time.After(timeout):
		err = fmt.Errorf("Operation timed out.")
	case result = <-r:
		err = nil
	}
	return
}

func CombineErrors(errors ...error) error {
	var err error
	for _, e := range errors {
		if e == nil {
			continue
		}
		if err == nil {
			err = e
		} else {
			err = fmt.Errorf("%w; %v", err, e.Error())
		}
	}
	return err
}

type NatsStdout struct {
	subject string
	nc      *nats.Conn
}

func CreateNatsStdout(subject string) *NatsStdout {
	var n NatsStdout
	n.subject = "stdout." + subject
	n.nc, _ = nats.Connect(nats.DefaultURL)
	return &n
}

func (n *NatsStdout) Close() {
	n.nc.Close()
}

func (n *NatsStdout) Write(data []byte) (int, error) {
	n.nc.Publish(n.subject, data)
	return len(data), nil
}

func (j *ProcessJob) Start() error {
	if j.service.Executable == "" {
		return fmt.Errorf("Service %s has no configured executable.", j.name)
	}
	if j.service.Enabled != nil && *j.service.Enabled == false {
		return nil
	}
	if j.cmd != nil {
		return fmt.Errorf("Process %s is already running.", j.name)
	}
	j.StartCount += 1
	log.Printf("Starting %s. (%d)\n", j.name, j.StartCount)

	j.restart = *j.service.AutoRestart == true
	prog := j.service
	cmd := exec.Command(prog.Executable, prog.Arguments...)
	ref := fmt.Sprintf("%s=%s", shell.SERVICE_ENV_KEY, j.name)
	env := os.Environ()
	env = append(env, prog.Environment...)
	env = append(env, ref)
	cmd.Env = env
	cmd.Dir = prog.WorkingDirectory
	natsStdout := CreateNatsStdout(j.name)
	natsStderr := CreateNatsStdout(j.name)
	cmd.Stdout = natsStdout
	cmd.Stderr = natsStderr

	if j.service.Visible == false {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.HideWindow = true
	}

	err := cmd.Start()

	if err != nil {
		log.Printf("Process %s failed to start. %v\n", j.name, err.Error())
		j.restart = false
		return err
	}

	j.cmd = cmd

	go func() {
		defer natsStdout.Close()
		defer natsStderr.Close()
		err := cmd.Wait()
		ex := cmd.ProcessState.ExitCode()
		j.cmd = nil
		if err != nil {
			log.Printf("Process %s exited. (%d: %v)\n", j.name, ex, err)
		} else {
			log.Printf("Process %s exited. (%d)\n", j.name, ex)
		}

		if j.restart {
			j.Start()
		}
	}()

	return err
}

func (j *ProcessJob) Stop() (err error) {
	j.restart = false
	if j.cmd == nil {
		return fmt.Errorf("Process %s is not running.", j.name)
	}
	killError := j.cmd.Process.Kill()
	waitErr, timeoutErr := withTimeout(j.cmd.Wait, time.Second*3)

	// the process is dead
	if timeoutErr == nil {
		j.cmd = nil
	}

	err = CombineErrors(killError, waitErr, timeoutErr)
	return
}

func NewProcessJob(name string, service shell.Service) *ProcessJob {
	s := ProcessJob{}
	s.service = &service
	s.name = name
	if service.AutoRestart == nil {
		b := false
		service.AutoRestart = &b
	}
	if service.Enabled == nil {
		b := true
		service.Enabled = &b
	}
	return &s
}
