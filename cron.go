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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-co-op/gocron/v2"
)

type Cron struct {
	Name      string     `work:"name"`
	As        string     `work:"as"`
	Schedules []Schedule `work:"schedules"`
}

// --------------------------
// Cron handler
// --------------------------
func (t *Work) Cron(ctx *Context) error {
	if t.Type != TypeCron {
		return errors.New("type is not cron")
	}

	cfg := Cron{}
	if err := cfg.parse(t, ctx); err != nil {
		return err
	}

	// Tạo scheduler mới
	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	// Đăng ký tất cả schedule
	for _, sched := range cfg.Schedules {

		def, err := sched.Definition()
		if err != nil {
			fmt.Println(err)
			return err
		}

		task := gocron.NewTask(t.Run, ctx, "cron")
		if _, err = s.NewJob(def, task); err != nil {
			return err
		}

	}

	// Bắt đầu scheduler async
	s.Start()

	fmt.Println("→ [cron] tất cả lịch đã đăng ký và chạy ...")
	return nil
}

// --------------------------
// Parse schedules từ w.Config
// --------------------------
func (c *Cron) parse(w *Work, ctx *Context) error {
	raw, ok := w.Config["schedules"]
	if !ok {
		return fmt.Errorf("schedules not found in config")
	}

	rawSlice, ok := raw.([]interface{})
	if !ok {
		return fmt.Errorf("schedules is not a list")
	}

	schedules := make([]Schedule, 0, len(rawSlice))

	// Regex patterns
	cronRegex := regexp.MustCompile(`^(\S+\s){4}\S+$`)
	dailyRegex := regexp.MustCompile(`^\d{1,2}:\d{2}$`)
	weeklyRegex := regexp.MustCompile(`^(Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday)\s\d{1,2}:\d{2}$`)
	monthlyRegex := regexp.MustCompile(`^\d{1,2}\s\d{1,2}:\d{2}$`)
	hourlyShortRegex := regexp.MustCompile(`^:(\d{1,2})$`)

	for _, item := range rawSlice {
		switch v := item.(type) {

		case string:
			v = strings.TrimSpace(v)
			s := Schedule{}

			switch {
			case cronRegex.MatchString(v):
				s.Type = "cron"
				s.Value = v
			case dailyRegex.MatchString(v):
				s.Type = "daily"
				s.Value = v
			case weeklyRegex.MatchString(v):
				s.Type = "weekly"
				s.Value = v
			case monthlyRegex.MatchString(v):
				s.Type = "monthly"
				s.Value = v
			case hourlyShortRegex.MatchString(v):
				s.Type = "hourly"
				matches := hourlyShortRegex.FindStringSubmatch(v)
				min := matches[1]
				if min == "00" {
					min = "0"
				}
				s.Value = min
			default:
				s.Type = "every"
				s.Value = v
			}

			schedules = append(schedules, s)

		case map[string]interface{}:
			s := Schedule{}
			for key, val := range v {
				switch key {
				case "daily":
					s.Type = key
					if value, ok := val.(string); ok {
						s.Value = value
					}
				case "weekly":
					s.Type = key
					if value, ok := val.(string); ok {
						s.Value = value
					}
				case "monthly":
					s.Type = key
					if value, ok := val.(string); ok {
						s.Value = value
					}
				case "every":
					s.Type = key
					if value, ok := val.(string); ok {
						s.Value = value
					}
				}

			}
			schedules = append(schedules, s)
		default:
			return fmt.Errorf("invalid schedule item: %T", item)
		}
	}

	c.Schedules = schedules
	fmt.Println("Parsed schedules:", schedules)
	return nil
}
