package keymap

import (
	"fmt"
	"sort"
	"strings"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/pkg/input"
	"github.com/operdies/windows-nats-shell/pkg/input/keyboard"
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
type config struct {
	Keymap []struct {
		Keys    string
		Actions []action
	}
}

type mod string

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
	vkeyModMap = map[input.VKEY]mod{
		input.VK_MENU:    alt,
		input.VK_CONTROL: ctrl,
		input.VK_LWIN:    win,
		input.VK_SHIFT:   shift,
	}
)

type BindingTree struct {
	HasAction bool
	Action    []action
	Subtrees  map[uint32]*BindingTree
}

type Keymap struct {
	Bindings   *BindingTree
	activeMods map[input.VKEY]bool
}

func parseKey(key string) input.VKEY {
	key = strings.ToLower(key)
	bs := []byte(key)

	if code, ok := input.VK_MAP[key]; ok {
		return code
	}
	if len(bs) > 1 || len(bs) < 1 {
		err := fmt.Errorf("Key not supported: %s", key)
		panic(err)
	}

	a, _, _ := VkKeyScanA.Call(uintptr(bs[0]))
	return input.VKEY(a)
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

func getBinding(k *Keymap, vkey input.VKEY) *BindingTree {
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

func handleKey(k *Keymap, kei keyboard.KeyboardEventInfo, bmap *BindingTree) {
	// We can't differentiate presses and holds from the event info.
	// But we know it's a hold if the key is already mapped in activeMods
	keyDown := kei.TransitionState == false

	// Update the modifier state if the key is a modifier
	vkey := input.VKEY(kei.VirtualKeyCode)
	// Clear the keymap when escape is pressed
	if vkey == input.VK_ESCAPE {
		k.activeMods = map[input.VKEY]bool{}
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
func (k *Keymap) ProcessEvent(kei keyboard.KeyboardEventInfo) bool {
	// This is a strange bug in consoles when a key is unmapped in the registry
	// I know this is a strange place to patch it but whatever
	if kei.ScanCode == 0 && kei.VirtualKeyCode == 255 {
		return true
	}
	// Return immediately after determining whether this key is handled.
	binding := getBinding(k, input.VKEY(kei.VirtualKeyCode))
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
	result.activeMods = map[input.VKEY]bool{}
	c := client.Default()
	cfg := client.GetConfig[config](c.Request)
	c.Close()

	hotkeys := make([]hotkey, 0, len(cfg.Keymap))

	for _, mapping := range cfg.Keymap {
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
