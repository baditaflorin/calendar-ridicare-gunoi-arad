package app

import (
	"time"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
)

type printData struct {
	Place     domain.Place
	Month     time.Time
	MonthName string
	Days      []calendarDay
	Legend    []legendItem
	UpdatedAt string
}

type calendarDay struct {
	Date    time.Time
	Day     int
	InMonth bool
	Events  []domain.Event
}

type legendItem struct {
	Type  domain.WasteType
	Label string
	Color string
}

func buildMonthGrid(month time.Time, events []domain.Event) []calendarDay {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	offset := int(first.Weekday() - time.Monday)
	if offset < 0 {
		offset = 6
	}
	start := first.AddDate(0, 0, -offset)
	eventByDate := map[string][]domain.Event{}
	for _, event := range events {
		eventByDate[event.Date.Format(time.DateOnly)] = append(eventByDate[event.Date.Format(time.DateOnly)], event)
	}
	days := make([]calendarDay, 0, 42)
	for i := 0; i < 42; i++ {
		date := start.AddDate(0, 0, i)
		days = append(days, calendarDay{
			Date:    date,
			Day:     date.Day(),
			InMonth: date.Month() == month.Month(),
			Events:  eventByDate[date.Format(time.DateOnly)],
		})
	}
	return days
}

func legend(events []domain.Event) []legendItem {
	order := []domain.WasteType{
		domain.WastePaper,
		domain.WastePlasticMetal,
		domain.WasteBio,
		domain.WasteResidual,
		domain.WasteBulky,
		domain.WasteTextile,
		domain.WasteHazardous,
	}
	seen := map[domain.WasteType]bool{}
	for _, event := range events {
		seen[event.WasteType] = true
	}
	items := make([]legendItem, 0, len(seen))
	for _, waste := range order {
		if !seen[waste] {
			continue
		}
		items = append(items, legendItem{Type: waste, Label: domain.WasteLabel(waste), Color: domain.WasteColor(waste)})
	}
	return items
}

func latestFetchedAt(events []domain.Event) string {
	for _, event := range events {
		if event.FetchedAt != "" {
			return event.FetchedAt
		}
	}
	return ""
}

func romanianMonth(month time.Month) string {
	switch month {
	case time.January:
		return "Ianuarie"
	case time.February:
		return "Februarie"
	case time.March:
		return "Martie"
	case time.April:
		return "Aprilie"
	case time.May:
		return "Mai"
	case time.June:
		return "Iunie"
	case time.July:
		return "Iulie"
	case time.August:
		return "August"
	case time.September:
		return "Septembrie"
	case time.October:
		return "Octombrie"
	case time.November:
		return "Noiembrie"
	default:
		return "Decembrie"
	}
}
