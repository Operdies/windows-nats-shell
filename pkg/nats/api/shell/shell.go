package shell

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
	// Get the currently loaded config
	Config = "Shell.Config"
	// Set a new config
	SetConfig = "Shell.SetConfig"
	// Add a new service
	AddService = "Shell.AddService"
	// Remove an existing service
	RemoveService = "Shell.RemoveService"
)

type Service struct {
	Custom map[string]string
	// The full path to the exectuable file
	Executable string
	Arguments  []string
	// Defaults to cwd
	WorkingDirectory string
	AutoRestart      *bool
	ForwardStdout    bool
	ForwardStderror  bool
	ForwardStdin     bool
	// Any environment variables that should be defined
	Environment []string
}

type Configuration struct {
	Services map[string]Service
}
