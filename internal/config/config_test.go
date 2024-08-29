// Пакет для работы с файлом конфига.

package config

import (
	"GoExamComments/internal/logger"
	"testing"
)

// TestMustLoad позволяет проверить корректность указания пути
// к файлу конфига в переменных окружения.
func TestMustLoad(t *testing.T) {
	logger.Discard()

	var got *Config = MustLoad()
	if got == nil {
		t.Fatalf("MustLoad() error = failed to load config")
	}
}
