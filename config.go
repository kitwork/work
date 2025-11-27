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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gofiber/fiber/v2"
)

type Config struct {
	secret   Source
	schedule Source
	router   Source

	data []Source

	err error
}

func New() *Config {
	return &Config{}
}

func (c *Config) Run() error {

	pipeline := NewPipe()
	accepts := []string{".work"}

	// 1. Secret
	secret, err := c.secret.Scanner(accepts...)
	if err != nil {
		return err
	}
	pipeline.As("secret", secret)

	// 2. Schedule
	schedules, err := c.schedule.Scanner(accepts...)
	if err != nil {
		return err
	}

	// Create a single scheduler for the application
	s, err := gocron.NewScheduler(gocron.WithLocation(time.Local))
	if err != nil {
		return err
	}

	for name, content := range schedules {

		scheduler, ok := content.(map[string]interface{})
		if !ok {
			fmt.Printf("⚠️  Warning: invalid schedule format: %s\n", name)
			continue
		}

		scheded := false

		for key := range scheduler {
			switch key {
			case "schedules", "schedule", "cron", "daily", "weekly", "monthly", "every":
				scheded = true
			}
			if scheded {
				break
			}
		}

		if scheded {
			jobs, err := NewScheduler(scheduler)
			if err != nil {
				return fmt.Errorf("failed to create scheduler for %s: %w", name, err)
			}

			if len(jobs) > 0 {
				task := gocron.NewTask(func() {
					work := NewWorker(TypeCron, scheduler)
					ctx := NewContext(pipeline.Clone())
					if err := work.Run(ctx, "cron"); err != nil {
						fmt.Printf("Job error in %s: %v\n", name, err)
					}
				})

				for _, job := range jobs {
					if _, err = s.NewJob(job, task); err != nil {
						return err
					}
				}
			}
		}
	}

	// 3. Router
	// Router
	router, err := c.router.Scanner(accepts...)
	if err != nil {
		return err
	}

	// Start scheduler
	s.Start()
	defer func() { _ = s.Shutdown() }()

	// If router exists, start Fiber alongside scheduler
	if len(router) > 0 {

		fmt.Println(router)
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
		})

		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Kitwork 🌱")
		})

		go func() {
			if err := app.Listen(":3000"); err != nil {
				fmt.Println("Fiber stopped:", err)
			}
		}()
	}

	fmt.Println("KitWork started. Press Ctrl+C to stop.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down...")
	return nil
}
