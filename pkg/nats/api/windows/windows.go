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
	// Move the window 
	Move = "Window.Move"
	// Resize the window 
	Resize = "Window.Resize"
	// Minimize the window 
	Minimize = "Window.Minimize"
	// Restore the window
	Restore = "Window.Restore"
	// Maximize the window
	Maximize = "Window.Maximize"
	// Focus the previous window 
	FocusPrevious = "Window.FocusPrevious"
	// Focus the next window 
	FocusNext = "Window.FocusNext"
)
