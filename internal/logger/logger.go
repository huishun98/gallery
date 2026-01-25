package logger

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	log := logrus.New()

	log.SetReportCaller(true)
	log.SetLevel(logrus.InfoLevel)

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableQuote:  true,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			// function name (last segment)
			funcName := filepath.Base(f.Function)

			// file:line
			fileName := fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)

			return funcName, fileName
		},
	})

	return log
}
