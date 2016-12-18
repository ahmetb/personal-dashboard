package task

import (
	"os"

	"github.com/go-kit/kit/log"
)

func LoggerWithTask(name string) *log.Context {
	return log.NewContext(log.NewSyncLogger(log.NewLogfmtLogger(os.Stdout))).
		With("time", log.DefaultTimestampUTC, "task", name)
}

func LogFatal(logger *log.Context, keyvals ...interface{}) {
	logger.Log(keyvals...)
	os.Exit(1)
}
