package rooms

import (
	"fmt"

	"github.com/soapboxsocial/soapbox/pkg/blocks"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Auth struct {
	rooms   *Repository
	blocked *blocks.Backend
}

func NewAuth(rooms *Repository, blocked *blocks.Backend) *Auth {
	return &Auth{
		rooms:   rooms,
		blocked: blocked,
	}
}

// CanJoin returns whether a user can join a specific room.
func (a *Auth) CanJoin(room string, user int) bool {
	r, err := a.rooms.Get(room)
	if err != nil {
		return false
	}

	if !a.canJoin(r, user) {
		return false
	}

	return a.containsBlockers(r, user)
}

// FilterWhoCanJoin checks for a set of users who can join a room.
func (a *Auth) FilterWhoCanJoin(room string, users []int64) []int64 {
	r, err := a.rooms.Get(room)
	if err != nil {
		return []int64{}
	}

	res := make([]int64, 0)

	for _, user := range users {
		if !a.canJoin(r, int(user)) {
			continue
		}

		if a.containsBlockers(r, int(user)) {
			continue
		}

		res = append(res, user)
	}

	return res
}

func (a *Auth) canJoin(room *Room, user int) bool {
	if room.IsKicked(user) {
		return false
	}

	if room.Visibility() == pb.Visibility_VISIBILITY_PRIVATE {
		return room.IsInvited(user)
	}

	return true
}

func (a *Auth) containsBlockers(room *Room, user int) bool {
	blockingUsers, err := a.blocked.GetUsersWhoBlocked(user)
	if err != nil {
		fmt.Printf("failed to get blocked users who blocked: %+v", err)
	}

	return room.ContainsUsers(blockingUsers)
}
