package worker

import "sync"

type Job struct {
	UserID    int
	WaitGroup *sync.WaitGroup
}
