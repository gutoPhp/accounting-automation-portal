package app

import (
	"net/http"
	"sync"
)

type App struct {
	store *Store

	sessionsMu sync.RWMutex
	sessions   map[string]User
}

func New(dataFile string) (*App, error) {
	store, err := NewStore(dataFile)
	if err != nil {
		return nil, err
	}

	app := &App{
		store:    store,
		sessions: make(map[string]User),
	}

	app.ensureMonthlySC02()
	app.ensureMonthlySC20()
	go app.runScheduler()

	return app, nil
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()

	a.registerAuthenticationRoutes(mux)
	a.registerDashboardRoutes(mux)
	a.registerSC02Routes(mux)
	a.registerSC05Routes(mux)
	a.registerSC06Routes(mux)
	a.registerSC20Routes(mux)

	mux.Handle("/", newSPAHandler("web/dist"))

	return withSecurityHeaders(mux)
}
