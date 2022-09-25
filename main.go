package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
)

type Services struct {
	Name             string
	Executable       string
	Arguments        []string
	WorkingDirectory string
	AutoRestart      *bool
	ForwardStdout    bool
	ForwardStdin     bool
	Environment      []string
}
type Config struct {
	Services []Services
}

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

func myFilter[T1 any](source []T1, filter func(T1) bool) []T1 {
	cp := make([]T1, len(source))
	k := 0
	for i := 0; i < len(source); i = i + 1 {
		item := source[i]
		if filter(item) {
			cp[k] = item
			k = k + 1
		}
	}
	return cp[:k]
}

func (f forwardedStdout) Write(p []byte) (n int, err error) {
	fmt.Printf("%s: [%s]\n", f.name, string(p))
	n = len(p)
	err = nil
	return
}

func startProgram(prog Services) {
	start := func() {
		cmd := exec.Command(prog.Executable, prog.Arguments...)
		cmd.Env = append(cmd.Env, prog.Environment...)
		cmd.Dir = prog.WorkingDirectory
		if prog.ForwardStdin {
			cmd.Stdin = os.Stdin
		}
		if prog.ForwardStdout {
			var f forwardedStdout
			f.name = prog.Name
			cmd.Stdout = os.Stdout
		}
		cmd.Run()
	}

	i := 0
  autoRestart := valueOrDefault(prog.AutoRestart, true)
	fmt.Printf("Starting Process: %v %v in %v (%v) with AutoRestart: %v\n", prog.Executable, prog.Arguments, prog.WorkingDirectory, prog.Name, autoRestart)
  if autoRestart {
		for {
			start()
			i += 1
			fmt.Printf("Starting Process(%d): %v %v in %v (%v)\n", i, prog.Executable, prog.Arguments, prog.WorkingDirectory, prog.Name)
		}
	} else {
		start()
	}
}

func parseCfg(path string) *Config {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg Config
	json.Unmarshal(content, &cfg)
	return &cfg
}

func loadConfig() *Config {
	fileExists := func(f string) bool {
		_, err := os.Stat(f)
		return err == nil
	}

	for _, cand := range getConfigPaths() {
		if fileExists(cand) {
			return parseCfg(cand)
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

func applyConfig(cfg *Config) {
	for _, prog := range cfg.Services {
		go startProgram(prog)
	}
}

func superFlusher(){
	buf := make([]byte, 1)
	for {
		// We need to flush the stdin buffer in order to other processes to be able to read it
		_, eof := os.Stdin.Read(buf)
		if eof == io.EOF {
			return
		}
	}
}

func main() {
  home, _ := os.UserHomeDir()
  os.Chdir(home)
	config := loadConfig()
	if config == nil {
		panic("No config file found")
	}

	applyConfig(config)
  go superFlusher()
  select {}
}
