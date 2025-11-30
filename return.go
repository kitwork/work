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
	"encoding/json"
	"fmt"
	"strings"
)

type Return struct {
	// Name    string `work:"name"`
	ctx     *Context
	Type    string      `work:"type" default:"string"`
	Content interface{} `work:"content"`
}

func (r *Return) String() string {
	if r.Type == "string" || r.Type == "html" {
		if s, ok := r.Content.(string); ok {
			v, _ := r.ctx.render(s)
			return v
		}
	}
	return fmt.Sprintf("%v", r.Content) // fallback
}

func (r *Return) JSON() interface{} {
	if r.Type != "json" {
		return nil
	}

	switch c := r.Content.(type) {

	// 1️⃣ Nếu là string → render + parse JSON
	case string:
		v, err := r.ctx.render(c)
		if err != nil {
			fmt.Println("Render error:", err)
			return nil
		}

		if !IsJSON(v) {
			fmt.Println("Value not JSON:", v)
			return nil
		}

		result, err := ParseJSON(v)
		if err != nil {
			fmt.Println("JSON unmarshal error:", err)
			return nil
		}
		return result

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

func IsPipe(s string) bool {
	s = strings.TrimSpace(s) // loại bỏ khoảng trắng đầu/cuối
	return strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}")
}

func IsJSON(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	// JSON phải bắt đầu bằng { hoặc [
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

func ParseJSON(s string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RenderJSONMap recursively renders a map or slice
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
		rendered, _ := ctx.render(str)
		if IsJSON(rendered) {
			parsed, _ := ParseJSON(rendered)
			return parsed
		}
		return rendered

	// Các type khác → trả nguyên
	default:
		return v
	}
}

func (t *Work) Return(ctx *Context) error {

	cfg := Return{ctx: ctx}

	switch t.Kind {

	case KindValue:

		cfg.Type = "string"
		cfg.Content = t.value()

	case KindFull:
		if len(t.Config) == 1 {
			for k, v := range t.Config {
				switch k {
				case "html":
					cfg.Type = k    // gán type
					cfg.Content = v // gán value
				case "string":
					cfg.Type = k    // gán type
					cfg.Content = v // gán value
				case "json":
					cfg.Type = k    // gán type
					cfg.Content = v // gán value
				case "file":
					cfg.Type = k    // gán type
					cfg.Content = v // gán value
				default:
					return fmt.Errorf("unknown type: %s", k)
				}
			}

		}

	default:
		return fmt.Errorf("type is not a request/fetch type: %s", t.Type)
	}

	ctx.Return = &cfg

	return nil
}
