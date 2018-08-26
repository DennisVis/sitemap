package main

import (
	"flag"
	"fmt"
	"net/url"
	"time"

	"github.com/DennisVis/sitemap/students/dennisvis/sitemap"
)

var (
	domain = flag.String("domain", "", "The domain to generate a sitemap for")
	scheme = flag.String("scheme", "https", "The scheme to use while crawling the domain (http or https)")
)

func main() {
	flag.Parse()

	if *domain == "" {
		panic("Please provide a domain to generate the sitemap for")
	}

	url, err := url.Parse(fmt.Sprintf("%s://%s", *scheme, *domain))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Going to generate sitemap for [%s]...\n", url.String())

	start := time.Now()
	sitemap, err := sitemap.Generator{URL: url}.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sitemap generated for [%s] in %.2f seconds:\n\n", url.String(), time.Now().Sub(start).Seconds())

	fmt.Println(sitemap)
}
