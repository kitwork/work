// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
)

type Mapping struct {
	Name string      `work:"name"`
	As   string      `work:"as"`
	From interface{} `work:"url,required,fallback"` // URL endpoint

	Rename map[string]interface{}
}

func (t *Work) Mapping(ctx *Context) error {

	cfg := Mapping{
		Name: t.Name,
	}

	// 1️⃣ Lấy dữ liệu từ 'from' nếu có
	if fromRaw, ok := t.Config["from"]; ok {
		switch from := fromRaw.(type) {
		case string:
			// Lấy từ pipe hoặc template
			val, has := ctx.Getter(from)
			if !has {
				return fmt.Errorf("Mapping 'from' empty: %s", from)
			}
			cfg.From = val
		default:
			// Nếu là object/array trực tiếp
			cfg.From = from
		}
	}

	// 2️⃣ Lấy rename nếu có
	if renameRaw, ok := t.Config["rename"]; ok {
		if renameMap, ok := renameRaw.(map[string]interface{}); ok {
			cfg.Rename = renameMap
		}
	}

	var result interface{}
	// 3️⃣ Áp dụng RenameKeys nếu cần
	if cfg.Rename != nil && cfg.From != nil {
		if cfg.Rename != nil {
			result = RenameKeys(cfg.From, cfg.Rename)
		} else {
			result = cfg.From
		}
	}

	// fmt.Println(result)

	// 4️⃣ Lưu kết quả vào Context

	// 2️⃣ Lấy rename nếu có
	if as, ok := t.Config["as"].(string); ok {
		cfg.As = as
	}

	if cfg.As != "" {
		ctx.As(cfg.As, ToSliceMap(result))
	}

	return nil
}

func RenameKeys(val interface{}, rename map[string]interface{}) interface{} {
	switch v := val.(type) {

	case map[string]interface{}:
		newMap := make(map[string]interface{})

		for oldKey, itemValue := range v {
			newKey := oldKey

			// Nếu rename có dạng: newKey: oldKey
			for rNew, rOld := range rename {
				if rOld == oldKey {
					newKey = rNew
					break
				}
			}

			// ❗ GIỮ NGUYÊN itemValue, không đệ quy đổi key sâu
			newMap[newKey] = itemValue
		}
		return newMap

	case []interface{}:
		newArr := make([]interface{}, len(v))
		for i, item := range v {
			// ❗ Áp dụng rename từng phần tử cấp 1
			newArr[i] = RenameKeys(item, rename)
		}
		return newArr

	default:
		return val
	}
}

func ToSliceMap(val interface{}) []map[string]interface{} {
	arr, ok := val.([]interface{})
	if !ok {
		return nil
	}

	out := make([]map[string]interface{}, 0)
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}
