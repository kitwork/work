// KitWork - Work Engine Core
// Copyright (C) 2025 Huỳnh Nhân Quốc

// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License version 3 (AGPL-3.0).

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the AGPL-3.0 License along with this program.
// If not, see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type Schedule struct {
	Type  string `work:"type"`  // cron | daily | hourly | weekly | monthly | every
	Value string `work:"value"` // tuỳ theo type

	Tag     string     `work:"tag"`
	Limit   int        `work:"limit"`
	StartAt *time.Time `work:"start_at"`
	EndAt   *time.Time `work:"end_at"`
}

func (cfg *Config) Schedule(source ...string) *Config {
	cfg.schedule = cfg.source("tasks", source...)
	return cfg
}

func (s *Schedule) Definition() (gocron.JobDefinition, error) {

	switch s.Type {
	case "cron":
		return gocron.CronJob(s.Value, false), nil

	case "daily":
		// Value có thể "HH:MM" hoặc "HH:MM:SS"

		atTime, err := parseAtTime(s.Value)

		if err != nil {
			return nil, err
		}
		return gocron.DailyJob(1, gocron.NewAtTimes(atTime)), nil

	case "weekly":
		// Value có thể "Monday HH:MM" hoặc "Monday HH:MM:SS"
		parts := strings.Split(strings.TrimSpace(s.Value), " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid weekly format: %s", s.Value)
		}

		dayStr, timeStr := parts[0], parts[1]

		atTime, err := parseAtTime(timeStr)
		if err != nil {
			return nil, err
		}
		weekday, err := WeekdayFromString(dayStr)
		if err != nil {
			return nil, err
		}
		return gocron.WeeklyJob(1, gocron.NewWeekdays(weekday), gocron.NewAtTimes(atTime)), nil

	case "monthly":
		// Value có thể "5 10:00" hoặc "5 10:00,15 14:00"
		parts := strings.Split(strings.TrimSpace(s.Value), " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid monthly format: %s", s.Value)
		}
		dayStr, timeStr := atoiOrZero(parts[0]), parts[1]

		atTime, err := parseAtTime(timeStr)
		if err != nil {
			return nil, err
		}
		return gocron.MonthlyJob(1, gocron.NewDaysOfTheMonth(int(dayStr)), gocron.NewAtTimes(atTime)), nil

	case "every":
		d, err := time.ParseDuration(s.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: %s", s.Value)
		}
		return gocron.DurationJob(d), nil

	default:
		return nil, fmt.Errorf("unknown schedule type: %s", s.Type)
	}
}

func job(typing string, value string) (gocron.JobDefinition, error) {
	// value, ok := val.(string)
	// if !ok {
	// 	return nil, fmt.Errorf("value schedule type: %s", typing)
	// }
	switch typing {
	case "cron":
		return gocron.CronJob(value, false), nil

	case "daily":
		// Value có thể "HH:MM" hoặc "HH:MM:SS"

		atTime, err := parseAtTime(value)

		if err != nil {
			return nil, err
		}
		return gocron.DailyJob(1, gocron.NewAtTimes(atTime)), nil

	case "weekly":
		// Value có thể "Monday HH:MM" hoặc "Monday HH:MM:SS"
		parts := strings.Split(strings.TrimSpace(value), " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid weekly format: %s", value)
		}

		dayStr, timeStr := parts[0], parts[1]

		atTime, err := parseAtTime(timeStr)
		if err != nil {
			return nil, err
		}
		weekday, err := WeekdayFromString(dayStr)
		if err != nil {
			return nil, err
		}
		return gocron.WeeklyJob(1, gocron.NewWeekdays(weekday), gocron.NewAtTimes(atTime)), nil

	case "monthly":
		// Value có thể "5 10:00" hoặc "5 10:00,15 14:00"
		parts := strings.Split(strings.TrimSpace(value), " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid monthly format: %s", value)
		}
		dayStr, timeStr := atoiOrZero(parts[0]), parts[1]

		atTime, err := parseAtTime(timeStr)
		if err != nil {
			return nil, err
		}
		return gocron.MonthlyJob(1, gocron.NewDaysOfTheMonth(int(dayStr)), gocron.NewAtTimes(atTime)), nil

	case "every":
		d, err := time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: %s", value)
		}
		return gocron.DurationJob(d), nil

	default:
		return nil, fmt.Errorf("unknown schedule type: %s", typing)
	}
}

func WeekdayFromString(s string) (time.Weekday, error) {
	switch strings.ToLower(s) {
	case "sunday":
		return time.Sunday, nil
	case "monday":
		return time.Monday, nil
	case "tuesday":
		return time.Tuesday, nil
	case "wednesday":
		return time.Wednesday, nil
	case "thursday":
		return time.Thursday, nil
	case "friday":
		return time.Friday, nil
	case "saturday":
		return time.Saturday, nil
	default:
		return time.Sunday, fmt.Errorf("invalid weekday: %s", s)
	}
}

// parseAtTime parses a string like "HH:MM" or "HH:MM:SS" to gocron.AtTime
func parseAtTime(value string) (gocron.AtTime, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return gocron.NewAtTime(00, 00, 00), fmt.Errorf("invalid time format: %s", value)
	}

	hour := atoiOrZero(parts[0])
	minute := atoiOrZero(parts[1])
	second := uint(0)
	if len(parts) == 3 {
		second = atoiOrZero(parts[2])
	}

	return gocron.NewAtTime(hour, minute, second), nil
}

func atoiOrZero(s string) uint {
	i, err := strconv.Atoi(s)
	if err != nil || i < 0 {
		return 0
	}
	return uint(i)
}
