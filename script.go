// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"errors"
	"fmt"
)

type Script struct {
	Run   string `work:"url,required,fallback"`       // URL endpoint
	Agent Type   `work:"agent,ignore" default:"http"` // http | client

}

func (t *Work) Script(ctx *Context) error {
	if t.Type != TypeScript {
		return errors.New("type is not script")
	}

	fmt.Println("→ [script] chạy script ...")
	return nil
}
