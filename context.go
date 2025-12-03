// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

type Context struct {
	Return *Return // dữ liệu cuối cùng trả về khi workflow kết thúc

	returned *Return
	debug    bool
	version  string

	Debug bool

	Version string

	templ *template.Template
	pipes *Pipeline

	database *Database
}

func NewContext(pipes *Pipeline) *Context {
	return &Context{pipes: pipes}
}

func (ctx *Context) db(database *Database) *Context {
	ctx.database = database
	return ctx
}

func (c *Context) As(key string, val interface{}) error {
	if key != "" {
		return c.pipes.As(key, val)
	}
	return nil
}

func (c *Context) Getter(path string) (interface{}, bool) {
	return c.pipes.Getter(path)
}

func (c *Context) template(aliases ...map[string]interface{}) *template.Template {
	if c.templ == nil {
		c.templ = template.New("work")
	}
	if len(aliases) == 0 {
		return c.templ.Funcs(c.pipes.Functions())
	}
	pipes := c.pipes.Clone()

	for _, alias := range aliases {
		for k, v := range alias {
			pipes.As(k, v)
		}
	}

	return c.templ.Funcs(pipes.Functions())
}

func (c *Context) render(val string, aliases ...map[string]interface{}) (string, error) {
	if val == "" {
		return "", nil
	}

	var text string
	tmpl := c.template(aliases...)

	if strings.HasSuffix(val, ".tmpl") {
		// Nếu là file template
		content, err := os.ReadFile(val)
		if err != nil {
			return "", err
		}

		text = string(content)

	} else {
		text = val
	}

	tmple, err := tmpl.Parse(text)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmple.Execute(&buf, c); err != nil {
		return "", err
	}

	return buf.String(), nil
}
