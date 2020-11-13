package search

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/gorilla/mux"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Response struct {
	Users  []*users.User   `json:"users"`
	Groups []*groups.Group `json:"groups"`
}

type Endpoint struct {
	client *elasticsearch.Client
}

func NewEndpoint() *Endpoint {
	return &Endpoint{}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/search").Methods("GET").HandlerFunc(e.Search)

	return r
}

func (e *Endpoint) Search(w http.ResponseWriter, r *http.Request) {
	indexes, err := types(r.URL.Query())
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	response := Response{}

	var wg sync.WaitGroup
	for _, index := range indexes {
		if index == "users" {
			wg.Add(1)

			go func() {
				list, err := e.searchUsers("", 10, 10) // @todo
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
				list, err := e.searchGroups("", 10, 10) // @todo
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

	return nil, nil
}

func (e *Endpoint) searchUsers(query string, limit, offset int) ([]*users.User, error) {
	return nil, nil
	//response, err := e.search(query, "users", limit, offset)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return nil, nil
}

func (e *Endpoint) searchGroups(query string, limit, offset int) ([]*groups.Group, error) {
	return nil, nil
}

func (e *Endpoint) search(query, index string, limit, offset int) (*esapi.Response, error) {
	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(index),
		e.client.Search.WithQuery(query),
		e.client.Search.WithSize(limit),
		e.client.Search.WithFrom(offset),
		e.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		return nil, err
	}

	return res, nil
}
