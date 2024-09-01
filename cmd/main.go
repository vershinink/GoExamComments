package main

import (
	"GoExamComments/internal/config"
	"GoExamComments/internal/logger"
	"GoExamComments/internal/server"
	"GoExamComments/internal/stopsignal"
	"GoExamComments/internal/storage/mongodb"
	"log/slog"
)

func main() {

	// Инициализируем конфиг файл и логгер.
	logger.SetupLogger()
	cfg := config.MustLoad()
	slog.Debug("config file and logger initialized")

	// Инициализируем базу данных.
	st := mongodb.New(cfg)
	slog.Debug("storage initialized")
	defer st.Close()

	// Инициализируем сервер, объявляем обработчики API и запускаем сервер.
	srv := server.New(cfg)
	srv.Start(cfg, st)
	slog.Info("Server started")

	// Блокируем выполнение основной горутины и ожидаем сигнала прерывания.
	stopsignal.Stop()

	// После сигнала прерывания останавливаем парсер и сервер.
	srv.Shutdown()

	slog.Info("Server stopped")
}
