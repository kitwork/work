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
