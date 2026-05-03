package etl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/search"
	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/store"
)

const programYear = 2026

var dayPattern = regexp.MustCompile(`\d{1,2}`)

func ParseZona1(sourceID int64, html string, b *builder) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return err
	}
	parseWeeklyRules(sourceID, doc, b)
	parseExactMonthTable(sourceID, doc, b, "table#tablepress-10", domain.WastePaper)
	parseExactMonthTable(sourceID, doc, b, "table#tablepress-62", domain.WastePlasticMetal)
	return nil
}

func parseWeeklyRules(sourceID int64, doc *goquery.Document, b *builder) {
	doc.Find("table#tablepress-56 tbody tr").Each(func(_ int, row *goquery.Selection) {
		cells := cells(row)
		if len(cells) < 5 || allEmpty(cells) {
			return
		}
		placeKey := b.addPlace(sourceID, cells[0], cells[1])
		if placeKey == "" {
			b.addIssue(sourceID, "warning", strings.Join(cells, " | "), "weekly rule row has empty street")
			return
		}
		for _, item := range []struct {
			waste domain.WasteType
			day   string
		}{
			{domain.WasteResidual, cells[2]},
			{domain.WasteBio, cells[3]},
		} {
			weekday, ok := parseWeekday(item.day)
			if !ok {
				b.addIssue(sourceID, "warning", strings.Join(cells, " | "), "unknown weekday "+item.day)
				continue
			}
			b.data.Rules = append(b.data.Rules, store.RuleRecord{
				PlaceKey:       placeKey,
				WasteType:      string(item.waste),
				RecurrenceKind: "weekly",
				Weekday:        weekday,
				SourceID:       sourceID,
			})
		}
	})
}

func parseExactMonthTable(sourceID int64, doc *goquery.Document, b *builder, selector string, waste domain.WasteType) {
	doc.Find(selector + " tbody tr").Each(func(_ int, row *goquery.Selection) {
		cells := cells(row)
		if len(cells) < 16 || allEmpty(cells) {
			return
		}
		placeKey := b.addPlace(sourceID, cells[0], cells[1])
		if placeKey == "" {
			b.addIssue(sourceID, "warning", strings.Join(cells, " | "), "exact date row has empty street")
			return
		}
		weekLabel := cells[2]
		weekday := cells[3]
		for monthIndex := 1; monthIndex <= 12; monthIndex++ {
			for _, day := range parseMonthDays(cells[3+monthIndex]) {
				eventDate := time.Date(programYear, time.Month(monthIndex), day, 0, 0, 0, 0, time.UTC)
				if eventDate.Month() != time.Month(monthIndex) {
					b.addIssue(sourceID, "warning", strings.Join(cells, " | "), fmt.Sprintf("invalid day %d for month %d", day, monthIndex))
					continue
				}
				title := domain.WasteLabel(waste)
				if weekLabel != "" || weekday != "" {
					title = compact(title + " " + weekLabel + " " + weekday)
				}
				b.data.Events = append(b.data.Events, store.EventRecord{
					PlaceKey:    placeKey,
					CartierNorm: search.Normalize(cleanCartier(cells[0])),
					WasteType:   string(waste),
					EventDate:   eventDate.Format(time.DateOnly),
					StartTime:   "07:00",
					Title:       title,
					Kind:        "exact",
					SourceID:    sourceID,
					Confidence:  1,
				})
			}
		}
	})
}

func ParseCampaign1(sourceID int64, html string, b *builder) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return err
	}
	parsePointCampaign(sourceID, doc, b, "table#tablepress-65", domain.WasteTextile, "Campania 1 textile")
	parsePointCampaign(sourceID, doc, b, "table#tablepress-66", domain.WasteHazardous, "Campania 1 periculoase")
	parseBulky(sourceID, doc, b)
	return nil
}

func parsePointCampaign(sourceID int64, doc *goquery.Document, b *builder, selector string, waste domain.WasteType, campaignName string) {
	doc.Find(selector + " tbody tr").Each(func(_ int, row *goquery.Selection) {
		cells := cells(row)
		if len(cells) < 4 || allEmpty(cells) {
			return
		}
		start, end, ok := parseDateRange(cells[1])
		if !ok {
			b.addIssue(sourceID, "warning", strings.Join(cells, " | "), "campaign range could not be parsed")
			return
		}
		startTime, endTime := parseTimeInterval(cells[3])
		cartier := cleanCartier(cells[0])
		b.data.Campaigns = append(b.data.Campaigns, store.CampaignRecord{
			Name:         campaignName,
			YearLabel:    strconv.Itoa(start.Year()),
			StartDate:    start.Format(time.DateOnly),
			EndDate:      end.Format(time.DateOnly),
			WasteType:    string(waste),
			LocationType: "fixed_point",
			SourceID:     sourceID,
		})
		for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
			if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
				continue
			}
			b.data.Events = append(b.data.Events, store.EventRecord{
				CartierNorm: search.Normalize(cartier),
				WasteType:   string(waste),
				EventDate:   current.Format(time.DateOnly),
				StartTime:   startTime,
				EndTime:     endTime,
				Location:    compact(cells[2]),
				Title:       domain.WasteLabel(waste) + " - punct fix",
				Kind:        "campaign",
				SourceID:    sourceID,
				Confidence:  1,
			})
		}
	})
}

func parseBulky(sourceID int64, doc *goquery.Document, b *builder) {
	doc.Find("table#tablepress-67 tbody tr").Each(func(_ int, row *goquery.Selection) {
		cells := cells(row)
		if len(cells) < 3 || allEmpty(cells) {
			return
		}
		date, ok := parseRomanianLongDate(cells[2])
		if !ok {
			b.addIssue(sourceID, "warning", strings.Join(cells, " | "), "bulky date could not be parsed")
			return
		}
		placeKey := b.addPlace(sourceID, cells[0], cells[1])
		if placeKey == "" {
			b.addIssue(sourceID, "warning", strings.Join(cells, " | "), "bulky row has empty street")
			return
		}
		b.data.Events = append(b.data.Events, store.EventRecord{
			PlaceKey:   placeKey,
			WasteType:  string(domain.WasteBulky),
			EventDate:  date.Format(time.DateOnly),
			Title:      domain.WasteLabel(domain.WasteBulky),
			Kind:       "campaign",
			SourceID:   sourceID,
			Confidence: 1,
		})
	})
}

func cells(row *goquery.Selection) []string {
	out := make([]string, 0, 16)
	row.Find("th,td").Each(func(_ int, cell *goquery.Selection) {
		out = append(out, compact(cell.Text()))
	})
	return out
}

func allEmpty(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func parseMonthDays(cell string) []int {
	matches := dayPattern.FindAllString(cell, -1)
	days := make([]int, 0, len(matches))
	for _, match := range matches {
		day, err := strconv.Atoi(match)
		if err == nil && day > 0 {
			days = append(days, day)
		}
	}
	return days
}

func parseWeekday(value string) (time.Weekday, bool) {
	switch search.Normalize(value) {
	case "luni":
		return time.Monday, true
	case "marti":
		return time.Tuesday, true
	case "miercuri":
		return time.Wednesday, true
	case "joi":
		return time.Thursday, true
	case "vineri":
		return time.Friday, true
	case "sambata":
		return time.Saturday, true
	case "duminica":
		return time.Sunday, true
	default:
		return time.Sunday, false
	}
}

var rangePattern = regexp.MustCompile(`(?i)(\d{1,2})\.(\d{1,2})-(\d{1,2})\.(\d{1,2})\.(\d{4})`)

func parseDateRange(value string) (time.Time, time.Time, bool) {
	match := rangePattern.FindStringSubmatch(value)
	if len(match) != 6 {
		return time.Time{}, time.Time{}, false
	}
	startDay, _ := strconv.Atoi(match[1])
	startMonth, _ := strconv.Atoi(match[2])
	endDay, _ := strconv.Atoi(match[3])
	endMonth, _ := strconv.Atoi(match[4])
	year, _ := strconv.Atoi(match[5])
	start := time.Date(year, time.Month(startMonth), startDay, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.Month(endMonth), endDay, 0, 0, 0, 0, time.UTC)
	return start, end, start.Month() == time.Month(startMonth) && end.Month() == time.Month(endMonth) && !end.Before(start)
}

var intervalPattern = regexp.MustCompile(`(?i)(\d{1,2}:\d{2})\s*-\s*(\d{1,2}:\d{2})`)

func parseTimeInterval(value string) (string, string) {
	match := intervalPattern.FindStringSubmatch(value)
	if len(match) != 3 {
		return "", ""
	}
	return match[1], match[2]
}

var romanianDatePattern = regexp.MustCompile(`(?i)(\d{1,2})\s+([a-zăâîșşțţ]+)\s+(\d{4})`)

func parseRomanianLongDate(value string) (time.Time, bool) {
	match := romanianDatePattern.FindStringSubmatch(strings.ToLower(value))
	if len(match) != 4 {
		return time.Time{}, false
	}
	day, _ := strconv.Atoi(match[1])
	year, _ := strconv.Atoi(match[3])
	month, ok := romanianMonths[search.Normalize(match[2])]
	if !ok {
		return time.Time{}, false
	}
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return date, date.Day() == day
}

var romanianMonths = map[string]time.Month{
	"ianuarie":   time.January,
	"februarie":  time.February,
	"martie":     time.March,
	"aprilie":    time.April,
	"mai":        time.May,
	"iunie":      time.June,
	"iulie":      time.July,
	"august":     time.August,
	"septembrie": time.September,
	"octombrie":  time.October,
	"noiembrie":  time.November,
	"decembrie":  time.December,
}
