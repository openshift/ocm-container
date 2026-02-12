package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func InitializeLogger() error {
	logLevel, err := setLogLevel()
	if err != nil {
		return err
	}
	noColor := (viper.GetBool("no-color") || (viper.IsSet("log.color") && !viper.GetBool("log.color")))
	logrus.SetFormatter(&TextFormatter{
		DisableTimestamp: true,
		DisableColors:    noColor,
	})
	logrus.SetLevel(logLevel)

	return nil
}

func setLogLevel() (logrus.Level, error) {
	// if the flag is set, use that var
	if viper.IsSet("log-level") {
		return parseLogLevel(viper.GetString("log-level"))
	}
	// otherwise look for the config key
	if viper.IsSet("log.level") {
		return parseLogLevel(viper.GetString("log.level"))
	}
	// otherwise return warning as a default
	return logrus.WarnLevel, nil
}

func parseLogLevel(level string) (logrus.Level, error) {
	switch level {
	case "debug":
		return logrus.DebugLevel, nil

	case "info":
		return logrus.InfoLevel, nil

	case "warn", "warning":
		return logrus.WarnLevel, nil

	case "err", "error":
		return logrus.ErrorLevel, nil
	}

	return logrus.WarnLevel, fmt.Errorf("Invalid log level. Valid values are 'debug', 'info', 'warning', 'error'")

}
