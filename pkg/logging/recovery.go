package logging

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

func RecoverGoroutine(logger *slog.Logger, goroutineName string) {
	if r := recover(); r != nil {
		logger.Error("goroutine panicked",
			"goroutine", goroutineName,
			"panic", fmt.Sprintf("%v", r),
			"stack", string(debug.Stack()),
		)
	}
}
