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

func (r *Return) JSON() map[string]interface{} {
	if r.Type == "json" {
		if m, ok := r.Content.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
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
