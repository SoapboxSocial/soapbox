package minis

type Mini struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}
