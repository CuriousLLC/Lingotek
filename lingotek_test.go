package lingotek

import (
	"bytes"
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

func createTestServer(requestChan chan *http.Request, getFileName func(*http.Request) string) (*httptest.Server, http.Client) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var testData io.Reader
		var err error

		testData, err = os.Open(getFileName(r))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		io.Copy(w, testData)

		requestChan <- r
	})

	server := httptest.NewServer(handler)
	testUrl, _ := url.Parse(server.URL)

	client := http.Client{
		Transport: RewriteTransport{
			URL: testUrl,
		},
	}

	return server, client
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
	f := func(r *http.Request) string {
		return "test_data/test_communitys.json"
	}
	rCh := make(chan *http.Request, 1)
	server, client := createTestServer(rCh, f)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)

	r := Response{}

	selfLink := Link{
		Rel:  []string{"self"},
		Href: "test?limit=10&offset=0",
	}
	nextLink := Link{
		Rel:  []string{"next"},
		Href: "test?limit=10&offset=20",
	}

	r.Links = append(r.Links, selfLink)
	r.Links = append(r.Links, nextLink)

	api.getNextPage(&r)

	req := <-rCh
	offset := req.URL.Query().Get("offset")
	if offset != "20" {
		t.Errorf("Expected 20, got: %s", offset)
	}

	nextLink.Href = "test?limit=10&offset=30"
	r.Links[1] = nextLink
	api.getNextPage(&r)
	req = <-rCh
	offset = req.URL.Query().Get("offset")
	if offset != "30" {
		t.Errorf("Expected 30, got: %s", offset)
	}

}

func TestGetCommunities(t *testing.T) {
	f := func(r *http.Request) string {
		return "test_data/test_communitys.json"
	}
	rCh := make(chan *http.Request, 1)
	server, client := createTestServer(rCh, f)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)
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

}

func TestListCommunities(t *testing.T) {
	p := func(r *http.Request) (fileName string) {
		offset := r.URL.Query().Get("offset")

		if offset == "10" {
			fileName = "test_data/test_communities_justone.json"
		} else {
			fileName = "test_data/test_communitys.json"
		}

		return
	}

	rCh := make(chan *http.Request, 3)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)

	doneChan := make(chan bool)
	communityChan, _ := api.ListCommunities(doneChan)

	cNum := 0
	for _ = range communityChan {
		cNum += 1
	}

	if cNum != 11 {
		t.Errorf("Expected 11 communities, got %d", cNum)
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
	p := func(r *http.Request) (fileName string) {
		return "test_data/test_projects.json"
	}

	rCh := make(chan *http.Request, 1)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)
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

func TestListProjects(t *testing.T) {
	p := func(r *http.Request) (fileName string) {
		offset := r.URL.Query().Get("offset")

		if offset == "10" {
			fileName = "test_data/test_projects_three.json"
		} else {
			fileName = "test_data/test_projects.json"
		}

		return
	}

	rCh := make(chan *http.Request, 3)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)

	community := Community{}
	community.Property.Id = "dummycommunity"

	doneChan := make(chan bool)
	projectChan, _ := api.ListProjects(&community, doneChan)

	cNum := 0
	for _ = range projectChan {
		cNum += 1
	}

	if cNum != 13 {
		t.Errorf("Expected 13 projects, got %d", cNum)
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

func TestListDocuments(t *testing.T) {
	p := func(r *http.Request) (fileName string) {
		offset := r.URL.Query().Get("offset")

		if offset == "10" {
			fileName = "test_data/test_documents_two.json"
		} else {
			fileName = "test_data/test_documents.json"
		}

		return
	}

	rCh := make(chan *http.Request, 3)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)
	doneChan := make(chan bool)
	documentChan, _ := api.ListDocuments(doneChan)

	cNum := 0
	for _ = range documentChan {
		cNum += 1
	}

	if cNum != 12 {
		t.Errorf("Expected 12 documents, got %d", cNum)
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

func TestUploadString(t *testing.T) {
	p := func(r *http.Request) (fileName string) {
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded;charset=utf-8" {
			t.Error("Expected Content-Type x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		err := r.ParseForm()
		if err != nil {
			t.Error(err)
		}

		if r.Form.Get("locale_code") != "en-US" {
			t.Errorf("Expected en-US, got %s", r.PostForm.Get("locale_code"))
		}

		if r.Form.Get("project_id") != "12345" {
			t.Errorf("Expected 12345, got %s", r.PostForm.Get("project_id"))
		}

		return "test_data/status.json"
	}

	rCh := make(chan *http.Request, 1)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)
	prop := ProjectProperty{}
	prop.Id = "12345"
	project := Project{}
	project.Property = prop
	resp, err := api.UploadString("My API Test", "Let's go to the shoe store", "en-US", project)
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

func TestAddTranslation(t *testing.T) {
	p := func(r *http.Request) (fileName string) {
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded;charset=utf-8" {
			t.Error("Expected Content-Type x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		err := r.ParseForm()
		if err != nil {
			t.Error(err)
		}

		if r.PostForm.Get("locale_code") != "es-ES" {
			t.Errorf("Expected es-ES, got %s", r.PostForm.Get("locale_code"))
		}

		return "test_data/document_translate_post.json"
	}

	rCh := make(chan *http.Request, 1)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)
	prop := DocumentProperty{}
	document := Document{}
	document.Property = prop

	translation, err := api.AddTranslation(&document, "es-ES")
	if err != IdRequired {
		t.Error("Looking for IdRequired error")
	}

	prop.Id = "12345"
	document.Property = prop

	translation, err = api.AddTranslation(&document, "es-ES")
	if err != nil {
		t.Error(err)
	}

	if translation.Property.PercentComplete != 0 {
		t.Errorf("Expected 0, got %d", translation.Property.PercentComplete)
	}

}

func TestListTranslations(t *testing.T) {
	p := func(r *http.Request) (fileName string) {
		offset := r.URL.Query().Get("offset")

		if offset == "10" {
			fileName = "test_data/document_translate_get_empty.json"
		} else {
			fileName = "test_data/document_translate_get.json"
		}

		return
	}

	rCh := make(chan *http.Request, 3)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)
	doneChan := make(chan bool)
	prop := DocumentProperty{}
	prop.Id = "12345"
	document := Document{}
	document.Property = prop

	translationChan, _ := api.ListTranslations(&document, doneChan)

	cNum := 0
	for _ = range translationChan {
		cNum += 1
	}

	if cNum != 4 {
		t.Errorf("Expected 4 projects, got %d", cNum)
	}
}

func TestCheckStatus(t *testing.T) {
	p := func(r *http.Request) (fileName string) {
		return "test_data/document.json"
	}

	rCh := make(chan *http.Request, 1)
	server, client := createTestServer(rCh, p)
	defer server.Close()
	defer close(rCh)

	api := NewApi("dummyToken", &client)
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

func TestGetTranslatedDocument(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "test_data/big_document.bin")
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

	var buf bytes.Buffer

	n, err := api.GetTranslatedDocument(&document, "es-ES", &buf)

	if err != nil {
		t.Error(err)
	}

	if len(buf.Bytes()) != 102400 {
		t.Errorf("Expected len 102400, got %d", len(buf.Bytes()))
	}

	if n != int64(len(buf.Bytes())) {
		t.Errorf("Expected len(buf)(%d) to equal n(%d)", len(buf.Bytes()), n)
	}
}
