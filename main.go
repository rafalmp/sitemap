package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rafalmp/link"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type urlset struct {
	Xmlns string `xml:"xmlns,attr"`
	Urls  []loc  `xml:"url"`
}

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "the url of the site to build the map for")
	maxDepth := flag.Int("depth", 3, "maximum traversal depth")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)

	toXml := urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{page})
	}

	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	fmt.Print(xml.Header)
	if err := enc.Encode(toXml); err != nil {
		panic(err)
	}
	fmt.Println()
}

func bfs(urlStr string, maxDepth int) []string {
	// go has no `set` data structure; one common way to implement it is to use
	// map[string]struct{} as an empty struct's size is 0 so it occupies no memory.
	seen := make(map[string]struct{})
	var q map[string]struct{}
	nq := map[string]struct{}{
		urlStr: {},
	}

	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		for url := range q {
			if _, found := seen[url]; found {
				continue
			}
			seen[url] = struct{}{}

			links, _ := get(url)
			for _, link := range links {
				nq[link] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(seen))
	for url := range seen {
		result = append(result, url)
	}
	return result
}

func get(urlStr string) ([]string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reqUrl := resp.Request.URL
	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}
	base := baseUrl.String()

	return filter(hrefs(resp.Body, base), withPrefix(base)), nil
}

func hrefs(r io.Reader, base string) []string {
	var result []string
	links, _ := link.Parse(r)
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			result = append(result, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			result = append(result, l.Href)
		}
	}

	return result
}

func filter(links []string, keepFn func(string) bool) []string {
	var result []string
	for _, link := range links {
		if keepFn(link) {
			result = append(result, link)
		}
	}
	return result
}

func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}
