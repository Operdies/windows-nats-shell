package system

const (
	// Get executables in path, and any program from the startmenu
	GetPrograms = "System.GetPrograms"
	/* Launch a given program. If the argument is an absolute path, it will be launched like 'start $1'
	   if it is a base name (e.g. pwsh), then we will attempt to match against a program in PATH or in a startmenu entry  */
	LaunchProgram = "System.LaunchProgram"
)
