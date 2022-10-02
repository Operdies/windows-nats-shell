package files

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	knownFileExtensions = map[string]bool{}
)

type WatchedDir struct {
	files     map[string]string
	root      string
	recursive bool
	isIndexed bool
	indexLock sync.Mutex
	watcher   *fsnotify.Watcher
}

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
				w.files[key] = evt.Name
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

var (
	fileExtensionFilter = map[string]bool{
		"exe":  true,
		"lnk":  true,
		"bat":  true,
		"ps1":  true,
		"url":  true,
		"html": true,
		"png":  true,
		"jpg":  true,
		"gif":  true,
	}
)

func extLowerNoDot(s string) string {
	s = filepath.Ext(s)
	if len(s) > 0 {
		return strings.ToLower(s)[1:]
	}
	return s
}

func includeFile(f fs.DirEntry) bool {
	if f.IsDir() {
		return false
	}
	v, ok := fileExtensionFilter[extLowerNoDot(f.Name())]
	return ok && v
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
			if pathEqual(w.root, path) {
				return nil
			}
			if d.IsDir() && w.recursive == false {
				return filepath.SkipDir
			}
			if includeFile(d) {
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
	if recursive {
	} else {
		watcher.Add(root)
	}
	w.watcher = watcher
	w.indexFiles()
	if watch {
		go w.watchEvents()
	}
	return &w
}
