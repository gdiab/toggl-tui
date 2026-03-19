package api

import (
	"fmt"
	"time"
)

// DaySummary holds time entries for a single day with a total.
type DaySummary struct {
	Date    time.Time
	Entries []TimeEntry
	Total   time.Duration
}

// WeekSummary groups a week's time entries by day.
type WeekSummary struct {
	Days  []DaySummary  // Monday through today (or through Friday if week is complete)
	Total time.Duration // Sum of all day totals
}

// GetWeekEntries fetches the current week's time entries (Monday through today)
// grouped by day with per-day and week totals.
func (c *Client) GetWeekEntries() (WeekSummary, error) {
	return c.getWeekEntries(time.Now())
}

// getWeekEntries is the testable implementation that accepts a reference time.
func (c *Client) getWeekEntries(now time.Time) (WeekSummary, error) {
	monday := weekMonday(now)
	endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	startDate := monday.UTC().Format(time.RFC3339)
	endDate := endOfToday.UTC().Format(time.RFC3339)

	entries, err := c.GetTimeEntries(startDate, endDate)
	if err != nil {
		return WeekSummary{}, fmt.Errorf("fetch week entries: %w", err)
	}

	return buildWeekSummary(monday, now, entries)
}

// weekMonday returns midnight on the Monday of the week containing t.
func weekMonday(t time.Time) time.Time {
	weekday := t.Weekday()
	// time.Sunday == 0, time.Monday == 1, ...
	offset := (int(weekday) - int(time.Monday) + 7) % 7
	monday := t.AddDate(0, 0, -offset)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}

// buildWeekSummary groups entries by day from Monday through today.
func buildWeekSummary(monday, now time.Time, entries []TimeEntry) (WeekSummary, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	numDays := int(today.Sub(monday).Hours()/24) + 1

	days := make([]DaySummary, numDays)
	for i := range days {
		days[i].Date = monday.AddDate(0, 0, i)
	}

	for _, e := range entries {
		t, err := time.Parse(time.RFC3339, e.Start)
		if err != nil {
			return WeekSummary{}, fmt.Errorf("parse entry start %q: %w", e.Start, err)
		}
		// Convert to local time for day grouping.
		t = t.In(now.Location())
		entryDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		idx := int(entryDate.Sub(monday).Hours() / 24)
		if idx < 0 || idx >= numDays {
			continue // entry outside the range
		}

		dur := e.Duration
		if dur < 0 {
			// Running timer: compute elapsed from start to now.
			dur = int(now.Sub(t).Seconds())
		}

		days[idx].Entries = append(days[idx].Entries, e)
		days[idx].Total += time.Duration(dur) * time.Second
	}

	var weekTotal time.Duration
	for _, d := range days {
		weekTotal += d.Total
	}

	return WeekSummary{Days: days, Total: weekTotal}, nil
}
