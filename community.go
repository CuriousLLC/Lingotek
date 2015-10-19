package lingotek

import (
	"encoding/json"
	"net/url"
	"strconv"
)

func (l *Lingotek) GetCommunity(communityId string) (*Community, error) {
	var community Community

	err := l.getEntity("community/"+communityId, nil, &community)
	return &community, err
}

func (l *Lingotek) GetCommunitiesPage(offset, limit int) ([]Community, error) {
	var communities []Community
	v := url.Values{}
	v.Set("offset", strconv.Itoa(offset))
	v.Set("limit", strconv.Itoa(limit))

	err := l.getEntityCollectionPage("community", &v, &communities)
	return communities, err
}

func (l *Lingotek) ListCommunities(doneChan <-chan bool) (<-chan Community, <-chan error) {
	resultChan := make(chan Community)
	errChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		response := l.createDummyResponse("community", nil)

		var communities []Community
		var totalRead = int32(0)

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

			err = json.Unmarshal(response.Entities, &communities)
			if err != nil {
				errChan <- err
				return
			}

			for i := 0; i < len(communities); i++ {
				totalRead += 1
				select {
				case <-doneChan:
					return
				default:
					resultChan <- communities[i]
				}
			}

			if totalRead == response.Properties.Total {
				return
			}
		}
	}()

	return resultChan, errChan
}
