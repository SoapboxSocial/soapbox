package me

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/stories"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type MeEndpoint struct {
	users       *users.UserBackend
	groups      *groups.Backend
	ns          *notifications.Storage
	oauthConfig *oauth1.Config
	la          *linkedaccounts.Backend
	stories     *stories.Backend
}

// Notification that the API returns.
// @TODO IN THE FUTURE WE MAY WANT TO BE ABLE TO SEND NOTIFICATIONS WITHOUT A USER, AND OTHER DATA?
// For example:
//   - group invites
//   - terms of service updates?
type Notification struct {
	Timestamp int64                              `json:"timestamp"`
	From      *users.NotificationUser            `json:"from"`
	Group     *groups.Group                      `json:"group"`
	Category  notifications.NotificationCategory `json:"category"`
}

func NewMeEndpoint(users *users.UserBackend, groups *groups.Backend, ns *notifications.Storage, config *oauth1.Config, la *linkedaccounts.Backend, backend *stories.Backend) *MeEndpoint {
	return &MeEndpoint{
		users:       users,
		groups:      groups,
		ns:          ns,
		oauthConfig: config,
		la:          la,
		stories:     backend,
	}
}

func (m *MeEndpoint) GetMe(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	user, err := m.users.FindByID(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetUser, "failed to get self")
		return
	}

	err = httputil.JsonEncode(w, user)
	if err != nil {
		log.Printf("failed to write me response: %s\n", err.Error())
	}
}

func (m *MeEndpoint) GetNotifications(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetUserIDFromContext(r.Context())
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

		if notification.Category == notifications.GROUP_INVITE {
			id, err := getId(notification, "group")
			if err != nil {
				log.Printf("getId err: %v\n", err)
				continue
			}

			group, err := m.groups.FindById(id)
			if err != nil {
				log.Printf("users.NotificationUserFor err: %v\n", err)
				continue
			}

			populatedNotification.Group = group
		}

		populated = append(populated, populatedNotification)
	}

	err = httputil.JsonEncode(w, populated)
	if err != nil {
		log.Printf("failed to write me response: %s\n", err.Error())
	}
}

func (m *MeEndpoint) AddTwitter(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, ok := auth.GetUserIDFromContext(r.Context())
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

func (m *MeEndpoint) RemoveTwitter(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetUserIDFromContext(r.Context())
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

func (m *MeEndpoint) GetFeed(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetUserIDFromContext(r.Context())
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
