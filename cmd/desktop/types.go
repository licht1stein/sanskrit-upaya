package main

import (
	"sync"
	"time"

	"github.com/licht1stein/sanskrit-upaya/pkg/search"
)

// DictEntry represents articles from one dictionary
type DictEntry struct {
	DictCode string
	DictName string
	Articles []search.Result
}

// GroupedResult represents all articles for a word across all dictionaries
type GroupedResult struct {
	Word    string
	Entries []DictEntry // All dictionaries that have this word
}

// debouncer for search-as-you-type
type debouncer struct {
	mu       sync.Mutex
	timer    *time.Timer
	duration time.Duration
}

func newDebouncer(d time.Duration) *debouncer {
	return &debouncer{duration: d}
}

func (d *debouncer) Do(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.duration, f)
}
