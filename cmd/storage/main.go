package main

import (
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"radio-storage/internal/app"
	"radio-storage/internal/config"
	"radio-storage/internal/lib/logger/slogpretty"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env, cfg.LogPath)

	log.Info("starting radio", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	application := app.New(
		log,
		cfg.GRPC.Port,
		getAllowedIPs(),
		cfg.Source.SourcePath,
		cfg.Source.NestingDepth,
		cfg.Source.IdLength,
	)

	// Start app
	go application.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	// Stop app
	application.Stop()

	log.Info("application stopped")
}

func setupLogger(env, logPath string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envProd:
		var logWriter io.Writer

		if logPath == "" {
			logWriter = os.Stdout
		} else {
			var err error
			logWriter, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic("failed to open log file. Error: " + err.Error())
			}
		}

		log = slog.New(
			slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

func getAllowedIPs() []string {
	str := os.Getenv("ALLOWED_IPS")

	if str == "" {
		panic("secret not specified")
	}

	allowedIPs := strings.FieldsFunc(str, func(r rune) bool { return r == ':' })

	if len(allowedIPs) == 0 {
		panic("slice if allowed ips is empty")
	}

	return allowedIPs
}
