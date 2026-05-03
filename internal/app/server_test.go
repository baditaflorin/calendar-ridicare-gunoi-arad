package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/config"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/metrics"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
)

func TestServerSmokeEndpoints(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	sourceID, err := db.InsertSource(ctx, domain.Source{
		URL:           "https://retim.ro/utile-arad/zona-1/",
		SourceType:    "test",
		FetchedAt:     time.Date(2026, 5, 3, 20, 0, 0, 0, time.UTC),
		ContentHash:   "abc",
		RawPath:       "fixture.html",
		ParserVersion: "test",
		HTTPStatus:    200,
	})
	if err != nil {
		t.Fatal(err)
	}
	data := store.ImportData{
		Places: []store.PlaceRecord{{
			Key:         "ARAD|muresel|nicolae densuseanu||",
			UAT:         "ARAD",
			Cartier:     "Mureșel",
			CartierNorm: "muresel",
			StreetRaw:   "Nicolae Densușeanu",
			StreetNorm:  "nicolae densuseanu",
			AliasesJSON: `["nicolae densusianu"]`,
			SourceID:    sourceID,
		}},
		Rules: []store.RuleRecord{{
			PlaceKey:       "ARAD|muresel|nicolae densuseanu||",
			WasteType:      "residual",
			RecurrenceKind: "weekly",
			Weekday:        time.Monday,
			SourceID:       sourceID,
		}},
		Events: []store.EventRecord{{
			PlaceKey:   "ARAD|muresel|nicolae densuseanu||",
			WasteType:  "plastic_metal",
			EventDate:  "2026-05-12",
			StartTime:  "07:00",
			Kind:       "exact",
			SourceID:   sourceID,
			Confidence: 1,
		}},
	}
	if err := db.ReplaceImportedData(ctx, data); err != nil {
		t.Fatal(err)
	}

	server, err := NewServer(db, config.Config{PublicBaseURL: "http://example.test"}, metrics.New())
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	assertOK(t, ts.URL+"/readyz")
	assertContains(t, ts.URL+"/metrics", "gunoi_arad_http_requests_total")

	searchBody := get(t, ts.URL+"/api/search?q=densusianu")
	var searchPayload struct {
		Results []domain.SearchResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(searchBody), &searchPayload); err != nil {
		t.Fatal(err)
	}
	if len(searchPayload.Results) != 1 {
		t.Fatalf("search results = %d, want 1", len(searchPayload.Results))
	}

	placesBody := get(t, ts.URL+"/api/neighborhoods")
	assertStringContains(t, placesBody, `"norm":"muresel"`)

	eventsBody := get(t, ts.URL+"/api/events?place_id=1&from=2026-05-01&to=2026-05-31")
	assertStringContains(t, eventsBody, `"waste_type":"plastic_metal"`)
	assertStringContains(t, eventsBody, `"waste_type":"residual"`)
}

func assertOK(t *testing.T, url string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s status = %d, want 200", url, resp.StatusCode)
	}
}

func assertContains(t *testing.T, url string, want string) {
	t.Helper()
	body := get(t, url)
	assertStringContains(t, body, want)
}

func assertStringContains(t *testing.T, body string, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("body does not contain %q:\n%s", want, body)
	}
}

func get(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s status = %d body=%s", url, resp.StatusCode, buf.String())
	}
	return buf.String()
}
