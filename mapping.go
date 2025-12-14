// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

type Mapping map[string]interface{}

func (t *Work) Mapping(ctx *Context) error {
	cfg := Mapping{}

	switch t.Kind {
	case KindValue:

	case KindFull:
		cfg = t.Config
		if t.Name != "" {
			cfg["name"] = t.Name
		}

	}

	from, err := ctx.result()
	if err != nil {
		return err
	}
	// 4️⃣ Áp dụng transform + render deep với alias item/index
	result := cfg.Run(ctx, from)

	return ctx.as(result)
}

// Transformer: áp dụng transform + render deep cho From
func (m Mapping) Run(ctx *Context, from interface{}) []interface{} {
	var result []interface{}

	switch items := from.(type) {
	case []interface{}:
		for i, item := range items {

			aliases := map[string]interface{}{
				"item":  item,
				"index": i,
			}
			newItem := ApplyTransformDeep(ctx, m, aliases)
			result = append(result, newItem)
		}
	case map[string]interface{}:
		aliases := map[string]interface{}{
			"item":  items,
			"index": 0,
		}
		newItem := ApplyTransformDeep(ctx, m, aliases)
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
			rendered, err := ctx.temp(aliases).evaluate(vv)
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
