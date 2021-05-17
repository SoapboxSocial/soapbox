package worker

import (
	"log"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/providers"
)

type Config struct {
	Twitter         *providers.Twitter
	Recommendations *follows.Backend
	Queue           *pubsub.Queue
}

type Worker struct {
	jobs    chan *Job
	workers chan<- chan *Job
	quit    chan bool

	config *Config
}

func NewWorker(pool chan<- chan *Job, config *Config) *Worker {
	w := &Worker{
		workers: pool,
		jobs:    make(chan *Job),
		quit:    make(chan bool),
		config:  config,
	}

	return w
}

func (w *Worker) Start() {
	go func() {
		for {
			w.workers <- w.jobs

			select {
			case job := <-w.jobs:
				// Receive a work request.
				w.handle(job)
			case <-w.quit:
				// We have been asked to stop.
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (w *Worker) handle(job *Job) {
	defer job.WaitGroup.Done()

	id := job.UserID

	last, err := w.config.Recommendations.LastUpdatedFor(id)
	if err != nil {
		log.Printf("backend.LastUpdatedFor err: %s", err)
		return
	}

	if time.Now().Sub(*last) >= 14*(24*time.Hour) {
		log.Println("week not passed")
		return
	}

	users, err := w.config.Twitter.FindUsersToFollowFor(id)
	if err != nil {
		log.Printf("twitter.FindUsersToFollowFor err: %s", err)
		return
	}

	if len(users) == 0 {
		return
	}

	err = w.config.Recommendations.AddRecommendationsFor(id, users)
	if err != nil {
		log.Printf("backend.AddRecommendationsFor err: %s", err)
		return
	}

	err = w.config.Queue.Publish(pubsub.UserTopic, pubsub.NewFollowRecommendationsEvent(id))
	if err != nil {
		log.Printf("w.queue.Publish err: %s", err)
	}
}
