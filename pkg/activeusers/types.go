package activeusers

import "github.com/soapboxsocial/soapbox/pkg/users"

type ActiveUser struct {
	users.User

	Room *string `json:"room,omitempty"`
}
