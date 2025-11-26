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
	"time"
)

func NewWorker(typing Type, data map[string]interface{}) *Work {
	work := &Work{
		Config:  make(map[string]interface{}),
		Works:   []*Work{},
		Success: []*Work{},
		Error:   []*Work{},
		Kind:    KindUnknown,
		Type:    typing,
	}

	for key, val := range data {
		switch key {
		case "name":
			if s, ok := val.(string); ok {
				work.Name = s
			}

		case "works", "work":
			work.Works = NewWorks(val)
		case "success":
			work.Success = NewWorks(val)
		case "error":
			work.Error = NewWorks(val)
		case "timeout":
			if tStr, ok := val.(string); ok {
				d, err := time.ParseDuration(tStr)
				if err == nil {
					work.Timeout = d
				} else {
					fmt.Printf("⚠️  Warning: cannot parse timeout %v: %v\n", val, err)
				}
			}
		default:
			work.Config[key] = val
		}

	}

	return work
}

// Parse converts YAML-decoded map into an Work struct (recursive).
func NewWork(data map[string]interface{}) *Work {
	work := &Work{
		Config:  make(map[string]interface{}),
		Works:   []*Work{},
		Success: []*Work{},
		Error:   []*Work{},
		Kind:    KindUnknown,
	}

	for key, val := range data {

		switch v := val.(type) {

		// ================================
		// 1) FULL FORM
		//    log:
		//      name: hello
		//      works: [...]
		// ================================
		case map[string]interface{}:
			work.Type = TypeParseSafe(key)
			work.Kind = KindFull

			for k, vv := range v {
				switch k {
				case "name":
					if s, ok := vv.(string); ok {
						work.Name = s
					}

				case "works", "work":
					work.Works = NewWorks(vv)
				case "success":
					work.Success = NewWorks(vv)
				case "error":
					work.Error = NewWorks(vv)
				case "timeout":
					if tStr, ok := vv.(string); ok {
						d, err := time.ParseDuration(tStr)
						if err == nil {
							work.Timeout = d
						} else {
							fmt.Printf("⚠️  Warning: cannot parse timeout %v: %v\n", vv, err)
						}
					}
				default:
					work.Config[k] = vv
				}
			}

		// ================================
		// 2) LIST FORM
		//    works:
		//      - log
		//      - fetch: {...}
		// ================================
		case []interface{}:
			work.Type = TypeParseSafe(key)
			work.Kind = KindList
			work.Works = NewWorks(v)

		// ================================
		// 3) VALUE FORM
		//    log: "hello"
		// ================================
		case string, int, float64, bool:
			work.Type = TypeParseSafe(key)
			work.Kind = KindValue
			work.Config["value"] = v

		// ================================
		// 4) FALLBACK / UNKNOWN
		// ================================
		default:
			work.Type = TypeCustom
			work.Kind = KindUnknown
			work.Config["value"] = v
		}
	}

	return work
}

// NewWorks parses:
// - short string: "log"
// - full form: { fetch: {...} }
// - unknown primitives
func NewWorks(v interface{}) []*Work {
	result := []*Work{}

	list, ok := v.([]interface{})
	if !ok {
		return result
	}

	for _, item := range list {

		switch x := item.(type) {

		// -----------------------
		// SHORT FORM
		// - log
		// -----------------------
		case string:
			result = append(result, &Work{
				Type:   TypeParseSafe(x),
				Kind:   KindShort,
				Config: map[string]interface{}{},
			})

		// -----------------------
		// FULL FORM
		// - log: { msg: "hi" }
		// -----------------------
		case map[string]interface{}:
			result = append(result, NewWork(x))

		// -----------------------
		// Unknown value in list
		// -----------------------
		default:
			result = append(result, &Work{
				Type:   TypeCustom,
				Kind:   KindUnknown,
				Config: map[string]interface{}{"value": x},
			})
		}
	}

	return result
}
