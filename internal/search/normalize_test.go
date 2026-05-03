package search

import (
	"testing"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
)

func TestNormalizeRemovesDiacriticsAndStreetNoise(t *testing.T) {
	got := Normalize("Strada Nicolae Densușeanu, Arad")
	want := "nicolae densuseanu"
	if got != want {
		t.Fatalf("Normalize() = %q, want %q", got, want)
	}
}

func TestRankPlacesHandlesCommonMisspelling(t *testing.T) {
	places := []domain.Place{
		{
			ID:          1,
			Cartier:     "Mureșel",
			CartierNorm: "muresel",
			StreetRaw:   "Nicolae Densușeanu",
			StreetNorm:  "nicolae densuseanu",
			Aliases:     []string{"nicolae densusianu"},
		},
		{
			ID:          2,
			Cartier:     "Centru",
			CartierNorm: "centru",
			StreetRaw:   "Desseanu",
			StreetNorm:  "desseanu",
		},
	}

	results := RankPlaces("densusianu", places, 5)
	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}
	if results[0].Place.ID != 1 {
		t.Fatalf("top result id = %d, want 1", results[0].Place.ID)
	}
}
