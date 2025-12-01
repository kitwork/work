// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"bytes"
	"text/template"
)

type Context struct {
	Return *Return // dữ liệu cuối cùng trả về khi workflow kết thúc

	Debug bool

	Version string

	templ *template.Template
	pipes *Pipeline
}

func NewContext(pipes *Pipeline) *Context {
	return &Context{pipes: pipes}
}

func (c *Context) As(key string, val interface{}) error {
	return c.pipes.As(key, val)
}

func (c *Context) Getter(path string) (interface{}, bool) {
	return c.pipes.Getter(path)
}

func (c *Context) template() *template.Template {
	if c.templ == nil {
		c.templ = template.New("work")
	}
	return c.templ.Funcs(c.pipes.Functions())
}

func (c *Context) render(val string) (string, error) {
	if val == "" {
		return "", nil
	}

	tmpl, err := c.template().Parse(val)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		return "", err
	}

	return buf.String(), nil
}
