package lingotek

import (
	"encoding/json"
	"net/url"
)

func (l *Lingotek) GetProjects(communityId string) ([]Project, error) {
	v := url.Values{}
	v.Set("community_id", communityId)

	var projects []Project

	err := l.getEntityCollectionPage("project", &v, &projects)
	return projects, err
}

func (l *Lingotek) ListProjects(community *Community, doneChan <-chan bool) (<-chan Project, <-chan error) {
	resultChan := make(chan Project)
	errChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		params := url.Values{}
		params.Set("community_id", community.Property.Id)
		response := l.createDummyResponse("project", &params)

		var projects []Project

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

			err = json.Unmarshal(response.Entities, &projects)
			if err != nil {
				errChan <- err
				return
			}

			for i := 0; i < len(projects); i++ {
				select {
				case <-doneChan:
					return
				default:
					resultChan <- projects[i]
				}

			}

		}
	}()
	return resultChan, errChan
}
