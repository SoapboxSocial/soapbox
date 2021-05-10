package types

type User struct {
	ID          int     `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	Image       string  `json:"image"`
	Bio         string  `json:"bio"`
	Email       *string `json:"email,omitempty"`
}
