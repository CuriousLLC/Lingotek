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

func TestResponseGetNext(t *testing.T) {
	r := Response{}

	selfLink := Link{
		Rel:  []string{"self"},
		Href: "test?limit=10&offset=0",
	}
	nextLink := Link{
		Rel:  []string{"next"},
		Href: "test?limit=10&offset=10",
	}

	r.Links = append(r.Links, selfLink)
	r.Links = append(r.Links, nextLink)

	path, params, err := r.GetNext()
	if err != nil {
		t.Errorf("Get error %s", err)
	}

	if path != "test" {
		t.Errorf("Expected test, got %s", path)
	}

	if params.Get("limit") != "10" {
		t.Errorf("Expected limit 10, got %s", params.Get("limit"))
	}

	if params.Get("offset") != "10" {
		t.Errorf("Expected offset 10, got %s", params.Get("offset"))
	}

}

func TestGetNextPage(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		offset := r.URL.Query().Get("offset")
		if offset != "99" {
			t.Errorf("Expected 99, got: %s", offset)
		}

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

	r := Response{}

	selfLink := Link{
		Rel:  []string{"self"},
		Href: "test?limit=10&offset=0",
	}
	nextLink := Link{
		Rel:  []string{"next"},
		Href: "test?limit=10&offset=99",
	}

	r.Links = append(r.Links, selfLink)
	r.Links = append(r.Links, nextLink)

	api.getNextPage(&r)

}

func TestGetCommunities(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		offset := r.URL.Query().Get("offset")

		if offset == "10" {
			testData, err = os.Open("test_data/test_communities_empty.json")
			if err != nil {
				t.Error("Could not find test data")
				return
			}
		} else {
			testData, err = os.Open("test_data/test_communitys.json")
			if err != nil {
				t.Error("Could not find test data")
				return
			}
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
	resp, err := api.GetCommunitiesPage(0, 10)
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

	doneChan := make(chan bool)
	communityChan, _ := api.ListCommunities(doneChan)

	cNum := 0
	for _ = range communityChan {
		cNum += 1
	}

	if cNum != 10 {
		t.Errorf("Expected 10 communities, got %d", cNum)
	}

	communityChan, errs := api.ListCommunities(doneChan)
	cNum = 0
	for c := range communityChan {
		cNum += 1

		if cNum == 8 {
			doneChan <- true
		}

		if cNum == 7 {
			if c.Property.Title != "asdf" {
				t.Errorf("Expected \"asdf\", got %s", c.Property.Title)
			}
		}
	}

	// We always range one more time, because we don't see the done message
	// until the next value has already been written to the channel
	if cNum != 8 {
		t.Errorf("Expected 8 communities, got %d", cNum)
	}

	// If we exited early due to an error, we can check here
	if err, ok := <-errs; ok && err != nil {
		t.Errorf("Given error %s", err)
	}
}

func TestGetProjects(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		offset := r.URL.Query().Get("offset")
		if offset == "10" {
			testData, err = os.Open("test_data/test_projects_empty.json")
			if err != nil {
				t.Error("Could not find test data")
				return
			}
		} else {
			testData, err = os.Open("test_data/test_projects.json")
			if err != nil {
				t.Error("Could not find test data")
				return
			}
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

	community := Community{}
	community.Property.Id = "dummycommunity"

	doneChan := make(chan bool)
	projectChan, _ := api.ListProjects(&community, doneChan)

	cNum := 0
	for _ = range projectChan {
		cNum += 1
	}

	if cNum != 10 {
		t.Errorf("Expected 10 projects, got %d", cNum)
	}

	projectChan, _ = api.ListProjects(&community, doneChan)
	cNum = 0
	for c := range projectChan {
		cNum += 1

		if cNum == 8 {
			doneChan <- true
		}

		if cNum == 7 {
			if c.Property.Title != "devzone.lingotek.com" {
				t.Errorf("Expected \"devzone.lingotek.com\", got %s", c.Property.Title)
			}
		}
	}

	// We always range one more time, because we don't see the done message
	// until the next value has already been written to the channel
	if cNum != 8 {
		t.Errorf("Expected 8 projects, got %d", cNum)
	}
}

func TestGetDocuments(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		offset := r.URL.Query().Get("offset")
		if offset == "10" {
			testData, err = os.Open("test_data/test_documents_empty.json")
			if err != nil {
				t.Error("Could not find test data")
				return
			}
		} else {
			testData, err = os.Open("test_data/test_documents.json")
			if err != nil {
				t.Error("Could not find test data")
				return
			}
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
	doneChan := make(chan bool)
	documentChan, _ := api.ListDocuments(doneChan)

	cNum := 0
	for _ = range documentChan {
		cNum += 1
	}

	if cNum != 10 {
		t.Errorf("Expected 10 projects, got %d", cNum)
	}

	documentChan, _ = api.ListDocuments(doneChan)
	cNum = 0
	for c := range documentChan {
		cNum += 1

		if cNum == 8 {
			doneChan <- true
		}

		if cNum == 7 {
			if c.Property.Title != "My New Document 1401064842" {
				t.Errorf("Expected \"My New Document 1401064842\", got %s", c.Property.Title)
			}
		}
	}

	// We always range one more time, because we don't see the done message
	// until the next value has already been written to the channel
	if cNum != 8 {
		t.Errorf("Expected 8 projects, got %d", cNum)
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
		t.Error(err)
	}

	if resp.Property.Title != "My Test" {
		t.Errorf("Expected \"My Test\", got %s", resp.Property.Title)
	}

	if resp.Property.Id != "59d28ae8-25bd-4f99-85fc-9fd4fbc2af87" {
		t.Errorf("Expected \"59d28ae8-25bd-4f99-85fc-9fd4fbc2af87\", got %s", resp.Property.Id)
	}

	if resp.Status.Property.Title != "Status of My Test" {
		t.Errorf("Expected \"Status of My Test\", got %s", resp.Status.Property.Title)
	}

	if resp.Status.Property.Progress != 0 {
		t.Errorf("Expected 0, got %d", resp.Status.Property.Progress)
	}

	if resp.Status.Property.Count.Word.Total != 5 {
		t.Errorf("Expected 5, got %d", resp.Status.Property.Count.Word.Total)
	}
}
