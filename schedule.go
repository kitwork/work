// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func (cfg *Config) Schedule(source ...string) *Config {
	cfg.schedule = NewSource("tasks", source...)
	return cfg
}

var (
	DurationRegex = regexp.MustCompile(`^\d+(ns|us|µs|ms|s|m|h)$`)
	TimeRegex     = regexp.MustCompile(`^\d{1,2}:\d{2}(:\d{2})?$`)
	WeeklyRegex   = regexp.MustCompile(`^(Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday)\s\d{1,2}:\d{2}(:\d{2})?$`)
	MonthlyRegex  = regexp.MustCompile(`^\d{1,2}\s\d{1,2}:\d{2}(:\d{2})?$`)
	CronRegex     = regexp.MustCompile(`^(\S+\s){4,5}\S+$`)
)

type Schedule struct {
	Duration *Durationly    `json:"duration,omitempty"`
	Daily    *Daily         `json:"daily,omitempty"`
	Weekly   *Weekly        `json:"weekly,omitempty"`
	Monthly  *Monthly       `json:"monthly,omitempty"`
	Cron     string         `json:"cron,omitempty"` // giữ nguyên string
	Once     *At            `json:"once,omitempty"`
	Fixed    *time.Duration `json:"fixed,omitempty"`
	After    *time.Duration `json:"after,omitempty"`
	Limit    *uint          `json:"limit,omitempty"`
}

type Durationly struct {
	Every   time.Duration `json:"every"`              // ví dụ 5s, 1m, 1h
	StartAt *At           `json:"start_at,omitempty"` // optional
}

func (d *Durationly) Job() gocron.JobDefinition {
	return gocron.DurationJob(d.Every)
}

type Daily struct {
	Interval uint `json:"interval"`           // số ngày: 1 = mỗi ngày
	Ats      []At `json:"at_times,omitempty"` // nhiều AtTime
}

func (d *Daily) AtTimes() gocron.AtTimes {
	return AtTimes(d.Ats)
}

func (d *Daily) Job() gocron.JobDefinition {
	if d.Interval == 0 {
		d.Interval = 1
	}
	return gocron.DailyJob(d.Interval, d.AtTimes())
}

type Weekly struct {
	Interval uint           `json:"interval"` // số tuần
	Weekdays []time.Weekday `json:"weekdays"` // nhiều thứ trong tuần
	Ats      []At           `json:"at_times,omitempty"`
}

func (w *Weekly) AtTimes() gocron.AtTimes {
	return AtTimes(w.Ats)
}

func (w *Weekly) DaysOfTheWeek() gocron.Weekdays {
	if len(w.Weekdays) == 0 {
		return nil
	}
	// gocron.NewWeekdays yêu cầu ít nhất 1 argument
	return gocron.NewWeekdays(w.Weekdays[0], w.Weekdays[1:]...)
}

func (d *Weekly) Job() gocron.JobDefinition {
	if d.Interval == 0 {
		d.Interval = 1
	}
	return gocron.WeeklyJob(d.Interval, d.DaysOfTheWeek(), d.AtTimes())
}

type Monthly struct {
	Interval uint  `json:"interval"` // mỗi N tháng
	Days     []Day `json:"days"`     // 1,2,3… hoặc -1 cho ngày cuối
	Ats      []At  `json:"at_times,omitempty"`
}

func (month Monthly) DaysOfTheMonth() gocron.DaysOfTheMonth {
	if len(month.Days) == 0 {
		return gocron.NewDaysOfTheMonth(1) // mặc định ngày 1 nếu không có
	}

	// Lấy ngày đầu tiên
	first := int(month.Days[0])
	// Các ngày còn lại
	moreDays := make([]int, len(month.Days)-1)
	for i, d := range month.Days[1:] {
		moreDays[i] = int(d)
	}

	return gocron.NewDaysOfTheMonth(first, moreDays...)
}

type Day int

func (day Day) OfTheMonth() gocron.DaysOfTheMonth {
	return gocron.NewDaysOfTheMonth(int(day))
}

type At struct {
	Hour    uint
	Minutes uint
	Seconds uint
}

func (at *At) Time() gocron.AtTime {
	return gocron.NewAtTime(at.Hour, at.Minutes, at.Seconds)
}

func AtTimes(ats []At) gocron.AtTimes {

	if len(ats) == 0 {
		return nil
	}

	first := ats[0].Time()
	if len(ats) == 1 {
		return gocron.NewAtTimes(first)
	}

	more := make([]gocron.AtTime, len(ats)-1)
	for i, at := range ats[1:] {
		more[i] = at.Time()
	}

	return gocron.NewAtTimes(first, more...)
}

func NewScheduler(scheduler map[string]interface{}) (jobs []gocron.JobDefinition, err error) {

	schedule := &Schedule{}

	for key, value := range scheduler {
		switch key {
		case "schedules", "schedule":
			switch v := value.(type) {
			case string:
				// tự động nhận biết string: "5s", "15:00", "Monday 10:30", v.v
				if err := schedule.smart("", v); err != nil {
					return nil, err
				}

			case map[string]interface{}:
				// mỗi key con: daily, weekly, monthly, every, cron
				for k, val := range v {
					if err := schedule.parseing(k, val); err != nil {
						return nil, fmt.Errorf("key %s: %w", k, err)
					}
				}

			case []interface{}:
				for i, item := range v {
					switch x := item.(type) {
					case string:
						if err := schedule.smart("", x); err != nil {
							return nil, fmt.Errorf("item %d: %w", i, err)
						}
					case map[string]interface{}:
						for k, val := range x {
							if err := schedule.parseing(k, val); err != nil {
								return nil, fmt.Errorf("item %d key %s: %w", i, k, err)
							}
						}
					default:
						return nil, fmt.Errorf("item %d unsupported type %T", i, x)
					}
				}

			default:
				return nil, fmt.Errorf("unsupported type %T for schedules", value)
			}

		case "daily", "weekly", "monthly", "every":
			// trực tiếp daily / weekly / monthly / every
			if err := schedule.parseing(key, value); err != nil {
				return nil, err
			}
		default:

		}

	}
	// schedule.Logging()

	// --- Áp dụng schedule vào gocron ---
	// Daily
	if schedule.Daily != nil {
		jobs = append(jobs, schedule.Daily.Job())

	}

	// Weekly
	if schedule.Weekly != nil {
		jobs = append(jobs, schedule.Weekly.Job())
	}

	// // Monthly
	// if schedule.Monthly != nil {
	// 	for i, day := range schedule.Monthly.Days {
	// 		for _, at := range schedule.Monthly.Ats {
	// 			job := s.Every(schedule.Monthly.Interval).MonthDay(day.OfTheMonth())
	// 			job.At(at.ToString())
	// 			if i == 0 {
	// 				// start scheduler
	// 			}
	// 		}
	// 	}
	// }

	// // Duration / Every
	if schedule.Duration != nil {

		jobs = append(jobs, schedule.Duration.Job())
		// every :=

		// work := NewWorker(TypeCron, scheduler)
		// tasker := gocron.NewTask(work.Run, ctx, "cron")
		// // logJSON(work)
		// if _, err = s.NewJob(def, tasker); err != nil {
		// 	fmt.Println(err)
		// 	return nil, err
		// }
	}

	// Cron (string)
	// if schedule.Cron != "" {
	// 	job := s.Cron(schedule.Cron)
	// 	_ = job
	// }

	return
}

func logJSON(input interface{}) {
	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal schedule:", err)
		return
	}
	fmt.Println(string(data))
}

func (s *Schedule) Logging() {
	logJSON(s)
}

// ------------------- Parse một key value -------------------
func (schedule *Schedule) parseing(kind string, value interface{}) error {
	switch v := value.(type) {
	case string:
		return schedule.smart(kind, v)
	case map[string]interface{}:
		return schedule.mapping(kind, v)
	case []interface{}:
		for i, item := range v {
			switch x := item.(type) {
			case string:
				if err := schedule.smart(kind, x); err != nil {
					return fmt.Errorf("item %d: %w", i, err)
				}
			case map[string]interface{}:
				if err := schedule.mapping(kind, x); err != nil {
					return fmt.Errorf("item %d: %w", i, err)
				}
			default:
				return fmt.Errorf("item %d unsupported type %T", i, x)
			}
		}
	default:
		return fmt.Errorf("value unsupported type %T", value)
	}
	return nil
}

// ------------------- Parse map -------------------
func (s *Schedule) mapping(kind string, m map[string]interface{}) error {
	for key, raw := range m {
		switch val := raw.(type) {
		case string:
			if err := s.smart(kind, val); err != nil {
				return err
			}
		case float64:
			// nếu là every interval
			if kind == "every" || key == "every" {
				if s.Duration == nil {
					s.Duration = &Durationly{}
				}
				s.Duration.Every = time.Duration(val) * time.Second
			}
		default:
			return fmt.Errorf("%s field must be string or number, got %T", key, raw)
		}
	}
	return nil
}

// ------------------- Smart parse string -------------------
func (s *Schedule) smart(kind, v string) error {
	switch kind {
	case "every":
		if !DurationRegex.MatchString(v) {
			return fmt.Errorf("invalid duration: %s", v)
		}
		if s.Duration == nil {
			s.Duration = &Durationly{}
		}
		d, _ := time.ParseDuration(v)
		s.Duration.Every = d
		return nil

	case "daily":
		if !TimeRegex.MatchString(v) {
			return fmt.Errorf("invalid daily time: %s", v)
		}
		at, err := parseAtTime(v)
		if err != nil {
			return err
		}
		if s.Daily == nil {
			s.Daily = &Daily{}
		}
		s.Daily.Ats = append(s.Daily.Ats, at)
		return nil

	case "weekly":
		parts := strings.Fields(v)
		if len(parts) != 2 {
			return fmt.Errorf("invalid weekly format: %s", v)
		}
		day, err := WeekdayFromString(parts[0])
		if err != nil {
			return err
		}
		at, err := parseAtTime(parts[1])
		if err != nil {
			return err
		}
		if s.Weekly == nil {
			s.Weekly = &Weekly{}
		}
		s.Weekly.Weekdays = append(s.Weekly.Weekdays, day)
		s.Weekly.Ats = append(s.Weekly.Ats, at)
		return nil

	case "monthly":
		parts := strings.Fields(v)
		if len(parts) != 2 {
			return fmt.Errorf("invalid monthly format: %s", v)
		}
		day, _ := strconv.Atoi(parts[0])
		at, err := parseAtTime(parts[1])
		if err != nil {
			return err
		}
		if s.Monthly == nil {
			s.Monthly = &Monthly{}
		}
		s.Monthly.Days = append(s.Monthly.Days, Day(day))
		s.Monthly.Ats = append(s.Monthly.Ats, at)
		return nil

	case "cron":
		if !CronRegex.MatchString(v) {
			return fmt.Errorf("invalid cron expression: %s", v)
		}
		s.Cron = v
		return nil

	default:
		// nếu kind rỗng, tự nhận biết
		switch {
		case DurationRegex.MatchString(v):
			return s.smart("every", v)
		case TimeRegex.MatchString(v):
			return s.smart("daily", v)
		case WeeklyRegex.MatchString(v):
			return s.smart("weekly", v)
		case MonthlyRegex.MatchString(v):
			return s.smart("monthly", v)
		case CronRegex.MatchString(v):
			return s.smart("cron", v)
		}
		return fmt.Errorf("unknown schedule format: %s", v)
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

// parseAt parses a string like "HH:MM" or "HH:MM:SS" to gocron.AtTime
func parseAtTime(value string) (At, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return At{}, fmt.Errorf("invalid time format: %s", value)
	}

	hour := atoiOrZero(parts[0])
	minute := atoiOrZero(parts[1])
	second := uint(0)
	if len(parts) == 3 {
		second = atoiOrZero(parts[2])
	}

	return At{Hour: hour, Minutes: minute, Seconds: second}, nil
}

func atoiOrZero(s string) uint {
	i, err := strconv.Atoi(s)
	if err != nil || i < 0 {
		return 0
	}
	return uint(i)
}
