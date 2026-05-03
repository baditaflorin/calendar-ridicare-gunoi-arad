package store

import "time"

type PlaceRecord struct {
	Key         string
	UAT         string
	Cartier     string
	CartierNorm string
	StreetRaw   string
	StreetNorm  string
	Side        string
	HouseParity string
	AliasesJSON string
	SourceID    int64
}

type RuleRecord struct {
	PlaceKey       string
	WasteType      string
	RecurrenceKind string
	Weekday        time.Weekday
	WeekLabel      string
	ValidFrom      string
	ValidTo        string
	SourceID       int64
}

type EventRecord struct {
	PlaceKey    string
	CartierNorm string
	WasteType   string
	EventDate   string
	StartTime   string
	EndTime     string
	Location    string
	Title       string
	Kind        string
	SourceID    int64
	Confidence  float64
}

type CampaignRecord struct {
	Name         string
	YearLabel    string
	StartDate    string
	EndDate      string
	WasteType    string
	LocationType string
	SourceID     int64
}

type IssueRecord struct {
	SourceID int64
	Severity string
	RowText  string
	Reason   string
}

type ImportData struct {
	Places    []PlaceRecord
	Rules     []RuleRecord
	Events    []EventRecord
	Campaigns []CampaignRecord
	Issues    []IssueRecord
}
