package me

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/stories"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Endpoint struct {
	users       *users.UserBackend
	ns          *notifications.Storage
	oauthConfig *oauth1.Config
	la          *linkedaccounts.Backend
	stories     *stories.Backend
	queue       *pubsub.Queue
}

// Me is returned to the user calling the `/me` endpoint.
// It contains the user and additional information.
type Me struct {
	*users.User

	HasNotifications bool `json:"has_notifications"`
}

// Notification that the API returns.
// @TODO IN THE FUTURE WE MAY WANT TO BE ABLE TO SEND NOTIFICATIONS WITHOUT A USER, AND OTHER DATA?
// For example:
//   - terms of service updates?
type Notification struct {
	Timestamp int64                              `json:"timestamp"`
	From      *users.NotificationUser            `json:"from"`
	Room      *string                            `json:"room,omitempty"`
	Category  notifications.NotificationCategory `json:"category"`
}

func NewEndpoint(
	users *users.UserBackend,
	ns *notifications.Storage,
	config *oauth1.Config,
	la *linkedaccounts.Backend,
	backend *stories.Backend,
	queue *pubsub.Queue,
) *Endpoint {
	return &Endpoint{
		users:       users,
		ns:          ns,
		oauthConfig: config,
		la:          la,
		stories:     backend,
		queue:       queue,
	}
}

func (m *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", m.me).Methods("GET")
	r.HandleFunc("/notifications", m.notifications).Methods("GET")
	r.HandleFunc("/profiles/twitter", m.addTwitter).Methods("POST")
	r.HandleFunc("/profiles/twitter", m.removeTwitter).Methods("DELETE")
	r.HandleFunc("/feed", m.feed).Methods("GET")

	return r
}

func (m *Endpoint) me(w http.ResponseWriter, r *http.Request) {
	id, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	user, err := m.users.FindByID(id)
	if err != nil {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeFailedToGetUser, "failed to get self")
		return
	}

	has := m.ns.HasNewNotifications(id)
	me := &Me{user, has}

	go func() {
		err := m.queue.Publish(pubsub.UserTopic, pubsub.NewUserHeartbeatEvent(id))
		if err != nil {
			log.Printf("queue.Publish err %v", err)
		}
	}()

	err = httputil.JsonEncode(w, me)
	if err != nil {
		log.Printf("failed to write me response: %s\n", err.Error())
	}
}

func (m *Endpoint) notifications(w http.ResponseWriter, r *http.Request) {
	id, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	list, err := m.ns.GetNotifications(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetUser, "failed to get self")
		return
	}

	populated := make([]Notification, 0)
	for _, notification := range list {
		populatedNotification := Notification{Timestamp: notification.Timestamp, Category: notification.Category}

		from, err := m.users.NotificationUserFor(notification.From)
		if err != nil {
			log.Printf("users.NotificationUserFor err: %v\n", err)
			continue
		}

		populatedNotification.From = from

		if notification.Category == notifications.WELCOME_ROOM {
			room := notification.Arguments["room"].(string)
			populatedNotification.Room = &room
		}

		populated = append(populated, populatedNotification)
	}

	m.ns.MarkNotificationsViewed(id)

	err = httputil.JsonEncode(w, populated)
	if err != nil {
		log.Printf("failed to write me response: %s\n", err.Error())
	}
}

func (m *Endpoint) addTwitter(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "unauthorized")
		return
	}

	token := r.Form.Get("token")
	if token == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "fuck")
		return
	}

	secret := r.Form.Get("secret")
	if secret == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "fuck1")
		return
	}

	access := oauth1.NewToken(token, secret)
	httpClient := m.oauthConfig.Client(oauth1.NoContext, access)

	client := twitter.NewClient(httpClient)
	user, _, err := client.Accounts.VerifyCredentials(nil)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, err.Error())
		return
	}

	err = m.la.LinkTwitterProfile(id, int(user.ID), token, secret, user.ScreenName)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, err.Error())
		return
	}

	httputil.JsonSuccess(w)
}

func (m *Endpoint) removeTwitter(w http.ResponseWriter, r *http.Request) {
	id, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "unauthorized")
		return
	}

	err := m.la.UnlinkTwitterProfile(id)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, err.Error())
		return
	}

	httputil.JsonSuccess(w)
}

func (m *Endpoint) feed(w http.ResponseWriter, r *http.Request) {
	id, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "unauthorized")
		return
	}

	s, err := m.stories.GetStoriesForFollower(id, time.Now().Unix())
	if err != nil {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "unauthorized")
		return
	}

	feeds := make([]stories.StoryFeed, 0)
	for id, results := range s {
		user, err := m.users.FindByID(id)
		if err != nil {
			continue
		}

		user.Bio = ""
		user.Email = nil

		feeds = append(feeds, stories.StoryFeed{
			User:    *user,
			Stories: results,
		})
	}

	err = httputil.JsonEncode(w, feeds)
	if err != nil {
		log.Printf("failed to write me response: %s\n", err.Error())
	}
}

func getId(event *notifications.Notification, field string) (int, error) {
	creator, ok := event.Arguments[field].(float64)
	if !ok {
		return 0, errors.New("failed to recover creator")
	}

	return int(creator), nil
}
