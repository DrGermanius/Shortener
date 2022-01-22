package app

import (
	"context"
	"strconv"
	"sync"

	"go.uber.org/zap"

	"github.com/DrGermanius/Shortener/internal/app/config"
)

type DeleteWorkerPool struct {
	workerCount int
	uid         string
	links       []string
	errCh       chan error
	context     context.Context
	quit        chan struct{}
	deleteFunc  func(context.Context, string, string) error
	logger      *zap.SugaredLogger
}

func NewDeleteWorkerPool(context context.Context, uid string, links []string, deleteFunc func(context.Context, string, string) error, logger *zap.SugaredLogger) DeleteWorkerPool {
	wc, err := strconv.Atoi(config.Config().WorkersCount)
	if err != nil {
		wc = 10
		logger.Errorf("error while reading config workers count: %f", err)
	}
	return DeleteWorkerPool{
		workerCount: wc,
		uid:         uid,
		links:       links,
		errCh:       make(chan error),
		context:     context,
		quit:        make(chan struct{}),
		deleteFunc:  deleteFunc,
		logger:      logger,
	}
}

func (w DeleteWorkerPool) Run() {
	go w.do()

	for {
		select {
		case <-w.errCh:
			{
				err := <-w.errCh
				w.logger.Error(err)
			}
		case <-w.quit:
			{
				return
			}
		case <-w.context.Done():
			{
				w.logger.Infof("context done")
				return
			}
		}
	}
}

func (w DeleteWorkerPool) do() {
	inCh := make(chan string, len(w.links))
	go w.fanIn(inCh)

	wg := sync.WaitGroup{}
	wg.Add(w.workerCount)
	for i := 0; i < w.workerCount; i++ {
		go w.runWorker(&wg, inCh)
	}
	wg.Wait()

	w.quit <- struct{}{}
	close(w.errCh)
}

func (w DeleteWorkerPool) fanIn(inCh chan<- string) {
	for _, v := range w.links {
		inCh <- v
	}
	close(inCh)
}

func (w DeleteWorkerPool) runWorker(wg *sync.WaitGroup, linksCh <-chan string) {
	for v := range linksCh {
		if err := w.deleteFunc(w.context, w.uid, v); err != nil {
			w.errCh <- err
		}
	}
	wg.Done()
}
