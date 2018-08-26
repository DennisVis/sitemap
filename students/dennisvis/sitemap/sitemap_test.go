package sitemap

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/DennisVis/link/students/dennisvis/link"
	"github.com/alecthomas/template"
)

var (
	testURL, _         = url.Parse("https://example.org")
	generatorUnderTest = Generator{testURL, 100}
)

func createTestAnchors() []link.Anchor {
	testAnchors := make([]link.Anchor, 3)
	testAnchors[0] = link.Anchor{Href: "https://example.org/about/", Text: "About example"}
	testAnchors[1] = link.Anchor{Href: "https://example.org/contact/", Text: "Contact example"}
	testAnchors[2] = link.Anchor{Href: "https://example.org/something-else/", Text: "Something completely different"}
	return testAnchors
}

func Test_toSiteMap_WhenGivenThreeAchnors_ShouldCreateSitemap(t *testing.T) {
	testAnchors := createTestAnchors()
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://example.org/about/</loc></url><url><loc>https://example.org/contact/</loc></url><url><loc>https://example.org/something-else/</loc></url></urlset>`

	result := generatorUnderTest.toSitemap(testAnchors)

	if result != expected {
		t.Errorf("Expected [%+v] to be equal to [%+v]", expected, result)
	}
}

func Test_contains_WhenGivenAnchorNotPresentInSlice_ShouldReturnFalse(t *testing.T) {
	if generatorUnderTest.contains(createTestAnchors(), link.Anchor{Href: "/foo", Text: "Bar"}) != false {
		t.Error("Should have returned false")
	}
}

func Test_contains_WhenGivenAnchorPresentInSlice_ShouldReturnTrue(t *testing.T) {
	if generatorUnderTest.contains(createTestAnchors(), createTestAnchors()[0]) != true {
		t.Error("Should have returned true")
	}
}

func Test_toAbsHrefAnchor_WhenGivenAbsoluteAnchor_ShouldReturnUnchangedAnchor(t *testing.T) {
	testAnchor := link.Anchor{Href: "http://example.org/about/"}
	expected := testAnchor

	result := generatorUnderTest.toAbsHrefAnchor(testAnchor)

	if result != expected {
		t.Errorf("Expected [%+v] to be equal to [%+v]", expected, result)
	}
}

func Test_toAbsHrefAnchor_WhenGivenSchemelessAnchor_ShouldReturnAnchorWithScheme(t *testing.T) {
	testAnchor := link.Anchor{Href: "//example.org/about"}
	expected := link.Anchor{Href: "https://example.org/about/"}

	result := generatorUnderTest.toAbsHrefAnchor(testAnchor)

	if result != expected {
		t.Errorf("Expected [%+v] to be equal to [%+v]", expected, result)
	}
}

func Test_toAbsHrefAnchor_WhenGivenRelativeAnchor_ShouldReturnAnchorWithSchemeAndHost(t *testing.T) {
	testAnchor := link.Anchor{Href: "/about"}
	expected := link.Anchor{Href: "https://example.org/about/"}

	result := generatorUnderTest.toAbsHrefAnchor(testAnchor)

	if result != expected {
		t.Errorf("Expected [%s] to be equal to [%s]", expected, result)
	}
}

func Test_appendOrIgnore_WhenGivenAnchorFromDifferentDomain_ShouldNotAppend(t *testing.T) {
	testAnchors := createTestAnchors()

	result := generatorUnderTest.appendOrIgnore(testAnchors, link.Anchor{Href: "https://otherdomain.com", Text: ""})

	if len(result) != len(testAnchors) {
		t.Error("Expected not to have appended")
	}
}

func Test_appendOrIgnore_WhenGivenAnchorFromSameDomainAlreadyPresentInSlice_ShouldNotAppend(t *testing.T) {
	testAnchors := createTestAnchors()

	result := generatorUnderTest.appendOrIgnore(testAnchors, createTestAnchors()[0])

	if len(result) != len(testAnchors) {
		t.Error("Expected not to have appended")
	}
}

func Test_appendOrIgnore_WhenGivenAnchorFromSameDomainNotPresentInSlice_ShouldAppend(t *testing.T) {
	testAnchors := createTestAnchors()

	result := generatorUnderTest.appendOrIgnore(testAnchors, link.Anchor{Href: "https://example.org/new", Text: ""})

	if len(result) == len(testAnchors) {
		t.Error("Expected to have appended")
	}
}

type anchorHolder struct {
	Anchors []link.Anchor
}

func serveTestSite() {
	t := template.New("test-template")
	t, err := t.Parse(`
		<!doctype html>
		<html>
			<head>
				<title>Test site</title>
			</head>
			<body>
				{{range .Anchors}}
					<a href="{{.Href}}">{{.Text}}</a>
				{{end}}
			</body>
		</html>
	`)
	if err != nil {
		panic(err)
	}

	homeAnchors := make([]link.Anchor, 5)
	homeAnchors[0] = link.Anchor{Href: "/", Text: "Home"}
	homeAnchors[1] = link.Anchor{Href: "/about", Text: "About example"}
	homeAnchors[2] = link.Anchor{Href: "http://localhost:8090/contact", Text: "Contact example"}
	homeAnchors[3] = link.Anchor{Href: "http://localhost:8090/something-else", Text: "Something completely different"}
	homeAnchors[4] = link.Anchor{Href: "http://otherhost:8090/somewhere-else", Text: "Somehere far away"}

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			if err := t.Execute(res, anchorHolder{homeAnchors}); err != nil {
				panic(err)
			}
		}
	})

	aboutAnchors := append([]link.Anchor(nil), homeAnchors[:len(homeAnchors)-2]...)
	aboutAnchors = append(aboutAnchors, link.Anchor{Href: "http://localhost:8090/about/more-info", Text: "More info"})

	http.HandleFunc("/about/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			if err := t.Execute(res, anchorHolder{aboutAnchors}); err != nil {
				panic(err)
			}
		}
	})

	http.HandleFunc("/contact/", func(res http.ResponseWriter, req *http.Request) {})
	http.HandleFunc("/something-else/", func(res http.ResponseWriter, req *http.Request) {})

	go func() {
		if err := http.ListenAndServe(":8090", nil); err != nil {
			panic(err)
		}
		fmt.Println("Running test site on [localhost:8090]")
	}()

	for {
		res, err := http.Get("http://localhost:8090")
		if err == nil && res.StatusCode == 200 {
			break
		}
		time.Sleep(2 * time.Second)
	}

	return
}

func Test_Generate_WhenGivenTestSite_ShouldGenerateSitemap(t *testing.T) {
	testURL, err := url.Parse("http://localhost:8090")
	if err != nil {
		panic(err)
	}
	generator := Generator{URL: testURL, MaxDepth: 5}
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>http://localhost:8090/</loc></url><url><loc>http://localhost:8090/about/</loc></url><url><loc>http://localhost:8090/contact/</loc></url><url><loc>http://localhost:8090/something-else/</loc></url><url><loc>http://localhost:8090/about/more-info/</loc></url></urlset>`

	serveTestSite()
	result, err := generator.Generate()
	if err != nil {
		t.Error(err)
	}

	if result != expected {
		t.Errorf("Expected [%+v] to equal [%+v]", result, expected)
	}
}
