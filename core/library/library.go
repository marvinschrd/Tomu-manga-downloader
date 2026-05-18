package library

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mangatool/core"
	_ "modernc.org/sqlite"
)

type Library struct {
	db *sql.DB
}

func Open(path string) (*Library, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	lib := &Library{db: db}
	if err := lib.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return lib, nil
}

func (l *Library) Close() error {
	return l.db.Close()
}

func (l *Library) migrate() error {
	_, err := l.db.Exec(`
		CREATE TABLE IF NOT EXISTS manga (
			id         TEXT PRIMARY KEY,
			source     TEXT NOT NULL,
			title      TEXT NOT NULL,
			cover_url  TEXT,
			output_dir TEXT NOT NULL,
			added_at   DATETIME NOT NULL
		);
		CREATE TABLE IF NOT EXISTS chapters (
			id             TEXT PRIMARY KEY,
			manga_id       TEXT NOT NULL REFERENCES manga(id),
			number         REAL NOT NULL,
			title          TEXT,
			volume         TEXT,
			language       TEXT NOT NULL,
			cbz_path       TEXT NOT NULL,
			downloaded_at  DATETIME NOT NULL
		);
	`)
	return err
}

func (l *Library) AddManga(m core.Manga, outputDir string) error {
	_, err := l.db.Exec(`
		INSERT OR IGNORE INTO manga (id, source, title, cover_url, output_dir, added_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		m.ID, m.Source, m.Title, m.CoverURL, outputDir, time.Now(),
	)
	return err
}

func (l *Library) GetManga(id string) (core.DownloadedManga, error) {
	row := l.db.QueryRow(`SELECT id, source, title, cover_url, output_dir, added_at FROM manga WHERE id = ?`, id)
	var m core.DownloadedManga
	var addedAt time.Time
	err := row.Scan(&m.ID, &m.Source, &m.Title, &m.CoverURL, &m.OutputDir, &addedAt)
	if err == sql.ErrNoRows {
		return m, fmt.Errorf("manga %q not found in library", id)
	}
	m.AddedAt = addedAt
	return m, err
}

func (l *Library) ListManga() ([]core.DownloadedManga, error) {
	rows, err := l.db.Query(`SELECT id, source, title, cover_url, output_dir, added_at FROM manga ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []core.DownloadedManga
	for rows.Next() {
		var m core.DownloadedManga
		if err := rows.Scan(&m.ID, &m.Source, &m.Title, &m.CoverURL, &m.OutputDir, &m.AddedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (l *Library) AddChapter(ch core.DownloadedChapter) error {
	_, err := l.db.Exec(`
		INSERT OR IGNORE INTO chapters (id, manga_id, number, title, volume, language, cbz_path, downloaded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ch.ID, ch.MangaID, ch.Number, ch.Title, ch.Volume, ch.Language, ch.CBZPath, ch.DownloadedAt,
	)
	return err
}

func (l *Library) GetChapters(mangaID string) ([]core.DownloadedChapter, error) {
	rows, err := l.db.Query(`
		SELECT id, manga_id, number, title, volume, language, cbz_path, downloaded_at
		FROM chapters WHERE manga_id = ? ORDER BY number`, mangaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []core.DownloadedChapter
	for rows.Next() {
		var ch core.DownloadedChapter
		if err := rows.Scan(&ch.ID, &ch.MangaID, &ch.Number, &ch.Title, &ch.Volume, &ch.Language, &ch.CBZPath, &ch.DownloadedAt); err != nil {
			return nil, err
		}
		list = append(list, ch)
	}
	return list, rows.Err()
}

func (l *Library) MaxChapterNumber(mangaID string) (float64, error) {
	row := l.db.QueryRow(`SELECT COALESCE(MAX(number), 0) FROM chapters WHERE manga_id = ?`, mangaID)
	var max float64
	return max, row.Scan(&max)
}

func (l *Library) HasChapter(chapterID string) (bool, error) {
	row := l.db.QueryRow(`SELECT COUNT(1) FROM chapters WHERE id = ?`, chapterID)
	var count int
	return count > 0, row.Scan(&count)
}
