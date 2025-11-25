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
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// scan quét thư mục dir, chỉ lấy các file có đuôi trong accepts, parse YAML
func scan(dir string, accepts ...string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// chuẩn hóa đuôi: luôn bắt đầu bằng "."
	for i, ext := range accepts {
		if !strings.HasPrefix(ext, ".") {
			accepts[i] = "." + ext
		}
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		for _, ext := range accepts {
			if strings.HasSuffix(d.Name(), ext) {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				var parsed interface{}
				if err := yaml.Unmarshal(data, &parsed); err != nil {
					return fmt.Errorf("failed to parse %s: %w", d.Name(), err)
				}

				// Lấy tên file không có đuôi
				name := strings.TrimSuffix(d.Name(), ext)
				result[name] = parsed
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func files(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func directory(source ...string) string {

	// Tạo đường dẫn từ source
	dir := path.Join(source...)
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(".", dir)
	}
	return dir
}

func readfile(path string) (map[string]interface{}, error) {
	// 1. Đọc file YAML
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	// 2. Parse YAML vào map[string]interface{}
	var workflow map[string]interface{}
	err = yaml.Unmarshal(data, &workflow)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return workflow, nil
}
