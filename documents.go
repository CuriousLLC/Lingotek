package lingotek

import (
	"encoding/json"
	"net/url"
)

func (l *Lingotek) ListDocuments(doneChan <-chan bool) (<-chan Document, <-chan error) {
	resultChan := make(chan Document)
	errChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		response := l.createDummyResponse("document", nil)

		var documents []Document

		for {
			resp, err := l.getNextPage(response)
			if err != nil {
				if err != EndOfList {
					errChan <- err
				}
				return
			}

			response = resp

			if response.Properties.Size == 0 {
				return
			}

			err = json.Unmarshal(response.Entities, &documents)
			if err != nil {
				errChan <- err
				return
			}

			for i := 0; i < len(documents); i++ {
				select {
				case <-doneChan:
					return
				default:
					resultChan <- documents[i]
				}

			}

		}
	}()

	return resultChan, errChan
}

func (l *Lingotek) TranslateString(title, content, localeCode string, project Project) (*Status, error) {
	var status Status
	v := url.Values{}
	v.Set("title", title)
	v.Set("content", content)
	v.Set("locale_code", localeCode)
	v.Set("project_id", project.Property.Id)

	err := l.postEntity("document", &v, &status)
	return &status, err
}

func (l *Lingotek) CheckStatus(doc Document) (*Document, error) {
	var document Document

	err := l.getEntity("document/"+doc.Property.Id, nil, &document)
	return &document, err
}
