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
