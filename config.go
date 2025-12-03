// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gofiber/fiber/v2"
)

type Config struct {
	secret   *Source
	schedule *Source
	router   *Source

	database *Source

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
	dbs, err := c.ConnectDB(accepts...)
	if err != nil {
		panic(err)
	}

	database := new(Database)

	for _, db := range dbs {
		fmt.Println("Connected DB:", db.Key)
		defer db.SQL.Close() // nhớ đóng connection khi app shutdown

		database.Add(db.Key, db.Gorm, db.IsDefault)

	}

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

	// Start scheduler
	s.Start()
	defer func() { _ = s.Shutdown() }()

	// 3. Router
	// Router
	routes, err := c.routing(accepts...)
	if err != nil {
		return err
	}

	// fmt.Println(c.router.Routing(".work"))

	var port string
	// If router exists, start Fiber alongside scheduler
	if len(routes) > 0 {

		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
		})

		for _, router := range routes {
			switch router.Method {
			case "GET", "POST", "PUT", "DELETE":

				app.Add(router.Method, router.Path, func(request *fiber.Ctx) error {

					ctx := NewContext(pipeline.Clone()).db(database)
					return router.Handle(request, ctx)
				})
				break
			}

		}

		port = ":80"

		go func() {
			if err := app.Listen(":80"); err != nil {
				fmt.Println("HTTP stopped:", err)
			}
		}()

		go func() {
			cfg := SSL()
			ln, err := tls.Listen("tcp", ":443", cfg)
			if err != nil {
				panic(err)
			}

			if err := app.Listener(ln); err != nil {
				fmt.Println("HTTPS stopped:", err)
			}
		}()

	}

	fmt.Printf("KitWork started port %s . Press Ctrl+C to stop.", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down...")
	return nil
}
