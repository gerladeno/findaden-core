package logging

import "github.com/sirupsen/logrus"

func GetLogger(verbose bool) *logrus.Logger {
	log := logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})
	if verbose {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("log level set to debug")
	}
	return log
}
