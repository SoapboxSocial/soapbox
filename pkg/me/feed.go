package me

import "github.com/soapboxsocial/soapbox/pkg/stories"

// FeedUser represents a user that is displayed on the feed.
type FeedUser struct {
	Name     string          `json:"name"`
	Username string          `json:"username"`
	Image    string          `json:"image"`
	IsActive bool            `json:"is_active"`
	Room     string          `json:"room"`
	Stories  []stories.Story `json:"stories,omitempty"`
}
