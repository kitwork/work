// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

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

	return &Router{Method: strings.ToUpper(method), Path: path, Source: raw}
}

func (r *Router) Pathfile() string {
	return path.Join("router", r.Source) + ".work"
}

func (r *Router) data() (map[string]interface{}, error) {
	return readfile(path.Join("router", r.Source) + ".work")
}

func (router *Router) Handle(request *fiber.Ctx, ctx *Context) error {
	data, err := router.data()
	if err != nil {
		return err
	}

	work := NewWorker(TypeRouter, data)

	params := request.AllParams()
	queries := request.Queries()

	ctx.Set("routes", request.App().GetRoutes(false))
	ctx.Set("request", request)
	ctx.Set("params", params)
	ctx.Set("queries", queries)

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
		case "page":
			request.Set("Content-Type", "text/html")
			return request.SendString(returned.Page())
		case "image/png", "png":

			return request.Type("png").Send(returned.Png())
		case "json":
			return request.JSON(returned.JSON())
		}
	}
	return request.JSON(data)
}
