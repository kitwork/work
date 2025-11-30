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
	"path"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (cfg *Config) Router(source ...string) *Config {
	cfg.router = NewSource("router", source...)
	return cfg
}

func (cfg *Config) routing(accepts ...string) (routes []Router, err error) {
	return cfg.router.Routing(accepts...)
}

type Router struct {
	Source string `work:"srouce"`
	Method string `work:"method"`
	Path   string `work:"path"`
}

func NewRouter(raw string) *Router {

	// Bỏ dấu / thừa
	raw = strings.Trim(raw, "/")

	parts := strings.Split(raw, "/")

	// Method là phần cuối
	method := parts[len(parts)-1]

	// Path parts (bỏ method)
	pathParts := parts[:len(parts)-1]

	// Convert {id} => :id, {$} => *
	for i, p := range pathParts {

		// wildcard
		if p == "{$}" {
			pathParts[i] = "*"
			continue
		}

		// param {id}
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			key := p[1 : len(p)-1] // bỏ 2 dấu {}
			pathParts[i] = ":" + key
		}
	}
	path := strings.Join(pathParts, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return &Router{Method: method, Path: path, Source: raw}
}

func (r *Router) Pathfile() string {
	return path.Join("router", r.Source) + ".work"
}

func (r *Router) data() (map[string]interface{}, error) {
	return readfile(path.Join("router", r.Source) + ".work")
}

func (router *Router) Handle(request *fiber.Ctx, pipeline *Pipeline) error {
	data, err := router.data()
	if err != nil {
		return err
	}

	params := request.AllParams()
	queries := request.Queries()
	pipeRouter := pipeline.Clone()
	pipeRouter.As("routes", request.App().GetRoutes(false))
	pipeRouter.As("request", request)
	pipeRouter.As("param", params)
	pipeRouter.As("query", queries)
	work := NewWorker(TypeRouter, data)
	ctx := NewContext(pipeRouter)
	if err := work.Run(ctx, "router"); err != nil {
		return err
	}
	returned := ctx.Return

	if ctx.Return != nil {
		switch returned.Type {
		case "string":
			result := returned.String()
			if IsJSON(result) {
				json, err := ParseJSON(result)
				if err == nil {
					return request.JSON(json)
				}
			}
			return request.SendString(result)
		case "html":
			request.Set("Content-Type", "text/html")
			return request.SendString(returned.String())
		case "json":
			return request.JSON(returned.JSON())
		}
	}
	return request.JSON(data)
}
