package metadata_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"mangatool/core"
	"mangatool/core/metadata"
)

func TestGenerateChapterComicInfo(t *testing.T) {
	ch := core.Chapter{
		Number:   5,
		Title:    "The Beginning",
		Volume:   "1",
		Language: "en",
	}
	manga := core.Manga{
		Title:  "Test Manga",
		Source: "mangadex",
	}

	data, err := metadata.GenerateComicInfo(manga, ch, 20)
	if err != nil {
		t.Fatalf("GenerateComicInfo: %v", err)
	}

	var ci metadata.ComicInfo
	if err := xml.Unmarshal(data, &ci); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if ci.Series != "Test Manga" {
		t.Errorf("expected series 'Test Manga', got %q", ci.Series)
	}
	if ci.Number != "5" {
		t.Errorf("expected number '5', got %q", ci.Number)
	}
	if ci.Volume != 1 {
		t.Errorf("expected volume 1, got %d", ci.Volume)
	}
	if ci.PageCount != 20 {
		t.Errorf("expected page count 20, got %d", ci.PageCount)
	}
	if !strings.HasPrefix(string(data), "<?xml") {
		t.Error("output should start with XML declaration")
	}
}

func TestGenerateMergedComicInfo(t *testing.T) {
	manga := core.Manga{Title: "Test Manga", Source: "mangadex"}
	data, err := metadata.GenerateMergedComicInfo(manga, "Vol.1", 1, 7, 180)
	if err != nil {
		t.Fatalf("GenerateMergedComicInfo: %v", err)
	}
	var ci metadata.ComicInfo
	if err := xml.Unmarshal(data, &ci); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if ci.PageCount != 180 {
		t.Errorf("expected 180 pages, got %d", ci.PageCount)
	}
}
