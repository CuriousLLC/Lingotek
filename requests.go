package lingotek

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

var ServerError = errors.New("Server returned an error")

// getEntityCollection converts an array of entities into the specified type
func (l *Lingotek) getEntityCollection(route string, params *url.Values, entity interface{}) error {
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
		return body, ServerError
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
