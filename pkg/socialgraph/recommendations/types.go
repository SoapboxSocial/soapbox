package recommendations

// FollowerRecommendations is a generic interface for returning a set of recommended users to follow
// for any specific user based on the algorithm.
type FollowerRecommendations interface {
	FindUsersToFollowFor(user int) ([]int, error)
}
