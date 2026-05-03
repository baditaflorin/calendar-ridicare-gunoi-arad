package ical

import (
	"fmt"
	"strings"
	"time"

	gocal "github.com/arran4/golang-ical"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
)

func Build(place domain.Place, events []domain.Event, publicBaseURL string) string {
	cal := gocal.NewCalendarFor("gunoi-arad")
	cal.SetName("Gunoi Arad - " + place.DisplayName())
	cal.SetDescription("Program de colectare pentru " + place.DisplayName())

	now := time.Now().UTC()
	for _, item := range events {
		id := fmt.Sprintf("%s-%d-%s@gunoikarad.local", item.Date.Format("20060102"), place.ID, item.WasteType)
		event := cal.AddEvent(id)
		event.SetDtStampTime(now)
		event.SetSummary(item.Label + " - " + place.StreetRaw)
		description := "Scoate recipientele pana la 07:00."
		if item.Location != "" {
			description += " Locatie: " + item.Location + "."
		}
		if item.SourceURL != "" {
			description += " Sursa: " + item.SourceURL
		}
		event.SetDescription(description)
		if publicBaseURL != "" {
			event.SetURL(strings.TrimRight(publicBaseURL, "/") + "/program?place_id=" + fmt.Sprint(place.ID))
		}
		event.SetAllDayStartAt(item.Date)
		event.SetAllDayEndAt(item.Date.AddDate(0, 0, 1))
	}
	return cal.Serialize()
}
