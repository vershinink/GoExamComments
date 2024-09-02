// Пакет для работы с цензором Censor.
package censor

import (
	"GoExamComments/internal/config"
	"GoExamComments/internal/storage"
	"context"
	"log/slog"
	"strings"
	"time"
)

// Censor - структура цензора.
type Censor struct {
	queue chan storage.Comment
	list  []string
}

// tm - таймаут на изменение статуса цензуры у поста в БД.
var tm time.Duration = 10 * time.Second

// New - конструктор цензора.
func New(cfg *config.Config) *Censor {
	cnr := &Censor{
		queue: make(chan storage.Comment, 12),
		list:  cfg.CensorList,
	}
	return cnr
}

// Start запускает цензор в отдельной горутине. Читает структуры
// комментариев из канала queue и проверяет их на наличие недопустимых
// слов. Если находит совпадение, то выставляет статус цензуры
// у комментария в БД.
func (c *Censor) Start(st storage.DB) {
	go func() {
		for comm := range c.queue {
			go c.check(comm, st)
		}
		slog.Debug("censor stopped")
	}()

}

// Push отправляет переданный комментарий в очередь.
func (c *Censor) Push(comm storage.Comment) {
	c.queue <- comm
}

// Shutdown останавливает цензор, закрывая канал queue.
func (c *Censor) Shutdown() {
	close(c.queue)
}

// isOffensive проверяет контент комментария на содержание недопустимых
// выражений.
func (c *Censor) isOffensive(text string) bool {
	for _, word := range c.list {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

// check проводит цензурирование комментария, вызывая метод isOffensive,
// и устанавливает статус цензуры к комментария в БД.
func (c *Censor) check(comm storage.Comment, st storage.DB) {
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
