package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/config"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/web"
)

type staticCatalog struct {
	GeneratedAt   string                `json:"generated_at"`
	Neighborhoods []domain.Neighborhood `json:"neighborhoods"`
	Places        []domain.Place        `json:"places"`
}

type staticPlaceFile struct {
	Place  domain.Place  `json:"place"`
	Events []staticEvent `json:"events"`
}

type staticEvent struct {
	WasteType  domain.WasteType `json:"waste_type"`
	Date       string           `json:"date"`
	StartTime  string           `json:"start_time,omitempty"`
	EndTime    string           `json:"end_time,omitempty"`
	Location   string           `json:"location,omitempty"`
	Kind       string           `json:"kind"`
	SourceURL  string           `json:"source_url,omitempty"`
	Confidence float64          `json:"confidence"`
	Generated  bool             `json:"generated,omitempty"`
}

type staticStats struct {
	GeneratedAt     string                    `json:"generated_at"`
	TotalPlaces     int                       `json:"total_places"`
	TotalEvents     int                       `json:"total_events"`
	WasteTypeCounts map[string]int            `json:"waste_type_counts"`
	DailyBreakdown  map[string]map[string]int `json:"daily_breakdown"`
}

func exportStatic(ctx context.Context, cfg config.Config, outDir string, fromValue string, toValue string) error {
	from, err := time.Parse(time.DateOnly, fromValue)
	if err != nil {
		return err
	}
	to, err := time.Parse(time.DateOnly, toValue)
	if err != nil {
		return err
	}
	db, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	empty, err := db.IsEmpty(ctx)
	if err != nil {
		return err
	}
	if empty {
		return fmt.Errorf("database is empty; run `gunoiarad etl --db %s` first", cfg.DBPath)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, dir := range []string{"data", "static", "program", "print", "map", "manifesto"} {
		if err := os.RemoveAll(filepath.Join(outDir, dir)); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(outDir, dir), 0o755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(outDir, ".nojekyll"), []byte(""), 0o644); err != nil {
		return err
	}
	if err := copyStaticAssets(outDir); err != nil {
		return err
	}
	if err := writeStaticPages(outDir); err != nil {
		return err
	}
	return writeStaticData(ctx, db, outDir, from, to)
}

func copyStaticAssets(outDir string) error {
	return fs.WalkDir(web.Assets, "static", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(outDir, path)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		content, err := web.Assets.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o644)
	})
}

func writeStaticPages(outDir string) error {
	pages := []struct {
		template string
		target   string
		prefix   string
	}{
		{"templates/index.html", "index.html", "./"},
		{"templates/index.html", "program/index.html", "../"},
		{"templates/map.html", "map/index.html", "../"},
		{"templates/manifesto.html", "manifesto/index.html", "../"},
		{"templates/ghid.html", "ghid/index.html", "../"},
		{"templates/analize.html", "analize/index.html", "../"},
	}
	for _, page := range pages {
		content, err := web.Assets.ReadFile(page.template)
		if err != nil {
			return err
		}
		html := staticHTML(string(content), page.prefix)
		targetPath := filepath.Join(outDir, page.target)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, []byte(html), 0o644); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(outDir, "404.html"), []byte(staticRedirectHTML()), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "print/index.html"), []byte(staticPrintHTML()), 0o644)
}

func staticHTML(html string, prefix string) string {
	cacheBust := fmt.Sprintf("?v=%d", time.Now().Unix())
	replacements := map[string]string{
		`href="/static/`:    `href="` + prefix + `static/`,
		`src="/static/`:     `src="` + prefix + `static/`,
		`href="/"`:          `href="` + prefix + `"`,
		`href="/map"`:       `href="` + prefix + `map/"`,
		`href="/manifesto"`: `href="` + prefix + `manifesto/"`,
		`href="/ghid"`:      `href="` + prefix + `ghid/"`,
		`href="/analize"`:   `href="` + prefix + `analize/"`,
		`href="https://retim.ro/utile-arad/zona-1/"`: `href="https://retim.ro/utile-arad/zona-1/"`,
		`<a href="/metrics">Metrics</a>`:             `<a href="https://retim.ro/utile-arad/zona-1/" target="_blank" rel="noreferrer">RETIM</a>`,
	}
	for old, newValue := range replacements {
		html = strings.ReplaceAll(html, old, newValue)
	}
	// Cache-bust CSS and JS
	html = strings.ReplaceAll(html, `.css"`, `.css`+cacheBust+`"`)
	html = strings.ReplaceAll(html, `.js"`, `.js`+cacheBust+`"`)
	html = strings.Replace(html, "</head>", "  <script>window.GUNOI_STATIC = true;</script>\n</head>", 1)
	return html
}

func staticRedirectHTML() string {
	return `<!doctype html>
<html lang="ro">
<meta charset="utf-8">
<title>Gunoi Arad</title>
<script>
  if (sessionStorage.getItem('redirect_loop') === location.href) {
    sessionStorage.removeItem('redirect_loop');
    document.write('<p>Pagina nu a fost gasita.</p>');
  } else {
    sessionStorage.setItem('redirect_loop', location.href);
    location.replace("./");
  }
</script>
<noscript><p><a href="./">Inapoi la Gunoi Arad</a></p></noscript>
</html>`
}

func staticPrintHTML() string {
	return `<!doctype html>
<html lang="ro">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Calendar de colectare - Gunoi Arad</title>
  <link rel="stylesheet" href="../static/styles.css">
  <script src="https://unpkg.com/html2canvas@1.4.1/dist/html2canvas.min.js"></script>
  <script>window.GUNOI_STATIC = true;</script>
  <script defer src="../static/print.js"></script>
</head>
<body class="print-body">
  <nav class="print-toolbar" aria-label="Actiuni calendar">
    <a id="print-back" href="../">Inapoi la program</a>
    <button type="button" onclick="downloadImage()">Descarca Imagine (PNG)</button>
  </nav>
  <main class="paper" id="print-paper">
    <p>Se incarca calendarul...</p>
  </main>
</body>
</html>`
}

func writeStaticData(ctx context.Context, db *store.Store, outDir string, from time.Time, to time.Time) error {
	places, err := db.AllPlaces(ctx)
	if err != nil {
		return err
	}
	neighborhoods, err := db.Neighborhoods(ctx)
	if err != nil {
		return err
	}
	generatedAt := time.Now().UTC().Format(time.RFC3339)
	if err := writeJSON(filepath.Join(outDir, "data/catalog.json"), staticCatalog{
		GeneratedAt:   generatedAt,
		Neighborhoods: neighborhoods,
		Places:        places,
	}); err != nil {
		return err
	}
	placesDir := filepath.Join(outDir, "data/places")
	if err := os.MkdirAll(placesDir, 0o755); err != nil {
		return err
	}
	var allEvents []staticEvent
	for _, place := range places {
		loadedPlace, events, err := db.EventsForPlace(ctx, place.ID, from, to)
		if err != nil {
			return err
		}
		payload := staticPlaceFile{
			Place:  loadedPlace,
			Events: compactEvents(events),
		}
		allEvents = append(allEvents, compactEvents(events)...)
		if err := writeJSON(filepath.Join(placesDir, fmt.Sprintf("%d.json", place.ID)), payload); err != nil {
			return err
		}
	}
	// Write precomputed stats
	stats := buildStats(allEvents, generatedAt, len(places))
	if err := writeJSON(filepath.Join(outDir, "data/stats.json"), stats); err != nil {
		return err
	}
	return nil
}

func compactEvents(events []domain.Event) []staticEvent {
	out := make([]staticEvent, 0, len(events))
	for _, event := range events {
		out = append(out, staticEvent{
			WasteType:  event.WasteType,
			Date:       event.DateISO,
			StartTime:  event.StartTime,
			EndTime:    event.EndTime,
			Location:   event.Location,
			Kind:       event.Kind,
			SourceURL:  event.SourceURL,
			Confidence: event.Confidence,
			Generated:  event.Generated,
		})
	}
	return out
}

func buildStats(events []staticEvent, generatedAt string, totalPlaces int) staticStats {
	wasteTypeCounts := make(map[string]int)
	dailyBreakdown := make(map[string]map[string]int)

	for _, e := range events {
		wasteTypeCounts[string(e.WasteType)]++
		if _, ok := dailyBreakdown[e.Date]; !ok {
			dailyBreakdown[e.Date] = make(map[string]int)
		}
		dailyBreakdown[e.Date][string(e.WasteType)]++
	}

	return staticStats{
		GeneratedAt:     generatedAt,
		TotalPlaces:     totalPlaces,
		TotalEvents:     len(events),
		WasteTypeCounts: wasteTypeCounts,
		DailyBreakdown:  dailyBreakdown,
	}
}

func writeJSON(path string, payload any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(payload)
}
