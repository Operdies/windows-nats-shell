package windows 
const (
	// Poll for visible windows
	GetWindows = "Windows.GetWindows"
	// Subscribe to window events
	WindowsUpdated = "Windows.Updated"
	// Poll for window focused state
	IsWindowFocused = "Window.Focused"
	// Attempt to bring the selected window to the foreground
	SetFocus = "Window.SetFocus"
)
