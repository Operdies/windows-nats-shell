package service

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
)

type Jobber interface {
	Start() error
	Stop() error
	Restart() error
}

type ProcessJob struct {
	restart      bool
	startCount int
	service      *shell.Service
	cmd          *exec.Cmd
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
  if j.cmd != nil {
    return fmt.Errorf("Process %s is already running.", j.service.Name)
  }
  j.startCount += 1
	fmt.Printf("Starting %s. (%d)\n", j.service.Name, j.startCount)

	j.restart = *j.service.AutoRestart == true
	prog := j.service
	cmd := exec.Command(prog.Executable, prog.Arguments...)
	ref := fmt.Sprintf("MINIMAL_SHELL_SERVICE_NAME=%s", prog.Name)
	env := os.Environ()
	env = append(env, prog.Environment...)
	env = append(env, ref)
	cmd.Env = env
	cmd.Dir = prog.WorkingDirectory

	if prog.ForwardStderror {
		cmd.Stderr = os.Stderr
	}
	if prog.ForwardStdin {
		cmd.Stdin = os.Stdin
	}

	if prog.ForwardStdout {
		cmd.Stdout = os.Stdout
	}

	err := cmd.Start()
	j.cmd = cmd

	if err != nil {
		fmt.Printf("Job %s failed to start. Auto-restart disabled.\n", prog.Name)
		j.restart = false
	}

	go func() {
		cmd.Wait()
		if j.restart {
			j.Start()
		}
	}()

	return err
}

func (j *ProcessJob) Stop() (err error) {
	j.restart = false
	if j.cmd == nil {
		return fmt.Errorf("Process %s is not running.", j.service.Name)
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
	stopError := j.Stop()
	startErr := j.Start()

	return CombineErrors(stopError, startErr)
}

func NewProcessJob(service shell.Service) *ProcessJob {
	s := ProcessJob{}
	s.service = &service
	return &s
}
