package users

import (
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v7"
)

type query struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
}

type hits struct {
	Total    map[string]interface{} `json:"total"`
	MaxScore float64                `json:"max_score"`
	Hits     []struct {
		Index  string  `json:"_index"`
		Type   string  `json:"_type"`
		ID     string  `json:"_id"`
		Score  float64 `json:"_score"`
		Source User    `json:"_source"`
	} `json:"hits"`
}

type result struct {
	Took     int            `json:"took"`
	TimedOut bool           `json:"timed_out"`
	Shards   map[string]int `json:"_shards"`
	Hits     hits           `json:"hits"`
}

type Search struct {
	client *elasticsearch.Client
}

func NewSearchBackend(client *elasticsearch.Client) *Search {
	return &Search{client: client}
}

func (s *Search) FindUsers(input string) ([]User, error) {
	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex("users"),
		s.client.Search.WithQuery(input),
		s.client.Search.WithTrackTotalHits(true),
		s.client.Search.WithPretty(),
	)

	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	result := &result{}
	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	val := make([]User, 0)

	for _, h := range result.Hits.Hits {
		val = append(val, h.Source)
	}

	return val, nil
}
