// Пакет для остановки вызывающей горутины.
package stopsignal

import (
	"os"
	"os/signal"
	"syscall"
)

// Stop блокирует выполнение горутины пока не поступит сигнал прерывания.
func Stop() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
