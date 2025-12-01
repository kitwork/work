// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

// type Data map[string]interface{}

// var data = map[string]interface{}{}

// func NewData() map[string]interface{} {
// 	return data
// }

// func (c *Config) Data(name string, source ...string) *Config {
// 	if strings.TrimSpace(name) == "" || len(source) == 0 {
// 		return c
// 	}

// 	// Tạo đường dẫn từ source
// 	dir := directory(source...)

// 	// Kiểm tra map data và gán
// 	if _, ok := data[name]; !ok {
// 		files, err := scan(dir, ".work")

// 		if err != nil {
// 			fmt.Println("Error:", err)
// 			return c
// 		}
// 		data[name] = files
// 	}
// 	return c
// }

// func (c *Config) Data(name string, source ...string) *Config {
// 	if strings.TrimSpace(name) == "" || len(source) == 0 {
// 		return c
// 	}

// 	// Tạo đường dẫn từ source
// 	dir := path.Join(source...)
// 	if !filepath.IsAbs(dir) {
// 		dir = filepath.Join(".", dir)
// 	}

// 	// Kiểm tra map data và gán
// 	if _, ok := data[name]; !ok {
// 		files, err := scan(dir, ".work")

// 		if err != nil {
// 			fmt.Println("Error:", err)
// 			return c
// 		}
// 		data[name] = files
// 	}
// 	return c
// }
