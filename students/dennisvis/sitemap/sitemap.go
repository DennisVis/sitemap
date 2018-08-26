package sitemap

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/DennisVis/link/students/dennisvis/link"
)

// Generator takes a url.URL and allows for generating a sitemap for the domain given in that url.URL
type Generator struct {
	URL      *url.URL
	MaxDepth int
}

func (g Generator) toSitemap(anchors []link.Anchor) string {
	sitemap := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	for _, a := range anchors {
		sitemap = fmt.Sprintf(`%s<url><loc>%s</loc></url>`, sitemap, a.Href)
	}
	return sitemap + `</urlset>`
}

func (g Generator) contains(anchors []link.Anchor, anchor link.Anchor) bool {
	for _, a := range anchors {
		if a.Href == anchor.Href {
			return true
		}
	}
	return false
}

func (g Generator) toAbsHrefAnchor(anchor link.Anchor) link.Anchor {
	href := anchor.Href
	if strings.HasPrefix(href, "//") {
		href = fmt.Sprintf("%s:%s", g.URL.Scheme, anchor.Href)
	} else if strings.HasPrefix(href, "/") {
		href = fmt.Sprintf("%s://%s%s", g.URL.Scheme, g.URL.Host, anchor.Href)
	}
	if !strings.HasSuffix(href, "/") {
		href = href + "/"
	}
	return link.Anchor{
		Href: href,
		Text: anchor.Text,
	}
}

func (g Generator) appendOrIgnore(anchors []link.Anchor, anchor link.Anchor) []link.Anchor {
	absAnchor := g.toAbsHrefAnchor(anchor)
	domain := fmt.Sprintf("%s://%s", g.URL.Scheme, g.URL.Host)
	if !g.contains(anchors, absAnchor) && strings.HasPrefix(absAnchor.Href, domain) {
		anchors = append(anchors, absAnchor)
	}
	return anchors
}

func (g Generator) getDomainAnchors(anchorsToVisit, anchorsVisited []link.Anchor, currentDepth int) []link.Anchor {
	anchorsToVisitNext := make([]link.Anchor, 0)

	for _, anchor := range anchorsToVisit {
		absAnchor := g.toAbsHrefAnchor(anchor)
		anchorsVisited = append(anchorsVisited, absAnchor)

		res, err := http.Get(absAnchor.Href)
		if err != nil {
			fmt.Printf("Could not visit [%s]: %v", absAnchor.Href, err)
			continue
		}
		defer res.Body.Close()

		anchorsOnPage, err := link.ParseAnchors(res.Body)
		if err != nil {
			fmt.Printf("Could not parse anchors from [%s]: %v", absAnchor.Href, err)
			continue
		}

		for _, anchorOnPage := range anchorsOnPage {
			absAnchorOnPage := g.toAbsHrefAnchor(anchorOnPage)
			if !g.contains(anchorsToVisit, absAnchorOnPage) && !g.contains(anchorsVisited, absAnchorOnPage) {
				anchorsToVisitNext = g.appendOrIgnore(anchorsToVisitNext, absAnchorOnPage)
			}
		}
	}

	if len(anchorsToVisitNext) > 0 && currentDepth+1 <= g.MaxDepth {
		return g.getDomainAnchors(anchorsToVisitNext, anchorsVisited, currentDepth+1)
	}

	return anchorsVisited
}

// Generate generates the sitemap from the url.URL given during instance creation
func (g Generator) Generate() (sitemap string, err error) {
	anchorsToVisit := make([]link.Anchor, 1)
	anchorsToVisit[0] = link.Anchor(g.toAbsHrefAnchor(link.Anchor{Href: "/", Text: ""}))

	allAnchors := g.getDomainAnchors(anchorsToVisit, make([]link.Anchor, 0), 0)

	sitemap = g.toSitemap(allAnchors)

	return
}
