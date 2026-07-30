package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// cronSchedule is a parsed 6-field cron expression (with seconds).
// Field order: second minute hour day-of-month month day-of-week.
// Each field is stored as a bitmask over its valid range.
type cronSchedule struct {
	second  uint64
	minute  uint64
	hour    uint64
	dom     uint64
	month   uint64
	dow     uint64
	domStar bool
	dowStar bool
}

type fieldRange struct {
	min int
	max int
}

// parseCron parses a 6-field cron expression (second minute hour dom month dow).
// Supported syntax per field: "*", "*/n", "a", "a-b", "a-b/n", and comma lists.
func parseCron(spec string) (*cronSchedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return nil, fmt.Errorf("invalid cron spec %q: expected 6 fields, got %d", spec, len(fields))
	}

	ranges := []fieldRange{
		{0, 59}, // second
		{0, 59}, // minute
		{0, 23}, // hour
		{1, 31}, // day of month
		{1, 12}, // month
		{0, 7},  // day of week (0 and 7 are Sunday)
	}

	masks := make([]uint64, 6)
	stars := make([]bool, 6)
	for i, f := range fields {
		mask, star, err := parseField(f, ranges[i].min, ranges[i].max)
		if err != nil {
			return nil, fmt.Errorf("invalid cron field %d (%q) in %q: %w", i, f, spec, err)
		}
		masks[i] = mask
		stars[i] = star
	}

	// Normalize day-of-week: treat 7 as Sunday (0).
	dow := masks[5]
	if dow&(1<<7) != 0 {
		dow = (dow &^ (1 << 7)) | (1 << 0)
	}

	return &cronSchedule{
		second:  masks[0],
		minute:  masks[1],
		hour:    masks[2],
		dom:     masks[3],
		month:   masks[4],
		dow:     dow,
		domStar: stars[3],
		dowStar: stars[5],
	}, nil
}

// parseField parses a single cron field into a bitmask. The returned bool
// reports whether the field was an unrestricted "*" (or "*/n").
func parseField(field string, min, max int) (uint64, bool, error) {
	if field == "" {
		return 0, false, fmt.Errorf("empty field")
	}

	var mask uint64
	star := false

	for _, part := range strings.Split(field, ",") {
		rangePart := part
		step := 1

		if idx := strings.Index(part, "/"); idx >= 0 {
			stepStr := part[idx+1:]
			rangePart = part[:idx]
			s, err := strconv.Atoi(stepStr)
			if err != nil || s <= 0 {
				return 0, false, fmt.Errorf("invalid step %q", stepStr)
			}
			step = s
		}

		var lo, hi int
		switch {
		case rangePart == "*":
			lo, hi = min, max
			if step == 1 {
				star = true
			}
		case strings.Contains(rangePart, "-"):
			bounds := strings.SplitN(rangePart, "-", 2)
			a, err := strconv.Atoi(bounds[0])
			if err != nil {
				return 0, false, fmt.Errorf("invalid range start %q", bounds[0])
			}
			b, err := strconv.Atoi(bounds[1])
			if err != nil {
				return 0, false, fmt.Errorf("invalid range end %q", bounds[1])
			}
			lo, hi = a, b
		default:
			v, err := strconv.Atoi(rangePart)
			if err != nil {
				return 0, false, fmt.Errorf("invalid value %q", rangePart)
			}
			lo, hi = v, v
		}

		if lo < min || hi > max || lo > hi {
			return 0, false, fmt.Errorf("value out of range [%d-%d]", min, max)
		}

		for v := lo; v <= hi; v += step {
			mask |= 1 << uint(v)
		}
	}

	if mask == 0 {
		return 0, false, fmt.Errorf("no values matched")
	}

	return mask, star, nil
}

// match reports whether the given time satisfies the schedule (second granularity).
func (c *cronSchedule) match(t time.Time) bool {
	if c.second&(1<<uint(t.Second())) == 0 {
		return false
	}
	if c.minute&(1<<uint(t.Minute())) == 0 {
		return false
	}
	if c.hour&(1<<uint(t.Hour())) == 0 {
		return false
	}
	if c.month&(1<<uint(int(t.Month()))) == 0 {
		return false
	}
	return c.dayMatch(t)
}

// dayMatch applies standard cron day-of-month / day-of-week semantics:
// when both fields are restricted, a match on either is sufficient;
// otherwise both must match (a "*" field always matches).
func (c *cronSchedule) dayMatch(t time.Time) bool {
	domMatch := c.dom&(1<<uint(t.Day())) != 0
	dowMatch := c.dow&(1<<uint(int(t.Weekday()))) != 0

	if !c.domStar && !c.dowStar {
		return domMatch || dowMatch
	}
	return domMatch && dowMatch
}

// next returns the earliest time strictly after "from" that matches the schedule.
// It advances field-by-field to avoid scanning every second across large gaps.
func (c *cronSchedule) next(from time.Time) time.Time {
	t := from.Truncate(time.Second).Add(time.Second)
	yearLimit := t.Year() + 5

	for {
		if t.Year() > yearLimit {
			return time.Time{}
		}
		if c.month&(1<<uint(int(t.Month()))) == 0 {
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0)
			continue
		}
		if !c.dayMatch(t) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, 1)
			continue
		}
		if c.hour&(1<<uint(t.Hour())) == 0 {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location()).Add(time.Hour)
			continue
		}
		if c.minute&(1<<uint(t.Minute())) == 0 {
			t = t.Truncate(time.Minute).Add(time.Minute)
			continue
		}
		if c.second&(1<<uint(t.Second())) == 0 {
			t = t.Add(time.Second)
			continue
		}
		return t
	}
}
