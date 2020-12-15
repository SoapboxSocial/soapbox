package stories

import "github.com/soapboxsocial/soapbox/pkg/users"

// Reaction represents the reactions users submitted to the story.
type Reaction struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

// Story represents a user story.
type Story struct {
	ID              string     `json:"id"`
	ExpiresAt       int64      `json:"expires_at"`
	DeviceTimestamp int64      `json:"device_timestamp"`
	Reactions       []Reaction `json:"reactions"`
}

// StoryFeed represents all of a users stories.
type StoryFeed struct {
	User    users.User `json:"user"`
	Stories []Story    `json:"stories"`
}
