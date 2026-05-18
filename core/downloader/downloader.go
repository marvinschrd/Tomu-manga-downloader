package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// CreateCBZ zips the given image files (in order) and a ComicInfo.xml into a CBZ file.
// If creation fails partway through, the partial file is removed.
func CreateCBZ(outPath string, imagePaths []string, comicInfo []byte) (retErr error) {
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

	for _, imgPath := range imagePaths {
		if err := addFileToZip(w, imgPath, filepath.Base(imgPath)); err != nil {
			return fmt.Errorf("adding %s: %w", imgPath, err)
		}
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

func addFileToZip(w *zip.Writer, srcPath, name string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}

// DownloadImages downloads image URLs in parallel into destDir.
// Returns ordered file paths matching the order of urls.
func DownloadImages(urls []string, destDir string) ([]string, error) {
	type result struct {
		idx  int
		path string
		err  error
	}

	results := make([]result, len(urls))
	ch := make(chan result, len(urls))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // max 5 concurrent downloads

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			path, err := downloadImage(u, destDir, fmt.Sprintf("%04d%s", idx+1, imageExt(u)))
			ch <- result{idx: idx, path: path, err: err}
		}(i, url)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		if r.err != nil {
			return nil, fmt.Errorf("downloading image %d: %w", r.idx+1, r.err)
		}
		results[r.idx] = r
	}

	paths := make([]string, len(urls))
	for i, r := range results {
		paths[i] = r.path
	}
	return paths, nil
}

func downloadImage(url, destDir, name string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	path := filepath.Join(destDir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return path, err
}

func imageExt(url string) string {
	ext := filepath.Ext(url)
	if ext == "" {
		return ".jpg"
	}
	return ext
}
