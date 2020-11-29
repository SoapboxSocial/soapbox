package internal

import "encoding/json"

type Hits struct {
	Total    map[string]interface{} `json:"total"`
	MaxScore float64                `json:"max_score"`
	Hits     []struct {
		Index  string          `json:"_index"`
		Type   string          `json:"_type"`
		ID     string          `json:"_id"`
		Score  float64         `json:"_score"`
		Source json.RawMessage `json:"_source"`
	} `json:"hits"`
}

type Result struct {
	Took     int            `json:"took"`
	TimedOut bool           `json:"timed_out"`
	Shards   map[string]int `json:"_shards"`
	Hits     Hits           `json:"hits"`
}
