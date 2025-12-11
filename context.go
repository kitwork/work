// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/kitwork/pipe"
)

type Context struct {
	Return *Return // dữ liệu cuối cùng trả về khi workflow kết thúc

	returned *Return
	debug    bool
	version  string

	Vars    map[string]interface{}
	Debug   bool
	Version string

	templ *template.Template
	pipes *Pipeline

	database *Database

	aliases []map[string]interface{}
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

func (c *Context) Set(key string, val interface{}) error {
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	// Remove leading '$'
	if key[0] == '$' {
		key = key[1:]
	}

	if key == "" {
		return fmt.Errorf("variable key cannot be empty after removing '$'")
	}

	if c.Vars == nil {
		c.Vars = make(map[string]interface{})
	}

	if _, exists := c.Vars[key]; exists {
		return fmt.Errorf("variable %q already exists", key)
	}

	c.Vars[key] = val
	return nil
}

func (c *Context) Get(key string, defaults ...interface{}) interface{} {
	if key == "" {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return ""
	}

	// Xóa dấu $ nếu có
	if key[0] == '$' {
		key = key[1:]
	}

	// Nếu sau khi xoá còn rỗng → trả default hoặc ""
	if key == "" {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return ""
	}

	// Lấy từ biến
	if c.Vars != nil {
		if v, ok := c.Vars[key]; ok {
			return v
		}
	}

	// fallback trả default
	if len(defaults) > 0 {
		return defaults[0]
	}

	return ""
}

func (c *Context) Getting(key string) (interface{}, bool) {
	if c.Vars == nil {
		return nil, false
	}

	// Không parse gì thêm, trả đúng key luôn
	v, ok := c.Vars[key]
	return v, ok
}

func (c *Context) Getter(path string) (interface{}, bool) {
	return c.pipes.Getter(path)
}

func (c *Context) Aliases(aliases ...map[string]interface{}) error {
	c.aliases = aliases
	return nil
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
	if len(c.aliases) > 0 {
		aliases = c.aliases
	}
	tmpl := c.template(aliases...)

	c.aliases = nil

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

	text = pipe.Preprocessor(text)

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

type ValueWriter struct {
	Value interface{}
}

func (w *ValueWriter) Write(p []byte) (int, error) {
	// Template sẽ ghi chuỗi vào Writer
	// Nhưng nếu chuỗi đó là một literal (số, bool...)
	// ta convert về raw value nếu parse được.
	s := strings.TrimSpace(string(p))

	// thử parse số
	if i, err := strconv.Atoi(s); err == nil {
		w.Value = i
		return len(p), nil
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		w.Value = f
		return len(p), nil
	}

	// true / false
	if s == "true" {
		w.Value = true
		return len(p), nil
	}
	if s == "false" {
		w.Value = false
		return len(p), nil
	}

	// nil
	if s == "nil" || s == "null" {
		w.Value = nil
		return len(p), nil
	}

	// fallback: giữ string
	w.Value = s
	return len(p), nil
}
