package minis

type Scores map[int]int

type Mini struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Slug        string `json:"slug"`
	Size        int    `json:"size"`
	Description string `json:"description"`
}
