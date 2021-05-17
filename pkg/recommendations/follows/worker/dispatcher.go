package worker

import "sync"

type Dispatcher struct {
	jobs chan *Job
	pool chan chan *Job
	quit chan bool

	maxWorkers int

	config *Config
	wg     *sync.WaitGroup
}

func NewDispatcher(maxWorkers int, config *Config) *Dispatcher {
	return &Dispatcher{
		jobs:       make(chan *Job),
		pool:       make(chan chan *Job),
		quit:       make(chan bool),
		maxWorkers: maxWorkers,
		config:     config,
		wg:         &sync.WaitGroup{},
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

func (d *Dispatcher) Stop() {
	go func() {
		d.quit <- true
	}()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobs:
			// a job request has been received
			go func(job *Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.pool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		case <-d.quit:
			// We have been asked to stop.
			return
		}
	}
}

func (d *Dispatcher) Wait() {
	d.wg.Wait()
}

func (d *Dispatcher) Dispatch(user int) {
	d.wg.Add(1)

	go func() {
		d.jobs <- &Job{UserID: user, WaitGroup: d.wg}
	}()
}
