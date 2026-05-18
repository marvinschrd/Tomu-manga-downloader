package mangadex_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mangatool/core"
	"mangatool/core/sources/mangadex"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manga":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "ok",
				"data": []map[string]interface{}{
					{
						"id": "manga-123",
						"attributes": map[string]interface{}{
							"title":  map[string]string{"en": "Test Manga"},
							"status": "ongoing",
						},
						"relationships": []interface{}{},
					},
				},
				"total": 1,
			})
		case "/manga/manga-123":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "ok",
				"data": map[string]interface{}{
					"id": "manga-123",
					"attributes": map[string]interface{}{
						"title":       map[string]string{"en": "Test Manga"},
						"description": map[string]string{"en": "A test manga"},
						"status":      "ongoing",
					},
					"relationships": []interface{}{},
				},
			})
		case "/manga/manga-123/feed":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "ok",
				"data": []map[string]interface{}{
					{
						"id": "ch-1",
						"attributes": map[string]interface{}{
							"chapter":            "1",
							"volume":             "1",
							"title":              "First Chapter",
							"translatedLanguage": "en",
							"pages":              20,
						},
					},
					{
						"id": "ch-2",
						"attributes": map[string]interface{}{
							"chapter":            "2",
							"volume":             "1",
							"title":              "Second Chapter",
							"translatedLanguage": "en",
							"pages":              18,
						},
					},
				},
				"total":  2,
				"limit":  500,
				"offset": 0,
			})
		}
	}))
}

func TestSearch(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	source := mangadex.NewWithBaseURL(srv.URL)
	results, err := source.Search("test", core.SearchOptions{Language: "en", Limit: 20})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Test Manga" {
		t.Errorf("expected 'Test Manga', got %q", results[0].Title)
	}
}

func TestGetManga(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	source := mangadex.NewWithBaseURL(srv.URL)
	manga, err := source.GetManga("manga-123")
	if err != nil {
		t.Fatalf("GetManga: %v", err)
	}
	if manga.Description != "A test manga" {
		t.Errorf("expected description 'A test manga', got %q", manga.Description)
	}
}

func TestGetChapters(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	source := mangadex.NewWithBaseURL(srv.URL)
	chapters, err := source.GetChapters("manga-123", core.ChapterOptions{Language: "en"})
	if err != nil {
		t.Fatalf("GetChapters: %v", err)
	}
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if chapters[0].Number != 1 {
		t.Errorf("expected chapter number 1, got %v", chapters[0].Number)
	}
}

func TestGetLatestChapters(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	source := mangadex.NewWithBaseURL(srv.URL)
	chapters, err := source.GetLatestChapters("manga-123", 1.0)
	if err != nil {
		t.Fatalf("GetLatestChapters: %v", err)
	}
	if len(chapters) != 1 {
		t.Fatalf("expected 1 new chapter, got %d", len(chapters))
	}
	if chapters[0].Number != 2 {
		t.Errorf("expected chapter 2, got %v", chapters[0].Number)
	}
}
