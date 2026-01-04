package cli

import "github.com/respawn-app/ksrc/internal/executil"

type App struct {
	Runner executil.Runner
}

func NewApp() *App {
	return &App{Runner: executil.OSRunner{}}
}
