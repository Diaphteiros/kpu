package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	rawTimeSubstringRegex = `(?P<time>\d+)(?P<unit>[smhdwMy]?)`
)

var (
	TimeStringRegex    = regexp.MustCompile(fmt.Sprintf(`^(%s)+$`, rawTimeSubstringRegex))
	TimeSubstringRegex = regexp.MustCompile(rawTimeSubstringRegex)

	TimeSubstringRegex_TimeIndex = -1
	TimeSubstringRegex_UnitIndex = -1
)

func init() {
	for i, name := range TimeSubstringRegex.SubexpNames() {
		switch name {
		case "time":
			TimeSubstringRegex_TimeIndex = i
		case "unit":
			TimeSubstringRegex_UnitIndex = i
		}
	}
	if TimeSubstringRegex_TimeIndex == -1 || TimeSubstringRegex_UnitIndex == -1 {
		panic("could not find capture groups in LabelKeyExistsRegex")
	}
}

// StringToDuration parses a string into a time.Duration.
// Accepted time units are:
// - s: seconds
// - m: minutes
// - h: hours
// - d: days (24 hours)
// - w: weeks (7 days)
// - M: months (30 days)
// - y: years (365 days)
func StringToDuration(s string) (time.Duration, error) {
	if !TimeStringRegex.MatchString(s) {
		return 0, fmt.Errorf("invalid duration string: %s", s)
	}
	matches := TimeSubstringRegex.FindAllStringSubmatch(s, -1)

	var d time.Duration
	seenUnits := sets.New[string]()
	for _, match := range matches {
		timeStr := match[TimeSubstringRegex_TimeIndex]
		unitStr := match[TimeSubstringRegex_UnitIndex]

		if unitStr == "" {
			unitStr = "s"
		}
		if seenUnits.Has(unitStr) {
			return 0, fmt.Errorf("duplicate unit '%s' in duration string: %s", unitStr, s)
		}
		seenUnits.Insert(unitStr)

		timeInt, err := strconv.Atoi(timeStr)
		if err != nil {
			return 0, fmt.Errorf("could not parse duration amount into int: %s", s)
		}

		var unit time.Duration
		switch unitStr {
		case "s":
			unit = time.Second
		case "m":
			unit = time.Minute
		case "h":
			unit = time.Hour
		case "d":
			unit = 24 * time.Hour
		case "w":
			unit = 7 * 24 * time.Hour
		case "M":
			unit = 30 * 24 * time.Hour
		case "y":
			unit = 365 * 24 * time.Hour
		default:
			return 0, fmt.Errorf("invalid unit in duration string: %s", s)
		}

		d += time.Duration(timeInt) * unit
	}

	return d, nil
}

// FormatDuration is the inverse to StringToDuration and formats a time.Duration into a human-readable string.
// Note that it neither prints months nor weeks, because a year is defined based on days and the corresponding number
// of weeks or months does not add up to the defined 365 days per year.
func FormatDuration(d time.Duration) string {
	durations := map[string]int{}
	durations["years"] = int(d.Hours() / (24 * 365))
	d -= time.Duration(durations["years"]) * 24 * 365 * time.Hour
	durations["days"] = int(d.Hours() / 24)
	d -= time.Duration(durations["days"]) * 24 * time.Hour
	durations["hours"] = int(d.Hours())
	d -= time.Duration(durations["hours"]) * time.Hour
	durations["minutes"] = int(d.Minutes())
	d -= time.Duration(durations["minutes"]) * time.Minute
	durations["seconds"] = int(d.Seconds())

	sb := strings.Builder{}
	for _, u := range []string{"years", "days", "hours", "minutes", "seconds"} {
		if durations[u] > 0 {
			fmt.Fprint(&sb, durations[u], u[:1])
		}
	}
	if sb.Len() == 0 {
		return "0s"
	}
	return sb.String()
}
