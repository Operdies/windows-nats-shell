package windowmanager

import (
	"context"
	"fmt"
	"sort"
	"sync"
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
	maplock      = sync.Mutex{}
)

func IsIgnored(hwnd wintypes.HWND) bool {
	if hwnd == 0 {
		return true
	}
	v, ok := ignoredCache[hwnd]
	if ok && v {
		return true
	}
	maplock.Lock()
	defer maplock.Unlock()

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
	Config             *Config
	subs               []*nats.Subscription
	cancelLayoutChange context.CancelFunc
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
		// Newly created windows should be inserted without changing focus
		if sei.ShellCode == shell.HSHELL_WINDOWCREATED {
			if IsIgnored(wintypes.HWND(sei.WParam)) {
				return
			}
			// wia.HideBorder(wintypes.HWND(sei.WParam))
			// wm.calculateLayout(wintypes.HWND(sei.WParam), false)
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

func (wm *WindowManager) selectSibling(siblingOf wintypes.HWND, reverse bool) {
	// Cancel any currently running animation
	if wm.cancelLayoutChange != nil {
		wm.cancelLayoutChange()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	wm.cancelLayoutChange = cancel
	done := false
	go func() {
		<-ctx.Done()
		done = true
	}()

	for i := 0; i < 50; i++ {
		if done {
			return
		}
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
			go wm.calculateLayout(subject, reverse, ctx)
			return
		}

		next := handles[0]
		if winapi.SuperFocusStealer(handles[0]) {
			go wm.calculateLayout(next, reverse, ctx)
			return
		}
		// If the operation failed, wait a bit and try again
		time.Sleep(time.Millisecond)
	}
}

func (wm *WindowManager) calculateLayout(mainWindow wintypes.HWND, reverse bool, ctx context.Context) {
	animationSteps := 15
	otherWindows := partition(mainWindow)
	cnt := len(otherWindows)
	screenRect := screen.GetScreenRect()

	for i, w := range otherWindows {
		// The position of this window on the perimeter of the monitor
		position := float64(i) / float64(cnt)
		point := screenRect.GetPointOnPerimeter(position)
		currentWindowRect := winapi.GetWindowRect(w)
		newWindowRect := currentWindowRect.CenterAround(point)
		anim := currentWindowRect.Animate(newWindowRect, animationSteps)
		go wia.AnimateRectWithContext(w, anim, ctx)
	}

	currentMainRect := winapi.GetWindowRect(mainWindow)
	newMainRect := screenRect.Scale(wm.Config.Ratio)
	go wia.AnimateRectWithContext(mainWindow, currentMainRect.Animate(newMainRect, animationSteps), ctx)
}

func (wm *WindowManager) PrevWindow(hwnd wintypes.HWND) {
	wm.selectSibling(hwnd, true)
}
func (wm *WindowManager) NextWindow(hwnd wintypes.HWND) {
	wm.selectSibling(hwnd, false)
}
