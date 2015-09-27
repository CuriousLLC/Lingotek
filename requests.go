package lingotek

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

var ServerError = errors.New("Server returned an error")
var EndOfList = errors.New("No next rel found")

// GetNext will return a path and query values pointing towards
// the next page of an entity.
func (r *Response) GetNext() (string, *url.Values, error) {
	var route string
	foundNext := false
	params := url.Values{}

	for _, link := range r.Links {
		parsed, err := url.Parse(link.Href)
		if err != nil {
			return "", nil, err
		}

		query := parsed.Query()

		if link.Rel[0] == "self" {
			route = parsed.Path

			// We don't want to risk overwriting the offset and limit
			// parameters. The links may not arrive in a guaranteed order.
			for p, _ := range query {
				if p != "offset" && p != "limit" {
					params.Set(p, query.Get(p))
				}
			}
		} else if link.Rel[0] == "next" {
			params.Set("offset", query.Get("offset"))
			params.Set("limit", query.Get("limit"))
			foundNext = true
		}
	}

	if foundNext == false {
		return "", nil, EndOfList
	}

	return route, &params, nil
}

// getEntityCollection converts an array of entities into the specified type
func (l *Lingotek) getEntityCollectionPage(route string, params *url.Values, entity interface{}) error {
	var jsonResponse Response

	resp, err := l.doRequest(route, "GET", params)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, &jsonResponse)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonResponse.Entities, entity)
	if err != nil {
		return err
	}

	return nil
}

// getEntityCollectionFull takes an initial response, and iterates through every
// page until the amount of returned entities is 0. Each page will be sent to
// entityChan until the done message is received, or entities run out.
func (l *Lingotek) getEntityCollectionFull(response *Response, doneChan <-chan bool) (<-chan json.RawMessage, chan error) {
	entityChan := make(chan json.RawMessage)
	errChan := make(chan error)

	go func() {
		defer close(entityChan)
		defer close(errChan)

		for {
			resp, err := l.getNextPage(response)
			if err != nil {
				errChan <- err
				return
			}

			response = resp

			if response.Properties.Size == 0 {
				return
			}

			select {
			case entityChan <- response.Entities:
			case <-doneChan:
				return
			}
		}

	}()

	return entityChan, errChan
}

// getNextPage takes a Response, and returns a new Response holding
// the next set of entities.
func (l *Lingotek) getNextPage(response *Response) (*Response, error) {
	var jsonResponse Response

	route, params, err := response.GetNext()
	if err != nil {
		return nil, err
	}

	resp, err := l.doRequest(route, "GET", params)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, nil
}

// getEntity converts a single response entity into the specified type
func (l *Lingotek) getEntity(route string, params *url.Values, entity interface{}) error {
	resp, err := l.doRequest(route, "GET", params)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, entity)
	if err != nil {
		return err
	}

	return nil
}

func (l *Lingotek) doRequest(route, method string, params *url.Values) ([]byte, error) {
	url := target + route
	if params != nil {
		url += "?" + params.Encode()
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", l.AccessToken)
	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return body, errors.New(method + url + ":" + resp.Status)
	}

	return body, nil
}

func (l *Lingotek) postEntity(route string, params *url.Values, entity interface{}) error {
	resp, err := l.doRequest(route, "POST", params)
	if err != nil {
		if err != ServerError {
			return err
		}
	}

	jsonErr := json.Unmarshal(resp, entity)
	if jsonErr != nil {
		return jsonErr
	}

	return err
}
