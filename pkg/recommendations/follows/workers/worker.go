package workers

import (
	"log"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/providers"
)

type Worker struct {
	twitter *providers.Twitter
	backend *follows.Backend
	queue   *pubsub.Queue
}

func (w *Worker) handle(job *Job) {
	id := job.UserID

	last, err := w.backend.LastUpdatedFor(id)
	if err != nil {
		log.Printf("backend.LastUpdatedFor err: %s", err)
		return
	}

	if time.Now().Sub(*last) >= 7*(24*time.Hour) {
		log.Println("week not passed")
		return
	}

	users, err := w.twitter.FindUsersToFollowFor(id)
	if err != nil {
		log.Printf("twitter.FindUsersToFollowFor err: %s", err)
		return
	}

	if len(users) == 0 {
		return
	}

	err = w.backend.AddRecommendationsFor(id, users)
	if err != nil {
		log.Printf("backend.AddRecommendationsFor err: %s", err)
		return
	}

	err = w.queue.Publish(pubsub.UserTopic, pubsub.NewFollowRecommendationsEvent(id))
	if err != nil {
		log.Printf("w.queue.Publish err: %s", err)
	}
}
