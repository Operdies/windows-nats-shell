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
	// Reload the shell config
	ReloadConfig = "Shell.ReloadConfig"
	// Restart the shell
	RestartShell = "Shell.Restart"
)
