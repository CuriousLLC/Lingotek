package lingotek

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"
	"time"
)

// RewriteTransport is an http.RoundTripper that rewrites requests
// using the provided URL's Scheme and Host, and its Path as a prefix.
// The Opaque field is untouched.
// If Transport is nil, http.DefaultTransport is used
// http://stackoverflow.com/a/27894872/61980
type RewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// note that url.URL.ResolveReference doesn't work here
	// since t.u is an absolute url
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func TestGetCommunities(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		testData, err = os.Open("test_data/test_communitys.json")
		if err != nil {
			t.Error("Could not find test data")
			return
		}

		io.Copy(w, testData)
	})

	// Setup our test server
	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	testUrl, _ := url.Parse(ts.URL)

	client := &http.Client{
		Transport: RewriteTransport{
			URL: testUrl,
		},
	}
	api := NewApi("dummyToken", client)
	resp, err := api.GetCommunities()
	if err != nil {
		t.Error(err)
	}

	if len(resp) != 10 {
		t.Errorf("Expected 10 communities, got %d\n", len(resp))
	}

	if resp[7].Property.Title != "Blah blah community" {
		t.Errorf("Expected \"Blah blah community\", got %s", resp[8].Property.Title)
	}

	if resp[2].Actions[0].Href != "community/25e1d7ad-40be-45b8-83b0-90d62447b865" {
		t.Errorf("Expected \"community/25e1d7ad-40be-45b8-83b0-90d62447b865\" got %s", resp[2].Actions[0].Href)
	}

}

func TestGetProjects(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		testData, err = os.Open("test_data/test_projects.json")
		if err != nil {
			t.Error("Could not find test data")
			return
		}

		io.Copy(w, testData)
	})

	// Setup our test server
	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	testUrl, _ := url.Parse(ts.URL)

	client := &http.Client{
		Transport: RewriteTransport{
			URL: testUrl,
		},
	}
	api := NewApi("dummyToken", client)
	resp, err := api.GetProjects("dummyCommunity")
	if err != nil {
		t.Error(err)
	}

	if len(resp) != 10 {
		t.Errorf("Expected 10 projects, got %d\n", len(resp))
	}

	if resp[0].Property.Title != "jobName" {
		t.Errorf("Expected \"jobName\", got %s", resp[0].Property.Title)
	}

	if resp[1].Property.Id != "72106daf-69f9-4366-8ad8-2c52af9ca3ee" {
		t.Errorf("Expected \"72106daf-69f9-4366-8ad8-2c52af9ca3ee\", got %s", resp[1].Property.Id)
	}

	creationDate := time.Unix(1433116800, 0)
	if resp[2].Property.CreationDate.Time != creationDate {
		t.Errorf("Expected \"%s\", got %s", creationDate, resp[2].Property.CreationDate)
	}
}

func TestTranslateString(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded;charset=utf-8" {
			t.Error("Expected Content-Type x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		testData, err = os.Open("test_data/status.json")
		if err != nil {
			t.Error("Could not find test data")
			return
		}

		io.Copy(w, testData)
	})

	// Setup our test server
	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	testUrl, _ := url.Parse(ts.URL)

	client := &http.Client{
		Transport: RewriteTransport{
			URL: testUrl,
		},
	}
	api := NewApi("dummyToken", client)
	prop := ProjectProperty{}
	prop.Id = "12345"
	project := Project{}
	project.Property = prop
	resp, err := api.TranslateString("My API Test", "Let's go to the shoe store", "es_ES", project)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Property.Title != "Status of My Test" {
		t.Fatal("Expected \"Status of My Test\", got %s", resp.Property.Title)
	}

	if resp.Property.Id != "59d28ae8-25bd-4f99-85fc-9fd4fbc2af87" {
		t.Fatal("Expected \"59d28ae8-25bd-4f99-85fc-9fd4fbc2af87\", got %s", resp.Property.Id)
	}
}

func TestCheckStatus(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		testData, err = os.Open("test_data/document.json")
		if err != nil {
			t.Error("Could not find test data")
			return
		}

		io.Copy(w, testData)
	})

	// Setup our test server
	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	testUrl, _ := url.Parse(ts.URL)

	client := &http.Client{
		Transport: RewriteTransport{
			URL: testUrl,
		},
	}
	api := NewApi("dummyToken", client)
	prop := DocumentProperty{}
	prop.Id = "12345"
	document := Document{}
	document.Property = prop
	resp, err := api.CheckStatus(document)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Property.Title != "My Test" {
		t.Fatalf("Expected \"My Test\", got %s", resp.Property.Title)
	}

	if resp.Property.Id != "59d28ae8-25bd-4f99-85fc-9fd4fbc2af87" {
		t.Fatal("Expected \"59d28ae8-25bd-4f99-85fc-9fd4fbc2af87\", got %s", resp.Property.Id)
	}

	if resp.Status.Property.Progress != 0 {
		t.Fatalf("Expected 0, got %d", resp.Status.Property.Progress)
	}
}
