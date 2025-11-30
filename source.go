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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Name string
	dir  *string
	// Dir   string
	files *[]string
	// Accepts []string // đuôi file chấp nhận

	list []string

	Embed bool
}

func NewSource(name string, source ...string) *Source {
	if len(source) == 0 {
		source = []string{name}
	}
	return &Source{Name: name, list: source}
}

// files — read 1 cấp thư mục (non-recursive)
func (src *Source) Directory() string {
	if src.dir != nil {
		return *src.dir
	}
	// Tạo đường dẫn từ source
	dir := directory(src.list...)
	src.dir = &dir
	return *src.dir
}

func (src *Source) PathFile(name string) string {
	return path.Join(src.Directory(), name+".work")
}

func (src *Source) HasFile(name string) bool {
	return fileExists(src.PathFile(name))
}

func (src *Source) HasFiles(names ...string) bool {
	for _, name := range names {
		if src.HasFile(name) {
			return true
		}
	}
	return false
}

// WalkPaths — read all files recursively (recursive)
func (src *Source) Paths(accepts ...string) ([]string, error) {
	exts := normalizeExts(accepts)
	var out []string

	err := filepath.WalkDir(src.Directory(), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		if matched := matchExt(entry.Name(), exts); matched != "" {
			out = append(out, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (src *Source) Routes(accepts ...string) (routes []Router, err error) {
	exts := normalizeExts(accepts)
	directory := src.Directory()

	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		if matched := matchExt(entry.Name(), exts); matched != "" {
			routes = append(routes, *NewRouter(path))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return routes, nil
}

func (src *Source) Routing(accepts ...string) (routes []Router, err error) {
	if len(accepts) == 0 {
		return nil, errors.New("error not using accepts...")
	}

	exts := normalizeExts(accepts)

	directory := src.Directory()

	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		matched := matchExt(name, exts)
		if matched == "" {
			return nil
		}

		// lấy relative path
		rel, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}

		// remove extension (.work)
		key := strings.TrimSuffix(rel, matched)

		// đổi slash dạng windows "\" thành "/"
		key = filepath.ToSlash(key)

		router := NewRouter(key)

		routes = append(routes, *router)
		// fmt.Println()

		return nil
	})

	if err != nil {
		return nil, err
	}

	return
}

// func (cfg *Config) source(name string, source ...string) Source {
// 	if len(source) == 0 {
// 		source = []string{name}
// 	}

// 	dir := directory(source...)

// 	return Source{
// 		Name: name,
// 		Dir:  dir,
// 	}

// }

// files — read 1 cấp thư mục (non-recursive)
func (src *Source) Files(accepts ...string) ([]string, error) {
	// cache
	if src.files != nil {
		return *src.files, nil
	}

	directory := src.Directory()
	exts := normalizeExts(accepts)

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var out []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if matched := matchExt(entry.Name(), exts); matched != "" {
			out = append(out, filepath.Join(directory, entry.Name()))
		}
	}

	src.files = &out
	return out, nil
}

// Scan — scan 1 cấp thư mục (non-recursive)
func (src *Source) Scan(accepts ...string) (map[string]interface{}, error) {
	if len(accepts) == 0 {
		return map[string]interface{}{}, errors.New("error not using accepts...")
	}
	result := make(map[string]interface{})
	exts := normalizeExts(accepts)

	directory := src.Directory()
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if matched := matchExt(name, exts); matched != "" {
			path := filepath.Join(directory, name)
			parsed, err := readfile(path)
			if err != nil {
				return nil, err
			}
			key := strings.TrimSuffix(name, matched)
			result[key] = parsed
		}
	}

	return result, nil
}

// Scanner — scan toàn bộ thư mục (recursive)
// func (src *Source) Scanner(accepts ...string) (map[string]interface{}, error) {
// 	if len(accepts) == 0 {
// 		return map[string]interface{}{}, errors.New("error not using accepts...")
// 	}
// 	result := make(map[string]interface{})
// 	exts := normalizeExts(accepts)

// 	err := filepath.WalkDir(src.Dir, func(path string, entry fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if entry.IsDir() {
// 			return nil
// 		}

// 		name := entry.Name()

// 		matched := matchExt(name, exts)
// 		if matched == "" {
// 			return nil
// 		}

// 		parsed, err := readYAML(path)
// 		if err != nil {
// 			return err
// 		}

// 		key := strings.TrimSuffix(name, matched)
// 		result[key] = parsed

// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

func (src *Source) Scanner(accepts ...string) (map[string]interface{}, error) {
	if len(accepts) == 0 {
		return map[string]interface{}{}, errors.New("error not using accepts...")
	}
	result := make(map[string]interface{})
	exts := normalizeExts(accepts)

	directory := src.Directory()

	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		matched := matchExt(name, exts)
		if matched == "" {
			return nil
		}

		parsed, err := readfile(path)
		if err != nil {
			return err
		}

		// lấy relative path
		rel, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}

		// remove extension (.work)
		key := strings.TrimSuffix(rel, matched)

		// đổi slash dạng windows "\" thành "/"
		key = filepath.ToSlash(key)

		result[key] = parsed
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ------------------------------
// Helpers
// ------------------------------

// chuẩn hoá đuôi file: yaml → .yaml
func normalizeExts(exts []string) []string {
	out := make([]string, 0, len(exts))
	for _, e := range exts {
		e = strings.ToLower(e)
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		out = append(out, e)
	}
	return out
}

func matchExt(name string, exts []string) (matched string) {
	lower := strings.ToLower(name)
	for _, ext := range exts {
		if strings.HasSuffix(lower, ext) {
			return ext
		}
	}
	return ""
}

// đọc file YAML
func readfile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return parsed, nil
}

func directory(source ...string) string {

	// Tạo đường dẫn từ source
	dir := path.Join(source...)
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(".", dir)
	}
	return dir
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // file tồn tại
	}
	return !os.IsNotExist(err) // false nếu file không tồn tại
}

// func router(raw string) (path string, method string) {

// 	// Bỏ dấu / thừa
// 	raw = strings.Trim(raw, "/")

// 	parts := strings.Split(raw, "/")

// 	// Method là phần cuối
// 	method = parts[len(parts)-1]

// 	// Path parts (bỏ method)
// 	pathParts := parts[:len(parts)-1]

// 	// Convert {id} => :id, {$} => *
// 	for i, p := range pathParts {

// 		// wildcard
// 		if p == "{$}" {
// 			pathParts[i] = "*"
// 			continue
// 		}

// 		// param {id}
// 		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
// 			key := p[1 : len(p)-1] // bỏ 2 dấu {}
// 			pathParts[i] = ":" + key
// 		}
// 	}

// 	// Gộp thành path
// 	path = "/" + strings.Join(pathParts, "/")

// 	return
// }
