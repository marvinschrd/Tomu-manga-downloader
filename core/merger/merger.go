package merger

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type ChapterEntry struct {
	Number  float64
	Volume  string
	CBZPath string
}

// MergeCBZs combines multiple CBZ files into one, renaming pages to maintain order.
func MergeCBZs(cbzPaths []string, outPath string, comicInfo []byte) (retErr error) {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		if retErr != nil {
			os.Remove(outPath)
		}
	}()

	w := zip.NewWriter(f)
	defer w.Close()

	pageIdx := 1
	for _, cbzPath := range cbzPaths {
		count, err := appendCBZPages(w, cbzPath, pageIdx)
		if err != nil {
			return fmt.Errorf("merging %s: %w", filepath.Base(cbzPath), err)
		}
		pageIdx += count
	}

	if comicInfo != nil {
		ci, err := w.Create("ComicInfo.xml")
		if err != nil {
			return err
		}
		if _, err := ci.Write(comicInfo); err != nil {
			return err
		}
	}

	return nil
}

func appendCBZPages(w *zip.Writer, cbzPath string, startIdx int) (int, error) {
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		if f.Name == "ComicInfo.xml" {
			continue
		}
		ext := filepath.Ext(f.Name)
		name := fmt.Sprintf("%05d%s", startIdx+count, ext)

		dst, err := w.Create(name)
		if err != nil {
			return count, err
		}
		src, err := f.Open()
		if err != nil {
			return count, err
		}
		_, copyErr := io.Copy(dst, src)
		src.Close()
		if copyErr != nil {
			return count, copyErr
		}
		count++
	}
	return count, nil
}

// GroupByVolume groups ChapterEntry slices by their Volume label.
// Chapters with an empty Volume are excluded (no official grouping).
func GroupByVolume(chapters []ChapterEntry) map[string][]ChapterEntry {
	groups := map[string][]ChapterEntry{}
	for _, ch := range chapters {
		if ch.Volume == "" {
			continue
		}
		groups[ch.Volume] = append(groups[ch.Volume], ch)
	}
	for k := range groups {
		sort.Slice(groups[k], func(i, j int) bool {
			return groups[k][i].Number < groups[k][j].Number
		})
	}
	return groups
}
