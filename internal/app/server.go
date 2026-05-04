package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/config"
	icalexport "github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/ical"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/metrics"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/web"
)

type Server struct {
	store     *store.Store
	config    config.Config
	metrics   *metrics.Metrics
	templates *template.Template
	version   string
}

func NewServer(store *store.Store, cfg config.Config, met *metrics.Metrics) (*Server, error) {
	versionBytes, _ := os.ReadFile("VERSION")
	version := strings.TrimSpace(string(versionBytes))
	if version == "" {
		version = "dev"
	}

	tpl := template.New("").Funcs(template.FuncMap{
		"VERSION": func() string { return version },
	})
	tpl, err := tpl.ParseFS(web.Assets, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{store: store, config: cfg, metrics: met, templates: tpl, version: version}, nil
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(s.instrument)

	staticFS, _ := fs.Sub(web.Assets, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	r.Get("/", s.home)
	r.Get("/program", s.home)
	r.Get("/manifesto", s.manifesto)
	r.Get("/ghid", s.ghid)
	r.Get("/map", s.mapPage)
	r.Get("/print", s.print)
	r.Get("/ics", s.ics)
	r.Get("/api/search", s.search)
	r.Get("/api/places", s.places)
	r.Get("/api/neighborhoods", s.neighborhoods)
	r.Get("/api/events", s.events)
	r.Get("/healthz", s.healthz)
	r.Get("/readyz", s.readyz)
	r.Handle("/metrics", promhttp.HandlerFor(s.metrics.Registry(), promhttp.HandlerOpts{}))
	return r
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.templates.ExecuteTemplate(w, "index.html", map[string]any{
		"PublicBaseURL": s.config.PublicBaseURL,
	})
}

func (s *Server) manifesto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.templates.ExecuteTemplate(w, "manifesto.html", nil)
}

func (s *Server) ghid(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.templates.ExecuteTemplate(w, "ghid.html", nil)
}

func (s *Server) mapPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.templates.ExecuteTemplate(w, "map.html", nil)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := intQuery(r, "limit", 8)
	results, err := s.store.SearchPlaces(r.Context(), query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (s *Server) places(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	cartierNorm := strings.TrimSpace(r.URL.Query().Get("cartier_norm"))
	limit := intQuery(r, "limit", 20)
	results, err := s.store.Places(r.Context(), cartierNorm, query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (s *Server) neighborhoods(w http.ResponseWriter, r *http.Request) {
	results, err := s.store.Neighborhoods(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"neighborhoods": results})
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	placeID, err := int64Query(r, "place_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	from := dateQuery(r, "from", time.Now())
	to := dateQuery(r, "to", from.AddDate(0, 2, 0))
	place, events, err := s.store.EventsForPlace(r.Context(), placeID, from, to)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"place":  place,
		"events": events,
	})
}

func (s *Server) ics(w http.ResponseWriter, r *http.Request) {
	placeID, err := int64Query(r, "place_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	from := time.Now()
	to := from.AddDate(1, 0, 0)
	place, events, err := s.store.EventsForPlace(r.Context(), placeID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gunoi-arad.ics"`)
	_, _ = w.Write([]byte(icalexport.Build(place, events, s.config.PublicBaseURL)))
}

func (s *Server) print(w http.ResponseWriter, r *http.Request) {
	placeID, err := int64Query(r, "place_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	month := monthQuery(r, "month", time.Now())
	from := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)
	place, events, err := s.store.EventsForPlace(r.Context(), placeID, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := printData{
		Place:      place,
		Month:      from,
		MonthName:  romanianMonth(from.Month()) + " " + strconv.Itoa(from.Year()),
		Days:       buildMonthGrid(from, events),
		Legend:     legend(events),
		UpdatedAt:  latestFetchedAt(events),
		ProgramURL: "/program?place_id=" + strconv.FormatInt(place.ID, 10),
		Sources:    sources(events),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.templates.ExecuteTemplate(w, "print.html", data)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	empty, err := s.store.IsEmpty(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if empty {
		writeError(w, http.StatusServiceUnavailable, errors.New("database has no imported places"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		route := r.URL.Path
		if chiRoute := chi.RouteContext(r.Context()).RoutePattern(); chiRoute != "" {
			route = chiRoute
		}
		s.metrics.HTTPRequests.WithLabelValues(r.Method, route, metrics.StatusCodeLabel(rec.status)).Inc()
		s.metrics.HTTPDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func intQuery(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func int64Query(r *http.Request, key string) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0, errors.New(key + " is required")
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, errors.New(key + " must be a positive integer")
	}
	return parsed, nil
}

func dateQuery(r *http.Request, key string, fallback time.Time) time.Time {
	value := r.URL.Query().Get(key)
	if value == "" {
		return dateOnly(fallback)
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return dateOnly(fallback)
	}
	return parsed
}

func monthQuery(r *http.Request, key string, fallback time.Time) time.Time {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.Parse("2006-01", value)
	if err != nil {
		return fallback
	}
	return parsed
}

func dateOnly(t time.Time) time.Time {
	parsed, _ := time.Parse(time.DateOnly, t.Format(time.DateOnly))
	return parsed
}
