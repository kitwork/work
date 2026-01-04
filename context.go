// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/kitwork/pipe"
)

type Context struct {
	Return *Return
	Error  error

	Vars map[string]any

	Errors []error

	Debug   bool
	Version string

	templ *template.Template
	pipes *Pipeline

	database *Database

	dir string

	temporary map[string]any
}

func NewContext(pipes *Pipeline) *Context {
	return &Context{
		pipes: pipes,
		Vars:  make(map[string]any),
	}
}

func (ctx *Context) db(database *Database) *Context {
	ctx.database = database
	return ctx
}

func (ctx *Context) directory(dir string) *Context {
	ctx.dir = dir
	return ctx
}

/* =========================
   Variable core
========================= */

func normalizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("variable key cannot be empty")
	}

	key = strings.TrimPrefix(strings.TrimSpace(key), "$")

	if idx := strings.IndexByte(key, '.'); idx >= 0 {
		key = key[:idx]
	}

	if key == "" {
		return "", fmt.Errorf("variable key cannot be empty")
	}

	return key, nil
}

func (c *Context) set(key string, val interface{}) error {
	k, err := normalizeKey(key)
	if err != nil {
		return err
	}

	if _, ok := c.Vars[k]; ok {
		return fmt.Errorf("variable %q already exists", k)
	}

	c.Vars[k] = val
	return nil
}

func (c *Context) target(target string, keys ...string) error {
	target = NormalizeTarget(target, keys...)
	result, err := c.evaluate(target)
	if err != nil {
		return err
	}

	return c.as(result)
}

func (c *Context) alias(alias string) error {
	result, err := c.result()
	if err != nil {
		return err
	}

	return c.as(result, alias)
}

func (c *Context) as(val any, keys ...string) error {
	// Luôn gán result
	c.Vars["result"] = val

	// Nếu không có keys, kết thúc luôn
	if len(keys) == 0 {
		return nil
	}

	// Duyệt keys và set
	for _, k := range keys {
		if k != "" {
			if err := c.set(k, val); err != nil {
				return err
			}
		}

	}

	return nil
}

func (c *Context) del(key string) {
	k, err := normalizeKey(key)
	if err != nil {
		return
	}
	delete(c.Vars, k)
}

/* =========================
   Getter (nested)
========================= */

func (c *Context) Getter(path string) (any, bool) {

	return GetterFromMap(c.Vars, path)
}

func GetterFromMap(m map[string]any, path string) (any, bool) {
	if m == nil || path == "" {
		return nil, false
	}

	path = strings.TrimPrefix(strings.TrimSpace(path), "$")

	parts := strings.Split(path, ".")

	var cur any = m

	for _, key := range parts {
		if key == "" {
			return nil, false
		}

		switch node := cur.(type) {
		case map[string]string:
			v, ok := node[key]
			if !ok {
				return nil, false
			}
			cur = v
		case map[string]any:
			v, ok := node[key]
			if !ok {
				return nil, false
			}
			cur = v
		case []any:
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(node) {
				return nil, false
			}

			cur = node[i]

		default:

			return nil, false
		}
	}

	return cur, true
}

/* =========================
   Temporary (sandbox vars)
========================= */

func (c *Context) temp(tmp map[string]any) *Context {
	if tmp == nil {
		c.temporary = nil
		return c
	}

	if c.temporary == nil {
		c.temporary = make(map[string]any)
	}

	for k, v := range tmp {
		nk, err := normalizeKey(k)
		if err != nil {
			continue
		}
		c.temporary[nk] = v
	}

	return c
}

func (c *Context) result(defaults ...any) (any, error) {
	return c.evaluate("$result", defaults...)
}

func (c *Context) evaluate(in interface{}, defaults ...any) (any, error) {
	var val string
	switch v := in.(type) {
	case string:
		val = v
	default:
		if len(defaults) == 0 {
			return v, nil
		}
		for _, def := range defaults {
			if def == nil {
				continue
			}
			if !reflect.ValueOf(def).IsZero() {
				return def, nil
			}
		}
	}

	val = strings.TrimSpace(val)
	if val == "" {
		return "", nil
	}

	if strings.HasPrefix(val, "$") {
		// ưu tiên temporary
		if c.temporary != nil {
			if v, ok := GetterFromMap(c.temporary, val); ok {
				return v, nil
			}
		}

		if v, ok := c.Getter(val); ok {

			return v, nil
		}
	}

	return val, nil
}

func (c *Context) evaluator(in interface{}, defaults ...any) (any, error) {
	var val string

	switch v := in.(type) {
	case string:
		val = strings.TrimSpace(v)
	default:
		if len(defaults) == 0 {
			return v, nil
		}
		for _, def := range defaults {
			if def != nil && !reflect.ValueOf(def).IsZero() {
				return def, nil
			}
		}
		return v, nil // 👈 fallback đúng
	}

	if val == "" {
		return "", nil
	}

	if strings.HasPrefix(val, "{{") && strings.HasSuffix(val, "}}") {
		val = strings.TrimSpace(val[2 : len(val)-2])
	}

	eval := NewEvaluator(c.scope(), c.pipes.Functions())

	// 1. Biến $
	if strings.HasPrefix(val, "$") || strings.HasPrefix(val, "($") {

		if !strings.ContainsAny(val, "|()+-*/") {
			if c.temporary != nil {
				if v, ok := GetterFromMap(c.temporary, val); ok {
					return v, nil
				}
			}

			if v, ok := c.Getter(val); ok {

				return v, nil
			}
		}

		return eval.Eval(val)
	}

	return eval.Template(val)
}

func (c *Context) scope() map[string]any {
	scope := make(map[string]any)

	for k, v := range c.Vars {
		scope[k] = v
	}

	if c.temporary != nil {
		for k, v := range c.temporary {
			scope[k] = v
		}
	}

	return scope
}

// func (c *Context) Resolve(v any) (any, error) {
// 	switch val := v.(type) {

// 	case string:
// 		val = strings.TrimSpace(val)
// 		if val == "" {
// 			return "", nil
// 		}

// 		needRender :=
// 			strings.HasPrefix(val, "$") ||
// 				strings.Contains(val, "{{")

// 		if !needRender {
// 			return val, nil
// 		}

// 		// $xxx -> {{ $xxx }}
// 		if strings.HasPrefix(val, "$") {
// 			val = "{{ " + val + " }}"
// 		}

// 		return c.Scalar(val)
// 	default:
// 		return v, nil
// 	}
// }

/* =========================
   Render (VIEW layer)
========================= */

func (c *Context) template() *template.Template {
	if c.templ == nil {
		c.templ = template.New("work").Funcs(c.pipes.Functions())
	}
	return c.templ
}

func loadRenderText(val string) (string, error) {
	if strings.HasSuffix(val, ".tmpl") {
		b, err := os.ReadFile(val)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return val, nil
}

func (c *Context) render(val string) (string, error) {
	if val == "" {
		return "", nil
	}

	text, err := loadRenderText(val)
	if err != nil {
		return "", err
	}

	text = pipe.Preprocessor(text)

	tpl, err := c.template().Parse(text)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = c.WithTemporary(c.temporary, func() error {
		return tpl.Execute(&buf, c)
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

/* =========================
   OPTIONAL: RenderScalar
========================= */

func (c *Context) Scalar(tpl string) (any, error) {

	out, err := c.render(tpl)
	if err != nil {
		return nil, err
	}

	var w ValueWriter
	_, _ = w.Write([]byte(out))
	return w.Resolve(), nil
}

/* =========================
   Sandbox execution
========================= */

func (c *Context) WithTemporary(tmp map[string]any, fn func() error) error {
	if len(tmp) == 0 {
		return fn()
	}

	if c.Vars == nil {
		c.Vars = make(map[string]any)
	}

	backup := make(map[string]any, len(tmp))
	applied := make([]string, 0, len(tmp))

	for k, v := range tmp {
		nk, err := normalizeKey(k)
		if err != nil {
			return err
		}

		if old, ok := c.Vars[nk]; ok {
			backup[nk] = old
		}

		c.Vars[nk] = v
		applied = append(applied, nk)
	}

	defer func() {
		for _, k := range applied {
			if old, ok := backup[k]; ok {
				c.Vars[k] = old
			} else {
				delete(c.Vars, k)
			}
		}
	}()

	return fn()
}

type ValueWriter struct {
	buf   bytes.Buffer
	Value any
	wrote bool
}

func (w *ValueWriter) Write(p []byte) (int, error) {
	w.wrote = true
	return w.buf.Write(p)
}

func (w *ValueWriter) Resolve() any {
	if !w.wrote {
		return ""
	}

	s := strings.TrimSpace(w.buf.String())

	// nil
	if s == "" || s == "nil" || s == "null" {
		return nil
	}

	// bool
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// int
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}

	// float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// fallback string
	return s
}

func (ctx *Context) json(val any) (any, error) {
	switch v := val.(type) {

	case string:
		res, err := ctx.evaluate(v)
		if err != nil {
			return nil, err
		}

		s, ok := res.(string)
		if !ok {
			return res, nil
		}

		s = strings.TrimSpace(s)
		if looksLikeJSON(s) {
			if parsed, ok := tryParseJSON(s); ok {
				return parsed, nil
			}
		}

		return res, nil

	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			r, err := ctx.json(val)
			if err != nil {
				return nil, err
			}
			out[k] = r
		}
		return out, nil

	case []any:
		out := make([]any, len(v))
		for i, val := range v {
			r, err := ctx.json(val)
			if err != nil {
				return nil, err
			}
			out[i] = r
		}
		return out, nil

	default:
		return v, nil
	}
}

func looksLikeJSON(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == '{' || s[0] == '['
}

func tryParseJSON(s string) (any, bool) {
	var v any
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()

	if err := dec.Decode(&v); err != nil {
		return nil, false
	}

	// ensure no trailing garbage
	if dec.More() {
		return nil, false
	}

	return v, true
}

func NormalizeTarget(path string, keys ...string) string {
	path = strings.TrimSpace(path)

	// 1. path absolute -> giữ nguyên
	if strings.HasPrefix(path, "$") {
		return path
	}

	// 2. resolve root từ keys
	root := "$result"
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if strings.HasPrefix(k, "$") {
			root = k
		} else {
			root = "$" + k
		}
		break
	}

	// 3. empty path
	if path == "" {
		return root
	}

	// 4. relative path
	if strings.HasPrefix(path, ".") {
		return root + path
	}

	// 5. normal path
	return root + "." + path
}
