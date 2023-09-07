package utils

import (
	"errors"
	"log/slog"
	"os"
	"runtime/debug"
	"slices"
	"strings"
)

var loglevel *slog.LevelVar

func ConfigureLogLevel(logLevel string) {
	var level, err = parseLogLevel(logLevel)
	if err != nil {
		slog.Warn("Erreur de configuration sur le loglevel", slog.String("cause", err.Error()))
		return
	}
	if loglevel == nil {
		InitLogger()
	}
	loglevel.Set(level)
}

func InitLogger() {
	loglevel = new(slog.LevelVar)
	loglevel.Set(slog.LevelInfo)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: loglevel,
	})

	parentLogger := slog.New(
		handler)
	buildInfo, _ := debug.ReadBuildInfo()
	sha1 := findBuildSetting(buildInfo.Settings, "vcs.revision")
	appLogger := parentLogger.With(
		slog.Group("app", slog.String("sha1", sha1)),
	)
	slog.SetDefault(appLogger)

	slog.Info(
		"initialisation",
		slog.String("go", buildInfo.GoVersion),
		slog.String("path", buildInfo.Path),
		slog.Any("any", buildInfo.Settings),
	)
}

func findBuildSetting(settings []debug.BuildSetting, search string) string {
	retour := "NOT FOUND"
	slices.SortFunc(settings, func(s1 debug.BuildSetting, s2 debug.BuildSetting) int {
		return strings.Compare(s1.Key, s2.Key)
	})
	index, found := slices.BinarySearchFunc(settings, search, func(input debug.BuildSetting, searched string) int {
		return strings.Compare(input.Key, searched)
	})
	if found {
		retour = settings[index].Value
	}
	return retour
}

func parseLogLevel(logLevel string) (slog.Level, error) {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New("log level inconnu : '" + logLevel + "'")
	}
}
