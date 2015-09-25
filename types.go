package lingotek

import (
	"encoding/json"
	"strconv"
	"time"
)

// Lingotek sends us a microsecond timestamp, so we
// need to add our own unmarshal logic
type LingoTime struct {
	time.Time
}

func (l *LingoTime) UnmarshalJSON(data []byte) error {
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}

	l.Time = time.Unix(i/1000, 0)

	return nil
}

// API Response
// An API response will have multiple unknown entites based
// on what method was called. These must be unmarshalled after
// the type has been determined.
type Response struct {
	Class      []string         `json:"class"`
	Properties ResponseProperty `json:"properties"`
	Entities   json.RawMessage  `json:"entities"`
	Links      []Link           `json:"links"`
}

type ResponseProperty struct {
	Title  string `json:"title"`
	Offset int32  `json:"offset"`
	Total  int32  `json:"total"`
	Limit  int32  `json:"limit"`
	Size   int32  `json:"size"`
}

type Entity struct {
	Class []string `json:"class"`
}

// When an error occurs, we are given some information
// from the server
type Messages struct {
	Messages []string `json:"messages"`
}

type CommunityProperty struct {
	Title string `json:"title"`
	Id    string `json:"id"`
}

type Community struct {
	Actions  []Action          `json:"actions"`
	Property CommunityProperty `json:"properties"`
	Rel      []string          `json:"rel"`
	Links    []Link            `json:"links"`
}

type ProjectProperty struct {
	CreationDate LingoTime `json:"creation_date"`
	WorkflowId   string    `json:"workflow_id"`
	CallbackUrl  string    `json:"callback_url"`
	DueDate      LingoTime `json:"due_date"`
	Title        string    `json:"title"`
	CommunityId  string    `json:"community_id"`
	Id           string    `json:"id"`
}

type Project struct {
	Property ProjectProperty `json:"properties"`
	Rel      []string        `json:"rel"`
	Links    []Link          `json:"links"`
}

type StatusCountPart struct {
	Total  int `json:"total"`
	Unique int `json:"unique"`
}

type StatusCount struct {
	Segment   StatusCountPart `json:"segment"`
	Word      StatusCountPart `json:"word"`
	FormatTag StatusCountPart `json:"format_tag"`
}

type StatusProperty struct {
	Title    string      `json:"title"`
	Id       string      `json:"id"`
	Progress int32       `json:"progress"`
	Count    StatusCount `json:"count"`
}

type Status struct {
	Property StatusProperty `json:"properties"`
	Links    []Link         `json:"links"`
	Messages
}

type Link struct {
	Rel  []string `json:"rel"`
	Href string   `json:"href"`
}

type Field struct {
	Name     string
	Type     string
	Required bool
}

type Action struct {
	Name   string  `json:"name"`
	Method string  `json:"method"`
	Href   string  `json:"href"`
	Title  string  `json:"title"`
	Type   string  `json:"type"`
	Fields []Field `json:"fields"`
}

type LocaleProperty struct {
	Code         string `json:"code"`
	LanguageCode string `json:"lanuage_code"`
	CountryCode  string `json:"country_code"`
	Title        string `json:"title"`
	Language     string `json:"language"`
	Country      string `json:"country"`
}

type Locale struct {
	Property LocaleProperty `json:"properties"`
	Rel      []string       `json:"rel"`
	Links    []Link         `json:"links"`
}

type DocumentProperty struct {
	ProjectId   string    `json:"project_id"`
	UploadDate  LingoTime `json:"upload_date"`
	Title       string    `json:"title"`
	ExternalUrl string    `json:"external_url"`
	Name        string    `json:"name"`
	Id          string    `json:"id"`
	Extension   string    `json:"extension"`
}

type Document struct {
	Property DocumentProperty
	Locale   Locale
	Status   Status
}

func (d *Document) UnmarshalJSON(data []byte) error {
	docObject := make(map[string]json.RawMessage)
	entities := make([]json.RawMessage, 2)

	err := json.Unmarshal(data, &docObject)
	if err != nil {
		return err
	}

	err = json.Unmarshal(docObject["properties"], &d.Property)
	if err != nil {
		return err
	}

	err = json.Unmarshal(docObject["entities"], &entities)
	if err != nil {
		return err
	}

	err = json.Unmarshal(entities[0], &d.Locale)
	if err != nil {
		return err
	}

	err = json.Unmarshal(entities[1], &d.Status)

	return err
}
