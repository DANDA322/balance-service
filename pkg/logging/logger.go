package logging

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func GetLogger(verbose string) *logrus.Logger {
	log := logrus.New()
	if strings.ToLower(verbose) == "true" {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("log level set to debug")
	}
	return log
}
