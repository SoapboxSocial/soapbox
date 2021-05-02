package worker

import "github.com/soapboxsocial/soapbox/pkg/notifications"

type Dispatcher struct {
	jobs chan Job
	pool chan chan Job

	maxWorkers int

	config *Config
}

func NewDispatcher(maxWorkers int, config *Config) *Dispatcher {
	return &Dispatcher{
		jobs:       make(chan Job),
		pool:       make(chan chan Job),
		maxWorkers: maxWorkers,
		config:     config,
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.pool, d.config)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobs:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.pool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}

func (d *Dispatcher) Dispatch(targets []notifications.Target, notification *notifications.PushNotification) {
	go func() {
		d.jobs <- Job{Targets: targets, Notification: notification}
	}()
}
