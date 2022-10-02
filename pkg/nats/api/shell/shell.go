package shell

import "gopkg.in/yaml.v3"

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
	ShellEvent = "Shell.Event"
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
	ForwardStdout    bool
	ForwardStderror  bool
	ForwardStdin     bool
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
	Services map[string]Service
}

type Event struct {
	Event  string
	NCode  int
	WParam uint64
	LParam uint64
}

// hshell codes
const (
	HSHELL_ACCESSIBILITYSTATE  = 11
	HSHELL_ACTIVATESHELLWINDOW = 3
	HSHELL_APPCOMMAND          = 12
	HSHELL_GETMINRECT          = 5
	HSHELL_LANGUAGE            = 8
	HSHELL_REDRAW              = 6
	HSHELL_TASKMAN             = 7
	HSHELL_WINDOWACTIVATED     = 4
	HSHELL_WINDOWCREATED       = 1
	HSHELL_WINDOWDESTROYED     = 2
	HSHELL_WINDOWREPLACED      = 13
)

func NewEvent(nCode int, wParam uintptr, lParam uintptr) Event {
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

	var e = Event{Event: evt, NCode: nCode, WParam: uint64(wParam), LParam: uint64(lParam)}
	return e
}
