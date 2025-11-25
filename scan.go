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
	"os"
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
