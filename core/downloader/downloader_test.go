package downloader_test

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"mangatool/core/downloader"
)

func TestCreateCBZ(t *testing.T) {
	dir := t.TempDir()

	imgPaths := []string{}
	for i, name := range []string{"001.jpg", "002.jpg", "003.jpg"} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(fmt.Sprintf("img%d", i)), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		imgPaths = append(imgPaths, p)
	}

	comicInfo := []byte("<ComicInfo></ComicInfo>")
	outPath := filepath.Join(dir, "chapter.cbz")

	if err := downloader.CreateCBZ(outPath, imgPaths, comicInfo); err != nil {
		t.Fatalf("CreateCBZ: %v", err)
	}

	r, err := zip.OpenReader(outPath)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()

	names := map[string]bool{}
	for _, f := range r.File {
		names[f.Name] = true
	}

	for _, img := range []string{"001.jpg", "002.jpg", "003.jpg"} {
		if !names[img] {
			t.Errorf("missing file %q in CBZ", img)
		}
	}
	if !names["ComicInfo.xml"] {
		t.Error("missing ComicInfo.xml in CBZ")
	}
}

func TestCreateCBZPartialFailCleansUp(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "chapter.cbz")

	err := downloader.CreateCBZ(outPath, []string{"/nonexistent/image.jpg"}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
		t.Error("partial CBZ file should have been deleted on failure")
	}
}
