package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
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
// onProgress is called after each successful image download; pass nil for no-op.
func DownloadImages(urls []string, destDir string, onProgress func()) ([]string, error) {
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
		if onProgress != nil {
			onProgress()
		}
	}

	paths := make([]string, len(urls))
	for i, r := range results {
		paths[i] = r.path
	}
	return paths, nil
}

const maxImageRetries = 4

func downloadImage(rawURL, destDir, name string) (string, error) {
	delay := time.Second
	for attempt := 0; attempt < maxImageRetries; attempt++ {
		resp, err := http.Get(rawURL)
		if err != nil {
			return "", err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := imageRetryAfter(resp.Header.Get("Retry-After"), delay)
			resp.Body.Close()
			if attempt == maxImageRetries-1 {
				return "", fmt.Errorf("rate limited downloading image — giving up after %d retries", maxImageRetries)
			}
			time.Sleep(wait)
			delay *= 2
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, rawURL)
		}

		path := filepath.Join(destDir, name)
		f, err := os.Create(path)
		if err != nil {
			resp.Body.Close()
			return "", err
		}
		_, copyErr := io.Copy(f, resp.Body)
		f.Close()
		resp.Body.Close()
		return path, copyErr
	}
	return "", fmt.Errorf("rate limited downloading image — giving up")
}

func imageRetryAfter(header string, fallback time.Duration) time.Duration {
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return fallback
}

func imageExt(url string) string {
	ext := filepath.Ext(url)
	if ext == "" {
		return ".jpg"
	}
	return ext
}
