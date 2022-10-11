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
	// Some CBT happened
	CBTEvent = "Shell.CBTEvent"
	// Some kyboard event happened
	KeyboardEvent = "Shell.KeyboardEvent"
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

func GetCustom[T any](s Service) (result T, err error) {
	custom := s.Custom
	buffer, err := yaml.Marshal(custom)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(buffer, &result)
	return
}

type Configuration struct {
	Path     string // Path to the file the config was loaded from
	Services map[string]Service
}

type CBTEventInfo struct {
	Event   string
	CBTCode WM_CBT_CODE
	WParam  uint64
	LParam  uint64
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
	HC_NOREMOVE              = 3
)

// hshell codes

type WM_SHELL_CODE = int

const (
	HSHELL_ACCESSIBILITYSTATE  WM_SHELL_CODE = 11
	HSHELL_ACTIVATESHELLWINDOW               = 3
	HSHELL_APPCOMMAND                        = 12
	HSHELL_GETMINRECT                        = 5
	HSHELL_LANGUAGE                          = 8
	HSHELL_REDRAW                            = 6
	HSHELL_TASKMAN                           = 7
	HSHELL_WINDOWACTIVATED                   = 4
	HSHELL_WINDOWCREATED                     = 1
	HSHELL_WINDOWDESTROYED                   = 2
	HSHELL_WINDOWREPLACED                    = 13
)

type WM_CBT_CODE = int

const (
	HCBT_ACTIVATE     WM_CBT_CODE = 5
	HCBT_CLICKSKIPPED             = 6
	HCBT_CREATEWND                = 3
	HCBT_DESTROYWND               = 4
	HCBT_KEYSKIPPED               = 7
	HCBT_MINMAX                   = 1
	HCBT_MOVESIZE                 = 0
	HCBT_QS                       = 2
	HCBT_SETFOCUS                 = 9
	HCBT_SYSCOMMAND               = 8
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

func WhCbtEvent(nCode WM_CBT_CODE, wParam uintptr, lParam uintptr) CBTEventInfo {
	var mapping = map[int]string{
		HCBT_ACTIVATE:     "HCBT_ACTIVATE",
		HCBT_CLICKSKIPPED: "HCBT_CLICKSKIPPED",
		HCBT_CREATEWND:    "HCBT_CREATEWND",
		HCBT_DESTROYWND:   "HCBT_DESTROYWND",
		HCBT_KEYSKIPPED:   "HCBT_KEYSKIPPED",
		HCBT_MINMAX:       "HCBT_MINMAX",
		HCBT_MOVESIZE:     "HCBT_MOVESIZE",
		HCBT_QS:           "HCBT_QS",
		HCBT_SETFOCUS:     "HCBT_SETFOCUS",
		HCBT_SYSCOMMAND:   "HCBT_SYSCOMMAND",
	}
	evt, ok := mapping[nCode]
	if !ok {
		evt = "UNKOWN_EVENT"
	}

	var e = CBTEventInfo{Event: evt, CBTCode: nCode, WParam: uint64(wParam), LParam: uint64(lParam)}
	return e
}

type KeyboardEventInfo struct {
	KeyboardEventCode KeyEventCode
	VirtualKeyCode    uint64
	// The repeat count. The value is the number of times the keystroke is repeated as a result of the user's holding down the key.
	// bit 0-15
	RepeatCount uint64
	// The scan code. The value depends on the OEM.
	// bit 16-23
	ScanCode uint64
	// Indicates whether the key is an extended key, such as a function key or a key on the numeric keypad. The value is 1 if the key is an extended key; otherwise, it is 0.
	// bit 24
	IsExtended bool
	// True if ALT is down, otherwise 0
	// bit 29
	ContextCode bool
	// The previous key state. The value is 1 if the key is down before the message is sent; it is 0 if the key is up.
	// bit 30
	PreviousKeyState bool
	// The transition state. The value is 0 if the key is being pressed and 1 if it is being released.
	// bit 31
	TransitionState bool
}

func bitRange(number uint64, start, end uint8) uint64 {
	var mask uint64
	var n uint64
	n = number >> start
	rng := end - start
	mask = (1 << (rng + 1)) - 1
	return n & mask
}

func WhKeyboardEvent(nCode KeyEventCode, wParam uintptr, lParam uintptr) KeyboardEventInfo {
	evt := KeyboardEventInfo{}
	evt.KeyboardEventCode = nCode
	evt.VirtualKeyCode = uint64(wParam)
	evt.RepeatCount = bitRange(uint64(lParam), 0, 15)
	evt.ScanCode = bitRange(uint64(lParam), 16, 23)
	evt.IsExtended = bitRange(uint64(lParam), 24, 24) == 1
	evt.ContextCode = bitRange(uint64(lParam), 29, 29) == 1
	evt.PreviousKeyState = bitRange(uint64(lParam), 30, 30) == 1
	evt.TransitionState = bitRange(uint64(lParam), 31, 31) == 1
	return evt
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
