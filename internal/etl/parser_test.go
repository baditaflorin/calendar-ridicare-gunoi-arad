package etl

import (
	"testing"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
)

func TestParseZonaAndCampaignInferMissingStreetFractions(t *testing.T) {
	zonaHTML := `
<table id="tablepress-56"><tbody>
<tr><td>Mureșel</td><td>Nicolae Horga</td><td>Luni</td><td>Miercuri</td><td>Miercuri</td></tr>
</tbody></table>
<table id="tablepress-10"><tbody>
<tr><td>Mureșel</td><td>Nicolae Horga</td><td>Saptamana 1</td><td>Miercuri</td><td></td><td></td><td></td><td></td><td>13;27</td><td></td><td></td><td></td><td></td><td></td><td></td><td></td></tr>
</tbody></table>
<table id="tablepress-62"><tbody>
<tr><td>Mureșel</td><td>Nicolae Densușeanu</td><td>Saptamana 1</td><td>Marți</td><td></td><td></td><td></td><td></td><td>12;26</td><td></td><td></td><td></td><td></td><td></td><td></td><td></td></tr>
</tbody></table>`
	campaignHTML := `
<table id="tablepress-65"><tbody></tbody></table>
<table id="tablepress-66"><tbody></tbody></table>
<table id="tablepress-67"><tbody>
<tr><td>Cartier Mureșel</td><td>Strada Nicolae Densușeanu</td><td>luni, 18 mai 2026</td></tr>
</tbody></table>`

	b := newBuilder()
	if err := ParseZona1(10, zonaHTML, b); err != nil {
		t.Fatal(err)
	}
	if err := ParseCampaign1(11, campaignHTML, b); err != nil {
		t.Fatal(err)
	}
	data := b.Data()

	var densuKey string
	for _, place := range data.Places {
		if place.StreetNorm == "nicolae densuseanu" {
			densuKey = place.Key
			break
		}
	}
	if densuKey == "" {
		t.Fatal("expected Densușeanu place to be extracted")
	}

	assertHasRule(t, data.Rules, densuKey, "residual")
	assertHasRule(t, data.Rules, densuKey, "bio")
	assertHasEvent(t, data.Events, densuKey, "plastic_metal", "2026-05-12", "exact")
	assertHasEvent(t, data.Events, densuKey, "paper", "2026-05-13", "inferred_cartier")
	assertHasEvent(t, data.Events, densuKey, "bulky", "2026-05-18", "campaign")
}

func assertHasRule(t *testing.T, rules []store.RuleRecord, placeKey string, waste string) {
	t.Helper()
	for _, rule := range rules {
		if rule.PlaceKey == placeKey && rule.WasteType == waste {
			return
		}
	}
	t.Fatalf("missing %s rule for %s", waste, placeKey)
}

func assertHasEvent(t *testing.T, events []store.EventRecord, placeKey string, waste string, date string, kind string) {
	t.Helper()
	for _, event := range events {
		if event.PlaceKey == placeKey && event.WasteType == waste && event.EventDate == date && event.Kind == kind {
			return
		}
	}
	t.Fatalf("missing %s %s event for %s on %s", kind, waste, placeKey, date)
}
