package windowmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/input"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
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
	ScaleY     float64
	ScaleX     float64
	// The scale of non-focused windows
	SmallScale float64
	// The location of the screen 'perimeter'
	Perimeter       float64
	AnimationFrames int
	AnimationTime   int
	// The number of windows that will be in the center
	Barrels int
	// The spacing between centered windows
	Padding int
}

type WindowManager struct {
	Config             *Config
	subs               []*nats.Subscription
	cancelLayoutChange context.CancelFunc
	cancelContextLock  sync.Mutex
	windowList         []wintypes.HWND
	windowListLock     sync.Mutex
}

func Create(cfg Config) *WindowManager {
	switch cfg.Layout {
	case revolver:
		break
	default:
		panic(fmt.Errorf("Unknown layout %v.", cfg.Layout))
	}

	if cfg.ScaleX > 1 || cfg.ScaleX < 0 {
		panic("ScaleX must be a number between 0 and 1")
	}
	if cfg.ScaleY > 1 || cfg.ScaleY < 0 {
		panic("ScaleY must be a number between 0 and 1")
	}

	if cfg.ScaleX == 0 {
		cfg.ScaleX = 0.8
	}
	if cfg.ScaleY == 0 {
		cfg.ScaleY = 0.8
	}

	if cfg.SmallScale == 0 {
		cfg.SmallScale = 1.0
	}
	if cfg.Perimeter == 0 {
		cfg.Perimeter = 0.9
	}
	if cfg.AnimationFrames == 0 {
		cfg.AnimationFrames = 15
	}
	if cfg.AnimationTime == 0 {
		cfg.AnimationTime = 200
	}
	if cfg.Barrels == 0 {
		cfg.Barrels = 1
	}
	if cfg.Barrels < 1 {
		panic("There must be at least one barrel")
	}

	ck := input.VK_MAP[cfg.CycleKey]
	ak := input.VK_MAP[cfg.ActionKey]
	if ck == ak {
		panic("cycleKey and actionKey cannot be identical")
	}
	cfg.CycleVKey = ck
	cfg.ActionVKey = ak

	var man WindowManager
	fmt.Printf("cfg: %+v\n", cfg)
	man.Config = &cfg
	return &man
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

func (wm *WindowManager) cancelAndCreateContext() context.Context {
	wm.cancelContextLock.Lock()
	defer wm.cancelContextLock.Unlock()
	// Cancel any currently running animation
	if wm.cancelLayoutChange != nil {
		wm.cancelLayoutChange()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(wm.Config.AnimationTime))
	wm.cancelLayoutChange = cancel
	return ctx
}

func reallyFocus(h wintypes.HWND, tries int, ctx context.Context) {
	done := false
	go func() {
		<-ctx.Done()
		done = true
	}()
	for i := 0; i < tries && !done; i++ {
		if winapi.SuperFocusStealer(h) {
			return
		}
		time.Sleep(time.Millisecond)
	}
}
func (wm *WindowManager) cycleWindows(reverse bool) {
	wm.windowListLock.Lock()
	defer wm.windowListLock.Unlock()
	wm.updateWindowList()
	ctx := wm.cancelAndCreateContext()

	if len(wm.windowList) > 1 {
		if reverse {
			// Shift the list and put the first element last
			wm.windowList = append(wm.windowList[1:], wm.windowList[0:1]...)
		} else {
			// Shift the list and put the last element first
			wm.windowList = append(wm.windowList[len(wm.windowList)-1:], wm.windowList[:len(wm.windowList)-1]...)
		}
	}
	go wm.calculateLayout(wm.windowList, ctx)
	reallyFocus(wm.windowList[0], 50, ctx)
}

func (wm *WindowManager) swapWindows(a, b wintypes.HWND) {
	if a == 0 || b == 0 {
		return
	}
	if a == b {
		return
	}

	wm.windowListLock.Lock()
	defer wm.windowListLock.Unlock()
	wm.updateWindowList()

	var aIdx, bIdx int
	for i, w := range wm.windowList {
		if w == a {
			aIdx = i
		}
		if w == b {
			bIdx = i
		}
	}

	ctx := wm.cancelAndCreateContext()
	go reallyFocus(a, 50, ctx)
	wm.windowList[aIdx], wm.windowList[bIdx] = wm.windowList[bIdx], wm.windowList[aIdx]
	r1 := winapi.GetWindowRect(a)
	r2 := winapi.GetWindowRect(b)
	go wia.AnimateRectWithContext(a, r1.Animate(r2, wm.Config.AnimationFrames, true), ctx)
	go wia.AnimateRectWithContext(b, r2.Animate(r1, wm.Config.AnimationFrames, true), ctx)
	wm.fixZOrder()
}

func split[T any](lst []T, index int) (fst, second []T) {
	if index < 0 {
		return make([]T, 0), lst
	}
	if len(lst) <= index {
		return lst, make([]T, 0)
	}
	return lst[:index], lst[index:]
}

func (wm *WindowManager) setPerimeterWindows(otherWindows []wintypes.HWND, screenRect windows.Rect, ctx context.Context) {
	perimeter := screenRect.Scale(wm.Config.Perimeter)
	cnt := len(otherWindows)
	for i, w := range otherWindows {
		// The position of this window on the perimeter of the monitor
		position := float64(i) / float64(cnt)
		point := perimeter.GetPointOnPerimeterCircleMethod(position)
		currentWindowRect := winapi.GetWindowRect(w)
		newWindowRect := screenRect.CenterAround(point).Scale(wm.Config.SmallScale)
		anim := currentWindowRect.Animate(newWindowRect, wm.Config.AnimationFrames, true)
		go wia.AnimateRectWithContext(w, anim, ctx)
	}
}

func (wm *WindowManager) updateWindowList() {
	windows := wia.GetVisibleWindows()

	handles := query.Select(windows, func(w wintypes.Window) wintypes.HWND { return w.Handle })
	handles = query.Filter(handles, func(hwnd wintypes.HWND) bool { return !IsIgnored(hwnd) && !wia.WindowMinimized(hwnd) })

	// Add any missing windows to the list
	for _, h := range handles {
		if query.Contains(wm.windowList, h) == false {
			fmt.Printf("added missing: %v\n", h)
			wm.windowList = append(wm.windowList, h)
		}
	}
	// Remove any superfluous windows from the list
	for i, w := range wm.windowList {
		if query.Contains(handles, w) == false {
			fmt.Printf("evicting: %v\n", w)
			wm.windowList = append(wm.windowList[:i], wm.windowList[i+1:]...)
		}
	}
}

func (wm *WindowManager) fixZOrder() {
	windowList := wm.windowList
	for i := 1; i < len(windowList); i++ {
		prev := windowList[i-1]
		this := windowList[i]
		wia.SetZOrder(prev, this)
	}
}
func (wm *WindowManager) calculateLayout(windowList []wintypes.HWND, ctx context.Context) {
	wm.fixZOrder()
	screenRect := screen.GetScreenRect()

	middleWindows, perimeterWindows := split(windowList, wm.Config.Barrels)

	go wm.setPerimeterWindows(perimeterWindows, screenRect, ctx)

	// cnt := len(middleWindows)
	mainArea := screenRect.ScaleX(wm.Config.ScaleX).ScaleY(wm.Config.ScaleY)

	p := int32(wm.Config.Padding)
	scale := 1.0 / float64(len(middleWindows))
	step := int(mainArea.Width()) / (len(middleWindows))
	for i, m := range middleWindows {
		current := winapi.GetWindowRect(m)
		clientArea := mainArea.ScaleX(scale).Align(mainArea, windows.TopLeft).Pad(p, p).Translate(step*i, 0)
		go wia.AnimateRectWithContext(m, current.Animate(clientArea, wm.Config.AnimationFrames, true), ctx)
	}
}

func (wm *WindowManager) FocusPrevWindow() {
	wm.cycleWindows(true)
}
func (wm *WindowManager) FocusNextWindow() {
	wm.cycleWindows(false)
}
func (wm *WindowManager) FocusThisWindow(next wintypes.HWND) {
	current := winapi.GetForegroundWindow()
	wm.swapWindows(next, current)
}
