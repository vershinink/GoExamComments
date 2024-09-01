package censor

import (
	"GoExamComments/internal/config"
	"GoExamComments/internal/storage"
	"context"
	"log/slog"
	"strings"
	"time"
)

type Censor struct {
	queue chan storage.Comment
	list  []string
}

var tm time.Duration = 10 * time.Second

func New(cfg *config.Config) *Censor {
	cnr := &Censor{
		queue: make(chan storage.Comment, 12),
		list:  cfg.CensorList,
	}
	return cnr
}

func (c *Censor) Start(st storage.DB) {
	go func() {
		for comm := range c.queue {
			if c.isOffensive(comm.Content) {
				ctx, cancel := context.WithTimeout(context.Background(), tm)

				err := st.SetOffensive(ctx, comm.ID)
				if err != nil {
					slog.Error("cannot set offensive flag", slog.String("id", comm.ID), slog.String("err", err.Error()))
				}
				if ctx.Err() == context.DeadlineExceeded {
					slog.Error("context deadline exceed", slog.String("id", comm.ID))
					c.Push(comm)
				}
				cancel()
			}
		}
		slog.Debug("censor stopped")
	}()

}

func (c *Censor) Push(comm storage.Comment) {
	c.queue <- comm
}

func (c *Censor) Shutdown() {
	close(c.queue)
}

func (c *Censor) isOffensive(text string) bool {
	for _, word := range c.list {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
