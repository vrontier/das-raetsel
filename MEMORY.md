# MEMORY

## Das Rätsel - Projektstand

### Zielsetzung (erster Vertikalschnitt)
- Web-basiertes Textadventure für Kinder mit Go
- Serverseitig gerendertes HTML ohne schweres Frontend-Framework
- Story aus lokalen Dateien laden
- Session speichern
- Lokal direkt startbar

### Umgesetzte Architektur
- CLI mit Subcommand `serve`
- Saubere Paketstruktur:
  - `cmd/das-raetsel`
  - `internal/app`
  - `internal/server`
  - `internal/db`
  - `internal/story`
  - `stories`
  - `web/templates`
  - `web/static`

### Backend-Funktionen
- HTTP-Server mit Routen:
  - `GET /`
  - `POST /choice`
  - `POST /puzzle`
  - `GET /static/*`
- SQLite-Initialisierung und Migration
- Session-Persistenz in SQLite
- Puzzle-Status pro Session+Szene in SQLite
- Fallback-Logik: ungültige gespeicherte Szene wird auf `start_scene` zurückgesetzt

### Story-System
- YAML-Loader mit Unterstützung für:
  - V1 (einfaches Modell)
  - V2 (`story`-Metadaten + Szenenliste)
- Validierung von Startszene und Choice-Zielen
- Aktive Story: `stories/intro.v2.yaml`

### Gameplay-Funktionen
- Erste Szene SSR im Browser rendern
- Interaktive Puzzle-Formulare mit serverseitiger Auswertung für:
  - `reading_question`
  - `combine_clues`
  - `fill_in_blank`
  - `sentence_order`
  - `dialog_choice`
- Szenen-Choices sind gesperrt bis das Szene-Rätsel gelöst ist
- Erfolg/Fehlschlag-Feedback aus Story-Daten

### Frontend/Assets
- Mobilfreundliches, serverseitiges HTML/CSS
- Statisches Bild-Serving unter `/static/...`
- Sichtbare SVG-Platzhalter für alle Intro-Szenenbilder

### Dokumentation und Qualität
- README mit Startanleitung, Struktur und CLI-Flags
- Tests für Story-Parsing (V1/V2)
- Tests für Puzzle-Auswertung
- Test für Static-Serving
- Test für Session-Fallback bei nicht mehr vorhandener Szene

### Zusammenarbeit / nächster Einstieg
- Kommunikation bevorzugt auf Deutsch.
- Der aktuelle Stand wurde als "guter Start" bestätigt.
- Vereinbarte sinnvolle nächste Ausbaustufen:
  - Spiel-State erweitern (Inventar/Fortschritt statt nur `solved` pro Szene)
  - Kapitelstruktur mit Story-Wechsel ohne Session-Brüche
  - Authoring-Tooling (Schema-Validierung + Vorschau)
