package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rafalmp/link"
)

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "the url of the site to build the map for")
	flag.Parse()

	resp, err := http.Get(*urlFlag)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	reqUrl := resp.Request.URL
	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}
	base := baseUrl.String()

	var hrefs []string
	links, _ := link.Parse(resp.Body)
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			hrefs = append(hrefs, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			hrefs = append(hrefs, l.Href)
		}
	}

	for _, href := range hrefs {
		fmt.Println(href)
	}
}
