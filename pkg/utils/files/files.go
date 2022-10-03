package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/fsnotify/fsnotify"
	"github.com/operdies/windows-nats-shell/pkg/utils/query"
	"github.com/operdies/windows-nats-shell/pkg/winapi"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

type WatchedDir struct {
	files     map[string]string
	root      string
	recursive bool
	isIndexed bool
	indexLock sync.Mutex
	watcher   *fsnotify.Watcher
}

var _extensions []string

func SetExtentions(s []string) { _extensions = s }

func (w *WatchedDir) Files() map[string]string {
	return w.files
}

func (w *WatchedDir) watchEvents() {
	for evt := range w.watcher.Events {
		if evt.Op == fsnotify.Create || evt.Op == fsnotify.Remove {
			w.indexLock.Lock()
			key := getKey(evt.Name)
			_, exists := w.files[key]
			// Overwrite if it exists
			if evt.Op == fsnotify.Create {
				if includeFile(evt.Name) {
					w.files[key] = evt.Name
				}
			} else if evt.Op == fsnotify.Remove && exists {
				delete(w.files, key)
			}
			w.indexLock.Unlock()
		}
	}
}

// Waits for indexing to complete if an indexing operation is running
// Otherwise return immediately
func (w *WatchedDir) WaitForIndex() {
	w.indexLock.Lock()
	defer w.indexLock.Unlock()
}

func extLower(s string) string {
	s = filepath.Ext(s)
	if len(s) > 0 {
		return strings.ToLower(s)
	}
	return s
}

func includeFile(name string) bool {
	// v, ok := fileExtensionFilter[extLower(f.Name())]
	ext := extLower(name)
	return query.Contains(_extensions, ext)
}

func calcExtHasAssoc(ext string) bool {
	bytes := []byte(ext)
	ptr := unsafe.Pointer(&bytes[0])
	var size wintypes.DWORD = 200
	sizePtr := unsafe.Pointer(&size)
	res := make([]byte, size)
	resultPtr := unsafe.Pointer(&res[0])
	hResult := winapi.AssocQueryString(wintypes.ASSOCF_NONE, wintypes.ASSOCSTR_FRIENDLYDOCNAME, wintypes.LPCSTR(ptr), 0, wintypes.LPSTR(resultPtr), uintptr(sizePtr))
	v := wintypes.SUCCEEDED(hResult)
	// resStr := res[:size]
	if v {
		// fmt.Printf("Extension %s supported: %v (%v) (%v)\n", ext, v, string(resStr), hResult)
	}
	return v
}

func pathEqual(s1, s2 string) bool {
	normalize := func(s string) string {
		s = strings.ToLower(strings.ReplaceAll(s, "\\", "/"))
		return strings.TrimRight(s, "/")
	}

	s1 = normalize(s1)
	s2 = normalize(s2)

	return s1 == s2
}

func (w *WatchedDir) indexFiles() {
	w.indexLock.Lock()
	go func() {
		defer w.indexLock.Unlock()
		items := map[string]string{}
		filepath.WalkDir(w.root, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				if w.recursive {
					return nil
				} else {
					if pathEqual(w.root, path) {
						return nil
					}
					return filepath.SkipDir
				}
			}
			if includeFile(d.Name()) {
				items[getKey(d.Name())] = path
			}
			return nil
		})
		w.files = items
	}()
}

func getKey(fullname string) string {
	bn := filepath.Base(fullname)
	return bn
}

func Create(root string, recursive, watch bool) *WatchedDir {
	var w = WatchedDir{root: root, recursive: recursive}
	watcher, _ := fsnotify.NewWatcher()

	if watch {
		if recursive {
			filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if d.IsDir() {
					err := watcher.Add(path)
					if err != nil {
						fmt.Printf("Error adding watch: %v", err.Error())
					}
				}
				return nil
			})
		} else {
			watcher.Add(root)
		}
	}

	w.watcher = watcher
	w.indexFiles()
	if watch {
		go w.watchEvents()
	}
	return &w
}
