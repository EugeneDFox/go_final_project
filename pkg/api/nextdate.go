package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// nextDayHandler handles GET /api/nextdate requests.
// Query parameters:
//   - now: reference date in "20060102" format (optional, defaults to current date)
//   - date: start date in "20060102" format
//   - repeat: repetition rule (e.g., "d 3", "y", "w 1,3,5", "m 1,15 1,3,5")
//
// Returns the next occurrence date after 'now' according to the repetition rule,
// formatted as "20060102". If the rule is invalid or parameters are malformed,
// returns an HTTP 400 error with a descriptive message.
func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	now := r.FormValue("now")
	if now == "" {
		now = time.Now().Format("20060102")
	}
	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(w, "неверный формат now", http.StatusBadRequest)
		return
	}
	dstart := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nextDate, err := NextDate(nowTime, dstart, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

// afterNow returns true if 'date' is strictly after 'now' when comparing only the calendar day
// (ignoring time components). Both dates are converted to UTC and truncated to midnight
// before comparison.
func afterNow(date, now time.Time) bool {
	// Convert both to UTC to ensure consistent date comparison
	dateUTC := date.UTC()
	nowUTC := now.UTC()
	// Truncate to beginning of day in UTC
	return dateUTC.Truncate(24 * time.Hour).After(nowUTC.Truncate(24 * time.Hour))
}

// parseMonth parses a monthly repetition rule of the form "days [months]".
// Example: "1,15 1,3,5" means days 1 and 15 of months January, March, and May.
// Special day values: -1 = last day of month, -2 = second‑last day of month.
// The function populates:
//   - dayFlags: boolean array where dayFlags[d] is true if day d (1‑31) is selected
//   - monthFlags: boolean array where monthFlags[m] is true if month m (1‑12) is selected
//   - hasLastDay, hasSecondLastDay: flags for the special day values
//
// If the months part is omitted, all months (1‑12) are selected.
func parseMonth(rep []string, dayFlags *[32]bool, monthFlags *[13]bool, hasLastDay *bool, hasSecondLastDay *bool) error {
	*hasLastDay = false
	*hasSecondLastDay = false
	for i := range dayFlags {
		dayFlags[i] = false
	}
	for i := range monthFlags {
		monthFlags[i] = false
	}

	dayParts := strings.Split(rep[0], ",")
	for _, part := range dayParts {
		day, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || day < -2 || day == 0 || day > 31 {
			return errors.New("неверный формат дня месяца")
		}
		if day > 0 {
			dayFlags[day] = true
		} else if day == -1 {
			*hasLastDay = true
		} else if day == -2 {
			*hasSecondLastDay = true
		}
	}

	if len(rep) == 2 {
		monthParts := strings.Split(rep[1], ",")
		for _, part := range monthParts {
			month, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || month < 1 || month > 12 {
				return errors.New("неверный формат месяца")
			}
			monthFlags[month] = true
		}
	} else {
		for m := 1; m <= 12; m++ {
			monthFlags[m] = true
		}
	}

	return nil
}

// daysInMonth returns the number of days in the given month of the given year.
func daysInMonth(year int, month time.Month) int {
	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear = year + 1
	}
	firstOfNext := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, time.UTC)
	lastOfCurrent := firstOfNext.AddDate(0, 0, -1)
	_, _, day := lastOfCurrent.Date()
	return day
}

// findNextDate finds the next date that matches the monthly rule defined by
// dayFlags, monthFlags, hasLastDay, and hasSecondLastDay.
// It starts from 'date' and iterates day by day until a date is found that:
//   - is after 'now' (according to afterNow)
//   - falls in a selected month (monthFlags)
//   - matches a selected day (positive day, last day, or second‑last day)
func findNextDate(date, now time.Time, dayFlags [32]bool, monthFlags [13]bool, hasLastDay, hasSecondLastDay bool) time.Time {
	current := date
	for !afterNow(current, now) {
		current = current.AddDate(0, 0, 1)
	}
	for {
		year, month, day := current.Date()
		monthNum := int(month)
		if !monthFlags[monthNum] {
			current = current.AddDate(0, 0, 1)
			continue
		}
		lastDay := daysInMonth(year, month)
		isMatch := false
		if day > 0 && day <= lastDay && dayFlags[day] {
			isMatch = true
		}
		if hasLastDay && day == lastDay {
			isMatch = true
		}
		if hasSecondLastDay && day == lastDay-1 {
			isMatch = true
		}
		if isMatch {
			return current
		}
		current = current.AddDate(0, 0, 1)
	}
}

// parseWeekDays parses a comma‑separated list of weekdays (1=Monday … 7=Sunday)
// and sets the corresponding flags in weekdayFlags[1..7].
func parseWeekDays(input string, weekdayFlags *[8]bool) error {
	for i := range weekdayFlags {
		weekdayFlags[i] = false
	}
	parts := strings.Split(input, ",")
	for _, part := range parts {
		day, err := strconv.Atoi(part)
		if err != nil || day < 1 || day > 7 {
			return errors.New("день недели должен быть от 1 до 7")
		}
		weekdayFlags[day] = true
	}
	return nil
}

// findNextWeekday finds the next date that falls on one of the selected weekdays.
// It starts from 'date' and iterates day by day until a date is found that:
//   - is after 'now' (according to afterNow)
//   - has a weekday (Monday=1 … Sunday=7) that is set in weekdayFlags
func findNextWeekday(date, now time.Time, weekdayFlags [8]bool) time.Time {
	current := date
	for !afterNow(current, now) {
		current = current.AddDate(0, 0, 1)
	}
	for {
		goWeekday := int(current.Weekday())
		if goWeekday == 0 { // воскресенье → 7
			goWeekday = 7
		}
		if weekdayFlags[goWeekday] {
			return current
		}
		current = current.AddDate(0, 0, 1)
	}
}

// NextDate calculates the next occurrence date based on a start date and a repetition rule.
// Parameters:
//   - now: reference date (the next occurrence must be after this date)
//   - dstart: start date in "20060102" format
//   - repeat: repetition rule string (see below)
//
// Supported rules:
//   - "d <interval>"   – every <interval> days (1‑400)
//   - "y"              – yearly (same month/day)
//   - "w <weekdays>"   – weekly on given weekdays (comma‑separated, 1‑7)
//   - "m <days> [months]" – monthly on given days (special values -1, -2) and optionally months
//
// Returns the next date formatted as "20060102", or an error if the rule is invalid.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse("20060102", dstart)
	if err != nil {
		return "", err
	}
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}
	rep := strings.Split(repeat, " ")
	switch rep[0] {
	case "d":
		if len(rep) != 2 {
			return "", errors.New("неверный формат правила d: ожидается d <интервал>")
		}
		interval, err := strconv.Atoi(rep[1])
		if err != nil || interval <= 0 || interval > 400 {
			return "", errors.New("интервал должен быть от 1 до 400")
		}
		for {
			date = date.AddDate(0, 0, interval)
			if afterNow(date, now) {
				break
			}
		}
	case "y":
		if len(rep) != 1 {
			return "", errors.New("неверный формат правила y: ожидается y")
		}
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				break
			}
		}
	case "w":
		if len(rep) != 2 {
			return "", errors.New("неверный формат правила w: ожидается w <дни>")
		}
		var weekdayFlags [8]bool
		err := parseWeekDays(rep[1], &weekdayFlags)
		if err != nil {
			return "", err
		}
		date = findNextWeekday(date, now, weekdayFlags)
	case "m":
		if len(rep) < 2 || len(rep) > 3 {
			return "", errors.New("неверный формат правила m: ожидается m <дни> [месяцы]")
		}
		var dayFlags [32]bool
		var monthFlags [13]bool
		var hasLastDay, hasSecondLastDay bool
		err := parseMonth(rep[1:], &dayFlags, &monthFlags, &hasLastDay, &hasSecondLastDay)
		if err != nil {
			return "", err
		}
		date = findNextDate(date, now, dayFlags, monthFlags, hasLastDay, hasSecondLastDay)
	default:
		return "", errors.New("неизвестное правило повторения")
	}
	return date.Format("20060102"), nil
}
