package netclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	http    *http.Client
	retries int
	ua      string
}

func New(retries, timeoutSec int, ua string) *Client {
	return &Client{
		http:    &http.Client{Timeout: 0},
		retries: retries,
		ua:      ua,
	}
}

func (c *Client) Get(ctx context.Context, rawURL string) ([]byte, int, error) {
	var last error
	for i := 0; i < c.retries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("User-Agent", c.ua)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain,*/*;q=0.8")

		resp, err := c.http.Do(req)
		if err != nil {
			last = err
			time.Sleep(time.Duration(i+1) * 300 * time.Millisecond)
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
		resp.Body.Close()
		if err != nil {
			last = err
			continue
		}
		return body, resp.StatusCode, nil
	}
	return nil, 0, fmt.Errorf("retries exhausted: %w", last)
}

type ProgressFunc func(written, total int64)

func (c *Client) Download(ctx context.Context, rawURL, dest string, existing int64, progress ProgressFunc) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", c.ua)
	if existing > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existing))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("404 not found")
	}
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("http %d", resp.StatusCode)
	}

	total := resp.ContentLength
	if existing > 0 && resp.StatusCode == http.StatusPartialContent {
		total += existing
	} else if existing > 0 {
		existing = 0
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return 0, err
	}

	flag := os.O_CREATE | os.O_WRONLY
	if existing > 0 {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	f, err := os.OpenFile(dest, flag, 0o644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	pw := &progressWriter{total: total, written: existing, cb: progress}
	_, err = io.Copy(f, io.TeeReader(resp.Body, pw))
	return pw.written, err
}

type progressWriter struct {
	total   int64
	written int64
	cb      ProgressFunc
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.written += int64(n)
	if pw.cb != nil {
		pw.cb(pw.written, pw.total)
	}
	return n, nil
}
