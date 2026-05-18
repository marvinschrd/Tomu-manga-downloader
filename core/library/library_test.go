package library_test

import (
	"fmt"
	"testing"
	"time"

	"mangatool/core"
	"mangatool/core/library"
)

func newTestDB(t *testing.T) *library.Library {
	t.Helper()
	lib, err := library.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { lib.Close() })
	return lib
}

func TestAddAndGetManga(t *testing.T) {
	lib := newTestDB(t)

	manga := core.Manga{
		ID:     "manga-1",
		Source: "mangadex",
		Title:  "One Piece",
	}
	if err := lib.AddManga(manga, "/manga/one-piece"); err != nil {
		t.Fatalf("AddManga: %v", err)
	}

	got, err := lib.GetManga("manga-1")
	if err != nil {
		t.Fatalf("GetManga: %v", err)
	}
	if got.Title != "One Piece" {
		t.Errorf("expected title 'One Piece', got %q", got.Title)
	}
	if got.OutputDir != "/manga/one-piece" {
		t.Errorf("expected output dir '/manga/one-piece', got %q", got.OutputDir)
	}
}

func TestListManga(t *testing.T) {
	lib := newTestDB(t)

	for _, m := range []core.Manga{
		{ID: "a", Source: "mangadex", Title: "A"},
		{ID: "b", Source: "mangadex", Title: "B"},
	} {
		if err := lib.AddManga(m, "/manga"); err != nil {
			t.Fatalf("AddManga: %v", err)
		}
	}

	list, err := lib.ListManga()
	if err != nil {
		t.Fatalf("ListManga: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 manga, got %d", len(list))
	}
}

func TestAddDuplicateMangaIsNoop(t *testing.T) {
	lib := newTestDB(t)
	manga := core.Manga{ID: "a", Source: "mangadex", Title: "A"}

	if err := lib.AddManga(manga, "/manga"); err != nil {
		t.Fatalf("first AddManga: %v", err)
	}
	if err := lib.AddManga(manga, "/manga"); err != nil {
		t.Fatalf("second AddManga should not error: %v", err)
	}

	list, _ := lib.ListManga()
	if len(list) != 1 {
		t.Errorf("expected 1 manga, got %d", len(list))
	}
}

func TestAddAndListChapters(t *testing.T) {
	lib := newTestDB(t)
	manga := core.Manga{ID: "manga-1", Source: "mangadex", Title: "Test"}
	_ = lib.AddManga(manga, "/manga")

	ch := core.DownloadedChapter{
		Chapter: core.Chapter{
			ID:       "ch-1",
			MangaID:  "manga-1",
			Number:   1,
			Volume:   "1",
			Language: "en",
		},
		CBZPath:      "/manga/test-ch1.cbz",
		DownloadedAt: time.Now(),
	}
	if err := lib.AddChapter(ch); err != nil {
		t.Fatalf("AddChapter: %v", err)
	}

	chapters, err := lib.GetChapters("manga-1")
	if err != nil {
		t.Fatalf("GetChapters: %v", err)
	}
	if len(chapters) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(chapters))
	}
	if chapters[0].Number != 1 {
		t.Errorf("expected chapter number 1, got %v", chapters[0].Number)
	}
}

func TestMaxChapterNumber(t *testing.T) {
	lib := newTestDB(t)
	manga := core.Manga{ID: "manga-1", Source: "mangadex", Title: "Test"}
	_ = lib.AddManga(manga, "/manga")

	for _, n := range []float64{1, 2, 3.5} {
		_ = lib.AddChapter(core.DownloadedChapter{
			Chapter:      core.Chapter{ID: fmt.Sprintf("ch-%v", n), MangaID: "manga-1", Number: n, Language: "en"},
			CBZPath:      "/manga/ch.cbz",
			DownloadedAt: time.Now(),
		})
	}

	max, err := lib.MaxChapterNumber("manga-1")
	if err != nil {
		t.Fatalf("MaxChapterNumber: %v", err)
	}
	if max != 3.5 {
		t.Errorf("expected max 3.5, got %v", max)
	}
}
