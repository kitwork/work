// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import "fmt"

type Log struct {
	Message string `work:"message,required,fallback"` // log message
}

func (t *Work) Log(ctx *Context) error {
	if t.Type != TypeLog {
		return fmt.Errorf("type is not log")
	}

	cfg := Log{}
	if err := t.classify(ctx, &cfg); err != nil {
		return err
	}

	fmt.Println("→ [log]", cfg.Message)
	return nil
}
