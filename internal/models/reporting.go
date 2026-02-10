package models

// ReportRequest is the request body for reporting endpoints.
type ReportRequest struct {
	StartTime        string   `json:"startTime"`
	EndTime          string   `json:"endTime"`
	Granularity      string   `json:"granularity,omitempty"` // HOURLY, DAILY, WEEKLY, MONTHLY
	GroupBy          []string `json:"groupBy,omitempty"`     // countryOrRegion, deviceClass, ageRange, gender, adminArea, locality
	Selector         *Selector `json:"selector,omitempty"`
	ReturnGrandTotals bool    `json:"returnGrandTotals,omitempty"`
	ReturnRecordsWithNoMetrics bool `json:"returnRecordsWithNoMetrics,omitempty"`
	ReturnRowTotals  bool    `json:"returnRowTotals,omitempty"`
	TimeZone         string  `json:"timeZone,omitempty"`
}

// ReportResponse wraps reporting response data.
type ReportResponse struct {
	ReportingDataResponse ReportingDataResponse `json:"reportingDataResponse"`
}

// ReportingDataResponse contains the actual report rows.
type ReportingDataResponse struct {
	Row        []ReportRow    `json:"row"`
	GrandTotals *ReportRow   `json:"grandTotals,omitempty"`
}

// ReportRow represents a single row in a report.
type ReportRow struct {
	Other    bool                   `json:"other,omitempty"`
	Total    *SpendRow              `json:"total,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Granularity []GranularityRow   `json:"granularity,omitempty"`
	Insights *InsightData           `json:"insights,omitempty"`
}

// SpendRow contains the metrics for a report row.
type SpendRow struct {
	Impressions       int64   `json:"impressions"`
	Taps              int64   `json:"taps"`
	TotalInstalls     int64   `json:"totalInstalls"`
	TapInstalls       int64   `json:"tapInstalls"`
	ViewInstalls      int64   `json:"viewInstalls"`
	TotalNewDownloads int64   `json:"totalNewDownloads"`
	TapNewDownloads   int64   `json:"tapNewDownloads"`
	ViewNewDownloads  int64   `json:"viewNewDownloads"`
	TotalRedownloads  int64   `json:"totalRedownloads"`
	TapRedownloads    int64   `json:"tapRedownloads"`
	ViewRedownloads   int64   `json:"viewRedownloads"`
	TTR               float64 `json:"ttr"`
	TotalInstallRate  float64 `json:"totalInstallRate"`
	TapInstallRate    float64 `json:"tapInstallRate"`
	AvgCPT            Money   `json:"avgCPT"`
	AvgCPM            Money   `json:"avgCPM"`
	TapInstallCPI     Money   `json:"tapInstallCPI"`
	TotalAvgCPI       Money   `json:"totalAvgCPI"`
	LocalSpend        Money   `json:"localSpend"`
}

// GranularityRow is a time-bucketed metrics row.
type GranularityRow struct {
	Date    string    `json:"date"`
	Metrics *SpendRow `json:"metrics,omitempty"`
}

// InsightData contains keyword-level insights.
type InsightData struct {
	BidRecommendation *BidRecommendation `json:"bidRecommendation,omitempty"`
}

// BidRecommendation for keyword bid suggestions.
type BidRecommendation struct {
	SuggestedBidAmount *Money `json:"suggestedBidAmount,omitempty"`
}

// SearchTermReportRow is a row in the search terms report.
type SearchTermReportRow struct {
	Other    bool                   `json:"other,omitempty"`
	Total    *SpendRow              `json:"total,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
