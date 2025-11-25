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
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Name  string
	Dir   string
	files *[]string
	// Accepts []string // đuôi file chấp nhận

	Embed bool
}

func (cfg *Config) source(name string, source ...string) Source {
	if len(source) == 0 {
		source = []string{name}
	}

	dir := directory(source...)

	return Source{
		Name: name,
		Dir:  dir,
	}

}

// files — read 1 cấp thư mục (non-recursive)
func (src *Source) Files(accepts ...string) ([]string, error) {
	// cache
	if src.files != nil {
		return *src.files, nil
	}

	exts := normalizeExts(accepts)

	entries, err := os.ReadDir(src.Dir)
	if err != nil {
		return nil, err
	}

	var out []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if matched := matchExt(entry.Name(), exts); matched != "" {
			out = append(out, filepath.Join(src.Dir, entry.Name()))
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

	entries, err := os.ReadDir(src.Dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if matched := matchExt(name, exts); matched != "" {
			path := filepath.Join(src.Dir, name)
			parsed, err := readYAML(path)
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
func (src *Source) Scanner(accepts ...string) (map[string]interface{}, error) {
	if len(accepts) == 0 {
		return map[string]interface{}{}, errors.New("error not using accepts...")
	}
	result := make(map[string]interface{})
	exts := normalizeExts(accepts)

	err := filepath.WalkDir(src.Dir, func(path string, entry fs.DirEntry, err error) error {
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

		parsed, err := readYAML(path)
		if err != nil {
			return err
		}

		key := strings.TrimSuffix(name, matched)
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
func readYAML(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var parsed interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return parsed, nil
}
