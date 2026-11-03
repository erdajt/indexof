package listing

import (
	"bytes"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Entry struct {
	Name      string
	URL       string
	IsDir     bool
	Size      string
	SizeBytes int64
}

var sizeRe = regexp.MustCompile(`(?i)^\s*([\d.]+)\s*([KMGT]?I?B?)\s*$`)

func Parse(base *url.URL, body []byte, showHidden bool) ([]Entry, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var entries []Entry
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := attr(n, "href")
			if href == "" || href == "../" || href == "/" || strings.HasPrefix(href, "?") {
				return
			}

			name := strings.TrimSpace(text(n))
			if name == "" {
				name = path.Base(href)
			}
			if !showHidden && strings.HasPrefix(name, ".") {
				return
			}

			u, err := base.Parse(href)
			if err != nil {
				return
			}

			isDir := strings.HasSuffix(href, "/") || strings.HasSuffix(name, "/")
			name = strings.TrimSuffix(name, "/")

			sizeStr, sizeBytes := extractSize(n)

			entries = append(entries, Entry{
				Name:      name,
				URL:       u.String(),
				IsDir:     isDir,
				Size:      sizeStr,
				SizeBytes: sizeBytes,
			})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return dedup(entries), nil
}

func extractSize(n *html.Node) (string, int64) {
	for sib := n.NextSibling; sib != nil; sib = sib.NextSibling {
		if sib.Type == html.TextNode {
			raw := strings.TrimSpace(sib.Data)
			if raw == "" {
				continue
			}
			return parseSize(raw)
		}
	}
	return "", 0
}

func parseSize(s string) (string, int64) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return "", 0
	}
	m := sizeRe.FindStringSubmatch(s)
	if m == nil {
		return s, 0
	}
	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return s, 0
	}
	mult := float64(1)
	switch unit := strings.ToUpper(m[2]); {
	case strings.HasPrefix(unit, "K"):
		mult = 1024
	case strings.HasPrefix(unit, "M"):
		mult = 1024 * 1024
	case strings.HasPrefix(unit, "G"):
		mult = 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "T"):
		mult = 1024 * 1024 * 1024 * 1024
	}
	return s, int64(val * mult)
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func text(n *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return b.String()
}

func dedup(in []Entry) []Entry {
	seen := make(map[string]struct{}, len(in))
	out := make([]Entry, 0, len(in))
	for _, e := range in {
		if _, ok := seen[e.Name]; ok {
			continue
		}
		seen[e.Name] = struct{}{}
		out = append(out, e)
	}
	return out
}
