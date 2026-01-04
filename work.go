// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
	"time"
)

// Work / work đại diện cho 1 work/workflow node
type Work struct {
	Name      string
	Condition bool
	As        string
	Async     bool
	Type      Type
	Kind      Kind // short, full, value, list, switch
	Config    map[string]interface{}
	Works     []*Work // chain mặc định
	Success   []*Work // nếu OK
	Error     []*Work // nếu lỗi
	Timeout   time.Duration
}

func (t *Work) value() string {
	s, _ := t.Config["value"].(string)
	return s
}

func (t *Work) alias() string {
	s, _ := t.Config["as"].(string)
	if s != "" {
		return s
	}
	alias, _ := t.Config["alias"].(string)
	return alias
}

// ========================
//  WORK RUNNER
// ========================

func (t *Work) Run(ctx *Context, skipTypes ...string) (err error) {

	if ctx.Return != nil {
		return
	}
	// fmt.Printf("\n→ Running Work: [%s] %s\n", t.Type, t.Name)

	// 1. chạy work chính
	switch t.Type {

	case TypeReturn:
		return t.Return(ctx)

	case TypeScript:
		err = t.Script(ctx)

	case TypeFetch, TypeHTTP, TypeClient, TypeRequest:
		err = t.Request(ctx)

	case TypeCron:
		// return nil

	case TypeRouter:
		// return nil

	case TypeScreenshot:
		// return nil
		err = t.Screenshot(ctx)

	case TypeMapping:
		// return nil
		err = t.Mapping(ctx)

	case TypeTarget:
		// return nil
		err = t.Target(ctx)

	case TypeAs, TypeAlias:
		// return nil
		err = t.Alias(ctx)

	case TypeLog:
		err = t.Log(ctx)

	case TypeParser:

		err = t.Parse(ctx)

	case TypeSql:

		err = t.Sql(ctx)

	case TypeForeach, TypeLoop:
		err = t.Loop(ctx)

	default:
		err = fmt.Errorf("unknown work type: %s", t.Type)
	}

	// 2. Nếu work chính OK → chạy chuỗi Works
	if err == nil && len(t.Works) > 0 {
		// fmt.Println("→ run works chain...")
		for _, a := range t.Works {
			if err = a.Run(ctx); err != nil {
				break
			}
		}
	}

	// 3. Nếu thành công → chạy Success branch
	if err == nil && len(t.Success) > 0 {
		// fmt.Println("→ Success → chạy success branch")
		for _, s := range t.Success {
			s.Run(ctx)
		}
	}

	// 4. Nếu lỗi → chạy Error branch
	if err != nil && len(t.Error) > 0 {
		fmt.Printf("→ Error in [%s]: %v. Running error branch...\n", t.Name, err)
		for _, e := range t.Error {
			e.Run(ctx)
		}
	}

	return
}

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
		case "success", "then":
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
			work.Condition = true
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
				case "success", "then":
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
