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
	"fmt"
	"reflect"
	"text/template"

	"github.com/kitwork/pipe"
)

type Context struct {
	Return interface{} // dữ liệu cuối cùng trả về khi workflow kết thúc
	Result interface{} // dữ liệu hiện tại của workflow, được các work đọc/ghi

	Debug bool

	Version string

	templ *template.Template
	pipe  *template.FuncMap
}

func NewContext() *Context {

	ctx := Context{}
	ctx.pipeData()

	return &ctx
}

// addPipe thêm một func mới vào pipes
func (c *Context) pipeAdd(name string, function interface{}) error {
	if name == "" {
		return fmt.Errorf("pipe name cannot be empty")
	}

	// check function có phải là func
	if function == nil || reflect.TypeOf(function).Kind() != reflect.Func {
		return fmt.Errorf("value for pipe '%s' is not a function", name)
	}

	// khởi tạo pipes nếu nil
	c.pipeline()

	// check key đã tồn tại chưa
	if _, exists := (*c.pipe)[name]; exists {
		return fmt.Errorf("pipe '%s' already exists", name)
	}

	// thêm vào pipes
	(*c.pipe)[name] = function
	return nil
}

func (c *Context) pipeValue(name string, val interface{}) error {
	return c.pipeAdd(name, func() reflect.Value {
		return reflect.ValueOf(val)
	})
}

func (c *Context) pipeline() (result template.FuncMap) {
	if c.pipe == nil {
		m := pipe.Functions()
		c.pipe = &m
	}
	return *c.pipe
}

func (c *Context) pipeData() (result template.FuncMap) {
	c.pipeline()
	for key, val := range data {

		if _, exists := (*c.pipe)[key]; !exists {
			(*c.pipe)[key] = func() reflect.Value {
				return reflect.ValueOf(val)
			}
		}

	}
	return *c.pipe
}

func (c *Context) render(val string) (string, error) {
	if val == "" {
		return "", nil
	}

	funcs := c.pipeline()

	tmpl, err := template.New("work").Funcs(funcs).Parse(val)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		return "", err
	}

	return buf.String(), nil
}
