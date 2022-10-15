package keymap

import (
	"fmt"
	"sort"
	"strings"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	"github.com/operdies/windows-nats-shell/pkg/utils"
)

var (
	user32 = syscall.MustLoadDLL("user32.dll")

	VkKeyScanA = user32.MustFindProc("VkKeyScanA")
)

type action struct {
	Nats struct {
		Subject string
		Payload any
	}
}
type keymap struct {
	Keymap []struct {
		Keys    string
		Actions []action
	}
}

type mod string
type VKEY uint32

const (
	alt   mod = "alt"
	none      = ""
	ctrl      = "ctrl"
	win       = "win"
	shift     = "shift"
)

var (
	modmap = map[string]mod{
		"alt":   alt,
		"ctrl":  ctrl,
		"win":   win,
		"shift": shift,
	}
	vkeyModMap = map[VKEY]mod{
		VK_MENU:    alt,
		VK_CONTROL: ctrl,
		VK_LWIN:    win,
		VK_SHIFT:   shift,
	}
)

type BindingTree struct {
	HasAction bool
	Action    []action
	Subtrees  map[uint32]*BindingTree
}

type Keymap struct {
	Bindings   *BindingTree
	activeMods map[VKEY]bool
}

func parseKey(key string) VKEY {
	key = strings.ToLower(key)
	bs := []byte(key)

	if code, ok := vk_map[key]; ok {
		return code
	}
	if len(bs) > 1 || len(bs) < 1 {
		err := fmt.Errorf("Key not supported: %s", key)
		panic(err)
	}

	a, _, _ := VkKeyScanA.Call(uintptr(bs[0]))
	return VKEY(a)
}

type hotkey struct {
	mods    []uint32
	actions []action
}

func sortMods(mods []uint32) {
	sort.Slice(mods, func(i, j int) bool {
		return mods[i] < mods[j]
	})
}

func ParseMod(parts []string) hotkey {
	mods := make([]uint32, 0, len(parts))

	for _, m := range parts {
		k := parseKey(m)
		if k == 0 {
			panic(fmt.Sprintf("I don't understand what '%s' is. (%v)", m, modmap))
		}
		mods = append(mods, uint32(k))
	}

	sortMods(mods)

	return hotkey{mods: mods}
}

var (
	nc *nats.Conn
)

func init() {
	var err error
	nc, err = nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
}

func unleash(m *BindingTree) {
	for i, act := range m.Action {
		msg := utils.EncodeAny(act.Nats.Payload)
		nc.Publish(act.Nats.Subject, msg)
		fmt.Printf("%d) Unleash: %+v\nWith payload:\n%v\n", i, act.Nats, string(msg))
	}
}

func getBinding(k *Keymap, vkey VKEY) *BindingTree {
	mods := make([]uint32, 0, len(k.activeMods)+1)
	for m, v := range k.activeMods {
		if m == vkey {
			continue
		}
		if v {
			mods = append(mods, uint32(m))
		}
	}
	mods = append(mods, uint32(vkey))
	sortMods(mods)

	var ok bool
	root := k.Bindings
	for _, m := range mods {
		root, ok = root.Subtrees[m]
		if ok == false || root == nil {
			return nil
		}
	}
	return root
}

func handleKey(k *Keymap, kei shell.KeyboardEventInfo, bmap *BindingTree) {
	// We can't differentiate presses and holds from the event info.
	// But we know it's a hold if the key is already mapped in activeMods
	keyDown := kei.TransitionState == false

	// Update the modifier state if the key is a modifier
	vkey := VKEY(kei.VirtualKeyCode)
	// Clear the keymap when escape is pressed
	if vkey == VK_ESCAPE {
		k.activeMods = map[VKEY]bool{}
		return
	}

	previousState, ok := k.activeMods[vkey]
	// If the value was not in the map, it should be considered 'false'
	previousState = ok && previousState
	isRelease := previousState && !keyDown
	isPress := !previousState && keyDown

	if isPress || isRelease {
		k.activeMods[vkey] = isPress
	}

	// Let's only ever fire events when keys are released
	if bmap != nil && bmap.HasAction && isRelease {
		go unleash(bmap)
	}
}
func (k *Keymap) ProcessEvent(kei shell.KeyboardEventInfo) bool {
	// Return immediately after determining whether this key is handled.
	binding := getBinding(k, VKEY(kei.VirtualKeyCode))
	ret := binding != nil && binding.HasAction
	go handleKey(k, kei, binding)
	return ret
}

func buildTree(keys []hotkey) *BindingTree {
	newTree := func() *BindingTree {
		var t BindingTree
		t.HasAction = false
		t.Subtrees = map[uint32]*BindingTree{}
		return &t
	}

	result := newTree()

	for _, k := range keys {
		node := result
		for _, m := range k.mods {
			if subtree, ok := node.Subtrees[m]; ok {
				node = subtree
				continue
			}
			subtree := newTree()
			node.Subtrees[m] = subtree
			node = subtree
		}
		node.HasAction = true
		node.Action = k.actions
	}

	return result
}

func Create() *Keymap {
	var result Keymap
	result.activeMods = map[VKEY]bool{}
	c := client.Default()
	cfg, _ := c.Request.Config("")
	custom, _ := shell.GetCustom[keymap](cfg)

	hotkeys := make([]hotkey, 0, len(custom.Keymap))

	for _, mapping := range custom.Keymap {
		keys := mapping.Keys
		parts := strings.Split(keys, "+")
		sanitizedParts := make([]string, 0, len(parts))
		for _, p := range parts {
			s := strings.ToLower(strings.TrimSpace(p))
			if s != "" {
				sanitizedParts = append(sanitizedParts, s)
			}
		}
		hotkey := ParseMod(sanitizedParts)
		hotkey.actions = mapping.Actions
		hotkeys = append(hotkeys, hotkey)
	}
	result.Bindings = buildTree(hotkeys)
	return &result
}

var (
	vk_map = map[string]VKEY{
		"backspace": VK_BACK,
		"tab":       VK_TAB,
		"return":    VK_RETURN,
		"pause":     VK_PAUSE,
		"enter":     VK_RETURN,
		"escape":    VK_ESCAPE,
		"space":     VK_SPACE,

		"pgdn":       VK_PRIOR,
		"pagedown":   VK_PRIOR,
		"pageup":     VK_NEXT,
		"pgup":       VK_NEXT,
		"end":        VK_END,
		"home":       VK_HOME,
		"numlock":    VK_NUMLOCK,
		"scrolllock": VK_SCROLL,

		"left":  VK_LEFT,
		"up":    VK_UP,
		"right": VK_RIGHT,
		"down":  VK_DOWN,

		"print":       VK_SNAPSHOT,
		"printscreen": VK_SNAPSHOT,
		"insert":      VK_INSERT,
		"del":         VK_DELETE,
		"delete":      VK_DELETE,

		"shift":    VK_LSHIFT,
		"lshift":   VK_LSHIFT,
		"rshift":   VK_RSHIFT,
		"lctrl":    VK_LCONTROL,
		"ctrl":     VK_LCONTROL,
		"control":  VK_LCONTROL,
		"lcontrol": VK_LCONTROL,
		"rctrl":    VK_RCONTROL,
		"rcontrol": VK_RCONTROL,
		"alt":      VK_LMENU,
		"menu":     VK_LMENU,
		"win":      VK_LWIN,
		"lwin":     VK_LWIN,
		"rwin":     VK_RWIN,

		"num0": VK_NUMPAD0,
		"num1": VK_NUMPAD1,
		"num2": VK_NUMPAD2,
		"num3": VK_NUMPAD3,
		"num4": VK_NUMPAD4,
		"num5": VK_NUMPAD5,
		"num6": VK_NUMPAD6,
		"num7": VK_NUMPAD7,
		"num8": VK_NUMPAD8,
		"num9": VK_NUMPAD9,

		"f1":  VK_F1,
		"f2":  VK_F2,
		"f3":  VK_F3,
		"f4":  VK_F4,
		"f5":  VK_F5,
		"f6":  VK_F6,
		"f7":  VK_F7,
		"f8":  VK_F8,
		"f9":  VK_F9,
		"f10": VK_F10,
		"f11": VK_F11,
		"f12": VK_F12,
	}
)

const (
	VK_BACK   VKEY = 0x08 //backspace
	VK_TAB         = 0x09
	VK_RETURN      = 0x0D
	VK_PAUSE       = 0x13
	VK_ESCAPE      = 0x1B
	VK_SPACE       = 0x20

	VK_PRIOR   VKEY = 0x21 //pageup
	VK_NEXT         = 0x22 //pagedown
	VK_END          = 0x23
	VK_HOME         = 0x24
	VK_NUMLOCK      = 0x90
	VK_SCROLL       = 0x91 // scroll lock

	VK_LEFT  VKEY = 0x25
	VK_UP         = 0x26
	VK_RIGHT      = 0x27
	VK_DOWN       = 0x28

	VK_SNAPSHOT VKEY = 0x2C // print screen
	VK_INSERT        = 0x2D
	VK_DELETE        = 0x2E

	VK_SHIFT   VKEY = 0x10
	VK_CONTROL      = 0x11
	VK_MENU         = 0x12 // alt
	VK_LWIN         = 0x5B
	VK_RWIN         = 0x5C

	/* Numpad keys */
	VK_NUMPAD0   VKEY = 0x60
	VK_NUMPAD1        = 0x61
	VK_NUMPAD2        = 0x62
	VK_NUMPAD3        = 0x63
	VK_NUMPAD4        = 0x64
	VK_NUMPAD5        = 0x65
	VK_NUMPAD6        = 0x66
	VK_NUMPAD7        = 0x67
	VK_NUMPAD8        = 0x68
	VK_NUMPAD9        = 0x69
	VK_MULTIPLY       = 0x6a
	VK_ADD            = 0x6b
	VK_SEPARATOR      = 0x6c
	VK_SUBTRACT       = 0x6d
	VK_DECIMAL        = 0x6e
	VK_DIVIDE         = 0x6f

	VK_F1  VKEY = 0x70
	VK_F2       = 0x71
	VK_F3       = 0x72
	VK_F4       = 0x73
	VK_F5       = 0x74
	VK_F6       = 0x75
	VK_F7       = 0x76
	VK_F8       = 0x77
	VK_F9       = 0x78
	VK_F10      = 0x79
	VK_F11      = 0x7a
	VK_F12      = 0x7b

	VK_LSHIFT   = 0xA0
	VK_RSHIFT   = 0xA1
	VK_LCONTROL = 0xA2
	VK_RCONTROL = 0xA3
	VK_LMENU    = 0xA4
	VK_RMENU    = 0xA5
)
