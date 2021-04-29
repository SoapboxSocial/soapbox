package worker

type Worker struct {
	Work        chan Job
	WorkerQueue chan chan Job
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work

			job := <-w.Work
			// @TODO
		}
	}()
}
