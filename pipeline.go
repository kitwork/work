// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"

	"github.com/kitwork/pipe"
)

type Pipeline struct {
	funcs template.FuncMap
}

func NewPipe() *Pipeline {
	return &Pipeline{funcs: pipe.New()}
}

// Add a new function to the pipeline
func (p *Pipeline) Add(key string, function interface{}) error {
	if key == "" {
		return fmt.Errorf("pipe name cannot be empty")
	}
	if function == nil || reflect.TypeOf(function).Kind() != reflect.Func {
		return fmt.Errorf("value for pipe '%s' is not a function", key)
	}
	if _, exists := p.funcs[key]; exists {
		return fmt.Errorf("pipe '%s' already exists", key)
	}
	p.funcs[key] = function
	return nil
}

// Emit a value as a function
func (p *Pipeline) As(key string, val interface{}) error {
	return p.Add(key, func() reflect.Value {
		return reflect.ValueOf(val)
	})
}

// Merge multiple key-value pairs into the pipeline
func (p *Pipeline) Merge(data map[string]interface{}) error {
	for k, v := range data {
		if err := p.As(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Return the underlying FuncMap
func (p *Pipeline) Functions() template.FuncMap {
	return p.funcs
}

func (p *Pipeline) Clone() *Pipeline {
	clone := make(template.FuncMap, len(p.funcs))
	for k, v := range p.funcs {
		clone[k] = v
	}
	return &Pipeline{funcs: clone}
}

func (p *Pipeline) Get(key string) (interface{}, bool) {
	fn, ok := p.funcs[key]
	if !ok {
		return nil, false
	}

	v := reflect.ValueOf(fn).Call(nil)
	if len(v) > 0 {
		// unwrap reflect.Value
		val := v[0].Interface()
		if rv, ok := val.(reflect.Value); ok {
			return rv.Interface(), true
		}
		return val, true
	}
	return nil, true
}

func (p *Pipeline) Getter(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, false
	}

	// Lấy giá trị gốc từ key đầu tiên
	val, ok := p.Get(parts[0])
	if !ok {
		return nil, false
	}

	// Duyệt các phần còn lại của path
	for _, key := range parts[1:] {
		switch v := val.(type) {
		case map[string]interface{}:
			val, ok = v[key]
			if !ok {
				return nil, false
			}
		case []interface{}:
			// Nếu key là số (index)
			idx, err := strconv.Atoi(key)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, false
			}
			val = v[idx]
		default:
			return nil, false
		}
	}

	return val, true
}
