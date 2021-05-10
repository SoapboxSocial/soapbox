package activeusers

import (
	"github.com/soapboxsocial/soapbox/pkg/users/types"
)

type ActiveUser struct {
	types.User

	Room *string `json:"room,omitempty"`
}
