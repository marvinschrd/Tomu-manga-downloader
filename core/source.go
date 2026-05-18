package core

import "time"

type Manga struct {
	ID          string
	Source      string
	Title       string
	Description string
	Status      string // ongoing, completed, hiatus, cancelled
	CoverURL    string
}

type Chapter struct {
	ID        string
	MangaID   string
	Source    string
	Number    float64 // supports 10.5 etc.
	Title     string
	Volume    string // official volume label, empty if unknown
	Language  string
	PageCount int
}

type SearchOptions struct {
	Language string
	Limit    int
}

type ChapterOptions struct {
	Language string
}

type Source interface {
	Search(query string, opts SearchOptions) ([]Manga, error)
	GetManga(id string) (Manga, error)
	GetChapters(mangaID string, opts ChapterOptions) ([]Chapter, error)
	DownloadChapter(chapter Chapter, destDir string) ([]string, error) // returns ordered image file paths
	GetLatestChapters(mangaID string, afterChapterNumber float64) ([]Chapter, error)
}

// DownloadedChapter represents a chapter stored in the local library.
type DownloadedChapter struct {
	Chapter
	CBZPath      string
	DownloadedAt time.Time
}

// DownloadedManga represents a manga tracked in the local library.
type DownloadedManga struct {
	Manga
	OutputDir string
	AddedAt   time.Time
	Chapters  []DownloadedChapter
}
