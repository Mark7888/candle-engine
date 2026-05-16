// Package models defines the JSON request and response types for the REST API.
package models

// UpdateSettingsRequest is the body for PATCH /api/v1/projects/:id/settings.
// Only non-zero fields are applied; omitted fields leave the existing values unchanged.
type UpdateSettingsRequest struct {
	// TimeFrame overrides the project's time frame (e.g. "1h", "4h", "1d").
	TimeFrame string `json:"timeFrame,omitempty"`
	// RangeFrom overrides the range start as an RFC3339 timestamp string.
	RangeFrom string `json:"rangeFrom,omitempty"`
	// RangeTo overrides the range end as an RFC3339 timestamp string.
	RangeTo string `json:"rangeTo,omitempty"`
	// CandleBullishColor overrides the bullish candle color (CSS hex, e.g. "#26a69a").
	CandleBullishColor string `json:"candleBullishColor,omitempty"`
	// CandleBearishColor overrides the bearish candle color (CSS hex, e.g. "#ef5350").
	CandleBearishColor string `json:"candleBearishColor,omitempty"`
}
