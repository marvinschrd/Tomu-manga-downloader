package engine

import (
	"context"
	"sync"

	"mangatool/core"
	"mangatool/core/library"
)

type EventKind int

const (
	EvtChapterStarted  EventKind = iota // chapter download beginning
	EvtImageProgress                    // one call per image downloaded
	EvtChapterSaved                     // CBZ written successfully
	EvtChapterSkipped                   // already in library
	EvtChapterExternal                  // no images — hosted on official publisher site
	EvtChapterFailed                    // non-fatal error
	EvtMangaDone                        // all chapters processed
)

type Event struct {
	Kind          EventKind
	MangaTitle    string
	Chapter       core.Chapter
	ChapterIndex  int // 1-based; set on all chapter-level events
	TotalChapters int // set on all chapter-level events
	TotalImages   int // EvtChapterStarted only
	CBZPath       string // EvtChapterSaved only
	Downloaded    int    // EvtMangaDone only
	Skipped       int    // EvtMangaDone only
	Err           error  // EvtChapterFailed / EvtMangaDone (fetch error) only
}

type JobOptions struct {
	OutputDir string
	Language  string
	Force     bool
	// Pre-filtered chapter list. If nil, runJob fetches all chapters newer than
	// the library's max chapter number for this manga (update mode).
	Chapters []core.Chapter
}

type Job struct {
	Manga   core.Manga
	Options JobOptions
}

type Engine struct {
	Source core.Source
	Lib    *library.Library
	Conc   int // max concurrent manga jobs; defaults to 1
}

// Run submits all jobs to the worker pool and blocks until all finish.
// onEvent may be called from multiple goroutines simultaneously — callers must
// synchronise any shared state (e.g. a sync.Mutex around fmt.Printf).
func (e *Engine) Run(ctx context.Context, jobs []Job, onEvent func(Event)) {
	conc := e.Conc
	if conc <= 0 {
		conc = 1
	}
	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			runJob(ctx, e.Source, e.Lib, job, onEvent)
		}(j)
	}
	wg.Wait()
}
