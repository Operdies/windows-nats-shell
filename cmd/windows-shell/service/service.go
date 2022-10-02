package service

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
)

type Jobber interface {
	Start() error
	Stop() error
	Restart() error
}

type ProcessJob struct {
	restart    bool
	startCount int
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

func (j *ProcessJob) Start() error {
	if j.service.Executable == "" {
		return fmt.Errorf("Service %s has no configured executable.", j.name)
	}
	// fmt.Printf("Start %s\n", j.name)

	if j.cmd != nil {
		return fmt.Errorf("Process %s is already running.", j.name)
	}
	j.startCount += 1
	fmt.Printf("Starting %s. (%d)\n", j.name, j.startCount)

	j.restart = *j.service.AutoRestart == true
	prog := j.service
	cmd := exec.Command(prog.Executable, prog.Arguments...)
	ref := fmt.Sprintf("%s=%s", shell.SERVICE_ENV_KEY, j.name)
	env := os.Environ()
	env = append(env, prog.Environment...)
	env = append(env, ref)
	cmd.Env = env
	cmd.Dir = prog.WorkingDirectory

	if prog.ForwardStdout {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
	}
	if prog.ForwardStdin {
		cmd.Stdin = os.Stdin
	} else {
		cmd.Stdin = os.NewFile(uintptr(syscall.Stdin), "/dev/stdin")
	}
	if prog.ForwardStderror {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.HideWindow = true

	err := cmd.Start()

	if err != nil {
		fmt.Printf("Process %s failed to start. %v\n", j.name, err.Error())
		j.restart = false
		return err
	}

	j.cmd = cmd

	go func() {
		err := cmd.Wait()
		ex := cmd.ProcessState.ExitCode()
		j.cmd = nil
		if err != nil {
			fmt.Printf("Process %s exited. (%d: %v)\n", j.name, ex, err)
		} else {
			fmt.Printf("Process %s exited. (%d)\n", j.name, ex)
		}

		if j.restart {
			j.Start()
		}
	}()

	return err
}

func (j *ProcessJob) Stop() (err error) {
	// fmt.Printf("Stop %s\n", j.name)
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

func (j *ProcessJob) Restart() error {
	// fmt.Printf("Restart %s\n", j.name)

	stopError := j.Stop()
	startErr := j.Start()

	return CombineErrors(stopError, startErr)
}

func NewProcessJob(name string, service shell.Service) *ProcessJob {
	s := ProcessJob{}
	s.service = &service
	s.name = name
	return &s
}
