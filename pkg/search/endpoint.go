package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/gorilla/mux"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/search/internal"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

// @TODO maybe do a type?
const (
	usersIndex  = "users"
	groupsIndex = "groups"
)

type Response struct {
	Users  []*users.User   `json:"users,omitempty"`
	Groups []*groups.Group `json:"groups,omitempty"`
}

type Endpoint struct {
	client *elasticsearch.Client
}

func NewEndpoint(client *elasticsearch.Client) *Endpoint {
	return &Endpoint{client: client}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/").Methods("GET").HandlerFunc(e.Search)

	return r
}

func (e *Endpoint) Search(w http.ResponseWriter, r *http.Request) {
	indexes, err := types(r.URL.Query())
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	response := Response{}

	var wg sync.WaitGroup
	for _, index := range indexes {
		if index == "users" {
			wg.Add(1)

			go func() {
				list, err := e.searchUsers(query, limit, offset)
				if err != nil {
					log.Printf("failed to search users: %s\n", err.Error())
					wg.Done()
					return
				}

				response.Users = list
				wg.Done()
			}()
		}

		if index == "groups" {
			wg.Add(1)

			go func() {
				list, err := e.searchGroups(query, limit, offset)
				if err != nil {
					log.Printf("failed to search groups: %s\n", err.Error())
					wg.Done()
					return
				}

				response.Groups = list
				wg.Done()
			}()
		}
	}

	wg.Wait()

	err = httputil.JsonEncode(w, response)
	if err != nil {
		log.Printf("failed to write search response: %s\n", err.Error())
	}
}

func types(query url.Values) ([]string, error) {
	indexes := query.Get("type")
	if indexes == "" {
		return nil, errors.New("no indexes")
	}

	vals := strings.Split(indexes, ",")
	for _, val := range vals {
		if val != usersIndex && val != groupsIndex {
			return nil, fmt.Errorf("invalid index %s", vals)
		}
	}

	return vals, nil
}

func (e *Endpoint) searchUsers(query string, limit, offset int) ([]*users.User, error) {
	res, err := e.search("users", query, limit, offset)
	if err != nil {
		return nil, err
	}

	data := make([]*users.User, 0)
	for _, hit := range res.Hits.Hits {
		user := &users.User{}
		err := json.Unmarshal(hit.Source, user)
		if err != nil {
			continue
		}

		data = append(data, user)
	}

	return data, nil
}

func (e *Endpoint) searchGroups(query string, limit, offset int) ([]*groups.Group, error) {
	res, err := e.search("groups", query, limit, offset)
	if err != nil {
		return nil, err
	}

	data := make([]*groups.Group, 0)
	for _, hit := range res.Hits.Hits {
		group := &groups.Group{}
		err := json.Unmarshal(hit.Source, group)
		if err != nil {
			continue
		}

		data = append(data, group)
	}

	return data, nil
}

func (e *Endpoint) search(index, query string, limit, offset int) (*internal.Result, error) {
	config := []func(*esapi.SearchRequest){
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(index),
		e.client.Search.WithQuery(query),
		e.client.Search.WithSize(limit),
		e.client.Search.WithFrom(offset),
		e.client.Search.WithTrackTotalHits(true),
	}

	if index == "users" && query == "*" {
		if query == "*" {
			config = append(config, e.client.Search.WithSort("room_time:desc", "followers:desc"))
		} else {
			config = append(config, e.client.Search.WithSort("_score"))
		}
	}

	res, err := e.client.Search(config...)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result := &internal.Result{}
	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
