// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import "fmt"

type Loop struct {
	Name string `work:"name"`

	Async bool        `work:"async"`
	As    string      `work:"as"`
	Item  string      `work:"item"`
	Index string      `work:"index"`
	From  interface{} `work:"url,required,fallback"`
}

func (t *Work) Loop(ctx *Context) error {
	cfg := Loop{
		Name:  t.Name,
		Item:  "item",
		Index: "idx",
	}

	// 1️⃣ Lấy dữ liệu đầu vào: from
	if fromRaw, ok := t.Config["from"]; ok {
		switch from := fromRaw.(type) {
		case string:
			val, has := ctx.Getter(from)
			if !has {
				return fmt.Errorf("Loop 'from' empty: %s", from)
			}
			cfg.From = val
		default:
			cfg.From = from
		}
	} else {
		return fmt.Errorf("Loop 'from' not provided")
	}

	// 2️⃣ Tên alias item (mặc định = item)
	if itemName, ok := t.Config["item"].(string); ok && itemName != "" {
		cfg.Item = itemName
	}

	// 3️⃣ Tên alias index (mặc định = idx)
	if indexName, ok := t.Config["index"].(string); ok && indexName != "" {
		cfg.Index = indexName
	}

	// // 6️⃣ Thực thi loop
	result, err := cfg.Run(ctx, t.Works)

	if err != nil {
		return err
	}

	// 5️⃣ Nếu có "as", lưu output vào context
	if as, ok := t.Config["as"].(string); ok {
		cfg.As = as
	}
	// logJSON(result)
	// // 7️⃣ Lưu kết quả vào context nếu có as
	if cfg.As != "" {
		ctx.As(cfg.As, result)
	}

	return nil
}

func (l *Loop) Run(ctx *Context, works []*Work) ([]interface{}, error) {
	if ctx.Return != nil {
		return nil, nil
	}

	// 1. Resolve source list
	var items []interface{}
	switch src := l.From.(type) {
	case []interface{}:
		items = src
	case map[string]interface{}:
		items = []interface{}{src}

	case string:

	default:
		return nil, fmt.Errorf("loop: unsupported From type %T", src)
	}

	itemKey := l.Item
	if itemKey == "" {
		itemKey = "item"
	}
	indexKey := l.Index
	if indexKey == "" {
		indexKey = "index"
	}

	asVal := l.As
	if asVal == "" {
		asVal = l.Name // fallback
	}

	var result []interface{}

	// 2. Loop
	for i, raw := range items {

		// alias cho từng vòng
		aliases := map[string]interface{}{
			itemKey:  raw,
			indexKey: i,
		}

		ctx.Aliases(aliases)

		// 3. Nếu async → chạy goroutine
		// if l.Async {
		// 	go func(raw interface{}, idx int, alias map[string]interface{}) {
		// 		// mỗi goroutine phải set lại alias
		// 		ctx.Aliases(alias)

		// 		for _, w := range l.Works {
		// 			_ = w.Run(ctx) // ignore error hoặc bạn thêm cơ chế gom lỗi
		// 		}
		// 	}(raw, i, aliases)

		// 	continue
		// }

		// 4. Run từng Work con theo thứ tự
		for _, w := range works {
			if err := w.Run(ctx, "loop"); err != nil {
				return nil, err
			}
		}

		result = append(result, raw)

	}

	// // 5. Success branch (Loop cấp Work)
	// for _, s := range l.Success {
	// 	s.Run(ctx)
	// }

	return result, nil
}
