package etl

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/search"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
)

type builder struct {
	data   store.ImportData
	places map[string]store.PlaceRecord
}

func newBuilder() *builder {
	return &builder{places: make(map[string]store.PlaceRecord)}
}

func (b *builder) Data() store.ImportData {
	b.fillMissingRecurringByCartier()
	b.fillMissingExactByCartier("paper")
	b.dedupeEvents()
	b.data.Places = b.data.Places[:0]
	for _, place := range b.places {
		b.data.Places = append(b.data.Places, place)
	}
	sort.SliceStable(b.data.Places, func(i, j int) bool {
		return b.data.Places[i].Key < b.data.Places[j].Key
	})
	return b.data
}

func (b *builder) dedupeEvents() {
	seen := map[string]struct{}{}
	out := b.data.Events[:0]
	for _, event := range b.data.Events {
		key := strings.Join([]string{
			event.PlaceKey,
			event.CartierNorm,
			event.WasteType,
			event.EventDate,
			event.StartTime,
			event.EndTime,
			event.Location,
			event.Kind,
		}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, event)
	}
	b.data.Events = out
}

func (b *builder) addPlace(sourceID int64, cartier string, street string) string {
	cartier = cleanCartier(cartier)
	street = cleanStreet(street)
	streetNorm := search.Normalize(street)
	cartierNorm := search.Normalize(cartier)
	if streetNorm == "" {
		return ""
	}
	key := strings.Join([]string{"ARAD", cartierNorm, streetNorm, "", ""}, "|")
	if _, exists := b.places[key]; exists {
		return key
	}
	aliases, _ := json.Marshal(search.Aliases(cartier, street))
	b.places[key] = store.PlaceRecord{
		Key:         key,
		UAT:         "ARAD",
		Cartier:     cartier,
		CartierNorm: cartierNorm,
		StreetRaw:   street,
		StreetNorm:  streetNorm,
		AliasesJSON: string(aliases),
		SourceID:    sourceID,
	}
	return key
}

func (b *builder) addIssue(sourceID int64, severity string, rowText string, reason string) {
	b.data.Issues = append(b.data.Issues, store.IssueRecord{
		SourceID: sourceID,
		Severity: severity,
		RowText:  rowText,
		Reason:   reason,
	})
}

func (b *builder) fillMissingRecurringByCartier() {
	type wasteDay struct {
		waste   string
		weekday int
	}
	counts := map[string]map[wasteDay]int{}
	hasRule := map[string]map[string]bool{}
	for _, rule := range b.data.Rules {
		place, ok := b.places[rule.PlaceKey]
		if !ok {
			continue
		}
		if hasRule[rule.PlaceKey] == nil {
			hasRule[rule.PlaceKey] = map[string]bool{}
		}
		hasRule[rule.PlaceKey][rule.WasteType] = true
		if counts[place.CartierNorm] == nil {
			counts[place.CartierNorm] = map[wasteDay]int{}
		}
		counts[place.CartierNorm][wasteDay{waste: rule.WasteType, weekday: int(rule.Weekday)}]++
	}

	dominant := map[string]map[string]wasteDay{}
	for cartier, byWasteDay := range counts {
		if dominant[cartier] == nil {
			dominant[cartier] = map[string]wasteDay{}
		}
		bestCount := map[string]int{}
		for candidate, count := range byWasteDay {
			if count > bestCount[candidate.waste] {
				bestCount[candidate.waste] = count
				dominant[cartier][candidate.waste] = candidate
			}
		}
	}

	for key, place := range b.places {
		for _, waste := range []string{"residual", "bio"} {
			if hasRule[key][waste] {
				continue
			}
			candidate, ok := dominant[place.CartierNorm][waste]
			if !ok {
				continue
			}
			b.data.Rules = append(b.data.Rules, store.RuleRecord{
				PlaceKey:       key,
				WasteType:      waste,
				RecurrenceKind: "weekly_inferred_cartier",
				Weekday:        timeWeekday(candidate.weekday),
				SourceID:       place.SourceID,
			})
		}
	}
}

func (b *builder) fillMissingExactByCartier(waste string) {
	type patternKey struct {
		cartier string
		place   string
	}
	patterns := map[patternKey][]store.EventRecord{}
	hasWaste := map[string]bool{}
	for _, event := range b.data.Events {
		if event.WasteType != waste || event.PlaceKey == "" {
			continue
		}
		hasWaste[event.PlaceKey] = true
		place, ok := b.places[event.PlaceKey]
		if !ok {
			continue
		}
		patterns[patternKey{cartier: place.CartierNorm, place: event.PlaceKey}] = append(patterns[patternKey{cartier: place.CartierNorm, place: event.PlaceKey}], event)
	}

	bestByCartier := map[string][]store.EventRecord{}
	for key, events := range patterns {
		if len(events) > len(bestByCartier[key.cartier]) {
			bestByCartier[key.cartier] = events
		}
	}

	for key, place := range b.places {
		if hasWaste[key] {
			continue
		}
		pattern := bestByCartier[place.CartierNorm]
		for _, event := range pattern {
			clone := event
			clone.PlaceKey = key
			clone.CartierNorm = place.CartierNorm
			clone.Kind = "inferred_cartier"
			clone.Confidence = 0.75
			clone.Title = compact(clone.Title + " inferat cartier")
			b.data.Events = append(b.data.Events, clone)
		}
	}
}

func timeWeekday(value int) time.Weekday {
	return time.Weekday(value)
}

func cleanStreet(value string) string {
	value = compact(value)
	prefixes := []string{"Strada ", "strada ", "Str. ", "str. ", "Bulevardul ", "bulevardul ", "Calea ", "calea ", "Aleea ", "aleea "}
	for _, prefix := range prefixes {
		value = strings.TrimPrefix(value, prefix)
	}
	return compact(value)
}

func cleanCartier(value string) string {
	value = compact(value)
	prefixes := []string{"Cartierul ", "cartierul ", "Cartier ", "cartier "}
	for _, prefix := range prefixes {
		value = strings.TrimPrefix(value, prefix)
	}
	return compact(value)
}

func compact(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
