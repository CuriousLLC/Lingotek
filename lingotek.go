package lingotek

import (
	"errors"
	"net/http"
	"net/url"
)

var ServerError = errors.New("Server returned an error")
var EndOfList = errors.New("No next rel found")
var IdRequired = errors.New("No ID given")

const target = "https://sandbox-api.lingotek.com/api/"

type Lingotek struct {
	AccessToken string
	client      *http.Client
}

func NewApi(accessToken string, client *http.Client) *Lingotek {
	api := Lingotek{"bearer " + accessToken, client}
	return &api
}

func (l *Lingotek) createDummyResponse(path string, params *url.Values) *Response {
	initialResponse := Response{}

	if params == nil {
		params = &url.Values{}
	}

	params.Set("limit", "10")
	params.Set("offset", "0")

	selfLink := Link{
		Rel:  []string{"self"},
		Href: path + "?" + params.Encode(),
	}
	nextLink := Link{
		Rel:  []string{"next"},
		Href: path + "?" + params.Encode(),
	}

	initialResponse.Links = append(initialResponse.Links, selfLink)
	initialResponse.Links = append(initialResponse.Links, nextLink)

	return &initialResponse
}
