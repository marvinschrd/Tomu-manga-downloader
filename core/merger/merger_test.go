package merger_test

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"mangatool/core/merger"
)

func makeCBZ(t *testing.T, dir, name string, pages int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, _ := os.Create(path)
	defer f.Close()
	w := zip.NewWriter(f)
	for i := 0; i < pages; i++ {
		e, _ := w.Create(fmt.Sprintf("%04d.jpg", i+1))
		e.Write([]byte("img"))
	}
	w.Close()
	return path
}

func TestMergeCBZs(t *testing.T) {
	dir := t.TempDir()
	cbz1 := makeCBZ(t, dir, "ch1.cbz", 3)
	cbz2 := makeCBZ(t, dir, "ch2.cbz", 2)

	outPath := filepath.Join(dir, "vol1.cbz")
	comicInfo := []byte("<ComicInfo></ComicInfo>")

	if err := merger.MergeCBZs([]string{cbz1, cbz2}, outPath, comicInfo); err != nil {
		t.Fatalf("MergeCBZs: %v", err)
	}

	r, err := zip.OpenReader(outPath)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()

	// 3 + 2 images + ComicInfo.xml = 6 files
	if len(r.File) != 6 {
		t.Errorf("expected 6 files in merged CBZ, got %d", len(r.File))
	}
}

func TestGroupByOfficialVolume(t *testing.T) {
	chapters := []merger.ChapterEntry{
		{Number: 1, Volume: "1", CBZPath: "ch1.cbz"},
		{Number: 2, Volume: "1", CBZPath: "ch2.cbz"},
		{Number: 3, Volume: "2", CBZPath: "ch3.cbz"},
		{Number: 4, Volume: "", CBZPath: "ch4.cbz"}, // no volume
	}

	groups := merger.GroupByVolume(chapters)
	if len(groups) != 2 {
		t.Errorf("expected 2 volume groups, got %d", len(groups))
	}
	if len(groups["1"]) != 2 {
		t.Errorf("expected 2 chapters in volume 1, got %d", len(groups["1"]))
	}
}
