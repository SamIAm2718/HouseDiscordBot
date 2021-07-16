package utils

import (
	"os"
	"time"

	"github.com/samuel-mokhtar/DiscordTwitchBot/constants"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

var (
	Log *logrus.Logger
)

func init() {
	Log = logrus.New()

	logLevel := logrus.InfoLevel
	if constants.Debug {
		logLevel = logrus.DebugLevel
	}

	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   constants.LogPath + "/bot.log",
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      logLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC822,
		},
	})
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to initialize file rotate hook.")
	}

	Log.SetLevel(logLevel)
	Log.SetOutput(os.Stderr)
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})
	Log.AddHook(rotateFileHook)
}
