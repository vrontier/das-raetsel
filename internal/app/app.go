package app

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"das-raetsel/internal/db"
	"das-raetsel/internal/server"
	"das-raetsel/internal/story"
)

type App struct {
	store  *db.Store
	server *server.Server
}

func New(dbPath string, storyPath string, templateGlob string, staticDir string) (*App, error) {
	store, err := db.Open(dbPath)
	if err != nil {
		return nil, err
	}

	s, err := story.Load(storyPath)
	if err != nil {
		_ = store.Close()
		return nil, err
	}

	tmpl, err := template.ParseGlob(templateGlob)
	if err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	h := server.New(store, s, tmpl, staticDir)

	return &App{store: store, server: h}, nil
}

func (a *App) Close() error {
	return a.store.Close()
}

func (a *App) Serve(ctx context.Context, addr string) error {
	httpServer := &http.Server{
		Addr:    addr,
		Handler: a.server.Routes(),
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on http://localhost%s", addr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx, httpServer); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
