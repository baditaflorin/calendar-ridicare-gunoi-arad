package etl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
)

const ParserVersion = "retim-html-v1"

type Target struct {
	URL        string
	SourceType string
	ParseKind  string
	Required   bool
}

var DefaultTargets = []Target{
	{
		URL:        "https://retim.ro/utile-arad/zona-1/",
		SourceType: "retim_zona_1",
		ParseKind:  "zona1",
		Required:   true,
	},
	{
		URL:        "https://retim.ro/utile-arad/campanii-periodice-2/anul-3/campania-1/",
		SourceType: "retim_campaign_year3_1",
		ParseKind:  "campaign1",
		Required:   true,
	},
	{
		URL:        "https://retim.ro/utile-arad/campanii-periodice-2/",
		SourceType: "retim_campaign_index",
		ParseKind:  "audit",
		Required:   false,
	},
	{
		URL:        "https://retim.ro/retim-si-adisigd-arad-anunta-modificarea-programului-de-colectare-pentru-deseurile-reciclabile-din-plastic-si-metal-pubela-galbena-in-municipiul-arad-si-orasele-din-zona-1/",
		SourceType: "retim_plastic_metal_change_notice",
		ParseKind:  "audit",
		Required:   false,
	},
}

type Service struct {
	Store   *store.Store
	RawDir  string
	Client  *http.Client
	Targets []Target
}

type Summary struct {
	Places       int
	Rules        int
	Events       int
	Campaigns    int
	Issues       int
	Sources      int
	Elapsed      time.Duration
	Warnings     []string
	LastFetchUTC time.Time
}

type snapshot struct {
	Target   Target
	SourceID int64
	HTML     string
}

func (s *Service) Run(ctx context.Context) (Summary, error) {
	started := time.Now()
	if s.Store == nil {
		return Summary{}, fmt.Errorf("etl store is nil")
	}
	if s.RawDir == "" {
		s.RawDir = "data/raw"
	}
	if s.Client == nil {
		s.Client = &http.Client{Timeout: 45 * time.Second}
	}
	if len(s.Targets) == 0 {
		s.Targets = DefaultTargets
	}

	var snapshots []snapshot
	var summary Summary
	for _, target := range s.Targets {
		snap, err := s.fetch(ctx, target)
		if err != nil {
			if target.Required {
				return summary, err
			}
			summary.Warnings = append(summary.Warnings, err.Error())
			continue
		}
		summary.Sources++
		summary.LastFetchUTC = time.Now().UTC()
		if target.ParseKind != "audit" {
			snapshots = append(snapshots, snap)
		}
	}

	builder := newBuilder()
	for _, snap := range snapshots {
		switch snap.Target.ParseKind {
		case "zona1":
			if err := ParseZona1(snap.SourceID, snap.HTML, builder); err != nil {
				return summary, err
			}
		case "campaign1":
			if err := ParseCampaign1(snap.SourceID, snap.HTML, builder); err != nil {
				return summary, err
			}
		}
	}
	data := builder.Data()
	if err := s.Store.ReplaceImportedData(ctx, data); err != nil {
		return summary, err
	}

	summary.Places = len(data.Places)
	summary.Rules = len(data.Rules)
	summary.Events = len(data.Events)
	summary.Campaigns = len(data.Campaigns)
	summary.Issues = len(data.Issues)
	summary.Elapsed = time.Since(started)
	return summary, nil
}

func (s *Service) fetch(ctx context.Context, target Target) (snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	if err != nil {
		return snapshot{}, err
	}
	req.Header.Set("User-Agent", "gunoi-arad/0.1 (+https://github.com/baditaflorin/calendar-ridicare-gunoi-arad)")
	resp, err := s.Client.Do(req)
	if err != nil {
		return snapshot{}, fmt.Errorf("fetch %s: %w", target.URL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return snapshot{}, fmt.Errorf("read %s: %w", target.URL, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return snapshot{}, fmt.Errorf("fetch %s: unexpected HTTP %d", target.URL, resp.StatusCode)
	}

	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])
	fetchedAt := time.Now().UTC()
	rawPath, err := writeSnapshot(s.RawDir, fetchedAt, target.SourceType, hash, body)
	if err != nil {
		return snapshot{}, err
	}
	sourceID, err := s.Store.InsertSource(ctx, domain.Source{
		URL:           target.URL,
		SourceType:    target.SourceType,
		FetchedAt:     fetchedAt,
		ContentHash:   hash,
		RawPath:       rawPath,
		ParserVersion: ParserVersion,
		HTTPStatus:    resp.StatusCode,
	})
	if err != nil {
		return snapshot{}, err
	}
	return snapshot{Target: target, SourceID: sourceID, HTML: string(body)}, nil
}

func writeSnapshot(rawDir string, fetchedAt time.Time, sourceType string, hash string, body []byte) (string, error) {
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		return "", err
	}
	slug := strings.NewReplacer("/", "-", ":", "-", " ", "-", "_", "-").Replace(sourceType)
	name := fmt.Sprintf("%s_%s_%s.html", fetchedAt.Format("20060102T150405Z"), hash[:16], slug)
	path := filepath.Join(rawDir, name)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
