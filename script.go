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
