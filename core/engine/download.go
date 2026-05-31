package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mangatool/core"
	"mangatool/core/downloader"
	"mangatool/core/library"
	"mangatool/core/metadata"
)

func runJob(ctx context.Context, src core.Source, lib *library.Library, job Job, onEvent func(Event)) {
	chapters := job.Options.Chapters
	if chapters == nil {
		max, _ := lib.MaxChapterNumber(job.Manga.ID)
		fetched, err := src.GetLatestChapters(job.Manga.ID, max)
		if err != nil {
			onEvent(Event{Kind: EvtMangaDone, MangaTitle: job.Manga.Title, Err: err})
			return
		}
		for _, ch := range fetched {
			if job.Options.Language == "" || ch.Language == job.Options.Language {
				chapters = append(chapters, ch)
			}
		}
	}

	if err := os.MkdirAll(job.Options.OutputDir, 0755); err != nil {
		onEvent(Event{Kind: EvtMangaDone, MangaTitle: job.Manga.Title, Err: err})
		return
	}

	total := len(chapters)
	var downloaded, skipped int

	for i, ch := range chapters {
		if ctx.Err() != nil {
			break
		}

		idx := i + 1

		if !job.Options.Force {
			if exists, _ := lib.HasChapter(ch.ID); exists {
				onEvent(Event{
					Kind: EvtChapterSkipped, MangaTitle: job.Manga.Title,
					Chapter: ch, ChapterIndex: idx, TotalChapters: total,
				})
				skipped++
				continue
			}
		}

		tmpDir, err := os.MkdirTemp("", "tomu-*")
		if err != nil {
			onEvent(Event{
				Kind: EvtChapterFailed, MangaTitle: job.Manga.Title,
				Chapter: ch, ChapterIndex: idx, TotalChapters: total, Err: err,
			})
			skipped++
			continue
		}

		imageURLs, err := src.DownloadChapter(ch, tmpDir)
		if err != nil {
			os.RemoveAll(tmpDir)
			onEvent(Event{
				Kind: EvtChapterFailed, MangaTitle: job.Manga.Title,
				Chapter: ch, ChapterIndex: idx, TotalChapters: total, Err: err,
			})
			skipped++
			continue
		}
		if len(imageURLs) == 0 {
			os.RemoveAll(tmpDir)
			onEvent(Event{
				Kind: EvtChapterExternal, MangaTitle: job.Manga.Title,
				Chapter: ch, ChapterIndex: idx, TotalChapters: total,
			})
			skipped++
			continue
		}

		onEvent(Event{
			Kind: EvtChapterStarted, MangaTitle: job.Manga.Title,
			Chapter: ch, ChapterIndex: idx, TotalChapters: total, TotalImages: len(imageURLs),
		})

		imagePaths, err := downloader.DownloadImages(imageURLs, tmpDir, func() {
			onEvent(Event{
				Kind: EvtImageProgress, MangaTitle: job.Manga.Title,
				Chapter: ch, ChapterIndex: idx, TotalChapters: total,
			})
		})
		if err != nil {
			os.RemoveAll(tmpDir)
			onEvent(Event{
				Kind: EvtChapterFailed, MangaTitle: job.Manga.Title,
				Chapter: ch, ChapterIndex: idx, TotalChapters: total, Err: err,
			})
			skipped++
			continue
		}

		ci, _ := metadata.GenerateComicInfo(job.Manga, ch, len(imagePaths))
		cbzPath := filepath.Join(job.Options.OutputDir, chapterFilename(job.Manga.Title, ch))
		if err := downloader.CreateCBZ(cbzPath, imagePaths, ci); err != nil {
			os.RemoveAll(tmpDir)
			onEvent(Event{
				Kind: EvtChapterFailed, MangaTitle: job.Manga.Title,
				Chapter: ch, ChapterIndex: idx, TotalChapters: total, Err: err,
			})
			skipped++
			continue
		}
		os.RemoveAll(tmpDir)

		_ = lib.AddChapter(core.DownloadedChapter{
			Chapter:      ch,
			CBZPath:      cbzPath,
			DownloadedAt: time.Now(),
		})

		onEvent(Event{
			Kind: EvtChapterSaved, MangaTitle: job.Manga.Title,
			Chapter: ch, ChapterIndex: idx, TotalChapters: total, CBZPath: cbzPath,
		})
		downloaded++
	}

	onEvent(Event{
		Kind:       EvtMangaDone,
		MangaTitle: job.Manga.Title,
		Downloaded: downloaded,
		Skipped:    skipped,
	})
}

func chapterFilename(title string, ch core.Chapter) string {
	if ch.Volume != "" {
		return fmt.Sprintf("%s - Vol.%s - Ch.%s.cbz", sanitize(title), ch.Volume, formatNum(ch.Number))
	}
	return fmt.Sprintf("%s - Ch.%s.cbz", sanitize(title), formatNum(ch.Number))
}

func formatNum(n float64) string {
	if n == float64(int(n)) {
		return fmt.Sprintf("%d", int(n))
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func sanitize(s string) string {
	r := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "", "?", "", `"`, "", "<", "", ">", "", "|", "")
	return r.Replace(s)
}
