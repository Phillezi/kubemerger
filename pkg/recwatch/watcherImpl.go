package recwatch

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	FilesCh <-chan []string
	Errors  <-chan error
	Close   func() error
}

type watcherImpl struct {
	watcher *fsnotify.Watcher
	exclude []string

	mu    sync.Mutex
	dirs  map[string]struct{}
	files map[string]struct{}

	filesCh  chan []string
	errorsCh chan error
	done     chan struct{}

	debounceMu    sync.RWMutex
	debounce      *time.Timer
	debounceDelay time.Duration
}

func New(root string, exclude []string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &watcherImpl{
		watcher:       fsw,
		exclude:       exclude,
		dirs:          make(map[string]struct{}),
		files:         make(map[string]struct{}),
		filesCh:       make(chan []string, 1),
		errorsCh:      make(chan error, 1),
		done:          make(chan struct{}),
		debounceDelay: 300 * time.Millisecond,
	}

	if err := w.addDirRecursive(root); err != nil {
		return nil, err
	}

	go w.loop()

	return &Watcher{
		FilesCh: w.filesCh,
		Errors:  w.errorsCh,
		Close:   w.close,
	}, nil
}

func (w *watcherImpl) handleEvent(ev fsnotify.Event) {
	path := ev.Name

	if w.isExcluded(path) {
		return
	}

	info, err := os.Stat(path)
	exists := err == nil

	switch {
	case ev.Op&(fsnotify.Create|fsnotify.Rename) != 0:
		if exists && info.IsDir() {
			_ = w.addDirRecursive(path)
		} else if exists {
			w.files[path] = struct{}{}
		}

	case ev.Op&(fsnotify.Remove) != 0:
		w.removeDir(path)
		delete(w.files, path)

	case ev.Op&(fsnotify.Write|fsnotify.Chmod) != 0:
		if exists && !info.IsDir() {
			w.files[path] = struct{}{}
		}
	}

	w.debounceMu.Lock()
	defer w.debounceMu.Unlock()
	if w.debounce != nil {
		w.debounce.Stop()
	}
	w.debounce = time.AfterFunc(w.debounceDelay, func() {
		w.emitFiles()
	})
}

func (w *watcherImpl) addDirRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if w.isExcluded(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return w.addDir(path)
		}
		w.files[path] = struct{}{}
		return nil
	})
}

func (w *watcherImpl) addDir(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.dirs[path]; ok {
		return nil
	}

	if err := w.watcher.Add(path); err != nil {
		return err
	}

	w.dirs[path] = struct{}{}
	return nil
}

func (w *watcherImpl) removeDir(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.dirs[path]; !ok {
		return
	}

	_ = w.watcher.Remove(path)
	delete(w.dirs, path)

	for f := range w.files {
		if isSubPath(path, f) {
			delete(w.files, f)
		}
	}
}

func (w *watcherImpl) loop() {
	for {
		select {
		case ev := <-w.watcher.Events:
			w.handleEvent(ev)
		case err := <-w.watcher.Errors:
			w.errorsCh <- err
		case <-w.done:
			return
		}
	}
}

func (w *watcherImpl) emitFiles() {
	w.mu.Lock()
	defer w.mu.Unlock()

	files := make([]string, 0, len(w.files))
	for f := range w.files {
		if w.isExcluded(f) {
			continue
		}
		files = append(files, f)
	}

	select {
	case w.filesCh <- files:
	default:
	}
}

func (w *watcherImpl) isExcluded(path string) bool {
	for _, ex := range w.exclude {
		if ex == path || isSubPath(ex, path) {
			return true
		}
	}
	return false
}

func isSubPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..")
}

func (w *watcherImpl) close() error {
	close(w.done)
	close(w.filesCh)
	close(w.errorsCh)
	return w.watcher.Close()
}
