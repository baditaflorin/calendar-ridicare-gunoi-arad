package search

import (
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
)

func RankPlaces(query string, places []domain.Place, limit int) []domain.SearchResult {
	q := Normalize(query)
	if q == "" {
		return nil
	}

	results := make([]domain.SearchResult, 0, limit)
	for _, place := range places {
		score := scorePlace(q, place)
		if score <= 0 {
			continue
		}
		results = append(results, domain.SearchResult{Place: place, Score: score})
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Place.DisplayName() < results[j].Place.DisplayName()
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func scorePlace(query string, place domain.Place) int {
	candidates := []string{place.StreetNorm, place.CartierNorm + " " + place.StreetNorm}
	candidates = append(candidates, place.Aliases...)

	best := 0
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		switch {
		case candidate == query:
			best = max(best, 120)
		case strings.Contains(candidate, query):
			best = max(best, 105-len(candidate)+len(query))
		case strings.Contains(query, candidate):
			best = max(best, 92-len(query)+len(candidate))
		default:
			distance := levenshtein.ComputeDistance(query, candidate)
			longest := max(len([]rune(query)), len([]rune(candidate)))
			if longest == 0 {
				continue
			}
			score := 90 - int(float64(distance)/float64(longest)*100)
			if distance <= 3 {
				score += 25
			}
			best = max(best, score)
		}
	}
	if best < 45 {
		return 0
	}
	return best
}
