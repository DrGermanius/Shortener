package app

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	"github.com/DrGermanius/Shortener/internal/app/config"
)

type WorkerPool struct {
	context context.Context
	inputCh chan input
	logger  *zap.SugaredLogger
}

type input struct {
	uid      string
	link     string
	context  context.Context
	function func(context.Context, string, string) error
}

func NewWorkerPool(context context.Context, logger *zap.SugaredLogger) WorkerPool {
	wc, err := strconv.Atoi(config.Config().WorkersCount)
	if err != nil {
		wc = 10
		logger.Errorf("error while reading config workers count: %f", err)
	}

	wp := WorkerPool{
		context: context,
		inputCh: make(chan input, 10),
		logger:  logger,
	}

	for i := 0; i < wc; i++ {
		go wp.listen()
	}
	go wp.listen()
	return wp
}

func (p WorkerPool) StartDeleteWorker(uid string, links []string, function func(context.Context, string, string) error, context context.Context) {
	for _, link := range links {
		i := input{
			uid:      uid,
			link:     link,
			function: function,
			context:  context,
		}
		p.inputCh <- i
	}
}

func (p WorkerPool) listen() {
	for {
		select {
		case v := <-p.inputCh:
			if err := v.function(v.context, v.uid, v.link); err != nil {
				p.logger.Error(err)
			}
		case <-p.context.Done():
			p.logger.Infof("context done")
			return
		}
	}
}
