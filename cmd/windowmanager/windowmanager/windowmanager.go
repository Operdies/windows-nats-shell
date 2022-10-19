package windowmanager

import (
	"fmt"
	"sort"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/input"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/utils/query"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/winapi/screen"
	wia "github.com/operdies/windows-nats-shell/pkg/winapi/winapiabstractions"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

type layout = string

const (
	revolver layout = "revolver"
)

var (
	actionKey    = input.VK_MAP["nullkey"]
	ignored      = map[string]bool{"Background": true}
	ignoredCache = map[wintypes.HWND]bool{}
)

func IsIgnored(hwnd wintypes.HWND) bool {
	if hwnd == 0 {
		return true
	}
	v, ok := ignoredCache[hwnd]
	if ok && v {
		return true
	}
	title, _ := wia.GetWindowTextEasy(hwnd)
	v, ok = ignored[title]
	ignoredCache[hwnd] = v && ok
	return v && ok
}

type Config struct {
	Layout     layout
	CycleKey   string
	CycleVKey  input.VKEY
	ActionKey  string
	ActionVKey input.VKEY
	// The gap between the centered window and a revolving window as a percentage of screen height
	Gap float64
	// The percentage of screen space used for the focused window
	Ratio float64
}

type WindowManager struct {
	Config *Config
	subs   []*nats.Subscription
}

func Create(cfg Config) WindowManager {
	switch cfg.Layout {
	case revolver:
		break
	default:
		panic(fmt.Errorf("Unknown layout %v.", cfg.Layout))
	}

	if cfg.Gap > 1 || cfg.Gap < 0 {
		panic("Gap must be a number between 0 and 1")
	}

	if cfg.Ratio > 1 || cfg.Ratio < 0 {
		panic("Ratio must be a number between 0 and 1")
	}

	ck := input.VK_MAP[cfg.CycleKey]
	ak := input.VK_MAP[cfg.ActionKey]
	if ck == ak {
		panic("cycleKey and actionKey cannot be identical")
	}
	cfg.CycleVKey = ck
	cfg.ActionVKey = ak

	var man WindowManager
	man.Config = &cfg
	return man
}

func (wm *WindowManager) Monitor() {
	nc := client.Default()
	sub, err := nc.Subscribe.WH_SHELL(func(sei shell.ShellEventInfo) {
		if sei.ShellCode == shell.HSHELL_WINDOWCREATED {
			if IsIgnored(wintypes.HWND(sei.WParam)) {
				return
			}
			// wia.HideBorder(wintypes.HWND(sei.WParam))
			wm.calculateLayout(wintypes.HWND(sei.WParam), false)
		} else if sei.ShellCode == shell.HSHELL_WINDOWDESTROYED {
			// current := winapi.GetForegroundWindow()
			// wm.calculateLayout(current)
		}
	})
	if err != nil {
		panic(err)
	}
	wm.subs = append(wm.subs, sub)
}

func (wm *WindowManager) Close() {
	for _, s := range wm.subs {
		s.Unsubscribe()
	}
}

func partition(around wintypes.HWND) []wintypes.HWND {
	windows := wia.GetVisibleWindows()
	handles := query.Select(windows, func(w wintypes.Window) wintypes.HWND { return w.Handle })
	handles = query.Filter(handles, func(hwnd wintypes.HWND) bool { return !IsIgnored(hwnd) })
	// Ensure the handles are always ordered the same way
	sort.Slice(handles, func(i, j int) bool {
		return handles[i] < handles[j]
	})

	// Remove 'siblingOf' from the list, and rearrange the list
	for i, w := range handles {
		if w == around {
			handles = append(handles[i+1:], handles[:i]...)
			break
		}
	}

	return handles
}

func (wm *WindowManager) SelectSibling(siblingOf wintypes.HWND, reverse bool) {
	for i := 0; i < 3; i++ {
		subject := siblingOf
		if subject == 0 {
			subject = winapi.GetForegroundWindow()
		}
		if IsIgnored(subject) {
			return
		}

		handles := partition(subject)
		// Reverse the array to cycle backwards instead of forwards
		if reverse {
			for i, j := 0, len(handles)-1; i < j; i, j = i+1, j-1 {
				handles[i], handles[j] = handles[j], handles[i]
			}
		}

		if len(handles) == 0 {
			go wm.calculateLayout(subject, reverse)
			return
		}

		for _, w := range handles {
			success := winapi.SuperFocusStealer(w)
			if success {
				go wm.calculateLayout(w, reverse)
				return
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func (wm *WindowManager) calculateLayout(mainWindow wintypes.HWND, reverse bool) {
	screenSize := screen.GetResolution()
	sw := int32(screenSize.Width)
	sh := int32(screenSize.Height)
	wMargin := int32(float64(sw) * wm.Config.Ratio)
	hMargin := int32(float64(sh) * wm.Config.Ratio)
	mainRect := wintypes.RECT{
		Left:   sw - wMargin,
		Top:    sh - hMargin,
		Right:  wMargin,
		Bottom: hMargin,
	}
	rem := partition(mainWindow)
	cnt := len(rem)
	screenRect := wintypes.RECT{
		Left: 0, Top: 0,
		Right: sw, Bottom: sh,
	}
	for i, w := range rem {
		position := float64(i) / float64(cnt)
		point := screenRect.GetPointOnPerimeter(position)
		tit, _ := wia.GetWindowTextEasy(w)
		fmt.Printf("Centering %v on %v\n", tit, point)
		winRect := winapi.GetWindowRect(w)
		newRect := winRect.CenterAround(point)
		if newRect.Width() == mainRect.Width() {
			newRect = newRect.Scale(0.85)
		}
		fmt.Printf("winRect: %v\n", winRect)
		fmt.Printf("newRect: %v\n", newRect)

		n := 50
		animationSteps := make([]wintypes.RECT, 0, n)
		var start, end float64
		if !reverse {
			start = float64(i+1) / float64(cnt)
		} else {
			start = float64(i-1) / float64(cnt)
		}
		end = float64(i) / float64(cnt)
		step := (end - start) / float64(n)
		for j := 1; j <= n; j++ {
			position = start + float64(j)*step
			point := screenRect.GetPointOnPerimeter(position)
			animationSteps = append(animationSteps, newRect.CenterAround(point))
		}
		go wia.AnimateRect(w, animationSteps, time.Millisecond*200)
	}

	fmt.Printf("mainRect: %+v\n", mainRect)
	wia.SetWindowRect(mainWindow, mainRect)
}

func (wm *WindowManager) PrevWindow(hwnd wintypes.HWND) {
	wm.SelectSibling(hwnd, true)
}
func (wm *WindowManager) NextWindow(hwnd wintypes.HWND) {
	wm.SelectSibling(hwnd, false)
}
