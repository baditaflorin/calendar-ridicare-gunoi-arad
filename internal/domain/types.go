package domain

import "time"

type WasteType string

const (
	WasteResidual     WasteType = "residual"
	WasteBio          WasteType = "bio"
	WastePaper        WasteType = "paper"
	WastePlasticMetal WasteType = "plastic_metal"
	WasteGlass        WasteType = "glass"
	WasteBulky        WasteType = "bulky"
	WasteTextile      WasteType = "textile"
	WasteHazardous    WasteType = "hazardous"
)

type Place struct {
	ID          int64    `json:"id"`
	UAT         string   `json:"uat"`
	Cartier     string   `json:"cartier"`
	CartierNorm string   `json:"cartier_norm"`
	StreetRaw   string   `json:"street_raw"`
	StreetNorm  string   `json:"street_norm"`
	Side        string   `json:"side,omitempty"`
	HouseParity string   `json:"house_parity,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
}

func (p Place) DisplayName() string {
	if p.Cartier == "" {
		return "Strada " + p.StreetRaw + ", Arad"
	}
	return "Strada " + p.StreetRaw + ", " + p.Cartier + ", Arad"
}

type Rule struct {
	ID             int64
	PlaceID        int64
	WasteType      WasteType
	RecurrenceKind string
	Weekday        time.Weekday
	WeekLabel      string
	ValidFrom      *time.Time
	ValidTo        *time.Time
	SourceID       int64
	SourceURL      string
	FetchedAt      string
	Confidence     float64
}

type Event struct {
	ID          int64     `json:"id"`
	PlaceID     int64     `json:"place_id,omitempty"`
	WasteType   WasteType `json:"waste_type"`
	Label       string    `json:"label"`
	Color       string    `json:"color"`
	Date        time.Time `json:"-"`
	DateISO     string    `json:"date"`
	StartTime   string    `json:"start_time,omitempty"`
	EndTime     string    `json:"end_time,omitempty"`
	Location    string    `json:"location,omitempty"`
	Title       string    `json:"title,omitempty"`
	Kind        string    `json:"kind"`
	SourceURL   string    `json:"source_url,omitempty"`
	FetchedAt   string    `json:"fetched_at,omitempty"`
	Confidence  float64   `json:"confidence"`
	Generated   bool      `json:"generated"`
	WeekdayName string    `json:"weekday"`
}

type Source struct {
	ID            int64
	URL           string
	SourceType    string
	FetchedAt     time.Time
	ContentHash   string
	RawPath       string
	ParserVersion string
	HTTPStatus    int
}

type SearchResult struct {
	Place Place `json:"place"`
	Score int   `json:"score"`
}

type Neighborhood struct {
	Name  string `json:"name"`
	Norm  string `json:"norm"`
	Count int    `json:"count"`
}

type Counts struct {
	Places      int
	Rules       int
	Events      int
	Sources     int
	ParseIssues int
}

func WasteLabel(w WasteType) string {
	switch w {
	case WasteResidual:
		return "Rezidual"
	case WasteBio:
		return "Bio"
	case WastePaper:
		return "Hartie & Carton"
	case WastePlasticMetal:
		return "Plastic & Metal"
	case WasteGlass:
		return "Sticla"
	case WasteBulky:
		return "Voluminoase"
	case WasteTextile:
		return "Textile"
	case WasteHazardous:
		return "Periculoase"
	default:
		return string(w)
	}
}

func WasteColor(w WasteType) string {
	switch w {
	case WasteResidual:
		return "#374151"
	case WasteBio:
		return "#16a34a"
	case WastePaper:
		return "#2563eb"
	case WastePlasticMetal:
		return "#eab308"
	case WasteGlass:
		return "#16a34a"
	case WasteBulky:
		return "#7c3aed"
	case WasteTextile:
		return "#db2777"
	case WasteHazardous:
		return "#dc2626"
	default:
		return "#64748b"
	}
}

func RomanianWeekday(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "Luni"
	case time.Tuesday:
		return "Marti"
	case time.Wednesday:
		return "Miercuri"
	case time.Thursday:
		return "Joi"
	case time.Friday:
		return "Vineri"
	case time.Saturday:
		return "Sambata"
	default:
		return "Duminica"
	}
}

func DecorateEvent(e Event) Event {
	e.DateISO = e.Date.Format(time.DateOnly)
	e.Label = WasteLabel(e.WasteType)
	e.Color = WasteColor(e.WasteType)
	e.WeekdayName = RomanianWeekday(e.Date.Weekday())
	if e.Confidence == 0 {
		e.Confidence = 1
	}
	if e.Title == "" {
		e.Title = e.Label
	}
	return e
}
