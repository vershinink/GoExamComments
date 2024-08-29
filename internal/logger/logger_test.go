// Пакет для создания логгера из пакета slog и работы с ним.

package logger

import (
	"errors"
	"testing"
)

var errTest = errors.New("test")

func TestErr(t *testing.T) {
	got := Err(errTest)
	if got.String() != "error=test" {
		t.Errorf("Err.String() = %v, want %s", got.String(), "error=test")
	}
}
