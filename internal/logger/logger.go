// Пакет для создания логгера из пакета slog и работы с ним.
package logger

import (
	"io"
	"log/slog"
	"os"
)

// SetupLogger инициализирует логгер из пакета slog с выводом в формате
// JSON и устанавливает его логгером по умолчанию, чтобы не передавать
// этот кастомный логгер другим объектам.
func SetupLogger() {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	slog.SetDefault(log)
}

// Err - обертка для ошибки, представляет ее как атрибут слоггера.
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

// Discard создает логгер из пакета slog, который никуда не пишет
// и устанавливает его логгером по умолчанию.
// Функция для использования в тестах.
func Discard() {
	log := slog.New(
		slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}),
	)
	slog.SetDefault(log)
}
