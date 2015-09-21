package lingotek

import (
	"net/http"
	"net/url"
)

const target = "https://sandbox-api.lingotek.com/api/"

type Lingotek struct {
	AccessToken string
	client      *http.Client
}

func NewApi(accessToken string, client *http.Client) *Lingotek {
	api := Lingotek{"bearer " + accessToken, client}
	return &api
}

func (l *Lingotek) GetProjects(communityId string) ([]Project, error) {
	v := url.Values{}
	v.Set("community_id", communityId)

	var projects []Project

	err := l.getEntityCollection("project", &v, &projects)
	return projects, err
}

func (l *Lingotek) GetCommunity(communityId string) (*Community, error) {
	var community Community

	err := l.getEntity("community/"+communityId, nil, &community)
	return &community, err
}

func (l *Lingotek) GetCommunities() ([]Community, error) {
	var communities []Community

	err := l.getEntityCollection("community", nil, &communities)
	return communities, err
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
