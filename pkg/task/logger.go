package task

import (
	"os"

	"github.com/go-kit/kit/log"
)

func LoggerWithTask(name, version string) *log.Context {
	if version == "" {
		version = "N/A"
	}
	return log.NewContext(log.NewSyncLogger(log.NewLogfmtLogger(os.Stdout))).
		With("task", name, "git", version, "time", log.DefaultTimestampUTC)
}

func LogFatal(logger *log.Context, keyvals ...interface{}) {
	logger.Log(keyvals...)
	os.Exit(1)
}
