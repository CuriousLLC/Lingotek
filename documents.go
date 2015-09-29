package lingotek

import (
	"encoding/json"
	"io"
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

func (l *Lingotek) UploadString(title, content, localeCode string, project Project) (*Status, error) {
	var status Status
	v := url.Values{}
	v.Set("title", title)
	v.Set("content", content)
	v.Set("locale_code", localeCode)
	v.Set("project_id", project.Property.Id)

	err := l.postEntity("document", &v, &status)
	return &status, err
}

func (l *Lingotek) AddTranslation(document *Document, localeCode string) (*Translation, error) {
	var translation Translation

	if document.Property.Id == "" {
		return nil, IdRequired
	}

	v := url.Values{}
	v.Set("locale_code", localeCode)

	err := l.postEntity("document/"+document.Property.Id+"/translation", &v, &translation)
	return &translation, err
}

func (l *Lingotek) GetTranslatedDocument(document *Document, localeCode string, writer io.Writer) (int64, error) {
	if document.Property.Id == "" {
		return 0, IdRequired
	}

	v := url.Values{}
	v.Set("locale_code", localeCode)

	return l.downloadContent("document/"+document.Property.Id+"/content", &v, writer)
}

func (l *Lingotek) ListTranslations(document *Document, doneChan <-chan bool) (<-chan Translation, <-chan error) {
	resultChan := make(chan Translation)
	errChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		response := l.createDummyResponse("document/"+document.Property.Id+"/translation", nil)

		var translations []Translation

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

			err = json.Unmarshal(response.Entities, &translations)
			if err != nil {
				errChan <- err
				return
			}

			for i := 0; i < len(translations); i++ {
				select {
				case <-doneChan:
					return
				default:
					resultChan <- translations[i]
				}
			}
		}
	}()

	return resultChan, errChan
}

func (l *Lingotek) CheckStatus(doc Document) (*Document, error) {
	if doc.Property.Id == "" {
		return nil, IdRequired
	}

	var document Document

	err := l.getEntity("document/"+doc.Property.Id, nil, &document)
	return &document, err
}

func (l *Lingotek) GetDocument(id string) (*Document, error) {
	var document Document

	err := l.getEntity("document/"+id, nil, &document)
	return &document, err
}
