package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/search"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err := store.ensureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) ensureSchema(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA foreign_keys=ON`,
		`CREATE TABLE IF NOT EXISTS sources (
			id INTEGER PRIMARY KEY,
			url TEXT NOT NULL,
			source_type TEXT NOT NULL,
			fetched_at TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			raw_path TEXT NOT NULL,
			parser_version TEXT NOT NULL,
			http_status INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sources_url_fetched ON sources(url, fetched_at DESC)`,
		`CREATE TABLE IF NOT EXISTS places (
			id INTEGER PRIMARY KEY,
			uat TEXT NOT NULL DEFAULT 'ARAD',
			cartier TEXT NOT NULL DEFAULT '',
			cartier_norm TEXT NOT NULL DEFAULT '',
			street_raw TEXT NOT NULL,
			street_norm TEXT NOT NULL,
			side TEXT NOT NULL DEFAULT '',
			house_parity TEXT NOT NULL DEFAULT '',
			aliases TEXT NOT NULL DEFAULT '[]',
			source_id INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			UNIQUE(uat, cartier_norm, street_norm, side, house_parity),
			FOREIGN KEY(source_id) REFERENCES sources(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_places_street_norm ON places(street_norm)`,
		`CREATE INDEX IF NOT EXISTS idx_places_cartier_norm ON places(cartier_norm)`,
		`CREATE TABLE IF NOT EXISTS rules (
			id INTEGER PRIMARY KEY,
			place_id INTEGER NOT NULL,
			waste_type TEXT NOT NULL,
			recurrence_kind TEXT NOT NULL,
			weekday INTEGER NOT NULL,
			week_label TEXT NOT NULL DEFAULT '',
			valid_from TEXT NOT NULL DEFAULT '',
			valid_to TEXT NOT NULL DEFAULT '',
			source_id INTEGER NOT NULL,
			FOREIGN KEY(place_id) REFERENCES places(id) ON DELETE CASCADE,
			FOREIGN KEY(source_id) REFERENCES sources(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rules_place ON rules(place_id)`,
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY,
			place_id INTEGER,
			cartier_norm TEXT NOT NULL DEFAULT '',
			waste_type TEXT NOT NULL,
			event_date TEXT NOT NULL,
			start_time TEXT NOT NULL DEFAULT '',
			end_time TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL DEFAULT '',
			kind TEXT NOT NULL DEFAULT 'exact',
			source_id INTEGER NOT NULL,
			confidence REAL NOT NULL DEFAULT 1.0,
			FOREIGN KEY(place_id) REFERENCES places(id) ON DELETE CASCADE,
			FOREIGN KEY(source_id) REFERENCES sources(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_place_date ON events(place_id, event_date)`,
		`CREATE INDEX IF NOT EXISTS idx_events_cartier_date ON events(cartier_norm, event_date)`,
		`CREATE TABLE IF NOT EXISTS campaigns (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			year_label TEXT NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			waste_type TEXT NOT NULL,
			location_type TEXT NOT NULL,
			source_id INTEGER NOT NULL,
			FOREIGN KEY(source_id) REFERENCES sources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS parse_issues (
			id INTEGER PRIMARY KEY,
			source_id INTEGER NOT NULL,
			severity TEXT NOT NULL,
			row_text TEXT NOT NULL,
			reason TEXT NOT NULL,
			resolved INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			FOREIGN KEY(source_id) REFERENCES sources(id)
		)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) InsertSource(ctx context.Context, src domain.Source) (int64, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO sources
		(url, source_type, fetched_at, content_hash, raw_path, parser_version, http_status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		src.URL,
		src.SourceType,
		src.FetchedAt.UTC().Format(time.RFC3339),
		src.ContentHash,
		src.RawPath,
		src.ParserVersion,
		src.HTTPStatus,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) ReplaceImportedData(ctx context.Context, data ImportData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, table := range []string{"parse_issues", "campaigns", "events", "rules", "places"} {
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return err
		}
	}

	placeIDs := make(map[string]int64, len(data.Places))
	for _, place := range data.Places {
		if place.Key == "" {
			continue
		}
		id, insertErr := upsertPlace(ctx, tx, place)
		if insertErr != nil {
			return insertErr
		}
		placeIDs[place.Key] = id
	}

	for _, rule := range data.Rules {
		placeID, ok := placeIDs[rule.PlaceKey]
		if !ok {
			return fmt.Errorf("rule refers to unknown place key %q", rule.PlaceKey)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO rules
			(place_id, waste_type, recurrence_kind, weekday, week_label, valid_from, valid_to, source_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			placeID,
			rule.WasteType,
			rule.RecurrenceKind,
			int(rule.Weekday),
			rule.WeekLabel,
			rule.ValidFrom,
			rule.ValidTo,
			rule.SourceID,
		); err != nil {
			return err
		}
	}

	for _, event := range data.Events {
		var placeID any
		if event.PlaceKey != "" {
			id, ok := placeIDs[event.PlaceKey]
			if !ok {
				return fmt.Errorf("event refers to unknown place key %q", event.PlaceKey)
			}
			placeID = id
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO events
			(place_id, cartier_norm, waste_type, event_date, start_time, end_time, location, title, kind, source_id, confidence)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			placeID,
			event.CartierNorm,
			event.WasteType,
			event.EventDate,
			event.StartTime,
			event.EndTime,
			event.Location,
			event.Title,
			event.Kind,
			event.SourceID,
			event.Confidence,
		); err != nil {
			return err
		}
	}

	for _, campaign := range data.Campaigns {
		if _, err = tx.ExecContext(ctx, `INSERT INTO campaigns
			(name, year_label, start_date, end_date, waste_type, location_type, source_id)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			campaign.Name,
			campaign.YearLabel,
			campaign.StartDate,
			campaign.EndDate,
			campaign.WasteType,
			campaign.LocationType,
			campaign.SourceID,
		); err != nil {
			return err
		}
	}

	for _, issue := range data.Issues {
		if _, err = tx.ExecContext(ctx, `INSERT INTO parse_issues
			(source_id, severity, row_text, reason)
			VALUES (?, ?, ?, ?)`,
			issue.SourceID,
			issue.Severity,
			issue.RowText,
			issue.Reason,
		); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func upsertPlace(ctx context.Context, tx *sql.Tx, place PlaceRecord) (int64, error) {
	row := tx.QueryRowContext(ctx, `INSERT INTO places
		(uat, cartier, cartier_norm, street_raw, street_norm, side, house_parity, aliases, source_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uat, cartier_norm, street_norm, side, house_parity) DO UPDATE SET
			cartier = excluded.cartier,
			street_raw = excluded.street_raw,
			aliases = excluded.aliases,
			source_id = excluded.source_id,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		RETURNING id`,
		place.UAT,
		place.Cartier,
		place.CartierNorm,
		place.StreetRaw,
		place.StreetNorm,
		place.Side,
		place.HouseParity,
		place.AliasesJSON,
		place.SourceID,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) IsEmpty(ctx context.Context) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM places`).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *Store) Counts(ctx context.Context) (domain.Counts, error) {
	var counts domain.Counts
	targets := []struct {
		table string
		dst   *int
	}{
		{"places", &counts.Places},
		{"rules", &counts.Rules},
		{"events", &counts.Events},
		{"sources", &counts.Sources},
		{"parse_issues", &counts.ParseIssues},
	}
	for _, target := range targets {
		if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+target.table).Scan(target.dst); err != nil {
			return counts, err
		}
	}
	return counts, nil
}

func (s *Store) AllPlaces(ctx context.Context) ([]domain.Place, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, uat, cartier, cartier_norm, street_raw, street_norm, side, house_parity, aliases
		FROM places
		ORDER BY cartier, street_raw`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []domain.Place
	for rows.Next() {
		place, scanErr := scanPlace(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		places = append(places, place)
	}
	return places, rows.Err()
}

func (s *Store) SearchPlaces(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	places, err := s.AllPlaces(ctx)
	if err != nil {
		return nil, err
	}
	return search.RankPlaces(query, places, limit), nil
}

func (s *Store) Places(ctx context.Context, cartierNorm string, query string, limit int) ([]domain.SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, uat, cartier, cartier_norm, street_raw, street_norm, side, house_parity, aliases
		FROM places
		WHERE (? = '' OR cartier_norm = ?)
		ORDER BY street_raw
		LIMIT 500`, cartierNorm, cartierNorm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []domain.Place
	for rows.Next() {
		place, scanErr := scanPlace(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		places = append(places, place)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if query != "" {
		return search.RankPlaces(query, places, limit), nil
	}
	results := make([]domain.SearchResult, 0, min(limit, len(places)))
	for index, place := range places {
		if index >= limit {
			break
		}
		results = append(results, domain.SearchResult{Place: place, Score: 100})
	}
	return results, nil
}

func (s *Store) Neighborhoods(ctx context.Context) ([]domain.Neighborhood, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT cartier, cartier_norm, COUNT(*)
		FROM places
		WHERE cartier_norm != ''
		GROUP BY cartier_norm
		ORDER BY cartier COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var neighborhoods []domain.Neighborhood
	for rows.Next() {
		var item domain.Neighborhood
		if err := rows.Scan(&item.Name, &item.Norm, &item.Count); err != nil {
			return nil, err
		}
		neighborhoods = append(neighborhoods, item)
	}
	return neighborhoods, rows.Err()
}

func (s *Store) Place(ctx context.Context, id int64) (domain.Place, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, uat, cartier, cartier_norm, street_raw, street_norm, side, house_parity, aliases
		FROM places WHERE id = ?`, id)
	place, err := scanPlace(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Place{}, err
	}
	return place, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPlace(row scanner) (domain.Place, error) {
	var place domain.Place
	var aliasesJSON string
	if err := row.Scan(&place.ID, &place.UAT, &place.Cartier, &place.CartierNorm, &place.StreetRaw, &place.StreetNorm, &place.Side, &place.HouseParity, &aliasesJSON); err != nil {
		return place, err
	}
	_ = json.Unmarshal([]byte(aliasesJSON), &place.Aliases)
	return place, nil
}

func (s *Store) EventsForPlace(ctx context.Context, placeID int64, from, to time.Time) (domain.Place, []domain.Event, error) {
	place, err := s.Place(ctx, placeID)
	if err != nil {
		return domain.Place{}, nil, err
	}
	events, err := s.exactEvents(ctx, place, from, to)
	if err != nil {
		return domain.Place{}, nil, err
	}
	rules, err := s.rulesForPlace(ctx, place.ID)
	if err != nil {
		return domain.Place{}, nil, err
	}
	events = append(events, generateRuleEvents(rules, from, to)...)
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Date.Equal(events[j].Date) {
			return events[i].WasteType < events[j].WasteType
		}
		return events[i].Date.Before(events[j].Date)
	})
	return place, events, nil
}

func (s *Store) exactEvents(ctx context.Context, place domain.Place, from, to time.Time) ([]domain.Event, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT
			e.id,
			COALESCE(e.place_id, 0),
			e.waste_type,
			e.event_date,
			e.start_time,
			e.end_time,
			e.location,
			e.title,
			e.kind,
			e.confidence,
			s.url,
			s.fetched_at
		FROM events e
		LEFT JOIN sources s ON s.id = e.source_id
		WHERE e.event_date BETWEEN ? AND ?
			AND (e.place_id = ? OR (e.place_id IS NULL AND e.cartier_norm = ?))
		ORDER BY e.event_date, e.waste_type`,
		from.Format(time.DateOnly),
		to.Format(time.DateOnly),
		place.ID,
		place.CartierNorm,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		var waste string
		var dateString string
		if err := rows.Scan(&event.ID, &event.PlaceID, &waste, &dateString, &event.StartTime, &event.EndTime, &event.Location, &event.Title, &event.Kind, &event.Confidence, &event.SourceURL, &event.FetchedAt); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.DateOnly, dateString)
		if err != nil {
			return nil, err
		}
		event.WasteType = domain.WasteType(waste)
		event.Date = parsed
		events = append(events, domain.DecorateEvent(event))
	}
	return events, rows.Err()
}

func (s *Store) rulesForPlace(ctx context.Context, placeID int64) ([]domain.Rule, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, place_id, waste_type, recurrence_kind, weekday, week_label, valid_from, valid_to, source_id
		FROM rules
		WHERE place_id = ?`, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var waste string
		var weekday int
		var validFrom, validTo string
		if err := rows.Scan(&rule.ID, &rule.PlaceID, &waste, &rule.RecurrenceKind, &weekday, &rule.WeekLabel, &validFrom, &validTo, &rule.SourceID); err != nil {
			return nil, err
		}
		rule.WasteType = domain.WasteType(waste)
		rule.Weekday = time.Weekday(weekday)
		if validFrom != "" {
			parsed, _ := time.Parse(time.DateOnly, validFrom)
			rule.ValidFrom = &parsed
		}
		if validTo != "" {
			parsed, _ := time.Parse(time.DateOnly, validTo)
			rule.ValidTo = &parsed
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func generateRuleEvents(rules []domain.Rule, from, to time.Time) []domain.Event {
	from = dateOnly(from)
	to = dateOnly(to)
	var events []domain.Event
	for _, rule := range rules {
		for current := from; !current.After(to); current = current.AddDate(0, 0, 1) {
			if current.Weekday() != rule.Weekday {
				continue
			}
			if rule.ValidFrom != nil && current.Before(dateOnly(*rule.ValidFrom)) {
				continue
			}
			if rule.ValidTo != nil && current.After(dateOnly(*rule.ValidTo)) {
				continue
			}
			events = append(events, domain.DecorateEvent(domain.Event{
				PlaceID:    rule.PlaceID,
				WasteType:  rule.WasteType,
				Date:       current,
				StartTime:  "07:00",
				Title:      domain.WasteLabel(rule.WasteType),
				Kind:       "recurring",
				Confidence: 1,
				Generated:  true,
			}))
		}
	}
	return events
}

func dateOnly(t time.Time) time.Time {
	parsed, _ := time.Parse(time.DateOnly, t.Format(time.DateOnly))
	return parsed
}
