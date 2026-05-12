package app

import "log/slog"

type App struct {
}

func New(log *slog.Logger) *App {
	return &App{}
}
