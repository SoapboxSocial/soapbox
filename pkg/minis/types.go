package minis

type Scores map[int]int

// AuthKeys maps an access token to a game ID.
type AuthKeys map[string]int

type Mini struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Slug        string `json:"slug"`
	Size        int    `json:"size"`
	Description string `json:"description"`
}
