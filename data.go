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
	"path"
	"path/filepath"
	"strings"
)

// type Data map[string]interface{}

var data = map[string]interface{}{}

func (c *Config) Data(name string, source ...string) *Config {
	if strings.TrimSpace(name) == "" || len(source) == 0 {
		return c
	}

	// Tạo đường dẫn từ source
	dir := path.Join(source...)
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(".", dir)
	}

	// Kiểm tra map data và gán
	if _, ok := data[name]; !ok {
		files, err := scan(dir, ".work")

		if err != nil {
			fmt.Println("Error:", err)
			return c
		}
		data[name] = files
	}
	return c
}
