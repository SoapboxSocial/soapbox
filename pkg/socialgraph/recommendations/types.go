package recommendations

type FollowerRecommendations interface {
	FindUsersToFollowFor(user int) ([]int, error)
}
