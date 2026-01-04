// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

type Return struct {
	ctx     *Context
	Type    string      `work:"type" default:"string"`
	Content interface{} `work:"content"`
	File    string      `work:"file"`
}

// String returns string or rendered template
func (r *Return) String() string {
	if r.Type == "string" || r.Type == "html" {
		if s, ok := r.Content.(string); ok {
			v, _ := r.ctx.render(s)
			return v
		}
	}
	return fmt.Sprintf("%v", r.Content)
}

// Png returns byte slice
func (r *Return) Png() []byte {
	if r.Type != "png" {
		return nil
	}

	switch v := r.Content.(type) {

	case []byte:
		return v

	case string:
		// chỉ resolve nếu là string template
		val, err := r.ctx.evaluate(v)
		if err != nil {
			return nil
		}

		b, _ := ToBytes(val)
		return b

	default:
		b, _ := ToBytes(v)
		return b
	}
}

// Page returns rendered page
func (r *Return) Page() string {
	if r.Type == "page" {
		v, _ := r.ctx.render(r.File)
		return v
	}
	return fmt.Sprintf("%v", r.Content)
}

// JSON returns fully resolved JSON structure
// func (r *Return) JSON() interface{} {
// 	if r.Type != "json" {
// 		return nil
// 	}
// 	val, _ := r.ctx.json(r.Content)
// 	return val
// }

func (r *Return) JSON() interface{} {
	if r.Type != "json" {
		return nil
	}

	switch c := r.Content.(type) {

	// 1️⃣ Nếu là string → render + parse JSON
	case string:
		v, err := r.ctx.evaluate(c)
		if err != nil {
			fmt.Println("Render error:", err)
			return nil
		}

		return v

	// 2️⃣ Nếu là map → trả trực tiếp
	case map[string]interface{}:

		return MapJSON(r.ctx, c)

	// 3️⃣ Nếu là slice → trả trực tiếp
	case []interface{}:
		return c
	}

	// 4️⃣ Trường hợp khác → cố gắng marshal → unmarshal
	b, err := json.Marshal(r.Content)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return nil
	}

	var result interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		fmt.Println("JSON unmarshal error:", err)
		return nil
	}

	return result
}

// Convert interface{} to []byte
func ToBytes(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	case int, int64, float64, float32, uint:
		return []byte(fmt.Sprint(val)), nil
	case bool:
		if val {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	default:
		return json.Marshal(val)
	}
}

// Check if string looks like JSON
func IsJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

// Parse string to interface{}
func ParseJSON(s string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	return result, err
}

func MapJSON(ctx *Context, val interface{}) interface{} {
	switch v := val.(type) {

	// Nếu là map → đệ quy xử lý từng key
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			newMap[key] = MapJSON(ctx, value)
		}
		return newMap

	// Nếu là slice → đệ quy từng phần tử
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, item := range v {
			newSlice[i] = MapJSON(ctx, item)
		}
		return newSlice

	// Nếu là string → kiểm tra pipe hoặc JSON
	case string:
		str := strings.TrimSpace(v)
		rendered, _ := ctx.evaluate(str)

		return rendered

	// Các type khác → trả nguyên
	default:
		return v
	}
}

// Return Work result
func (t *Work) Return(ctx *Context) error {
	cfg := &Return{ctx: ctx}

	switch t.Kind {
	case KindValue:

		val := t.value()
		val = strings.TrimSpace(val)
		if strings.HasSuffix(val, ".tmpl") {
			cfg.Type = "page"
			if val == ".tmpl" {
				cfg.File = path.Join(ctx.dir, "index"+val)
			} else {
				cfg.File = val
			}

		} else {
			if strings.HasPrefix(val, "$") {
				cfg.Type = "json"
			} else {
				cfg.Type = "string"
			}
			cfg.Content = val
		}

	case KindFull:
		if len(t.Config) == 1 {
			for k, v := range t.Config {
				switch k {
				case "string", "html", "png", "json":
					cfg.Type = k
					cfg.Content = v
				case "page":
					cfg.Type = k
					cfg.File = fmt.Sprintf("%v", v)
				case "file":
					cfg.Type = k
					cfg.Content = v
				default:
					return fmt.Errorf("unknown type: %s", k)
				}
			}
		}

	default:
		return fmt.Errorf("invalid work type: %s", t.Type)
	}

	ctx.Return = cfg
	return nil
}
