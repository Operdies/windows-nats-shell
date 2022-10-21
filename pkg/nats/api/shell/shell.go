package shell

import (
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// Restart a service by name
	RestartService = "Shell.RestartService"
	// Stop a service by name. AutoRestarting services
	// will not be restarted until the service is restarted,
	// or the shell is reloaded.
	StopService = "Shell.StopService"
	// Start a service by name
	StartService = "Shell.StartService"
	// Restart the shell
	RestartShell = "Shell.Restart"
	// Get the full shell config
	ShellConfig = "Shell.ShellConfig"
	// Get the config of a loaded service
	GetService = "Shell.Config"
	// Set a new config
	SetConfig = "Shell.SetConfig"
	// Add a new service
	AddService = "Shell.AddService"
	// Remove an existing service
	RemoveService = "Shell.RemoveService"
	// Quit the shell
	QuitShell = "Shell.Quit"
	// Some shell event happened
	ShellEvent = "Shell.ShellEvent"
	// Some keyboard event happened
	KeyboardEvent = "Shell.KeyboardEvent"
	// Some mouse event happened
	MouseEvent = "Shell.MouseEvent"
)

const (
	SERVICE_ENV_KEY = "_SHELL_SERVICE_NAME_"
)

type Service struct {
	Custom map[string]interface{}
	// The full path to the exectuable file
	Executable string
	Arguments  []string
	// Defaults to cwd
	WorkingDirectory string
	Enabled          *bool
	AutoRestart      *bool
	Visible          bool
	// Any environment variables that should be defined
	Environment []string
}

type Configuration struct {
	Path           string // Path to the file the config was loaded from
	Services       map[string]Service
	ServiceConfigs map[string]any
}

type cfg2 struct {
	Services map[string]any
}

type ShellEventInfo struct {
	Event     string
	ShellCode WM_SHELL_CODE
	WParam    uint64
	LParam    uint64
}

type KeyEventCode = int

const (
	HC_ACTION   KeyEventCode = 0
	HC_NOREMOVE KeyEventCode = 3
)

// hshell codes

type WM_SHELL_CODE = int

const (
	HSHELL_WINDOWCREATED       WM_SHELL_CODE = 1
	HSHELL_WINDOWDESTROYED     WM_SHELL_CODE = 2
	HSHELL_ACTIVATESHELLWINDOW WM_SHELL_CODE = 3
	HSHELL_WINDOWACTIVATED     WM_SHELL_CODE = 4
	HSHELL_GETMINRECT          WM_SHELL_CODE = 5
	HSHELL_REDRAW              WM_SHELL_CODE = 6
	HSHELL_TASKMAN             WM_SHELL_CODE = 7
	HSHELL_LANGUAGE            WM_SHELL_CODE = 8
	HSHELL_ACCESSIBILITYSTATE  WM_SHELL_CODE = 11
	HSHELL_APPCOMMAND          WM_SHELL_CODE = 12
	HSHELL_WINDOWREPLACED      WM_SHELL_CODE = 13
)

func WhShellEvent(nCode WM_SHELL_CODE, wParam uintptr, lParam uintptr) ShellEventInfo {
	var mapping = map[int]string{
		HSHELL_ACCESSIBILITYSTATE:  "HSHELL_ACCESSIBILITYSTATE",
		HSHELL_ACTIVATESHELLWINDOW: "HSHELL_ACTIVATESHELLWINDOW",
		HSHELL_APPCOMMAND:          "HSHELL_APPCOMMAND",
		HSHELL_GETMINRECT:          "HSHELL_GETMINRECT",
		HSHELL_LANGUAGE:            "HSHELL_LANGUAGE",
		HSHELL_REDRAW:              "HSHELL_REDRAW",
		HSHELL_TASKMAN:             "HSHELL_TASKMAN",
		HSHELL_WINDOWACTIVATED:     "HSHELL_WINDOWACTIVATED",
		HSHELL_WINDOWCREATED:       "HSHELL_WINDOWCREATED",
		HSHELL_WINDOWDESTROYED:     "HSHELL_WINDOWDESTROYED",
		HSHELL_WINDOWREPLACED:      "HSHELL_WINDOWREPLACED",
	}

	evt, ok := mapping[nCode]
	if !ok {
		evt = "UNKOWN_EVENT"
	}

	var e = ShellEventInfo{Event: evt, ShellCode: nCode, WParam: uint64(wParam), LParam: uint64(lParam)}
	return e
}

func parseCfg(path string) (config *Configuration, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg Configuration
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return
	}
	var cfgHelper cfg2
	err = yaml.Unmarshal(content, &cfgHelper)
	config = &cfg
	config.Path = path
	config.ServiceConfigs = cfgHelper.Services
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
	result := make([]string, 1)
	appdata, _ := os.UserHomeDir()
	result[0] = path.Join(appdata, "AppData", "Local", "windows-nats-shell", "config.yml")
	exeDir := getExeDir()
	fix := func(s string) string {
		return strings.ReplaceAll(s, `/`, `\`)
	}
	if exeDir != "" {
		result = append(result, path.Join(exeDir, "config.yml"))
	}

	wd, _ := os.Getwd()
	result = append(result, path.Join(wd, "config.yml"))

	for i := range result {
		result[i] = fix(result[i])
	}

	return result
}

func LoadDefault() *Configuration {
	path := loadConfig()
	if path == nil {
		panic("No config file found")
	}
	cfg, err := parseCfg(*path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *Configuration) Reload() (cfg *Configuration, err error) {
	cfg, err = parseCfg(c.Path)
	if err != nil {
		return c, err
	}
	return cfg, nil
}
