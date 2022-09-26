package api

const (
  // Poll for visible windows
	Windows        = "Windows"
  // Subscribe to window events
	WindowsUpdated = "Windows.Updated"
  // Poll for window focused state 
  IsWindowFocused = "Window.Focused"
  // Attempt to bring the selected window to the foreground
  SetFocus = "Window.SetFocus"
  // Get executables in path, and any program from the startmenu
  GetPrograms = "System.GetPrograms"
  /* Launch a given program. If the argument is an absolute path, it will be launched like 'start $1'
   if it is a base name (e.g. pwsh), then we will attempt to match against a program in PATH or in a startmenu entry  */
  LaunchProgram = "System.LaunchProgram"
)
