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

import "fmt"

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

	ctx := NewContext()

	accepts := []string{".work"}

	// 1. Secret
	secretData, err := c.secret.Scanner(accepts...)
	if err != nil {
		return err
	}
	ctx.emit("secret", secretData)

	// 2. Schedule
	scheduleData, err := c.schedule.Scanner(accepts...)
	if err != nil {
		return err
	}

	for name, content := range scheduleData {
		data, ok := content.(map[string]interface{})
		if !ok {
			fmt.Printf("⚠️  Warning: invalid schedule format: %s\n", name)
			continue
		}

		// fmt.Println(data)
		sched, exists := data["schedules"]
		if !exists {
			fmt.Printf("⚠️  Warning: missing key 'startTime' in schedule: %s\n", name)
			continue
		}

		switch v := sched.(type) {
		case string:
			job("every", v)

		case map[string]interface{}:

			for key, val := range v {
				switch key {
				case "daily", "weekly", "monthly", "every":
					// Chỉ lấy string, bạn có thể thay đổi type assertion theo nhu cầu
					if strVal, ok := val.(string); ok {
						job("every", strVal)
					} else {
						fmt.Printf("%s -> %s: value is not string, type: %T\n", name, key, val)
					}
				}

			}

		case []interface{}:
			for i, item := range v {
				switch s := item.(type) {
				case string:
					job("every", s)
				case map[string]interface{}:
					for key, val := range s {
						switch key {
						case "daily", "weekly", "monthly", "every":
							if strVal, ok := val.(string); ok {
								job("every", strVal)
							} else {
								fmt.Printf("%s -> item %d -> %s: value is not string, type: %T\n", name, i, key, val)
							}
						}
					}
				default:
					fmt.Printf("%s -> item %d: unknown type %T\n", name, i, item)
				}
			}
		default:
			fmt.Println("unknown type")
		}
		fmt.Println(sched)

	}

	// 3. Router
	// router, err := c.router.Scanner(accepts...)
	// if err != nil {
	// 	return err
	// }
	// for name, data := range router {
	// 	raw, ok := data.(map[string]interface{})
	// 	if !ok {
	// 		fmt.Printf("⚠️  Warning: invalid router format: %s\n", name)
	// 		continue
	// 	}
	// 	work := Parsing(raw)
	// 	if err := work.Run(ctx); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// func (c *Config) Run() error {

// 	ctx := NewContext()

// 	for _, file := range c.Files {

// 		// chỉ xử lý Work
// 		ext := filepath.Ext(file)
// 		if ext != ".work" {
// 			continue
// 		}
// 		// if ext != ".yaml" && ext != ".yml" {
// 		// 	continue
// 		// }

// 		// 1. đọc file yaml
// 		workflow, err := readfile(file)
// 		if err != nil {
// 			return fmt.Errorf("error reading %s: %w", file, err)
// 		}

// 		// 2. parse workflow
// 		root := Parsing(workflow)
// 		if root == nil {
// 			return fmt.Errorf("cannot parse workflow: %s", file)
// 		}

// 		// 3. run workflow
// 		fmt.Println(">> Running workflow:", file)

// 		if err := root.Run(ctx); err != nil {
// 			return err
// 		}

// 	}

// 	return nil
// }

// func Source(folder string) *Config {
// 	cfg := Config{}

// 	filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
// 		if err != nil {
// 			return nil
// 		}

// 		if !d.IsDir() {
// 			cfg.Files = append(cfg.Files, path)
// 		}

// 		return nil
// 	})

// 	return &cfg
// }
