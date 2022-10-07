package screen

const (
	// Get the current resolution
	GetResolution = "Screen.GetResolution"
	// Set the current resolution
	SetResolution = "Screen.SetResolution"
)

type Resolution struct {
	Width  uint32
	Height uint32
}
