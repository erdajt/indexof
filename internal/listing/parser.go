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

var (
	sizeRe = regexp.MustCompile(`(?i)([\d.]+)\s*([KMGT]I?B?)`)
	hrefRe = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)
)

func Parse(base *url.URL, body []byte, showHidden bool) ([]Entry, error) {
	var entries []Entry

	if doc, err := html.Parse(bytes.NewReader(body)); err == nil {
		entries = append(entries, walkAnchors(doc, base, showHidden)...)
	}
	entries = append(entries, rawHrefs(base, body, showHidden)...)

	return dedup(entries), nil
}

func walkAnchors(n *html.Node, base *url.URL, showHidden bool) []Entry {
	var out []Entry
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := ""
			for _, a := range n.Attr {
				if a.Key == "href" {
					href = a.Val
					break
				}
			}
			if e, ok := makeEntry(base, href, strings.TrimSpace(textContent(n)), showHidden); ok {
				// size from nearby text
				for sib := n.NextSibling; sib != nil; sib = sib.NextSibling {
					if sib.Type == html.TextNode {
						if s, b := parseSize(strings.TrimSpace(sib.Data)); b > 0 || s != "" {
							e.Size, e.SizeBytes = s, b
							break
						}
					}
				}
				out = append(out, e)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return out
}

func rawHrefs(base *url.URL, body []byte, showHidden bool) []Entry {
	var out []Entry
	for _, m := range hrefRe.FindAllStringSubmatch(string(body), -1) {
		href := m[1]
		if e, ok := makeEntry(base, href, "", showHidden); ok {
			out = append(out, e)
		}
	}
	return out
}

func makeEntry(base *url.URL, href, name string, showHidden bool) (Entry, bool) {
	href = strings.TrimSpace(href)
	if href == "" || href == "/" || href == "../" || href == "./" {
		return Entry{}, false
	}
	if strings.HasPrefix(href, "?") || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
		return Entry{}, false
	}
	if strings.EqualFold(href, "mailto:") || strings.Contains(href, "://") && base != nil {
		// allow absolute http(s) links on same host only
		if u, err := url.Parse(href); err == nil && base != nil {
			if u.Scheme != "" && u.Host != "" && u.Host != base.Host {
				return Entry{}, false
			}
		}
	}

	if name == "" {
		name = path.Base(strings.TrimSuffix(href, "/"))
	}
	name = strings.TrimSpace(strings.TrimSuffix(name, "/"))
	if name == "" || name == "." || name == ".." {
		return Entry{}, false
	}
	if !showHidden && strings.HasPrefix(name, ".") {
		return Entry{}, false
	}

	u, err := base.Parse(href)
	if err != nil {
		return Entry{}, false
	}

	return Entry{
		Name:  name,
		URL:   u.String(),
		IsDir: strings.HasSuffix(href, "/"),
	}, true
}

func parseSize(s string) (string, int64) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return "", 0
	}
	m := sizeRe.FindStringSubmatch(s)
	if m == nil {
		return "", 0
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
	return strings.TrimSpace(m[0]), int64(val * mult)
}

func textContent(n *html.Node) string {
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
		key := e.Name
		if e.IsDir {
			key += "/"
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}
	return out
}
