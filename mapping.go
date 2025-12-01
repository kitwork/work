// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
)

type Mapping struct {
	Name      string                 `work:"name"`
	As        string                 `work:"as"`
	Item      string                 `work:"item"`
	Index     string                 `work:"index"`
	From      interface{}            `work:"url,required,fallback"`
	Transform map[string]interface{} // key mới -> item field name
}

func (t *Work) Mapping(ctx *Context) error {
	cfg := Mapping{Name: t.Name}

	// 1️⃣ Lấy dữ liệu từ 'from'
	if fromRaw, ok := t.Config["from"]; ok {
		switch from := fromRaw.(type) {
		case string:
			val, has := ctx.Getter(from)
			if !has {
				return fmt.Errorf("Mapping 'from' empty: %s", from)
			}
			cfg.From = val
		default:
			cfg.From = from
		}
	} else {
		return fmt.Errorf("Mapping 'from' not provided")
	}

	// 2️⃣ Lấy transform
	transformRaw, ok := t.Config["transform"]
	if !ok {
		return fmt.Errorf("Mapping 'transform' empty")
	}
	cfg.Transform = transformRaw.(map[string]interface{})

	// 3️⃣ Lấy item và index từ config (mặc định item="item", index="idx")
	if itemName, ok := t.Config["item"].(string); ok && itemName != "" {
		cfg.Item = itemName
	} else {
		cfg.Item = "item"
	}
	if indexName, ok := t.Config["index"].(string); ok && indexName != "" {
		cfg.Index = indexName
	} else {
		cfg.Index = "idx"
	}

	// 4️⃣ Áp dụng transform + render deep với alias item/index
	result := cfg.Transformer(ctx)

	// 5️⃣ Lưu vào context
	if as, ok := t.Config["as"].(string); ok {
		cfg.As = as
	}
	if cfg.As != "" {
		ctx.As(cfg.As, result)
	}

	return nil
}

// Transformer: áp dụng transform + render deep cho From
func (m *Mapping) Transformer(ctx *Context) []interface{} {
	var result []interface{}
	itemName := m.Item
	if itemName == "" {
		itemName = "item"
	}
	indexName := m.Index
	if indexName == "" {
		indexName = "idx"
	}

	switch items := m.From.(type) {
	case []interface{}:
		for i, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			aliases := map[string]interface{}{
				itemName:  itemMap,
				indexName: i,
			}
			newItem := ApplyTransformDeep(ctx, m.Transform, aliases)
			result = append(result, newItem)
		}
	case map[string]interface{}:
		aliases := map[string]interface{}{
			itemName:  items,
			indexName: 0,
		}
		newItem := ApplyTransformDeep(ctx, m.Transform, aliases)
		result = append(result, newItem)
	default:
		// không hỗ trợ type khác
	}

	return result
}

// ApplyTransformDeep render template recursively với nested map
func ApplyTransformDeep(ctx *Context, transform map[string]interface{}, aliases map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range transform {
		switch vv := v.(type) {
		case string:
			rendered, err := ctx.render(vv, aliases)
			if err != nil {
				out[k] = vv // fallback giữ nguyên
			} else {
				out[k] = rendered
			}
		case map[string]interface{}:
			// deep nested
			out[k] = ApplyTransformDeep(ctx, vv, aliases)
		default:
			out[k] = vv
		}
	}
	return out
}
