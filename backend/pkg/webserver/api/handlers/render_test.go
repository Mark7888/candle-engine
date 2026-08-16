package handlers

import (
	"testing"
	"time"

	"candle-engine/pkg/webserver/api/models"
)

func at(minute int) *time.Time {
	t := time.Date(2024, 1, 1, 0, minute, 0, 0, time.UTC)
	return &t
}

// TestSortLogsByTime_InterleavesIndicators covers the log panel showing one
// indicator's lines in a block at the end instead of merged into a single
// timeline: logs are collected per indicator, so they arrive grouped.
func TestSortLogsByTime_InterleavesIndicators(t *testing.T) {
	logs := []models.LogEntry{
		{Indicator: "KAMA", Time: at(1), Message: "kama 1"},
		{Indicator: "KAMA", Time: at(3), Message: "kama 3"},
		{Indicator: "RSI", Time: at(2), Message: "rsi 2"},
		{Indicator: "RSI", Time: at(4), Message: "rsi 4"},
	}

	sortLogsByTime(logs)

	want := []string{"kama 1", "rsi 2", "kama 3", "rsi 4"}
	for i, msg := range want {
		if logs[i].Message != msg {
			t.Errorf("position %d: expected %q, got %q", i, msg, logs[i].Message)
		}
	}
}

func TestSortLogsByTime_UntimedEntriesComeFirst(t *testing.T) {
	logs := []models.LogEntry{
		{Indicator: "RSI", Time: at(2), Message: "rsi"},
		{Indicator: "ProjectErrors", Message: "boom"},
		{Indicator: "KAMA", Time: at(1), Message: "kama"},
	}

	sortLogsByTime(logs)

	if logs[0].Message != "boom" {
		t.Errorf("untimed entries should sort first, got %q", logs[0].Message)
	}
	if logs[1].Message != "kama" || logs[2].Message != "rsi" {
		t.Errorf("timed entries out of order: %q, %q", logs[1].Message, logs[2].Message)
	}
}

func TestSortLogsByTime_IsStableForEqualTimes(t *testing.T) {
	logs := []models.LogEntry{
		{Indicator: "A", Time: at(1), Message: "first"},
		{Indicator: "B", Time: at(1), Message: "second"},
		{Indicator: "C", Time: at(1), Message: "third"},
	}

	sortLogsByTime(logs)

	want := []string{"first", "second", "third"}
	for i, msg := range want {
		if logs[i].Message != msg {
			t.Errorf("position %d: expected %q, got %q", i, msg, logs[i].Message)
		}
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string][]int{"volume": nil, "": nil, "rsi": nil}

	got := sortedKeys(m)

	want := []string{"", "rsi", "volume"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
