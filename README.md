# Das Rätsel

Web-basiertes Textadventure für Kinder in Go.

## Projektstruktur

- `cmd/das-raetsel/` CLI-Einstiegspunkt mit Subcommand `serve`
- `internal/app/` Bootstrapping der Anwendung
- `internal/server/` HTTP-Routen, Session-Cookie, Rendering
- `internal/db/` SQLite-Initialisierung und Session-Persistenz
- `internal/story/` Laden und Validieren der Story-Datei (YAML)
- `stories/` Story-Dateien
- `web/templates/` serverseitige HTML-Templates
- `web/static/` statische Assets (z. B. Bilder unter `/static/...`)

## Voraussetzungen

- Go 1.22+

## Start

```bash
go mod tidy
go run ./cmd/das-raetsel serve
```

Dann im Browser öffnen:

- <http://localhost:8080>

## CLI

```bash
go run ./cmd/das-raetsel serve -h
```

Wichtige Flags:

- `-addr` HTTP-Adresse (Default `:8080`)
- `-db` SQLite-Datei (Default `data/das-raetsel.db`)
- `-story` Story-Datei (Default `stories/intro.v2.yaml`)
- `-templates` Template-Glob (Default `web/templates/*.html`)
- `-static` Verzeichnis für statische Dateien (Default `web/static`)

## Aktueller Umfang (Vertikalschnitt)

- CLI mit `serve`
- HTTP-Server mit SSR-Template
- SQLite-DB inklusive Schema-Migration
- Story-Import aus lokaler YAML-Datei
- Erste Szene wird im Browser gerendert
- Session per Cookie + DB gespeichert
- Interaktive Puzzle-Formulare mit serverseitiger Auswertung
